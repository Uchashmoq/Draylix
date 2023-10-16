package application

import (
	"draylix/proxy"
	"draylix/utils"
	"encoding/base64"
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
)

func newAddForm(k *Kernel, factory *proxy.RemoteServerFactory) *widget.Form {
	nameEntry := widget.NewEntry()
	countryEntry := widget.NewEntry()
	addrEntry := widget.NewEntry()
	ivEntry := widget.NewEntry()
	ivEntry.SetPlaceHolder("base64")
	ikeyEntry := widget.NewEntry()
	ikeyEntry.SetPlaceHolder("base64")
	tokenEntry := widget.NewEntry()

	form := widget.NewForm(
		&widget.FormItem{Text: "name", Widget: nameEntry},
		&widget.FormItem{Text: "country", Widget: countryEntry},
		&widget.FormItem{Text: "address", Widget: addrEntry},
		&widget.FormItem{Text: "initial vector", Widget: ivEntry},
		&widget.FormItem{Text: "key", Widget: ikeyEntry},
		&widget.FormItem{Text: "token", Widget: tokenEntry},
	)

	form.OnCancel = func() {
		defer k.window.Resize(fyne.NewSize(W, H))
		form.Hide()
		Manager.Scroll.Show()
		clearEntry(nameEntry, countryEntry, addrEntry, ivEntry, ikeyEntry, tokenEntry)

		nameEntry.SetPlaceHolder("")
		ivEntry.SetPlaceHolder("base 64")
		ikeyEntry.SetPlaceHolder("base 64")
		addrEntry.SetPlaceHolder("")
	}

	form.OnSubmit = func() {
		defer k.window.Resize(fyne.NewSize(W, H))
		iv, err := base64.StdEncoding.DecodeString(ivEntry.Text)
		if err != nil {
			ivEntry.SetText("")
			ivEntry.SetPlaceHolder("invalid base64 encoding")
			return
		}

		ikey, err := base64.StdEncoding.DecodeString(ikeyEntry.Text)
		if err != nil {
			ikeyEntry.SetText("")
			ikeyEntry.SetPlaceHolder("invalid base64 encoding")
			return
		}

		if factory.Contains(nameEntry.Text) {
			nameEntry.SetText("")
			nameEntry.SetPlaceHolder("name repeats")
			return
		}

		if !utils.CheckIpv4(addrEntry.Text) {
			addrEntry.SetText("")
			addrEntry.SetPlaceHolder(" incorrect address format")
			return
		}

		server := proxy.RemoteServer{
			Index:   factory.MaxIndex() + 1,
			Name:    nameEntry.Text,
			Country: countryEntry.Text,
			Token:   tokenEntry.Text,
			Address: addrEntry.Text,
			Iv:      iv,
			Ikey:    ikey,
		}

		factory.Add(server.Name, server)
		_ = factory.Store(NodesPath)

		Manager.Load(k)
		Manager.Scroll.Show()
		form.Hide()
		clearEntry(nameEntry, countryEntry, addrEntry, ivEntry, ikeyEntry, tokenEntry)
	}
	form.Hide()
	return form
}

func clearEntry(entries ...*widget.Entry) {
	for _, e := range entries {
		e.SetText("")
	}
}
