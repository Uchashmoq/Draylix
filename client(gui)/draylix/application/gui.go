package application

import (
	"draylix/application/resources"
	"draylix/dlog"
	"draylix/drlx"
	"draylix/proxy"
	"draylix/utils"
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"image/color"
	"os"
	"time"
)

const (
	APP_NAME = "Draylix"
	W        = 430
	H        = 600
)

type Kernel struct {
	Setting       Setting
	Client        *proxy.LocalServer
	remoteServers *proxy.RemoteServerFactory
	sigChan       chan os.Signal
	app           fyne.App
	window        fyne.Window
}

func (k *Kernel) Init() {
	err := k.init()
	k.initApp()
	k.initWindow()
	if err != nil {
		k.PopErrDialog(errors.Join(errors.New("can not load config "), err), func() {
			os.Exit(1)
		})
		k.window.ShowAndRun()
	}

}

func (k *Kernel) Run() {
	k.run()
	k.window.ShowAndRun()
}

func (k *Kernel) PopErrDialog(err error, onclose func()) {
	dg := dialog.NewError(err, k.window)
	dg.SetOnClosed(onclose)
	dg.Show()
}

func (k *Kernel) initApp() {
	err := os.Setenv("FYNE_FONT", "resources/msyh.ttc")
	if err != nil {
		dlog.Error("load font file : %v", err)
	}
	k.app = app.New()
}

func (k *Kernel) initWindow() {
	k.window = k.app.NewWindow(APP_NAME)
	k.window.Resize(fyne.NewSize(W, H))
	k.window.SetIcon(resources.Dtc50)
	k.window.SetFixedSize(true)
	k.window.SetTitle(APP_NAME)
	k.window.CenterOnScreen()
	container := k.initContainer()
	k.window.SetContent(container)
	dlog.LogToFile(k.Setting.LogPath)
	k.window.SetOnClosed(func() {
		_ = dlog.LogWriter.Close()
		k.cleanup()
	})
}

func (k *Kernel) initContainer() *fyne.Container {
	//标志，标题，流量表
	head := k.initHead()
	//控制台（端口，系统代理，日志等级）
	controller := k.initController()
	//节点选择
	nodeChooser := k.initNodeChooser()
	//添加节点的表单
	addForm := k.initAddForm()
	//底边菜单
	tools := k.initToolBar()
	//TODO 选择节点列表
	contain := fyne.NewContainerWithLayout(
		layout.NewVBoxLayout(),
		head,
		widget.NewSeparator(),
		controller,
		nodeChooser,
		addForm,
		layout.NewSpacer(),
		tools,
		layout.NewSpacer(),
	)
	return contain
}

func (k *Kernel) initHead() *fyne.Container {
	//标准
	symbol := canvas.NewImageFromResource(resources.Dtc80)
	symbol.FillMode = canvas.ImageFillOriginal
	//Draylix字样
	txt := canvas.NewText(APP_NAME, color.Black)
	txt.TextSize = 16

	//已用流量，上传，下载速度
	trafficPane := k.initTrafficPane()
	topBox := container.NewHBox(symbol, txt, layout.NewSpacer(), trafficPane)
	return topBox
}

var AllDataMonitor = &widget.Label{}

func (k *Kernel) initTrafficPane() *fyne.Container {
	outMonitor, outZero := getTrafficMonitor("↑")
	inMonitor, inZero := getTrafficMonitor("↓")
	AllDataMonitor = widget.NewLabel("0 B")
	go freshIOData(proxy.PerSecondOut, outMonitor, "↑", outZero)
	go freshIOData(proxy.PerSecondIn, inMonitor, "↓", inZero)
	go freshData(AllDataMonitor)
	trafficPane := container.NewHBox(outMonitor, inMonitor)
	return trafficPane
}

func (k *Kernel) initController() *fyne.Container {
	//设置监听端口
	portController := k.initPortController()
	//开启系统代理
	proxySwitcher := k.initProxySwitcher()
	//设置日志等级
	logController := initLogController()
	return container.NewVBox(
		portController,
		proxySwitcher,
		logController,
		widget.NewSeparator(),
	)
}

func initLogController() *fyne.Container {
	label := widget.NewLabel("log level")
	var levelSelector *widget.Select
	levelSelector = widget.NewSelect(dlog.LogStr, func(levelstr string) {
		level := dlog.LevelNameMap[levelstr]
		dlog.LogLevel = level
	})
	levelSelector.SetSelected(dlog.String(dlog.LogLevel))
	return container.NewHBox(label, layout.NewSpacer(), levelSelector)
}

func (k *Kernel) initPortController() *fyne.Container {
	label := widget.NewLabel("port")
	entry := widget.NewEntry()
	entry.SetText(k.Setting.ListenAddr)
	saveButton := widget.NewButton("restart", func() {
		k.Setting.ListenAddr = entry.Text
		err := k.StoreSetting(SettingPath)
		if err != nil {
			k.PopErrDialog(errors.Join(errors.New("store setting error"), err), func() {
				os.Exit(1)
			})
		}
		os.Exit(0)
	})
	saveButton.Hide()
	entry.OnChanged = func(addr string) {
		if utils.CheckIpv4(addr) && k.Setting.ListenAddr != addr {
			saveButton.Show()
		} else {
			saveButton.Hide()
		}
	}
	portController := container.NewHBox(label, layout.NewSpacer(), entry, saveButton)
	return portController
}

func (k *Kernel) initProxySwitcher() *fyne.Container {
	label := widget.NewLabel("system proxy")
	check := widget.NewCheck("", func(ok bool) {
		var err error
		if ok {
			err = k.ProxyOn()
			k.Setting.ProxyOn = 1
		} else {
			err = k.ProxyOff()
			k.Setting.ProxyOn = 0
		}
		if err != nil {
			k.PopErrDialog(err, nil)
		}
		_ = k.StoreSetting(SettingPath)
	})
	if k.Setting.ProxyOn == 1 {
		check.SetChecked(true)
	} else {
		check.SetChecked(false)
	}
	return container.NewHBox(label, layout.NewSpacer(), check)
}

var (
	Manager = ProxyManager{
		options: make([]*Option, 0),
	}
	AddForm = &widget.Form{}
)

func (k *Kernel) initNodeChooser() *fyne.Container {
	nodeScroll := container.NewScroll(nil)
	nodeScroll.SetMinSize(fyne.NewSize(W-20, H-250))
	Manager.Factory = k.remoteServers
	Manager.Scroll = nodeScroll
	Manager.Load(k)
	return container.NewVBox(nodeScroll)
}

func (k *Kernel) initToolBar() *fyne.Container {
	//测速
	delayTest := k.initDelayTestAction()
	//添加
	add := k.initAddAction()
	//删除
	del := k.initDelAction()
	exg := canvas.NewImageFromResource(resources.Exg)
	exg.FillMode = canvas.ImageFillOriginal
	return container.NewHBox(exg, AllDataMonitor, layout.NewSpacer(), widget.NewToolbar(del, add, delayTest))
}

func (k *Kernel) initDelayTestAction() widget.ToolbarItem {
	itm := widget.NewToolbarAction(resources.Spd, func() {
		Manager.DelayTest()
	})
	return itm
}

func (k *Kernel) initAddAction() widget.ToolbarItem {
	itm := widget.NewToolbarAction(theme.ContentAddIcon(), func() {
		Manager.Scroll.Hide()
		AddForm.Show()
	})
	return itm
}

func (k *Kernel) initAddForm() *widget.Form {
	AddForm = newAddForm(k, k.remoteServers)
	return AddForm
}

func (k *Kernel) initDelAction() widget.ToolbarItem {
	itm := widget.NewToolbarAction(theme.DeleteIcon(), func() {
		var dg *dialog.EntryDialog
		dg = dialog.NewEntryDialog("Delete", "name", func(name string) {
			if name == proxy.CurrentServer.Name {
				dg.SetText("")
				dg.SetPlaceholder("proxy in use")
				return
			}
			if !k.remoteServers.Contains(name) {
				dg.SetText("")
				dg.SetPlaceholder("unknown name")
				return
			}

			k.remoteServers.Delete(name)
			err := k.remoteServers.Store(NodesPath)
			if err != nil {
				dlog.Error("%v", err)
			}
			Manager.Load(k)
		}, k.window)
	})
	return itm
}

func getTrafficMonitor(ar string) (*widget.Label, string) {
	zero := fmt.Sprintf("%s 0 B/s", ar)
	monitor := widget.NewLabel(zero)
	return monitor, zero
}

var (
	zeroWait   = 2 * time.Second
	dataFreshD = 1 * time.Second
)

func freshData(label *widget.Label) {
	ticker := time.NewTicker(dataFreshD)
	for {
		select {
		case <-ticker.C:
			label.SetText(proxy.GetTraffic())
		}
	}
}

func freshIOData(ch chan int, label *widget.Label, ar, zero string) {
	var speed string
	for {
		select {
		case p := <-ch:
			speed = drlx.BytesFormat(int64(p))
			label.SetText(fmt.Sprintf("%s %s/s", ar, speed))
		case <-time.After(zeroWait):
			label.SetText(zero)
		}
	}
}
