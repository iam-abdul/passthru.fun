package client

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
)

func RunAsClient(port string, domain string) {
	// addr, err := net.ResolveTCPAddr("tcp", "localhost:8888")
	// 	log.Fatal(err)
	// }

	// conn, err := net.DialTCP("tcp", nil, addr)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	conf := &tls.Config{
		InsecureSkipVerify: true, // Set this to false in production!
	}

	conn, err := tls.Dial("tcp", "app.passthru.fun:443", conf)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	// will request the domain here
	_, err = conn.Write([]byte("domain " + domain))
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Println("connection ended ", err)
		} else {
			log.Fatal(err)
		}
	}

	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Println("Connection rejected ", err)
		} else {
			log.Fatal(err)
		}
	}

	// log.Println("Response: ", string(response))
	// log.Println("Byte representation of response: ", []byte(string(response)))
	if string(response[:n]) == "false" {
		log.Println("domain not available")
		return
	}

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("Server disconnected 2", err)
				break
			} else {
				log.Fatal(err)
			}
		}
		log.Println(string(buf[:n]))

		// create a tcp connection to the localhost 3000 server
		// and send forward the request
		localAddr, err := net.ResolveTCPAddr("tcp", "localhost:"+port)
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
