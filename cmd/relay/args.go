package main

import "os"

type ArgsType struct {
	MSGPACK bool
}

var Args = ArgsType{false}

func getArgs() {
	for _, arg := range os.Args {
		switch arg {
		case "--msgpack":
			Args.MSGPACK = true
		}
	}
}
