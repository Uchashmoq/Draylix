package utils

import (
	"draylix/dlog"
	"github.com/go-ping/ping"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	IPV4 = `^(((\d{1,2})|(1\d{2})|(2[0-4]\d)|(25[0-5]))\.){3}((\d{1,2})|(1\d{2})|(2[0-4]\d)|(25[0-5]))$`
)

func CheckIpv4(addstr string) bool {
	ipv4Regex, _ := regexp.Compile(IPV4)
	split := strings.Split(addstr, ":")
	if len(split) != 2 {
		return false
	}
	ip := split[0]
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return false
	}
	if port <= 0 || port > 65535 {
		return false
	}
	return ipv4Regex.MatchString(ip)
}

func Put(ch chan int, n int) {
	select {
	case ch <- n:
	default:
	}
}

func PingIP(ip string) (time.Duration, error) {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return 0, err
	}
	pinger.Count = 4
	pinger.Timeout = time.Second * 2
	pinger.SetPrivileged(true)
	err = pinger.Run()
	if err != nil {
		dlog.Error("ping failed %v", err)
		return 0, err
	}
	stats := pinger.Statistics()

	return stats.AvgRtt, nil
}
