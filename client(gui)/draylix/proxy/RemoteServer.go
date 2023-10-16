package proxy

import (
	"draylix/dlog"
	"draylix/utils"
	"fmt"
	"strings"
)

const (
	RECEIVE_LEN = 1024 * 32
)

type RemoteServer struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	Country string `json:"country"`
	Token   string `json:"token"`
	Address string `json:"address"`
	Iv      []byte `json:"initialVector"`
	Ikey    []byte `json:"initialKey"`
}

func (r *RemoteServer) Delay() (string, error) {
	hostPort := strings.Split(r.Address, ":")
	duration, err := utils.PingIP(hostPort[0])
	if err != nil {
		dlog.Error("remote server ping err,ip : %v,err :%v", hostPort[0], err)
		return "time out", err
	}
	milliseconds := duration.Milliseconds()
	millisecondsString := fmt.Sprintf("%d ms", milliseconds)
	return millisecondsString, nil
}
