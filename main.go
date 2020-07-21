package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func getIPList(addr string) ([]string, error) {
	defer wg.Done()
	timeout := 5000 * time.Millisecond
	d := net.Dialer{Timeout: timeout}
	tcpCoon, err := d.Dial("tcp", addr+":135") //建立连接
	if err != nil {
		//fmt.Println("conn erro")
		return nil, err
	}
	defer tcpCoon.Close()
	sendData := "\x05\x00\x0b\x03\x10\x00\x00\x00\x48\x00\x00\x00\x01\x00\x00\x00\xb8\x10\xb8\x10\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00\x01\x00\xc4\xfe\xfc\x99\x60\x52\x1b\x10\xbb\xcb\x00\xaa\x00\x21\x34\x7a\x00\x00\x00\x00\x04\x5d\x88\x8a\xeb\x1c\xc9\x11\x9f\xe8\x08\x00\x2b\x10\x48\x60\x02\x00\x00\x00"
	n, err := tcpCoon.Write([]byte(sendData))
	if err != nil {
		//fmt.Println("write1 erro")
		return nil, err
	}
	recvData := make([]byte, 4096)
	readTimeout := 5 * time.Second
	err = tcpCoon.SetReadDeadline(time.Now().Add(readTimeout))
	n, err = tcpCoon.Read(recvData)
	if err != nil {
		//fmt.Println("read1 erro")
		return nil, err
	}
	sendData2 := "\x05\x00\x00\x03\x10\x00\x00\x00\x18\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x05\x00"
	n, err = tcpCoon.Write([]byte(sendData2))
	if err != nil {
		//fmt.Println("write2 erro")
		return nil, err
	}
	err = tcpCoon.SetReadDeadline(time.Now().Add(readTimeout))
	n, err = tcpCoon.Read(recvData)
	if err != nil {
		//fmt.Println("read2 erro")
		return nil, err
	}
	recvStr := string(recvData[:n])
	recvStr_v2 := recvStr[42:]
	packet_v2_end := strings.Index(recvStr_v2, "\x09\x00\xff\xff\x00\x00")
	packet_v2 := recvStr_v2[:packet_v2_end]
	hostname_list := strings.Split(packet_v2, "\x00\x00")
	for _, value := range hostname_list {
		fmt.Println(value)
	}

	return hostname_list, nil
}

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

var (
	targetIP string
	help     bool
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("./scanIP -t CIDR")
}

func init() {
	flag.StringVar(&targetIP, "t", "", "CIDR")
	flag.BoolVar(&help, "h", false, "Show this help")
	flag.Usage = usage
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	fmt.Println(time.Now())
	ip, _ := Hosts(targetIP)
	//fmt.Println(ip)
	for _, value := range ip {
		wg.Add(1)
		go getIPList(value)
	}
	wg.Wait()
	fmt.Println(time.Now())
}
