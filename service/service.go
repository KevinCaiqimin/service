package service

import (
	"context"
	"net"

	"github.com/KevinCaiqimin/log"
)

type Service struct {
	addr     string
	listener net.Listener

	stat map[int]int64
}

func (c *Service) Start(ctx context.Context) {
	listener, err := net.Listen("tcp", c.addr)
	if err != nil {
		log.Error("service=%v listen error=%v", c.addr, err)
		return
	}
	c.listener = listener
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := c.listener.Accept()
				if err != nil {
					log.Error("service=%v accept error=%v", c.addr)
					continue
				}
				peer := NewPeer(conn, ctx)
				peer.Start()
			}
		}
	}()
	log.Info("service %v started", c.addr)
}

func NewService(addr string) *Service {
	return &Service{
		addr: addr,
		stat: make(map[int]int64),
	}
}
