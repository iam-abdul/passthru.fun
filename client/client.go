package client

import (
	"log"
	"net"
)

func RunAsClient() {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:8888")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	// will request the domain here
	response, err := conn.Write([]byte("domain abdul.com"))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Response: ", string(response))

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(buf[:n]))

	}
}
