package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/littleboycoding/fire/pkg/ip_utils"
	"github.com/littleboycoding/fire/pkg/stdwriter"
)

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

func showHelp() {
	fmt.Println("Fire, the fast and easy to use local network file transfers !")
	fmt.Println("Command syntax: fire [flags] [action]")
	fmt.Println("\nAvailable actions")
	fmt.Println("scan					Scan for local device")
	fmt.Println("send	[destination] [file path]	Send file to destination")
	fmt.Println("help	[action]			Show helps")
	fmt.Println("\nAdditional flags")
	fmt.Println("--port	[port]				Port for scan and send")
	fmt.Println("--msgpack				Return result as msgpack (to be use with external program)")
	fmt.Println("--name					Name to be shown for other users")
}

/*
scan all local-device within network
*/
func scanner(action []string) {
	var wg sync.WaitGroup

	//First part, getting all possible host address within network
	stdwriter.Write("Scanning...", stdwriter.NOT_MSP, Args.MSGPACK, false)
	ifaces, _ := net.Interfaces()
	ipList := []net.IP{}

	for i := range ifaces {
		addr, err := ifaces[i].Addrs()

		if err != nil {
			stdwriter.Write(err, stdwriter.BOTH, Args.MSGPACK, true)
		}

		for x := 0; x < len(addr); x++ {
			ip, ipnet, err := net.ParseCIDR(addr[x].String())

			if err != nil {
				stdwriter.Write(err, stdwriter.BOTH, Args.MSGPACK, true)
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
	}
	wg.Wait()

	//Second part, dialing to these address to see if port available
	// list := []string{}

	for i := range ipList {
		currentIp := ipList[i].String()
		addr := net.JoinHostPort(currentIp, Args.PORT)

		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", addr, time.Second*5)
			if err == nil {
				stdwriter.Write(fmt.Sprintf("Found %s", currentIp), stdwriter.NOT_MSP, Args.MSGPACK, false)
				stdwriter.Write(currentIp, stdwriter.ONLY_MSP, Args.MSGPACK, false)

				// list = append(list, currentIp)
				conn.Close()
			}
		}()
	}

	wg.Wait()

	// stdwriter.Write(list, stdwriter.BOTH, Args.MSGPACK, false)
}

//Transfer file to destination, multiple file allowed
func sender(action []string) {
	if len(action) < 3 {
		stdwriter.Write("Missing required arguments", stdwriter.BOTH, Args.MSGPACK, true)
	}

	target := "http://" + net.JoinHostPort(action[1], Args.PORT) + "/drop"

	var files []*os.File

	//Take the rest files
	for _, file := range action[2:] {
		stdwriter.Write(fmt.Sprintf("Reading %s", file), stdwriter.NOT_MSP, Args.MSGPACK, false)
		file, err := os.Open(file)
		if err != nil {
			stdwriter.Write(err, stdwriter.BOTH, Args.MSGPACK, true)
		}

		files = append(files, file)
	}

	b, w, err := createMultipartForm(files, strings.NewReader(Args.NAME))
	if err != nil {
		stdwriter.Write(err, stdwriter.BOTH, Args.MSGPACK, true)
	}

	stdwriter.Write(fmt.Sprintf("Transfering to %s", action[1]), stdwriter.NOT_MSP, Args.MSGPACK, false)
	client := &http.Client{}
	req, err := http.NewRequest("POST", target, &b)
	if err != nil {
		stdwriter.Write(err, stdwriter.BOTH, Args.MSGPACK, true)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		stdwriter.Write(err, stdwriter.BOTH, Args.MSGPACK, true)
	}

	defer resp.Body.Close()

	status, err := io.ReadAll(resp.Body)
	if err != nil {
		stdwriter.Write(err, stdwriter.BOTH, Args.MSGPACK, true)
	}

	stdwriter.Write(string(status), stdwriter.BOTH, Args.MSGPACK, false)
}

func main() {
	getArgs()

	if Args.ACTION_INDEX == 0 {
		showHelp()
		return
	}

	action := os.Args[Args.ACTION_INDEX:]

	switch action[0] {
	case "help":
		showHelp()
	case "scan":
		scanner(action)
	case "send":
		sender(action)
	}
}
