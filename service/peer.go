package service

import (
	"bytes"
	"context"
	"math/rand"
	"net"
	"time"

	"github.com/KevinCaiqimin/log"
	"github.com/KevinCaiqimin/service/pack"
)

type Peer struct {
	conn    net.Conn
	connted bool
	ctx     context.Context
	buf     *bytes.Buffer

	sz   int
	data []byte

	stat map[int]int64
}

func (p *Peer) addr() string {
	return p.conn.RemoteAddr().String()
}

func (c *Peer) response(sessid int, msg string, start_nano int64) {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(5)+1))

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
	elapse_nano := time.Now().UnixNano() - start_nano
	log.Info("elapse=%vms", elapse_nano/1000/1000)
	delete(c.stat, sessid)
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
	sessid, _, err := pack.Unpack(data)
	if err != nil {
		log.Error("dest=%v unpack data error=%v", c.addr(), err)
		return
	}
	log.Info("receive session: %v", sessid)
	start_nano := time.Now().UnixNano()
	go func(sessid int, msg string) {
		c.response(sessid, msg, start_nano)
	}(sessid, "pong")
}
func (p *Peer) Start() {
	go func() {
		for {
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
	log.Info("new peer of %v started", p.addr())
}

func NewPeer(conn net.Conn, ctx context.Context) *Peer {
	return &Peer{
		conn:    conn,
		connted: true,
		ctx:     ctx,
		buf:     bytes.NewBuffer([]byte{}),
		stat:    make(map[int]int64),
	}
}
