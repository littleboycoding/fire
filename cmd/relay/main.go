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
	"github.com/littleboycoding/fire/pkg/stdwriter"
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
	var fileList []string

	for _, file := range files {
		s := fmt.Sprintf("%s %v bytes", file.Filename, file.Size)
		fileList = append(fileList, s)
	}

	fileInfo := strings.Join(fileList, "\n")
	res := stdwriter.Response2{
		Title: "FILE_LIST",
		Data:  fileInfo,
	}
	stdwriter.Write2(res, stdwriter.NOT_MSP, Args.MSGPACK, false)

	reader := bufio.NewReader(os.Stdin)
	for {
		res := stdwriter.Response2{Title: "ACCEPT_FILE", Data: nil}
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
		stdwriter.Write2(res, stdwriter.BOTH, Args.MSGPACK, false)
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

		e := stdwriter.ErrorResponse{
			Title: "INCORRECT_ANSWER",
			Error: "Incorrect answer, try again",
		}
		stdwriter.Write2(e, stdwriter.BOTH, Args.MSGPACK, false)
	}
}

//Prompt for file destination and write file
func writeFile(files []*multipart.FileHeader) error {

	reader := bufio.NewReader(os.Stdin)

	res := stdwriter.Response2{
		Title: "FILE_DESTINATION",
		Data:  "File destination",
	}
	stdwriter.Write2(res, stdwriter.BOTH, Args.MSGPACK, false)

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

func errorResponse(w http.ResponseWriter, err error) {
	//Write error to sender
	fmt.Fprint(w, "Error occured on server")

	//Now display error on server
	{
		e := stdwriter.ErrorResponse{
			Title: "ERROR_OCCURED",
			Error: err,
		}
		stdwriter.Write2(e, stdwriter.BOTH, Args.MSGPACK, false)
	}

	{
		e := stdwriter.ErrorResponse{
			Title: "TRANSFER_FAILED",
			Error: "Transfer failed",
		}
		stdwriter.Write2(e, stdwriter.NOT_MSP, Args.MSGPACK, false)
	}
}

func main() {
	getArgs()

	router := httprouter.New()
	busy := false

	// println("Discoverable on port " + DEFAULT_PORT)
	res := stdwriter.Response2{
		Title: "DISCOVERABLE",
		Data:  "Discoverable on port " + DEFAULT_PORT,
	}
	stdwriter.Write2(res, stdwriter.NOT_MSP, Args.MSGPACK, false)

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
		res := stdwriter.Response2{
			Title: "INCOMING_FILE",
			Data:  fmt.Sprintf("File transfer incoming from %s", r.MultipartForm.Value["name"][0]),
		}
		stdwriter.Write2(res, stdwriter.NOT_MSP, Args.MSGPACK, false)
		files := r.MultipartForm.File["file"]

		accept, err := verifyFile(files, r.MultipartForm.Value["name"][0])

		if err != nil {
			errorResponse(w, err)
			return
		}

		if accept {
			err := writeFile(files)
			if err != nil {
				errorResponse(w, err)
				return
			}

			res := stdwriter.Response2{
				Title: "FILE_WRITTEN",
				Data:  "File written",
			}
			stdwriter.Write2(res, stdwriter.BOTH, Args.MSGPACK, false)
			fmt.Fprint(w, "Transfer success")
		} else {
			res := stdwriter.Response2{
				Title: "FILE_DENIED",
				Data:  "File denied",
			}
			stdwriter.Write2(res, stdwriter.BOTH, Args.MSGPACK, false)
			fmt.Fprint(w, "Transfer denied")
		}
	})

	e := stdwriter.ErrorResponse{
		Title: "SERVER_SHUTDOWN",
		Error: http.ListenAndServe(":"+DEFAULT_PORT, router),
	}
	stdwriter.Write2(e, stdwriter.BOTH, Args.MSGPACK, true)
}
