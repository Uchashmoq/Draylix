package main

import "draylix/application"

func main() {
	k := application.Kernel{}
	k.Init()
	k.Run()
}
