package main

import (
	"bufio"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/julienschmidt/httprouter"
	stdw "github.com/littleboycoding/fire/pkg/stdwriter"
)

type Action struct {
	Action string
	Data   interface{}
}

type ErrorAction struct {
	Title string
	Error error
}

const DEFAULT_PORT = "2001"

func verifyFile(files []*multipart.FileHeader, from string) (bool, error) {
	//Pretty display file list when not msgpack
	if !Args.MSGPACK {
		var fileList []string

		for _, file := range files {
			s := fmt.Sprintf("%s %v bytes", file.Filename, file.Size)
			fileList = append(fileList, s)
		}

		fileInfo := strings.Join(fileList, "\n")
		res := stdw.Response{
			Title: "FILE_LIST",
			Data:  fileInfo,
		}
		stdw.Write(res, stdw.NOT_MSP, false)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		res := stdw.Response{Title: "ACCEPT_FILE", Data: nil}
		if !Args.MSGPACK {
			res.Data = "Do you want to accept ? [y][N]"
		} else {
			type Data struct {
				From  string
				Files []*multipart.FileHeader
			}
			res.Data = Data{
				from,
				files,
			}
		}
		stdw.Write(res, stdw.BOTH, false)
		text, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		text = strings.TrimSpace(text)

		switch text {
		case "y", "Y":
			return true, nil
		case "n", "N", "":
			return false, nil
		}

		e := stdw.ErrorResponse{
			Title: "ERROR_INCORRECT_ANSWER",
			Error: "Incorrect answer, try again",
		}
		stdw.Write(e, stdw.BOTH, false)
	}
}

//Prompt for file destination and write file
func writeFile(files []*multipart.FileHeader) error {

	reader := bufio.NewReader(os.Stdin)

	res := stdw.Response{
		Title: "FILE_DESTINATION",
		Data:  "File destination",
	}
	stdw.Write(res, stdw.BOTH, false)

	rawDist, err := reader.ReadString('\n')

	if err != nil {
		return err
	}

	trimmedDist := strings.TrimSpace(rawDist)

	for _, file := range files {
		dist := path.Join(trimmedDist, file.Filename)
		r, err := file.Open()
		if err != nil {
			return err
		}

		b, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		err = os.WriteFile(dist, b, 0777)
		if err != nil {
			return err
		}
	}

	return nil
}

//Helper for writing and display readable error
func errorResponse(w http.ResponseWriter, title string, err error) {
	//Write error to sender
	fmt.Fprint(w, "Error occured on server")

	//Now display error on server
	{
		e := stdw.ErrorResponse{
			Title: title,
			Error: err,
		}
		stdw.Write(e, stdw.BOTH, false)
	}

	{
		e := stdw.ErrorResponse{
			Title: "TRANSFER_FAILED",
			Error: "Transfer failed",
		}
		stdw.Write(e, stdw.NOT_MSP, false)
	}
}

func main() {
	getArgs()

	stdw.SetMsgpack(Args.MSGPACK)

	router := httprouter.New()
	busy := false

	// println("Discoverable on port " + DEFAULT_PORT)
	res := stdw.Response{
		Title: "DISCOVERABLE",
		Data:  "Discoverable on port " + DEFAULT_PORT,
	}
	stdw.Write(res, stdw.NOT_MSP, false)

	router.POST("/drop", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if busy {
			fmt.Fprint(w, "Receiver is busy")
			return
		}

		busy = true
		defer func() {
			busy = false
			r.Body.Close()
		}()

		r.ParseMultipartForm(2048)
		res := stdw.Response{
			Title: "INCOMING_FILE",
			Data:  fmt.Sprintf("File transfer incoming from %s", r.MultipartForm.Value["name"][0]),
		}
		stdw.Write(res, stdw.NOT_MSP, false)
		files := r.MultipartForm.File["file"]

		accept, err := verifyFile(files, r.MultipartForm.Value["name"][0])

		if err != nil {
			errorResponse(w, "ERROR_VERIFY_FILE", err)
			return
		}

		if accept {
			err := writeFile(files)
			if err != nil {
				errorResponse(w, "ERROR_WRITE_FILE", err)
				return
			}

			res := stdw.Response{
				Title: "FILE_WRITTEN",
				Data:  "File written",
			}
			stdw.Write(res, stdw.BOTH, false)
			fmt.Fprint(w, "Transfer success")
		} else {
			res := stdw.Response{
				Title: "FILE_DENIED",
				Data:  "File denied",
			}
			stdw.Write(res, stdw.BOTH, false)
			fmt.Fprint(w, "Transfer denied")
		}
	})

	e := stdw.ErrorResponse{
		Title: "ERROR_SERVER_SHUTDOWN",
		Error: http.ListenAndServe(":"+DEFAULT_PORT, router),
	}
	stdw.Write(e, stdw.BOTH, true)
}
