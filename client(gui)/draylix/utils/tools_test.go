package utils

import (
	"fmt"
	"testing"
)

func TestPingIP(t *testing.T) {
	d, err := PingIP("45.76.162.86")
	if err != nil {
		println(err)
		t.Fatal(err)
	}
	fmt.Printf("%v", d)
}
