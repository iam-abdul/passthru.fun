package client

import (
	"fmt"
	"net"
)

func RunAsClient(clientPort string) {

	// start connection with tcp server
	// on port 8888

	conn, err := net.Dial("tcp", "192.168.80.45:8888")

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	// log the data received from the server

	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("server disconnected")
				break
			} else {
				panic(err)
			}
		}
		fmt.Println("on client --> ", string(buf[:n]))
	}

}
