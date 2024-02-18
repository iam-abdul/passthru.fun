package main

import (
	"flag"
	"fmt"

	"github.com/iam-abdul/go-tcp-tunnel/client"
	"github.com/iam-abdul/go-tcp-tunnel/server"
)

func main() {

	typeOf := flag.String("type", "client", "type of the program to run")
	port := flag.String("port", "8888", "port to run")
	domain := flag.String("domain", "", "domain to request")
	verbose := flag.Bool("verbose", false, "verbose mode")
	flag.Parse()

	if *typeOf == "client" && *domain == "" {
		fmt.Println("please provide a domain name")
		return
	}

	if *typeOf == "server" {
		server.StartNewServer(*port)
	} else {
		// fmt.Println("used as client")
		client.RunAsClientV2(*port, *domain, *verbose)
	}

}
