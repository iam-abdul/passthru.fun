package client

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
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

	httpConn, err := net.Dial("tcp", "localhost:"+clientPort)

	if err != nil {
		panic(err)
	}

	defer httpConn.Close()

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

		// create it as a http request
		// and send it to the http server after replacing the host header

		requestString := string(buf[:n])

		reader := strings.NewReader(requestString)

		// parsing the string to a http request object
		req, err := http.ReadRequest(bufio.NewReader(reader))

		if err != nil {
			fmt.Println("error parsing request string to http request object")
		} else {
			req.Host = "localhost:" + clientPort
			req.Write(httpConn)
		}

		fmt.Println("on client --> ", string(buf[:n]))
	}

}
