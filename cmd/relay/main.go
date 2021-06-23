package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
)

const DEFAULT_PORT = "2001"

//Prompt for file destination and write file
func saveFile(f *multipart.FileHeader) error {
	var err error
	r, err := f.Open()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Write %s to: ", f.Filename)
	text, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	trimedText := strings.TrimSpace(text)

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	err = os.WriteFile(trimedText, b, 0777)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	router := httprouter.New()
	busy := false

	println("Discoverable on port " + DEFAULT_PORT)

	// router.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// 	h := w.Header()
	// 	h.Add("Content-Type", "application/json")
	// 	fmt.Fprintf(w, `{ "device": "littlelaptop", "accept": %t, "host": "192.168.1.124", "port": "2001" }`, !busy)
	// })

	router.POST("/drop", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if busy {
			fmt.Fprint(w, "Receiver is busy")
			return
		}

		busy = true
		defer r.Body.Close()

		r.ParseMultipartForm(2048)
		fmt.Printf("File transfer incoming from %s\n", r.MultipartForm.Value["name"][0])
		files := r.MultipartForm.File["file"]
		for i := 0; i < len(files); i++ {
			if err := saveFile(files[i]); err != nil {
				fmt.Println(err)
				i--
			}
		}

		fmt.Println("File written")
		fmt.Fprint(w, "Success")
		busy = false
	})

	log.Fatal(http.ListenAndServe(":"+DEFAULT_PORT, router))
}
