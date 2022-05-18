package client

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/KevinCaiqimin/log"
	"github.com/KevinCaiqimin/service/pack"
)

type Client struct {
	addr     string
	conn     net.Conn
	connted  bool
	interval time.Duration
	sessid   int
	buf      *bytes.Buffer
	run      bool

	sz   int
	data []byte

	lock sync.Mutex
	stat map[int]int64
	ch   chan int
}

var ts int64
var cnt int64

//请求次数和总消耗时间
var t_times int64
var t_elapse int64

func (c *Client) new_session_id() int {
	c.sessid++
	return c.sessid
}

/*
func (c *Client) set_stat(k int, v int64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.stat[k] = v
}

func (c *Client) del_stat(k int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.stat, k)
}

func (c *Client) get_stat(k int) (int64, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	v, ok := c.stat[k]
	return v, ok
}
*/
func (c *Client) send_msg(msg string) {
	start_nano := time.Now().UnixNano()
	msg = fmt.Sprintf("%d", start_nano)
	sessid := c.new_session_id()
	bb, err := pack.Pack(sessid, []byte(msg))
	if err != nil {
		log.Error("dest=%v, pack message error=%v", c.addr, err)
		return
	}
	c.conn.Close() //TODO TEST
	k := 0
	for {
		n, err := c.conn.Write(bb[k:])
		if err != nil {
			log.Error("发送数据错误：%v", err)
			c.connted = false
		}
		k += n
		if k >= len(bb) {
			break
		}
	}
	// c.set_stat(sessid, start_nano)

	now := time.Now().Unix()
	if atomic.LoadInt64(&ts) != now {
		log.Error("qps=%d", atomic.LoadInt64(&cnt))
		atomic.SwapInt64(&ts, now)
		atomic.SwapInt64(&cnt, 0)
	} else {
		atomic.AddInt64(&cnt, 1)
	}
}

func (c *Client) proc_buf() {
	for {
		if c.sz == 0 {
			s, err := pack.ReadSize(c.buf)
			if s == 0 || err != nil {
				return
			}
			c.sz = s
		}
		data, err := pack.ReadData(c.buf, c.sz)
		if data == nil || err != nil {
			return
		}
		c.sz = 0
		_, data_bytes, err := pack.Unpack(data)
		if err != nil {
			log.Error("dest=%v unpack data error=%v", c.addr, err)
			return
		}
		msg := string(data_bytes)
		mmsg := strings.Split(msg, ",")

		start_nano_cli, _ := strconv.ParseInt(mmsg[0], 10, 64)
		// start_nano_srv, _ := strconv.ParseInt(mmsg[1], 10, 64)

		// start_nano, ok := c.get_stat(sessid)
		// if !ok {
		// 	log.Error("dest=%v no session=%v", c.addr, sessid)
		// 	return
		// }
		elapse_nano := time.Now().UnixNano() - start_nano_cli
		log.Info("elapse=%vms", elapse_nano/1000/1000)
		// c.del_stat(sessid)

		atomic.AddInt64(&t_elapse, elapse_nano/1000/1000)
		atomic.AddInt64(&t_times, 1)
		times := atomic.LoadInt64(&t_times)
		if times >= 1000 {
			log.Error("avg_elapse=%d", atomic.LoadInt64(&t_elapse)/atomic.LoadInt64(&t_times))
			atomic.StoreInt64(&t_elapse, 0)
			atomic.StoreInt64(&t_times, 0)
		}
	}
}
func (c *Client) recv_msg() {
	buf := make([]byte, 1024)
	n, err := c.conn.Read(buf)
	if n == 0 {
		log.Warn("dest=%v disconnected", c.addr)
		c.connted = false
		c.run = false
		return
	}
	if err != nil {
		log.Error("dest=%v read error=%v", c.addr, err)
		c.connted = false
		c.run = false
		return
	}
	c.buf.Write(buf[:n])
	c.proc_buf()
}

/*
func (c *Client) recv_msg() {
	sessid := <-c.ch
	start_nano, ok := c.get_stat(sessid)
	if !ok {
		log.Error("dest=%v no session=%v", c.addr, sessid)
		return
	}
	elapse_nano := time.Now().UnixNano() - start_nano
	log.Info("elapse=%vms", elapse_nano/1000/1000)
	c.del_stat(sessid)

	atomic.AddInt64(&t_elapse, elapse_nano/1000/1000)
	atomic.AddInt64(&t_times, 1)
	times := atomic.LoadInt64(&t_times)
	if times >= 10000 {
		log.Error("avg_elapse=%d", atomic.LoadInt64(&t_elapse)/atomic.LoadInt64(&t_times))
		atomic.StoreInt64(&t_elapse, 0)
		atomic.StoreInt64(&t_times, 0)
	}
}

func (c *Client) send_msg(msg string) {
	sessid := c.new_session_id()
	start_nano := time.Now().UnixNano()
	c.set_stat(sessid, start_nano)
	c.ch <- sessid

	now := time.Now().Unix()
	if atomic.LoadInt64(&ts) != now {
		log.Error("qps=%d", atomic.LoadInt64(&cnt))
		atomic.SwapInt64(&ts, now)
		atomic.SwapInt64(&cnt, 0)
	} else {
		atomic.AddInt64(&cnt, 1)
	}
}
*/
func (c *Client) Start(ctx context.Context) {
	go func() {
		conn, err := net.Dial("tcp", c.addr)
		if err != nil {
			log.Error("setup client connect to remote %v error=%v", c.addr, err)
			c.run = false
			return
		}
		c.conn = conn
		c.connted = true
		for c.run {
			select {
			case <-ctx.Done():
				return
			default:
				c.send_msg("ping")
				time.Sleep(time.Millisecond * 500)
			}
		}
	}()

	go func() {
		for c.run {
			select {
			case <-ctx.Done():
				return
			default:
				if !c.connted {
					continue
				}
				c.recv_msg()
			}
		}
	}()

	log.Info("client started")
}

// var stat_ch chan Stat

func StartStat() {
	go func() {
		for {

		}
	}()
}

func NewClient(addr string) *Client {
	return &Client{
		addr:     addr,
		connted:  false,
		interval: 3,
		sessid:   0,
		buf:      bytes.NewBuffer([]byte{}),
		stat:     make(map[int]int64),
		run:      true,
		ch:       make(chan int, 100000),
	}
}
