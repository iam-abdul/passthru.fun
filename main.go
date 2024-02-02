package main

import (
	"flag"

	"github.com/iam-abdul/go-tcp-tunnel/server"
)

func main() {

	typeOf := flag.String("type", "server", "type of the program to run")

	if *typeOf == "server" {
		server.StartNewServer("localhost:8888")
	}

}
