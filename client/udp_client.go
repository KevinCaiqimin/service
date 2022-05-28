package client

import (
	"fmt"
	"net"

	"github.com/KevinCaiqimin/log"
)

type UDPClient struct {
}

func (c *UDPClient) Start() {
	raddr, err := net.ResolveUDPAddr("udp", "81.71.15.127:7001")
	if err != nil {
		log.Error("resolve remote addr error=%v", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Error("dial udp error=%v", err)
		return
	}
	BUF_LEN := 1024 * 1
	buf := make([]byte, BUF_LEN)
	go func() {
		for {
			in_n := 0
			n, _ := fmt.Scan(&in_n)
			if n == 0 {
				continue
			}
			bbb := make([]byte, in_n)
			for i := 0; i < in_n; i++ {
				bbb[i] = 'C'
			}
			n, err := conn.Write(bbb)
			if err != nil {
				log.Error("Write error=%v", err)
				continue
			}
			log.Info("write %d bytes", n)
		}
	}()
	go func() {
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Error("Read from error=%v", err)
				continue
			}
			log.Info("Read %d bytes from server", n)
		}
	}()
}

func NewUDPClient() *UDPClient {
	return &UDPClient{}
}
