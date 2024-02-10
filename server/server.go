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
)

type clientConnection struct {
	conn     *net.TCPConn
	response chan []byte
}

type Server struct {
	connections map[string]clientConnection
}

func handleClientResponse(conn *net.TCPConn, response chan []byte) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Client disconnected ", err)
				break
			}
		}

		response <- buf[:n]
	}
}

func (s *Server) handleConnection(conn *net.TCPConn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Client disconnected ", err)
				delete(s.connections, conn.RemoteAddr().String())
				break
			}
		}

		if strings.HasPrefix(string(buf[:n]), "domain") {
			parts := strings.SplitN(string(buf[:n]), "domain", 2)
			if len(parts) > 1 {
				contentFollowingDomain := parts[1]

				contentFollowingDomain = strings.TrimSpace(contentFollowingDomain)
				// check if the domain is already in the map
				exists := s.connections[contentFollowingDomain]
				if exists.conn != nil {
					// we will forward the message to the other client
					_, err := conn.Write([]byte("false"))
					if err != nil {
						fmt.Println("Error writing to client: ", err)
					}
				} else {
					s.connections[contentFollowingDomain] = clientConnection{
						conn:     conn,
						response: make(chan []byte),
					}

					// since this is a client connection and it will be in connected state all
					// the time, we can start a goroutine to listen for the response from the client
					// and pipe it to the response channel

					go handleClientResponse(conn, s.connections[contentFollowingDomain].response)
					break

				}

				// trim the content
				fmt.Println("Content following domain:", contentFollowingDomain)
				// Now contentFollowingDomain contains the contents following the word "domain"
			}
		} else {
			// it is a http request that needs to be sent to proper client downstream
			// converting the buffer to http request

			defer conn.Close()
			request, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(buf[:n]))))
			if err != nil {
				fmt.Println("Error reading request: ", err)
				// write back some message
				_, err := conn.Write([]byte("Error reading request"))
				if err != nil {
					fmt.Println("Error writing to client: ", err)
				}
				// TODO should disconnect the client
				break
			}

			// find the host from the request
			host := request.Host
			fmt.Println("Host: ", host)

			// writing the request to the proper client
			clientConn := s.connections[host].conn
			if clientConn != nil {
				_, err := clientConn.Write(buf[:n])
				if err != nil {
					fmt.Println("Error writing to client: ", err)
				}

				// read the response from the client and write it back
				response := <-s.connections[host].response
				_, err = conn.Write(response)
				if err != nil {
					fmt.Println("Error writing to client: ", err)
				}
			}
		}

		fmt.Println("Received from client: ", string(buf[:n]))
		fmt.Println("Number of connections: ", len(s.connections))

	}
}

func (s *Server) start() {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:8888")
	if err != nil {
		// handle error
		log.Fatal("Error resolving TCP address: ", err)
	}

	ln, err := net.ListenTCP("tcp", addr)
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
		fmt.Println("New connection from: ", conn.RemoteAddr().String())

		go s.handleConnection(conn)
	}
}

func StartNewServer(address string) {
	server := &Server{
		connections: make(map[string]clientConnection),
	}
	server.start()
}
