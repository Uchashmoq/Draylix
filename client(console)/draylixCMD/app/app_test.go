package app

import (
	"draylix/proxy"
	"fmt"
	"testing"
)

func TestApp_LoadSetting(t *testing.T) {
	factory := proxy.NewRemoteServerFactory()
	factory.Add("vultr1", proxy.RemoteServer{
		Name:    "vultr1",
		Country: "singapore",
		Token:   "test",
		Address: "127.0.0.1:9945",
		Iv:      []byte("CEwjBM3inZuRqo1B"),
		Ikey:    []byte("H5rruxqFyIf0UdUBhJVrd3Bk8F272KPY"),
	})
	err := factory.Store(NodesPath)
	if err != nil {
		fmt.Println(err)
	}
}

func Test_winHttpProxy(t *testing.T) {
	err := winHttpProxy("0.0.0.0:9944", 1)
	if err != nil {
		fmt.Println(err)
	}
}
