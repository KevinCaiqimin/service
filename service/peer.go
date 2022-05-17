package service

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	"github.com/KevinCaiqimin/log"
	"github.com/KevinCaiqimin/service/pack"
)

type Session struct {
	sessid     int
	start_nano int64
	msg        []byte
}

type Peer struct {
	conn    net.Conn
	connted bool
	ctx     context.Context
	buf     *bytes.Buffer

	sz   int
	data []byte

	stat       map[int]int64
	ch         chan *Session
	cnt        int
	cnt_elapse int64
}

var ts int64
var cnt int64
var t_e int64
var t_t int64

func (p *Peer) addr() string {
	return p.conn.RemoteAddr().String()
}

func (c *Peer) response(sessid int, msg string, start_nano int64) {
	now_nano := time.Now().UnixNano()
	msg = fmt.Sprintf("%s,%d", msg, now_nano)
	times := rand.Intn(50000) + 1000
	kk := 0
	for i := 0; i < times; i++ {
		kk += i
	}
	log.Debug("kkkkk=%v", kk)

	bb, err := pack.Pack(sessid, []byte(msg))
	if err != nil {
		log.Error("dest=%v, pack message error", c.addr())
		return
	}
	k := 0
	for {
		n, err := c.conn.Write(bb[k:])
		if err != nil {
			c.connted = false
		}
		k += n
		if k >= len(bb) {
			break
		}
	}
	elapse_nano := now_nano - start_nano
	log.Info("elapse=%vms", elapse_nano/1000/1000)
	delete(c.stat, sessid)

	tt := atomic.AddInt64(&t_t, 1)
	ee := atomic.AddInt64(&t_e, elapse_nano/1000/1000)
	if tt >= 2000 {
		log.Error("avt_elapse=%vms", ee/tt)
		atomic.StoreInt64(&t_t, 0)
		atomic.StoreInt64(&t_e, 0)
	}

	now := time.Now().Unix()
	if atomic.LoadInt64(&ts) != now {
		log.Error("qps=%d", atomic.LoadInt64(&cnt))
		atomic.SwapInt64(&ts, now)
		atomic.SwapInt64(&cnt, 0)
	} else {
		atomic.AddInt64(&cnt, 1)
	}
}

func (c *Peer) exec_buf() {
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
		sessid, msg, err := pack.Unpack(data)
		if err != nil {
			log.Error("dest=%v unpack data error=%v", c.addr(), err)
			return
		}
		log.Info("receive session: %v", sessid)
		start_nano := time.Now().UnixNano()
		ss := &Session{
			sessid:     sessid,
			start_nano: start_nano,
			msg:        msg,
		}
		c.ch <- ss
	}
}

func (c *Peer) request() {
	buf := make([]byte, 1024)
	n, err := c.conn.Read(buf)
	if n == 0 {
		log.Warn("dest=%v disconnected", c.addr())
		c.connted = false
		return
	}
	if err != nil {
		log.Error("dest=%v read error=%v", c.addr(), err)
		c.connted = false
		return
	}
	log.Info("receive %d bytes", n)

	c.buf.Write(buf[:n])
	c.exec_buf()

}
func (p *Peer) Start() {
	go func() {
		for p.connted {
			select {
			case <-p.ctx.Done():
				return
			default:
				if !p.connted {
					continue
				}
				p.request()
			}
		}
	}()
	go func() {
		for p.connted {
			ss := <-p.ch
			sessid := ss.sessid
			start_nano := ss.start_nano
			p.response(sessid, string(ss.msg), start_nano)
		}
	}()
	log.Info("new peer of %v started", p.addr())
}

func NewPeer(conn net.Conn, ctx context.Context) *Peer {
	return &Peer{
		conn:       conn,
		connted:    true,
		ctx:        ctx,
		buf:        bytes.NewBuffer([]byte{}),
		stat:       make(map[int]int64),
		ch:         make(chan *Session, 10000),
		cnt:        0,
		cnt_elapse: 0,
	}
}
