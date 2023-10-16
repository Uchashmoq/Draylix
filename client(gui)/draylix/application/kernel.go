package application

import (
	"draylix/dlog"
	"draylix/drlx"
	"draylix/proxy"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	SettingPath = "./configs/settings.json"
	NodesPath   = "./configs/nodes.json"

	UndefinedErr    = errors.New("undefined command")
	TooShortErr     = errors.New("too few command parameters")
	InvalidParamErr = errors.New("invalid command parameters")
)

type Setting struct {
	LogPath       string `json:"log_path"`
	SelectedProxy string `json:"selected_proxy"`
	ProxyOn       int    `json:"proxy_on"`
	ListenAddr    string `json:"listen_addr"`
	ValidTime     int    `json:"valid_time"`
	HandShakeTime int    `json:"hand_shake_time"`
}

func (k *Kernel) LoadSetting(path string) error {
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
	err2 := json.Unmarshal(buf[0:n], &k.Setting)
	if err2 != nil {
		return err2
	}
	buf = nil
	return nil
}

func (k *Kernel) StoreSetting(path string) error {
	bytes, err := json.MarshalIndent(&k.Setting, " ", " ")
	if err != nil {
		return err
	}
	file, err1 := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	defer file.Close()

	if err1 != nil {
		return err1
	}
	_, err2 := file.Write(bytes)
	if err2 != nil {
		return err2
	}
	return nil
}

func (k *Kernel) init() error {
	err := k.LoadSetting(SettingPath)
	if err != nil {
		return err
	}
	factory := proxy.NewRemoteServerFactory()
	err = factory.Load(NodesPath)
	if err != nil {
		fmt.Printf("loading proxy nodes error : %v\n", err)
		os.Exit(1)
	}
	factory.RandServer()
	k.Client = &proxy.LocalServer{RemoteServers: factory}
	k.remoteServers = factory
	if server, ok := factory.RemoteServers[k.Setting.SelectedProxy]; ok {
		proxy.CurrentServer = server
	} else {
		k.Setting.SelectedProxy = ""
		_ = k.StoreSetting(SettingPath)
	}
	k.sigChan = make(chan os.Signal, 1)
	proxy.StartCollectPerSecondData()
	dlog.Debug("application initialized successfully")
	return nil
}

func (k *Kernel) run() {
	drlx.HANDSHAKE_TIMELIMIT = time.Duration(k.Setting.HandShakeTime) * time.Second
	drlx.RESP_DELAY = time.Duration(k.Setting.ValidTime) * time.Second
	go func() {
		err := k.Client.Bind(k.Setting.ListenAddr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()
	go k.waitClose()
}

const (
	llTip = "ll <logLevel>\t :change log level (0:TRACE, 1:DEBUG, 2:INFO, 3:WARN, 4:ERROR, 5:FATAL ,6:OFF)"
	cnTip = "cn <nodeName>\t :change proxy node"
)

func (k *Kernel) SetLogLevel(cmd []string) error {
	if len(cmd) < 2 {
		return errors.Join(TooShortErr, errors.New(llTip))
	}
	level, err := strconv.Atoi(cmd[1])
	if err != nil {
		return errors.Join(InvalidParamErr, err)
	}
	dlog.LogLevel = level
	return nil
}

func (k *Kernel) ShowNodes() error {
	factory := k.remoteServers
	if factory.Quantity() == 0 {
		fmt.Println("no proxy nodes")
		return nil
	}
	i := 1
	for _, s := range factory.RemoteServers {
		if s.Name == proxy.CurrentServer.Name {
			fmt.Printf("%d. [name:%s ,country: %s ,address:%s ,token:%s] (using)\n", i, s.Name, s.Country, s.Address, s.Token)
		} else {
			fmt.Printf("%d. [name:%s ,country: %s ,address:%s ,token:%s]\n", i, s.Name, s.Country, s.Address, s.Token)
		}
		i++
	}
	return nil
}

func (k *Kernel) ChangeNode(cmd []string) error {
	if len(cmd) < 2 {
		return errors.Join(TooShortErr, errors.New(cnTip))
	}
	if ok := k.remoteServers.Contains(cmd[1]); !ok {
		return fmt.Errorf("can not find node : %s", cmd[1])
	}
	proxy.CurrentServer = k.remoteServers.RemoteServers[cmd[1]]
	return nil
}

func (k *Kernel) waitClose() {
	signal.Notify(
		k.sigChan,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGKILL,
	)
	go func() {
		sig := <-k.sigChan
		dlog.Debug("received signal : %v", sig)
		k.cleanup()
		os.Exit(0)
	}()
}

func (k *Kernel) cleanup() {
	err := winHttpProxy(k.Setting.ListenAddr, 0)
	if err != nil {
		dlog.Error("closing proxy failed when cleaning up : %v ", err)
	}
}

func (k *Kernel) ProxyOn() error {
	err := winHttpProxy(k.Setting.ListenAddr, 1)
	if err == nil {
		dlog.Info("system proxy started at %s", k.Setting.ListenAddr)
	}
	return err
}

func (k *Kernel) ProxyOff() error {
	err := winHttpProxy(k.Setting.ListenAddr, 0)
	if err == nil {
		dlog.Info("system proxy closed")
	}
	return err
}

var (
	keyPath            = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`
	keyNameProxyServer = `ProxyServer`
	keyNameProxyEnable = `ProxyEnable`
)

func winHttpProxy(address string, enableProxy int) (err error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.WRITE)
	if err != nil {
		return
	}
	defer key.Close()
	err = key.SetStringValue(keyNameProxyServer, address)
	if err != nil {
		return
	}
	err = key.SetDWordValue(keyNameProxyEnable, uint32(enableProxy))
	if err != nil {
		return
	}
	return nil
}
