package main

import (
	"flag"
	"fmt"

	"github.com/iam-abdul/go-tcp-tunnel/client"
	"github.com/iam-abdul/go-tcp-tunnel/server"
)

func main() {

	typeOf := flag.String("type", "server", "type of the program to run")

	flag.Parse()

	if *typeOf == "server" {
		server.StartNewServer("localhost:8888")
	} else {
		fmt.Println("used as client")
		client.RunAsClient()
	}

}
