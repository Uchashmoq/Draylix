package application

import (
	"draylix/application/resources"
	"draylix/dlog"
	"draylix/proxy"
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"sync"
)

type ProxyManager struct {
	Scroll  *container.Scroll
	Factory *proxy.RemoteServerFactory
	options []*Option
}

func (pm *ProxyManager) DelayTest() {
	for i := 0; i < len(pm.options); i++ {
		pm.options[i].SetDelay()
	}
}

func (pm *ProxyManager) Load(k *Kernel) {
	vb := container.NewVBox()
	for _, server := range pm.Factory.SortedServers() {
		op := NewOption(server)
		op.initIcon()
		pm.options = append(pm.options, op)
		pm.initOpSelectedBtn(op, k)
		vb.Add(op.box)
		vb.Add(widget.NewSeparator())
		op.SetDelay()
	}
	pm.Scroll.Content = vb
	pm.Scroll.Refresh()
}

func (pm *ProxyManager) initOpSelectedBtn(op *Option, k *Kernel) {
	op.selectedBtn.OnTapped = func() {
		op.selectedBtn.SetIcon(theme.CheckButtonCheckedIcon())
		for i := 0; i < len(pm.options); i++ {
			if pm.options[i] != op {
				pm.options[i].selectedBtn.SetIcon(theme.CheckButtonIcon())
			}
		}
		proxy.CurrentServer = *op.server
		k.Setting.SelectedProxy = op.server.Name
		err := k.StoreSetting(SettingPath)
		if err != nil {
			k.PopErrDialog(errors.Join(errors.New("can not store config "), err), nil)
		}
		dlog.Debug("OpSelectedBtn tapped,change node to %v", proxy.CurrentServer)
	}
}

type Option struct {
	mu          sync.Mutex
	server      *proxy.RemoteServer
	box         *fyne.Container
	delayLabel  *widget.Label
	selectedBtn *widget.Button
}

func NewOption(server *proxy.RemoteServer) *Option {
	nodeNameIcon := canvas.NewImageFromResource(resources.NodeName)
	nodeNameIcon.FillMode = canvas.ImageFillOriginal

	delayIcon := canvas.NewImageFromResource(resources.Delay)
	delayIcon.FillMode = canvas.ImageFillOriginal

	nameCountryBox := container.NewHBox(nodeNameIcon, widget.NewLabel(fmt.Sprintf("[%s]  %s  %s", server.Country, server.Name, server.Address)))
	delayLabel := widget.NewLabel("testing")
	delayBox := container.NewHBox(delayIcon, delayLabel)
	infoBox := container.NewVBox(nameCountryBox, delayBox)
	selectedBtn := widget.NewButtonWithIcon("", theme.CheckButtonIcon(), nil)
	box := container.NewHBox(selectedBtn, infoBox)
	return &Option{
		server:      server,
		box:         box,
		delayLabel:  delayLabel,
		selectedBtn: selectedBtn,
	}
}

func (op *Option) SetDelay() {
	go func() {
		op.mu.Lock()
		op.delayLabel.SetText("testing")
		delay, _ := op.server.Delay()
		op.delayLabel.SetText(delay)
		op.mu.Unlock()
	}()
}

func (op *Option) initIcon() {
	if op.server.Name == proxy.CurrentServer.Name {
		op.selectedBtn.SetIcon(theme.CheckButtonCheckedIcon())
	} else {
		op.selectedBtn.SetIcon(theme.CheckButtonIcon())
	}
}
