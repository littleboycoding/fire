package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/littleboycoding/fire/pkg/ip_utils"
)

const DEFAULT_PORT = "2001"

type Action []string
type ActionOption struct {
	port    string
	msgpack bool
}

func showHelp() {
	println("Fire, the fast and easy to use local network file transfers !")
	println("Command syntax: fire [flags] [action]")
	println("\nAvailable actions")
	println("scan					Scan for local device")
	println("send	[destination] [file path]	Send file to destination")
	println("help	[action]			Show helps")
	println("\nAdditional flags")
	println("--port	[port]				Port for scan and send")
	println("--msgpack				Return result as msgpack (to be use with external program)")
}

/*
scan all local-device within network
*/
func scanner(action Action, opt *ActionOption) {

	var wg sync.WaitGroup

	//First part, getting all possible host address within network
	println("Scanning...")
	ifaces, _ := net.Interfaces()
	ipList := []net.IP{}

	for i := range ifaces {
		addr, err := ifaces[i].Addrs()

		if err != nil {
			log.Fatal(err)
		}

		ip, ipnet, err := net.ParseCIDR(addr[1].String())

		if err != nil {
			log.Fatal(err)
		}

		if ip.IsLoopback() || ip.IsUnspecified() {
			continue
		}

		if ip4 := ip.To4(); ip4 != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ipList = append(ipList, ip_utils.LookupHost(ipnet)...)
			}()
		}
	}
	wg.Wait()

	//Second part, dialing to these address to see if port available
	list := []string{}

	for i := range ipList {
		addr := net.JoinHostPort(ipList[i].String(), opt.port)

		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", addr, time.Second*5)
			if err == nil {
				fmt.Printf("Found %s\n", addr)

				list = append(list, addr)
				conn.Close()
			}
		}()
	}

	wg.Wait()

	fmt.Println(list)
}

func sender(action Action, opt *ActionOption) {
	if len(action) < 2 {
		log.Fatal("Missing required arguments")
	}

	target := "http://" + action[1] + ":" + opt.port
	fmt.Printf("Connecting to %s\n", target)

	if resp, err := http.Get(target); err == nil {
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body := resp.Body
			b, _ := io.ReadAll(body)

			fmt.Println(string(b))
		}
	} else {
		log.Fatal(err)
	}
}

func main() {
	var actionIndex int
	var port string = DEFAULT_PORT
	var msgpack bool

Loop:
	for i := range os.Args {
		switch os.Args[i] {
		case "--port":
			port = os.Args[i+1]
		case "--msgpack":
			msgpack = true
		case "send", "scan", "help":
			actionIndex = i
			break Loop
		}
	}

	if actionIndex == 0 {
		showHelp()
		return
	}

	action := os.Args[actionIndex:]
	opt := &ActionOption{
		port:    port,
		msgpack: msgpack,
	}

	switch action[0] {
	case "help":
		showHelp()
	case "scan":
		scanner(action, opt)
	case "send":
		sender(action, opt)
	}
}
