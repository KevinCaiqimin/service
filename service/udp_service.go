package service

import (
	"bytes"
	"net"

	"github.com/KevinCaiqimin/log"
)

type UDPService struct {
	addr string
	conn *net.UDPConn
	buf  bytes.Buffer

	sz   int
	data []byte
	stat map[int]int64
}

func (s *UDPService) address() string {
	return s.addr
}

func (s *UDPService) StartUDP() error {
	udp_addr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		log.Error("service=%v resolve udp addr error=%v", err)
		return err
	}
	conn, err := net.ListenUDP("udp", udp_addr)
	if err != nil {
		log.Error("service=%v listen error=%v", err)
		return err
	}
	s.conn = conn
	go func() {
		buf := make([]byte, 1024*1024)
		for {
			n, raddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Error("readfrom error=%v", err)
				continue
			}
			log.Info("receive from %s %d bytes", raddr, n)
			nn, err := conn.WriteToUDP(buf[:n], raddr)
			log.Info("send to remote %v %d bytes, err=%v", raddr, nn, err)
		}
	}()
	return nil
}

func NewUDPService(addr string) *UDPService {
	return &UDPService{
		addr: addr,
		stat: make(map[int]int64),
	}
}
