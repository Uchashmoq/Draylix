package app

import (
	"bufio"
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
	"strings"
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
	ListenAddr    string `json:"listen_addr"`
	ValidTime     int    `json:"valid_time"`
	HandShakeTime int    `json:"hand_shake_time"`
}

func (app *App) LoadSetting(path string) error {
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
	err2 := json.Unmarshal(buf[0:n], &app.Setting)
	if err2 != nil {
		return err2
	}
	buf = nil
	return nil
}

func (app *App) StoreSetting(path string) error {
	bytes, err := json.MarshalIndent(&app.Setting, " ", " ")
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

type App struct {
	Setting       Setting
	Client        *proxy.LocalServer
	remoteServers *proxy.RemoteServerFactory
	sigChan       chan os.Signal
}

func (app *App) Init() {
	err := app.LoadSetting(SettingPath)
	if err != nil {
		fmt.Printf("loading settings error : %v\n", err)
		os.Exit(1)
	}
	factory := proxy.NewRemoteServerFactory()
	err = factory.Load(NodesPath)
	if err != nil {
		fmt.Printf("loading proxy nodes error : %v\n", err)
		os.Exit(1)
	}
	factory.RandServer()
	app.Client = &proxy.LocalServer{RemoteServers: factory}
	app.remoteServers = factory
	app.sigChan = make(chan os.Signal, 1)
	dlog.Debug("application initialized successfully")
}

func (app *App) Run() {
	drlx.HANDSHAKE_TIMELIMIT = time.Duration(app.Setting.HandShakeTime) * time.Second
	drlx.RESP_DELAY = time.Duration(app.Setting.ValidTime) * time.Second
	go func() {
		err := app.Client.Bind(app.Setting.ListenAddr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()
	app.getCMD()
	go app.waitClose()
}

func (app *App) getCMD() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSuffix(input, "\r\n")
		if len(input) == 0 {
			continue
		}
		app.execute(splitCMD(input))
	}
}

func splitCMD(cmd string) []string {
	ss := make([]string, 0)
	for _, s := range strings.Split(cmd, " ") {
		if len(s) > 0 && []byte(s)[0] != ' ' {
			ss = append(ss, s)
		}
	}
	return ss
}

func (app *App) execute(cmd []string) {
	if len(cmd) == 0 {
		return
	}
	var err error
	switch cmd[0] {
	case "help":
		ShowHelp()
	case "ll":
		err = app.SetLogLevel(cmd)
	case "cn":
		err = app.ChangeNode(cmd)
	case "node":
		err = app.ShowNodes()
	case "on":
		err = app.ProxyOn()
	case "off":
		err = app.ProxyOff()
	case "data":
		seeData()
	default:
		err = TooShortErr
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("**done**\n")
}

func seeData() {
	fmt.Println(proxy.GetTraffic())
}

func ShowHelp() {
	fmt.Println("help :")
	fmt.Println(llTip)
	fmt.Println(cnTip)
	fmt.Println(nodeTip)
	fmt.Println(onTip)
	fmt.Println(offTip)
	fmt.Println(dataTip)
}

const (
	llTip   = "ll <logLevel>\t :change log level (0:TRACE, 1:DEBUG, 2:INFO, 3:WARN, 4:ERROR, 5:FATAL ,6:OFF)"
	cnTip   = "cn <nodeName>\t :change proxy node"
	nodeTip = "node\t :list all proxy nodes"
	onTip   = "on\t :proxy on"
	offTip  = "off\t :proxy off"
	dataTip = "data\t :see data"
)

func (app *App) SetLogLevel(cmd []string) error {
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

func (app *App) ShowNodes() error {
	factory := app.remoteServers
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

func (app *App) ChangeNode(cmd []string) error {
	if len(cmd) < 2 {
		return errors.Join(TooShortErr, errors.New(cnTip))
	}
	if ok := app.remoteServers.Contains(cmd[1]); !ok {
		return fmt.Errorf("can not find node : %s", cmd[1])
	}
	proxy.CurrentServer = app.remoteServers.RemoteServers[cmd[1]]
	return nil
}

func (app *App) waitClose() {
	signal.Notify(
		app.sigChan,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGKILL,
	)
	go func() {
		sig := <-app.sigChan
		dlog.Debug("received signal : %v", sig)
		app.cleanup()
		os.Exit(0)
	}()
}

func (app *App) cleanup() {
	err := winHttpProxy(app.Setting.ListenAddr, 0)
	if err != nil {
		dlog.Error("closing proxy failed when cleaning up : %v ", err)
	}
}

func (app *App) ProxyOn() error {
	err := winHttpProxy(app.Setting.ListenAddr, 1)
	if err == nil {
		dlog.Info("system proxy started at %s", app.Setting.ListenAddr)
	}
	return err
}

func (app *App) ProxyOff() error {
	err := winHttpProxy(app.Setting.ListenAddr, 0)
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
