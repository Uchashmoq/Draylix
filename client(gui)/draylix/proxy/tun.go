package proxy

import (
	"draylix/dlog"
	"draylix/utils"
	"errors"
	"net"
	"strings"
	"sync"
)

const (
	BUFFER_SIZE = 1024 * 64
)

type Info struct {
	Host string
}

type Tunnel interface {
	GetType() byte
	Handshake(initialReq []byte) (*Info, error)
	StartProxy() (in, out int64)
}

type HttpsTunnel struct {
	Local  net.Conn
	Remote net.Conn
}

type HttpTunnel struct {
	Local  net.Conn
	Remote net.Conn
}

func (h HttpTunnel) GetType() byte {
	return HTTP_PROXY
}

func (h HttpTunnel) Handshake(initialReq []byte) (*Info, error) {
	host, err := parseHttpRequest(initialReq)
	if err != nil {
		return nil, err
	}
	info := &Info{Host: host}
	_, err = h.Remote.Write(initialReq)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func parseHttpRequest(initialReq []byte) (string, error) {
	req := string(initialReq)
	args := strings.Split(req, " ")
	if len(args) < 2 {
		return "", errors.New("too few parameters in http head")
	}
	return args[1], nil
}

func (h HttpTunnel) StartProxy() (in, out int64) {
	g := sync.WaitGroup{}
	g.Add(2)
	go transmit(h.Local, h.Remote, &g, &in, InData)
	go transmit(h.Remote, h.Local, &g, &out, OutData)
	g.Wait()
	return
}

func (h *HttpsTunnel) StartProxy() (in, out int64) {
	g := sync.WaitGroup{}
	g.Add(2)
	go transmit(h.Local, h.Remote, &g, &in, InData)
	go transmit(h.Remote, h.Local, &g, &out, OutData)
	g.Wait()
	return
}

func transmit(dst, src net.Conn, g *sync.WaitGroup, flow *int64, dataChan chan int) {
	buf := make([]byte, BUFFER_SIZE)
	for {
		n, err := src.Read(buf)
		if err != nil {
			break
		}
		*flow += int64(n)
		utils.Put(dataChan, n)
		TrafficIncrease(int64(n))
		_, err = dst.Write(buf[:n])
		if err != nil {
			break
		}
	}
	_ = src.Close()
	_ = dst.Close()
	g.Done()
	buf = nil
}

type Socks5Tunnel struct {
	Local  net.Conn
	Remote net.Conn
}

func (s *Socks5Tunnel) StartProxy() (in, out int64) {
	g := sync.WaitGroup{}
	g.Add(2)
	go transmit(s.Local, s.Remote, &g, &in, InData)
	go transmit(s.Remote, s.Local, &g, &out, OutData)
	g.Wait()
	return
}

func (s *Socks5Tunnel) Handshake(initialReq []byte) (*Info, error) {
	_, err := s.Remote.Write(initialReq)
	if err != nil {
		return nil, err
	}
	resp := make([]byte, 64)
	n, err := s.Remote.Read(resp)
	if err != nil {
		return nil, err
	}
	_, err = s.Local.Write(resp[:n])
	if err != nil {
		return nil, err
	}
	cmdReq := make([]byte, 64)
	n, err = s.Local.Read(cmdReq)
	if err != nil {
		return nil, err
	}
	host, ok := parseSocks5Request(cmdReq[:n])
	if !ok {
		dlog.Warn("failed to parse socks5 cmd")
	}
	info := &Info{Host: host}
	_, err = s.Remote.Write(cmdReq[:n])
	if err != nil {
		return nil, err
	}
	cmdResp := make([]byte, 64)
	n, err = s.Remote.Read(cmdResp)
	if err != nil {
		return nil, err
	}
	_, err = s.Local.Write(cmdResp[:n])
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (s *Socks5Tunnel) GetType() byte {
	return SOCKS5_PROXY
}

func (h *HttpsTunnel) GetType() byte {
	return HTTPS_PROXY
}

func (h *HttpsTunnel) Handshake(initialReq []byte) (*Info, error) {
	host, ok := parseConnectRequest(initialReq)
	if !ok {
		dlog.Warn("malformed http connect request : %s", string(initialReq))
	}
	info := &Info{Host: host}
	_, err := h.Remote.Write(initialReq)
	if err != nil {
		return nil, err
	}
	resp := make([]byte, 128)
	n, err := h.Remote.Read(resp)
	if err != nil {
		return nil, err
	}
	_, err = h.Local.Write(resp[:n])
	if err != nil {
		return nil, err
	}
	return info, nil
}

func NewProxyTunnel(local net.Conn, remote net.Conn, proxyType byte) Tunnel {
	switch proxyType {
	case HTTPS_PROXY:
		p := &HttpsTunnel{
			Local:  local,
			Remote: remote,
		}
		return p
	case SOCKS5_PROXY:
		p := &Socks5Tunnel{
			Local:  local,
			Remote: remote,
		}
		return p
	case HTTP_PROXY:
		return &HttpTunnel{
			Local:  local,
			Remote: remote,
		}
	}
	return nil
}
