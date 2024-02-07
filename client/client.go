package client

import (
	"fmt"
	"io"
	"log"
	"net"
)

func joinConnections(conn1 net.Conn, conn2 net.Conn) {
	fmt.Println("joining connections")
	// Copy data from conn1 to conn2
	go func() {
		_, err := io.Copy(conn1, conn2)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Copy data from conn2 to conn1
	go func() {
		_, err := io.Copy(conn2, conn1)
		if err != nil {
			log.Fatal(err)
		}
	}()
}

func RunAsClient() {
	// start tcp connection to the server

	serverConnection, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}

	// requesting the domain
	serverConnection.Write([]byte("domain abdul.com"))

	// reading the response
	isDomainAvailable := make([]byte, 2048)
	n, err := serverConnection.Read(isDomainAvailable)

	if err != nil {
		panic(err)
	}

	fmt.Println("domain availability ", string(isDomainAvailable[:n]))

	if string(isDomainAvailable[:n]) == "true" {

		// client tcp connection
		clientConnection, err := net.Dial("tcp", "localhost:3000")
		if err != nil {
			panic(err)
		}

		// join the connections
		joinConnections(serverConnection, clientConnection)

		// for {
		// 	// read the response from the server
		// 	buf := make([]byte, 2048)
		// 	n, err := serverConnection.Read(buf)
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	fmt.Println("server says ", string(buf[:n]))

		// 	//

		// }
		for {

		}

	} else {
		fmt.Println("domain is not available")
	}

}
