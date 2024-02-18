package server

import (
	"bufio"
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
	conn     *net.TCPConn
	response chan []byte
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

		// if _, ok := net.LookupHost(label); ok != nil {
		// 	return false
		// }

		if label == "app" {
			return false
		}
	}

	return true
}

func handleClientResponse(conn *net.TCPConn, response chan []byte, thisSubdomain string, connections *map[string]clientConnection, connectionsLock sync.RWMutex) {
	defer conn.Close()
	buf := make([]byte, 10000024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {

				// remove the connection from the map
				// delete(s.connections, thisSubdomain)
				connectionsLock.Lock()
				delete(*connections, thisSubdomain)
				connectionsLock.Unlock()

				fmt.Println("removed closed client ", err)
				fmt.Println("Number of connections: ", len(*connections))
				break
			}
		}

		response <- buf[:n]
	}
}

func (s *Server) handleConnection(conn *net.TCPConn) {

	for {
		buf := make([]byte, 10000024)
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Client disconnected ", err)
				break
			} else {
				fmt.Println("Error reading from client: ", err)
			}
		}

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
					break
				}

				// check if the domain is already in the map
				exists := s.connections[contentFollowingDomain]
				fmt.Println("Exists: ", exists)
				if exists.conn != nil {
					// we will forward the message to the other client
					_, err := conn.Write([]byte("false"))
					if err != nil {
						fmt.Println("Error writing to client: ", err)
					}
					conn.Close()
					break
				} else {
					_, err := conn.Write([]byte("true"))
					if err != nil {
						fmt.Println("Error writing to client: ", err)
					}

					s.connectionsLock.Lock()
					s.connections[contentFollowingDomain] = clientConnection{
						conn:     conn,
						response: make(chan []byte),
					}
					s.connectionsLock.Unlock()

					// since this is a client connection and it will be in connected state all
					// the time, we can start a goroutine to listen for the response from the client
					// and pipe it to the response channel

					go handleClientResponse(conn, s.connections[contentFollowingDomain].response, contentFollowingDomain, &s.connections, s.connectionsLock)
					break

				}

			}
		} else {
			// fmt.Println("Length of buf: ", n)
			// it is a http request that needs to be sent to proper client downstream
			// converting the buffer to http request

			// defer conn.Close()
			fmt.Println("Request: from outside server ")
			_, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(buf[:n]))))
			if err != nil {
				fmt.Println("Error reading request: ", err)
				if errors.Is(err, io.EOF) {
					fmt.Println("Client disconnected ", err)
					for k, v := range s.connections {
						if v.conn == conn {
							delete(s.connections, k)
							break
						}
					}
					break
				}
				// write back some message
				_, err := conn.Write([]byte("Error reading request"))
				if err != nil {
					fmt.Println("Error writing to client: ", err)
				}
				break
			}

			// find the host from the request
			// host := request.Host
			host := "test.passthru.fun"
			// fmt.Println("Host: ", host)

			// writing the request to the proper client
			clientConn := s.connections[host].conn
			if clientConn != nil {
				_, err := clientConn.Write(buf[:n])
				if err != nil {
					fmt.Println("Error writing to client: ", err)
				}

				// read the response from the client and write it back
				response := <-s.connections[host].response

				// fmt.Println("Response from client hehe: ", string(response))
				nono, err := conn.Write(response)
				if err != nil {
					fmt.Println("Error writing to client: ", err)
				}
				fmt.Println("wrote back to client: ", nono)
				fmt.Println("Number of connections: ", len(s.connections))
				conn.Close()
				fmt.Println("Closed connection")
				break

			} else {
				// write back some message
				_, err := conn.Write([]byte("No client found for the host"))
				if err != nil {
					fmt.Println("Error writing to client when no client is found: ", err)
				}

				fmt.Println("No client found for the host")
				conn.Close()
				break
			}
		}

	}
}

func (s *Server) start(port string) {
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

		go s.handleConnection(conn)
	}
}

func StartNewServer(port string) {
	server := &Server{
		connections: make(map[string]clientConnection),
	}
	server.start(port)
}
