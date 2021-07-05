package main

import "os"

type ArgsType struct {
	MSGPACK bool
	NAME    string
	PORT    string
}

var Args = ArgsType{false, "Anonymous", "2001"}

func getArgs() {
	for i, arg := range os.Args {
		switch arg {
		case "--msgpack":
			Args.MSGPACK = true
		case "--name":
			Args.NAME = os.Args[i+1]
		case "--port":
			Args.PORT = os.Args[i+1]
		}
	}
}
