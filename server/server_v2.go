package server

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

type clientConnection struct {
	conn *net.TCPConn
}

type Server struct {
	connectionsLock sync.RWMutex
	connections     map[string]clientConnection
}

func isValidSubdomain(subdomain string) bool {
	if len(subdomain) > 255 {
		return false
	}

	labels := strings.Split(subdomain, ".")
	for _, label := range labels {
		if len(label) > 63 || len(label) < 1 {
			return false
		}

		if label == "app" {
			return false
		}
	}

	return true
}

func (s *Server) start_v2(port string) {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+port)

	if err != nil {
		// handle error
		log.Fatal("Error resolving TCP address: ", err)
	}

	ln, err := net.ListenTCP("tcp", addr)
	fmt.Println()
	if err != nil {
		// handle error
		log.Fatal("Error listening on TCP: ", err)
	}

	for {
		conn, err := ln.AcceptTCP()

		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Err accepting connection from client: ", err)
				continue
			}
		}
		// s.connections[conn.RemoteAddr().String()] = conn
		// fmt.Println("New connection from: ", conn.RemoteAddr().String())

		// we will do the domain assignment part here
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			// handle error
			fmt.Println("Error reading from client: ", err)
		}

		// now check if the first word is domain
		if strings.HasPrefix(string(buf[:n]), "domain") {
			parts := strings.SplitN(string(buf[:n]), "domain", 2)
			if len(parts) > 1 {
				contentFollowingDomain := parts[1]
				contentFollowingDomain = strings.TrimSpace(contentFollowingDomain) + ".passthru.fun"

				if !isValidSubdomain(contentFollowingDomain) {
					_, err := conn.Write([]byte("Invalid subdomain" + " " + contentFollowingDomain))
					if err != nil {
						fmt.Println("Error writing to client: ", err)
					}
					conn.Close()
					continue
				} else {
					// check if the domain is already assigned
					s.connectionsLock.RLock()
					_, ok := s.connections[contentFollowingDomain]
					s.connectionsLock.RUnlock()
					if ok {
						_, err := conn.Write([]byte("false"))
						if err != nil {
							fmt.Println("Error writing to client: ", err)
						}
						conn.Close()
						continue
					} else {
						s.connectionsLock.Lock()
						s.connections[contentFollowingDomain] = clientConnection{
							conn: conn,
						}
						s.connectionsLock.Unlock()
						_, err := conn.Write([]byte("true"))
						if err != nil {
							fmt.Println("Error writing to client: ", err)
						}

						continue

					}
				}

			}

		} else {
			go func() {

				// its a http request then
				// extract the host and forward to the client
				_, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(buf[:n]))))
				if err != nil {
					fmt.Println("Error reading the buffer into http request: ", err)
				}

				host := "test.passthru.fun"
				s.connectionsLock.RLock()
				clientConn, ok := s.connections[host]
				s.connectionsLock.RUnlock()
				if ok {
					_, err := clientConn.conn.Write(buf[:n])
					if err != nil {
						fmt.Println("Error writing to client: ", err)
					}

					fmt.Println("before copy code")

					defer conn.Close()
					// after writing the request we will stream the response back to the client
					// num, err := io.Copy(conn, clientConn.conn)
					// if err != nil {
					// 	fmt.Println("Error copying response to client: ", err)
					// }
					// fmt.Println("Copied: ", num)

					reader := bufio.NewReader(clientConn.conn)
					resp, err := http.ReadResponse(reader, nil)
					if err != nil {
						fmt.Println("Error reading response from client: ", err)
					}

					// fmt.Println("Status: ", resp.Status)
					// fmt.Println("Proto: ", resp.Proto)
					// fmt.Println("Header: ", resp.Header)
					// fmt.Println("Body: ", resp.Body)

					defer resp.Body.Close()
					statusLine := fmt.Sprintf("%s %s\r\n", resp.Proto, resp.Status)

					// Format the headers
					headers := new(bytes.Buffer)
					err = resp.Header.Write(headers)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Println("Response headers ", headers.String())

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

					// Instead of copy we will use CopyN, the N will the number of
					// bytes to be read from the response body we will get this N from content-length
					// header
					// stream the body
					bo, err := io.Copy(conn, resp.Body)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Println("Wrote back to client: ", bo)

				} else {
					fmt.Println("No client connection found for host: ", host)
					defer conn.Close()
				}

			}()
		}
	}
}

func StartNewServerV2(port string) {
	server := &Server{
		connections: make(map[string]clientConnection),
	}
	server.start_v2(port)
}
