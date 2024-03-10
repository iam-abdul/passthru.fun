// we will build the client to handle concurrent requests

package client

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

func giveResponse(conn *net.TCPConn, port string, buffer []byte) {

	// converting the buffer to a http request
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buffer)))
	if err != nil {
		log.Println("Error reading request: ", err)
		if errors.Is(err, io.EOF) {
			log.Println("Server disconnected ", err)
			return
		}
	}

	// changing the URL to the local server
	req.URL.Scheme = "http"
	req.URL.Host = "localhost:" + port

	// clearing the requestURI field
	req.RequestURI = ""

	fmt.Println("Request: ", req.URL.String())

	// we will do this:
	// since this runs on client we will load the whole response body to
	// a buffer and then write it to the client along with the content length
	// header if we don't get the content length directly from the client server

	// writing the request to the local server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making request to local server: ", err)
	}
	defer resp.Body.Close()

	// ###############################################################
	statusLine := fmt.Sprintf("%s %s\r\n", resp.Proto, resp.Status)

	// Format the headers
	headers := new(bytes.Buffer)
	err = resp.Header.Write(headers)
	if err != nil {
		log.Fatal(err)
	}

	// buffer to store the response
	var responseBuffer bytes.Buffer
	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		responseBuffer.Write(body)
		// add the content length header
		headers.WriteString("Content-Length: " + fmt.Sprint(len(body)) + "\r\n")
	}

	// fmt.Println("Response headers ", headers.String())

	// Write the status line
	_, err = io.WriteString(conn, statusLine)
	if err != nil {
		log.Fatal(err)
	}

	// Write the headers
	_, err = io.WriteString(conn, headers.String()+"\r\n")
	if err != nil {
		log.Fatal(err)
	}

	// stream the body
	if contentLength == "" {
		copied, err := io.Copy(conn, &responseBuffer)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Wrote back to client: ", copied)
	} else {
		copied, err := io.CopyN(conn, resp.Body, resp.ContentLength)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Wrote back to client: ", copied)
	}
	// bo, err := io.Copy(conn, resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// ###############################################################
	// // Write the status line
	// statusLine := fmt.Sprintf("%s %s\r\n", resp.Proto, resp.Status)

	// // Format the headers
	// headers := new(bytes.Buffer)
	// err = resp.Header.Write(headers)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("Response headers ", headers.String())

	// // Read the body
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Combine the status line, headers, and body into one string
	// response := statusLine + headers.String() + "\r\n" + string(body)

	// fmt.Println("Response: ", response)
	// // Write the response to the client
	// _, err = io.WriteString(conn, response)
	// if err != nil {
	// 	log.Fatal(err)
	// }

}

func RunAsClientV2(port string, domain string, verbose bool) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:8888")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	// defer conn.Close()

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

	if string(response[:n]) == "false" {
		log.Println("domain not available")
		return
	} else if string(response[:n]) == "true" {
		log.Println("\033[1m\033[31mlocalhost:" + port + "\033[0m <===> \033[1m\033[32m" + domain + ".passthru.fun\033[0m\033[0m")
	}

	for {

		buf := make([]byte, 4084)
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("Server disconnected 2", err)
				break
			} else {
				log.Fatal(err)
			}
		}
		if verbose {
			log.Println(string(buf[:n]))
		}

		go giveResponse(conn, port, buf[:n])

	}

}
