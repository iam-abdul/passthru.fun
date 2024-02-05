package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func RunAsClient() {
	// start tcp connection to the server

	_, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}

	// send the request to get a domain name
	// request should be a post request with the body containing the domain name
	req, err := http.NewRequest("POST", "http://localhost:8888", strings.NewReader("tester.com"))

	if err != nil {
		panic(err)
	}

	req.Host = "abdul.com"

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	bodyString := string(bodyBytes)

	fmt.Println("response from server ", resp.Status, resp.StatusCode, bodyString)

	defer resp.Body.Close()

	// now we will listen to the server connection

	// buf := make([]byte, 2048)

	for {

		// n, err := serverConnection.Read(buf)
		// if err != nil {
		// 	panic(err)
		// }

		// fmt.Println("received from server ", string(buf[:n]))

	}

}
