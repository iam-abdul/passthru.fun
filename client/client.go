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

		// create a tcp connection to the localhost 3000 server
		// and send forward the request
		localAddr, err := net.ResolveTCPAddr("tcp", "localhost:3000")
		if err != nil {
			log.Fatal(err)
		}

		localConn, err := net.DialTCP("tcp", nil, localAddr)
		if err != nil {
			log.Fatal(err)
		}

		// write the request to the local server
		_, err = localConn.Write(buf[:n])
		if err != nil {
			log.Fatal(err)
		}

		// read the response from the local server
		n, err = localConn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		// write the response back to the client
		_, err = conn.Write(buf[:n])
		if err != nil {
			log.Fatal(err)
		}

		localConn.Close()

	}
}
