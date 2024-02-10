package main

import (
	"flag"
	"fmt"

	"github.com/iam-abdul/go-tcp-tunnel/client"
	"github.com/iam-abdul/go-tcp-tunnel/server"
)

func main() {

	typeOf := flag.String("type", "server", "type of the program to run")
	port := flag.String("port", "8888", "port to run")
	domain := flag.String("domain", "test", "domain to request")
	flag.Parse()

	if *typeOf == "server" {
		server.StartNewServer(*port)
	} else {
		fmt.Println("used as client")
		client.RunAsClient(*port, *domain)
	}

}
