package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/iam-abdul/go-tcp-tunnel/client"
)

func GenerateRandomSubDomain() string {
	domain := "abd.com"

	rand.NewSource(time.Now().UnixNano())

	subdomain := fmt.Sprintf("%x", rand.Int63())

	return subdomain + "." + domain
}

func ServerSide() {

	// start a tcp server on port 8888
	// listen for incoming connections

	sln, err := net.Listen("tcp", "192.168.80.45:8888")

	if err != nil {
		panic(err)
	}

	defer sln.Close()
	// a map to store the connection objects
	// with the randomSubDomain as the key

	connections := make(map[string]net.Conn)

	// accept incoming connections concurrently

	for {
		conn, err := sln.Accept()

		if err != nil {
			panic(err)
		}

		go func(c net.Conn) {
			// defer conn.Close()

			// on first connection, send a welcome message
			// to the client

			randomSubDomain := GenerateRandomSubDomain()
			// store the connection object in a map
			// with the randomSubDomain as the key
			connections[randomSubDomain] = c

			// read if there is any header with host field

			c.Write([]byte(randomSubDomain + "\n"))

			buf := make([]byte, 1024)

			for {
				n, err := c.Read(buf)

				if err != nil {
					if err.Error() == "EOF" {
						fmt.Println("client disconnected")
						break
					} else {
						panic(err)
					}
				}

				requestString := string(buf[:n])
				reader := strings.NewReader(requestString)

				// parsing the string to a http request object
				req, err := http.ReadRequest(bufio.NewReader(reader))

				if err != nil {
					panic(err)
				}

				// get the host field from the header
				host := req.Host

				// get the connection object from the map
				// using the host field as the key
				conn, ok := connections[host]

				if !ok {
					fmt.Println("connection not found")
					break
				}

				// write the request to the connection object
				// which is the client connection

				// convert request string to byte array
				// and write to the connection object

				_, err = conn.Write([]byte(requestString))

				if err != nil {
					panic(err)
				}

			}

		}(conn)
	}

}

func main() {

	usedAs := flag.String("usedAs", "", "used as server or client")

	flag.Parse()

	fmt.Println("used as --> ", *usedAs)

	if *usedAs == "server" {
		fmt.Println("used as server")
		ServerSide()
	} else {
		client.RunAsClient("3000")
	}

}
