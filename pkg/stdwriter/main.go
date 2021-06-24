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

type Response struct {
	Success bool
	Value   interface{}
}

type Response2 struct {
	Title string
	Data  interface{}
}

type ErrorResponse struct {
	Title string
	Error interface{}
}

func (recv Response2) Write() string {
	return fmt.Sprint(recv.Data)
}

func (recv ErrorResponse) Write() string {
	return fmt.Sprint(recv.Error)
}

type StdWriter interface {
	Write() string
}

/*
stdWriter write value to stdout/stderr depending on value

mode is int range from 0 to 2, that indicate to write out if in msgpack or not msgpack mode

0 -> write both msgpack and not msgpack

1 -> only when not msgpack

2 -> only msgpack
*/
func Write(value interface{}, mode int, isMsgpack bool, exitWhenError bool) {
	if mode == 1 && isMsgpack {
		return
	} else if mode == 2 && !isMsgpack {
		return
	}

	switch t := value.(type) {
	case error:
		if isMsgpack {
			res := Response{false, t.Error()}
			msp, err := msgpack.Marshal(res)
			if err != nil {
				log.Fatal(err)
			}
			os.Stderr.Write(msp)
		} else {
			os.Stderr.Write([]byte(t.Error()))
		}

		if exitWhenError {
			os.Exit(1)
		}
	default:
		if isMsgpack {
			res := Response{true, t}
			msp, err := msgpack.Marshal(res)
			if err != nil {
				log.Fatal(err)
			}
			os.Stdout.Write(msp)
		} else {
			fmt.Println(t)
		}
	}
}

func Write2(value interface{}, mode int, isMsgpack bool, exitWhenError bool) {
	if mode == 1 && isMsgpack {
		return
	} else if mode == 2 && !isMsgpack {
		return
	}

	switch t := value.(type) {
	case ErrorResponse:
		if isMsgpack {
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
	case Response2:
		if isMsgpack {
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

func Write3(value StdWriter, mode int, isMsgpack bool, exitWhenError bool) {
	if mode == 1 && isMsgpack {
		return
	} else if mode == 2 && !isMsgpack {
		return
	}

	switch t := value.(type) {
	case ErrorResponse:
		if isMsgpack {
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
	case Response2:
		if isMsgpack {
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
