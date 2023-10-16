package proxy

import (
	"bufio"
	"bytes"
	"draylix/dlog"
	"draylix/drlx"
	"fmt"
	"net"
	"strings"
)

const (
	DEBUG = iota
	HTTPS_PROXY
	SOCKS5_PROXY
	HTTP_PROXY
	UNDEFINED_PROXY
)

type LocalServer struct {
	listener      net.Listener
	RemoteServers *RemoteServerFactory
}

func (ls *LocalServer) Bind(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	dlog.Info("listen at %s", address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			dlog.Error("accepting error %v", err)
		}
		dlog.Debug("%s connected", conn.RemoteAddr().String())
		go ls.proxy(conn)
	}
}

func (ls *LocalServer) proxy(conn net.Conn) {
	initialReq := make([]byte, 512)
	n, err := conn.Read(initialReq)
	if err != nil {
		_ = conn.Close()
		return
	}
	proxyType := parseServiceType(initialReq[:n])

	if proxyType == UNDEFINED_PROXY {
		dlog.Warn("unknown proxy type : \n%s", string(initialReq[:n]))
		_ = conn.Close()
		return
	}
	server, err := ls.RemoteServers.Auto()
	if err != nil {
		dlog.Error("%v", err)
		_ = conn.Close()
		return
	}
	drlxConn, err := drlx.Dial(server.Address, server.Token, server.Iv, server.Ikey, proxyType)
	if err != nil {
		dlog.Error("failed to dial %s : %v", server.Address, err)
		_ = conn.Close()
		return
	}
	tun := NewProxyTunnel(conn, drlxConn, proxyType)
	info, err := tun.Handshake(initialReq[:n])
	if err != nil {
		dlog.Error("proxy handshake error : %v", err)
		_ = conn.Close()
		return
	}
	dlog.Info("proxy  %s -> %s ", conn.RemoteAddr().String(), info.Host)
	in, out := tun.StartProxy()
	dlog.Debug("proxy %s->%s done, %s transmitted", conn.RemoteAddr().String(), info.Host, drlx.BytesFormat(in+out))
}

func parseServiceType(p []byte) byte {
	if p[0] == 5 {
		return SOCKS5_PROXY
	} else if strings.HasPrefix(string(p), "CONNECT") {
		return HTTPS_PROXY
	} else if isHttpProxy(p) {
		return HTTP_PROXY
	} else {
		return UNDEFINED_PROXY
	}
}

func isHttpProxy(p []byte) bool {
	var req string
	if len(p) > 10 {
		req = string(p[:10])
	} else {
		req = string(p)
	}
	return strings.HasPrefix(req, "GET") || strings.HasPrefix(req, "POST") || strings.HasPrefix(req, "PUT") || strings.HasPrefix(req, "HEAD") || strings.HasPrefix(req, "DELETE") || strings.HasPrefix(req, "OPTIONS") || strings.HasPrefix(req, "TRACE")
}

func parseConnectRequest(p []byte) (string, bool) {
	if p[0] != 'C' {
		return "", false
	}
	buf := bytes.Buffer{}
	buf.Write(p)
	// 从连接中创建一个带有缓冲区的读取器
	reader := bufio.NewReader(&buf)
	// 读取请求行
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return "", false
	}
	// 检查请求是否以"CONNECT"开头
	if !strings.HasPrefix(requestLine, "CONNECT ") {
		return "", false
	}
	// 解析主机名
	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		return "", false
	}
	return parts[1], true
}

func parseSocks5Request(data []byte) (string, bool) {
	if len(data) < 5 {
		return "", false
	}
	// Check SOCKS version (must be 5)
	if data[0] != 0x05 {
		return "", false
	}
	// Get the command field (CONNECT, BIND, UDP ASSOCIATE)
	command := data[1]
	// Check if it's a CONNECT command (value 0x01)
	if command != 0x01 {
		return "", false
	}
	addressType := data[3]

	var targetHost string

	switch addressType {
	case 0x01: // IPv4 address
		if len(data) < 10 {
			return "", false
		}
		targetHost = fmt.Sprintf("%d.%d.%d.%d", data[4], data[5], data[6], data[7])
	case 0x03: // Domain name
		domainLength := int(data[4])
		if len(data) < 5+domainLength+2 {
			return "", false
		}
		targetHost = string(data[5 : 5+domainLength])
	case 0x04: // IPv6 address (unsupported in this example)
		return "", false
	default:
		return "", false
	}
	return targetHost, true
}
