package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/littleboycoding/fire/pkg/ip_utils"
)

const DEFAULT_PORT = "2001"
const DEFAULT_NAME = "Anonymous"

type Action []string
type ActionOption struct {
	port    string
	msgpack bool
	name    string
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
	println("--name					Name to be shown for other users")
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

func createMultipartForm(files []*os.File, name io.Reader) (bytes.Buffer, *multipart.Writer, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for _, file := range files {
		fw, err := w.CreateFormFile("file", file.Name())
		io.Copy(fw, file)

		if err != nil {
			return b, nil, err
		}
	}

	fw, err := w.CreateFormField("name")
	io.Copy(fw, name)

	if err != nil {
		return b, nil, err
	}

	w.Close()

	return b, w, nil
}

//Transfer file to destination, multiple file allowed
func sender(action Action, opt *ActionOption) {
	if len(action) < 3 {
		log.Fatal("Missing required arguments")
	}

	target := "http://" + net.JoinHostPort(action[1], opt.port) + "/drop"

	var files []*os.File

	//Take the rest files
	for _, file := range action[2:] {
		fmt.Printf("Reading file %s\n", file)
		file, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}

		files = append(files, file)
	}

	b, w, err := createMultipartForm(files, strings.NewReader(opt.name))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Transfering to %s\n", action[1])
	client := &http.Client{}
	req, err := http.NewRequest("POST", target, &b)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	status, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(status))
}

func main() {
	var port string = DEFAULT_PORT
	var name string = DEFAULT_NAME
	var actionIndex int
	var msgpack bool

Loop:
	for i := range os.Args {
		switch os.Args[i] {
		case "--port":
			port = os.Args[i+1]
		case "--msgpack":
			msgpack = true
		case "--name":
			name = os.Args[i+1]
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
		name:    name,
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
