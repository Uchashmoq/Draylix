package proxy

import (
	"encoding/json"
	"errors"
	"os"
)

type RemoteServerFactory struct {
	RemoteServers map[string]RemoteServer
}

func NewRemoteServerFactory() *RemoteServerFactory {
	return &RemoteServerFactory{RemoteServers: make(map[string]RemoteServer)}
}

var (
	CurrentServer = RemoteServer{}
)

func (rsf *RemoteServerFactory) Auto() (RemoteServer, error) {
	if rsf.Quantity() == 0 {
		return RemoteServer{}, errors.New("no available proxy node")
	}
	rsf.RandServer()
	return CurrentServer, nil
}

func (rsf *RemoteServerFactory) RandServer() {
	if len(CurrentServer.Address) == 0 {
		for _, s := range rsf.RemoteServers {
			CurrentServer = s
			break
		}
	}
}
func (rsf *RemoteServerFactory) SortedServers() []*RemoteServer {
	servers := make([]*RemoteServer, 0)
	for name, _ := range rsf.RemoteServers {
		s := rsf.RemoteServers[name]
		servers = append(servers, &s)
	}
	n := len(servers)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if servers[i].Index > servers[j].Index {
				servers[i], servers[j] = servers[j], servers[i]
			}
		}
	}
	return servers
}

func (rsf *RemoteServerFactory) MaxIndex() int {
	max := -1
	for _, s := range rsf.RemoteServers {
		if s.Index > max {
			max = s.Index
		}
	}
	return max
}

func (rsf *RemoteServerFactory) Quantity() int {
	return len(rsf.RemoteServers)
}

func (rsf *RemoteServerFactory) Contains(name string) bool {
	_, ok := rsf.RemoteServers[name]
	return ok
}

func (rsf *RemoteServerFactory) Delete(name string) {
	delete(rsf.RemoteServers, name)
}

func (rsf *RemoteServerFactory) Add(name string, server RemoteServer) {
	rsf.RemoteServers[name] = server
}

func (rsf *RemoteServerFactory) Load(path string) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}
	buf := make([]byte, 1024*1024*4)
	n, err1 := file.Read(buf)
	if err1 != nil {
		return err1
	}
	err2 := json.Unmarshal(buf[0:n], &rsf.RemoteServers)
	if err2 != nil {
		return err2
	}
	buf = nil
	return nil
}

func (rsf *RemoteServerFactory) Store(path string) error {
	bytes, err := json.MarshalIndent(rsf.RemoteServers, " ", " ")
	if err != nil {
		return err
	}
	file, err1 := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
	defer file.Close()

	if err1 != nil {
		return err1
	}

	err = file.Truncate(0)
	if err != nil {
		return err
	}
	_, err2 := file.Write(bytes)
	if err2 != nil {
		return err2
	}
	return nil
}
