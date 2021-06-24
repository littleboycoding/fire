package stdwriter

import (
	"fmt"
	"log"
	"os"

	"github.com/vmihailenco/msgpack"
)

//Constant for stdWriter
const (
	BOTH     = 0
	NOT_MSP  = 1
	ONLY_MSP = 2
)

var MSGPACK = false

type Response struct {
	Title string
	Data  interface{}
}

type ErrorResponse struct {
	Title string
	Error interface{}
}

/*
stdWriter write value to stdout/stderr depending on value

mode is constant int range from 0 to 2, that indicate to write out if in msgpack or not msgpack mode

0 -> write both msgpack and not msgpack

1 -> only when not msgpack

2 -> only msgpack
*/
func Write(value interface{}, mode int, exitWhenError bool) {
	if mode == 1 && MSGPACK {
		return
	} else if mode == 2 && !MSGPACK {
		return
	}

	switch t := value.(type) {
	case ErrorResponse:
		if MSGPACK {
			msp, err := msgpack.Marshal(t)
			if err != nil {
				log.Fatal(err)
			}
			os.Stderr.Write(msp)
		} else {
			os.Stderr.Write([]byte(fmt.Sprintln(t.Error)))
		}

		if exitWhenError {
			os.Exit(1)
		}
	case Response:
		if MSGPACK {
			msp, err := msgpack.Marshal(t)
			if err != nil {
				log.Fatal(err)
			}
			os.Stdout.Write(msp)
		} else {
			fmt.Println(t.Data)
		}
	default:
		log.Fatal("Invalid response type")
	}
}

//Set if msgpack or not
func SetMsgpack(b bool) {
	MSGPACK = b
}
