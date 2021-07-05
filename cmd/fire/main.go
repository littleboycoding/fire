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
	stdw "github.com/littleboycoding/fire/pkg/stdwriter"
)

func createMultipartForm(files []*os.File, name io.Reader) (bytes.Buffer, *multipart.Writer, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for _, file := range files {
		stat, err := file.Stat()
		if err != nil {
			return b, nil, err
		}

		fw, err := w.CreateFormFile("file", stat.Name())
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
	fmt.Println("--dev	[name]				Device name to scan on, Default is to scan all")
	fmt.Println("--msgpack				Return result as msgpack (to be use with external program)")
	fmt.Println("--name					Name to be shown for other users")
	fmt.Println("--include				Include yourself in scan result, Useful for development")
}

/*
scan all local-device within network
*/
func scanner(action []string, ifname string) {
	var wg sync.WaitGroup

	res := stdw.Response{
		Title: "SCANNING",
		Data:  "Scanning...",
	}
	stdw.Write(res, stdw.NOT_MSP, false)
	var ifaces []net.Interface
	if ifname == "" {
		ifs, _ := net.Interfaces()
		ifaces = ifs
	} else {
		ifs, _ := net.InterfaceByName(ifname)
		ifaces = []net.Interface{*ifs}
	}
	ipList := []net.IP{}

	for i := range ifaces {
		addr, err := ifaces[i].Addrs()

		if err != nil {
			e := stdw.ErrorResponse{
				Title: "ERROR_GETTING_ADDRESS",
				Error: err,
			}
			stdw.Write(e, stdw.BOTH, true)
		}

		for x := 0; x < len(addr); x++ {
			ip, ipnet, err := net.ParseCIDR(addr[x].String())

			if err != nil {
				e := stdw.ErrorResponse{
					Title: "ERROR_PARSING_CIDR",
					Error: err,
				}
				stdw.Write(e, stdw.BOTH, true)
			}

			if ip.IsLoopback() || ip.IsUnspecified() {
				continue
			}

			if ip4 := ip.To4(); ip4 != nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ipList = append(ipList, ip_utils.LookupHost(ip4, ipnet, Args.INCLUDE)...)
				}()
				break
			}
		}
	}
	wg.Wait()

	for i := range ipList {
		currentIp := ipList[i].String()
		addr := net.JoinHostPort(currentIp, Args.PORT)

		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", addr, time.Second*5)
			if err == nil {
				res := stdw.Response{
					Title: "FOUND_ADDRESS",
					Data:  nil,
				}
				if !Args.MSGPACK {
					res.Data = fmt.Sprintf("Found %s", currentIp)
				} else {
					res.Data = currentIp
				}
				stdw.Write(res, stdw.BOTH, false)

				conn.Close()
			}
		}()
		time.Sleep(time.Microsecond)
	}

	wg.Wait()
}

//Transfer file to destination, multiple file allowed
func sender(action []string) {
	if len(action) < 3 {
		e := stdw.ErrorResponse{
			Title: "ERROR_MISSING_ARGUMENTS",
			Error: "Missing required arguments",
		}
		stdw.Write(e, stdw.BOTH, true)
	}

	target := "http://" + net.JoinHostPort(action[1], Args.PORT) + "/drop"

	var files []*os.File

	//Take the rest files
	for _, file := range action[2:] {
		res := stdw.Response{
			Title: "READING_FILE",
			Data:  fmt.Sprintf("Reading %s", file),
		}
		stdw.Write(res, stdw.NOT_MSP, false)
		file, err := os.Open(file)
		if err != nil {
			e := stdw.ErrorResponse{
				Title: "ERROR_OPEN_FILE",
				Error: err,
			}
			stdw.Write(e, stdw.BOTH, true)
		}

		files = append(files, file)
	}

	b, w, err := createMultipartForm(files, strings.NewReader(Args.NAME))
	if err != nil {
		e := stdw.ErrorResponse{
			Title: "ERROR_CREATE_MULTIPART_FORM",
			Error: err,
		}
		stdw.Write(e, stdw.BOTH, true)
	}

	res := stdw.Response{
		Title: "TRANSFERING_FILE",
		Data:  fmt.Sprintf("Transfering to %s", action[1]),
	}
	stdw.Write(res, stdw.NOT_MSP, false)
	client := &http.Client{}
	req, err := http.NewRequest("POST", target, &b)
	if err != nil {
		e := stdw.ErrorResponse{
			Title: "ERROR_NEW_POST_REQUEST",
			Error: err,
		}
		stdw.Write(e, stdw.BOTH, true)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		e := stdw.ErrorResponse{
			Title: "ERROR_CLIENT_DO",
			Error: err,
		}
		stdw.Write(e, stdw.BOTH, true)
	}

	defer resp.Body.Close()

	status, err := io.ReadAll(resp.Body)
	if err != nil {
		e := stdw.ErrorResponse{
			Title: "ERROR_READ_BODY",
			Error: err,
		}
		stdw.Write(e, stdw.BOTH, true)
	}

	res = stdw.Response{
		Title: "BODY_DATA",
		Data:  string(status),
	}
	stdw.Write(res, stdw.BOTH, false)
}

func main() {
	getArgs()
	stdw.SetMsgpack(Args.MSGPACK)

	if Args.ACTION_INDEX == 0 {
		showHelp()
		return
	}

	action := os.Args[Args.ACTION_INDEX:]

	switch action[0] {
	case "help":
		showHelp()
	case "scan":
		scanner(action, Args.DEVICE)
	case "send":
		sender(action)
	}
}
