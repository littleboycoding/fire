package main

import "os"

const DEFAULT_PORT = "2001"
const DEFAULT_NAME = "Anonymous"

type ArgsType = struct {
	PORT         string
	NAME         string
	ACTION_INDEX int
	MSGPACK      bool
}

var Args = ArgsType{DEFAULT_PORT, DEFAULT_NAME, 0, false}

//Return current program execution arguments. Also set as global variable
func getArgs() ArgsType {
Loop:
	for i := range os.Args {
		switch os.Args[i] {
		case "--port":
			Args.PORT = os.Args[i+1]
		case "--msgpack":
			Args.MSGPACK = true
		case "--name":
			Args.NAME = os.Args[i+1]
		case "send", "scan", "help":
			Args.ACTION_INDEX = i
			break Loop
		}
	}

	return Args
}
