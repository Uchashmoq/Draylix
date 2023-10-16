package drlx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

/*

HEAD LEN DATA

client_hello:
HEAD LEN  TIMESTAMP ANNOUNCE TYPE TOKEN

server_hello:
HEAD LEN TIMESTAMP ANNOUNCE RES SESSION_KEY


HEAD LEN DATA

*/

var (
	HANDSHAKE_TIMELIMIT = 5 * time.Second
	RESP_DELAY          = 20 * time.Second
	USED_ANNOUNCE       = make(map[int64]interface{})
	CHECK_TIME          = true
	CHECK_ANNOUNCE      = true

	_head     = []byte("drlx")
	_head_len = len(_head)

	mu = sync.Mutex{}
)

const (
	CONN_PERMIT = iota
	CONN_DENIED
	NONSUPPORT_SERVICE
)

type Conn struct {
	TcpConn net.Conn
	iv      []byte
	key     []byte
}

func (c *Conn) Close() error {
	c.key = nil
	c.iv = nil
	return c.TcpConn.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.TcpConn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.TcpConn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.TcpConn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.TcpConn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.TcpConn.SetWriteDeadline(t)
}

func (c *Conn) Write(p []byte) (int, error) {
	if c.key == nil {
		return 0, errors.New("conn is not in connection")
	}
	cipherText := encode(p, c.iv, c.key)
	return c.TcpConn.Write(cipherText)
}

func (c *Conn) Read(p []byte) (int, error) {
	if c.key == nil {
		return 0, errors.New("conn is not in connection")
	}
	frame, err := readDecryptedFrame(c.TcpConn, c.iv, c.key)
	if err != nil {
		return 0, err
	}
	n := copy(p, frame)
	return n, nil
}

func Dial(address, token string, iv, ikey []byte, serviceType byte) (*Conn, error) {
	tcpConn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	key, err1 := handshake(tcpConn, token, serviceType, iv, ikey)
	if err1 != nil {
		return nil, err1
	}
	conn := &Conn{
		TcpConn: tcpConn,
		iv:      iv,
		key:     key,
	}
	return conn, err
}

func handshake(conn net.Conn, token string, serviceType byte, iv []byte, ikey []byte) ([]byte, error) {
	hello := clientHelloData(token, serviceType, iv, ikey)
	err0 := conn.SetReadDeadline(time.Now().Add(HANDSHAKE_TIMELIMIT))
	defer conn.SetReadDeadline(time.Time{})
	if err0 != nil {
		return nil, err0
	}
	_, err := conn.Write(hello)
	if err != nil {
		return nil, err
	}
	response, err1 := readDecryptedFrame(conn, iv, ikey)
	if err1 != nil {
		return nil, err1
	}
	return analyzeResponse(response)
}

func analyzeResponse(response []byte) ([]byte, error) {
	if len(response) < 8+8+1 {
		return nil, errors.New("connection response is too short")
	}

	if t, ok := checkTime(response[:8]); !ok && CHECK_TIME {
		return nil, errors.New(fmt.Sprintf("timestamp error ,duration: %d sec ", time.Duration(t)/time.Second))
	}

	if !checkAnno(response[8:16]) && CHECK_ANNOUNCE {
		return nil, errors.New("announce repeats")
	}

	if response[16] != CONN_PERMIT {
		return nil, errors.New("connection failed : " + getReason(response[16]))
	}
	key := make([]byte, len(response)-8-8-1)
	copy(key, response[8+8+1:])
	return key, nil
}

func getReason(code byte) string {
	switch code {
	case CONN_DENIED:
		return "server denied connection"
	case NONSUPPORT_SERVICE:
		return "nonsupport service"
	default:
		return "unknown cause"
	}
}

func checkAnno(p []byte) bool {
	a := bytesToInt64(p)
	mu.Lock()
	defer mu.Unlock()
	_, ok := USED_ANNOUNCE[a]
	if !ok {
		USED_ANNOUNCE[a] = struct{}{}
		return true
	} else {
		return false
	}
}

func checkTime(p []byte) (int64, bool) {
	t := bytesToInt64(p)
	sendTime := time.Unix(t, 0)
	recvTime := time.Unix(time.Now().Unix(), 0)
	d := recvTime.Sub(sendTime)
	if d < 0 {
		d = -d
	}
	if d > RESP_DELAY {
		return int64(d), false
	} else {
		return int64(d), true
	}
}

func clientHelloData(token string, serviceType byte, iv, ikey []byte) []byte {
	timestamp := time.Now().Unix()
	announce := rand.Int63()
	buf := bytes.Buffer{}
	buf.Write(int64ToBytes(timestamp))
	buf.Write(int64ToBytes(announce))
	buf.WriteByte(serviceType)
	buf.Write([]byte(token))
	return encode(buf.Bytes(), iv, ikey)
}

func encode(data, iv, key []byte) []byte {
	cipherText := AesEncrypt(data, key, iv)
	buf := bytes.Buffer{}
	buf.Write(_head)
	buf.Write(intToBytes(len(cipherText)))
	buf.Write(cipherText)
	return buf.Bytes()
}

func readDecryptedFrame(reader io.Reader, iv, key []byte) ([]byte, error) {
	cipherText, err := readFrame(reader)
	if err != nil {
		return nil, err
	}
	decrypt, err := AesDecrypt(cipherText, key, iv)
	return decrypt, err
}

func readFrame(reader io.Reader) ([]byte, error) {
	head, err := readn(reader, _head_len+4)
	if err != nil {
		return nil, err
	}
	if isInvalidHead(head[:_head_len]) {
		return nil, errors.New("read invalid data")
	}
	dataLen := getHeadLen(head)
	data, err1 := readn(reader, dataLen)
	if err1 != nil {
		return nil, err1
	}
	return data, nil
}

func getHeadLen(head []byte) int {
	return bytesToInt(head[_head_len:])
}

func isInvalidHead(head []byte) bool {
	return !bytes.Equal(head, _head)
}

func readn(reader io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	for k := 0; k < n; {
		d, err := reader.Read(buf[k:])
		if err != nil {
			return nil, err
		}
		k += d
	}
	return buf, nil
}

/*
func printB(p []byte) int {
	mu.Lock()
	for i := 0; i < len(p); i++ {
		fmt.Printf("[%d] %d ", i, p[i])
		if (i+1)%10 == 0 {
			fmt.Println("")
		}
	}
	fmt.Println("\n\n")
	mu.Unlock()
	return len(p)
}*/
