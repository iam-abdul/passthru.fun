package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

type Server struct {
	Addr        string
	Connections map[string]net.Conn
}

func (s *Server) handleConnection(conn net.Conn) {
	// we will look at the first few words of data
	// to determine if it is a request or a tcp connection

	for {
		buf := make([]byte, 2048)
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				// clear up the connection
				fmt.Println("connection closed")
				disconnectedDomain := ""
				for domain, connection := range s.Connections {
					if connection == conn {
						disconnectedDomain = domain
						break
					}
				}

				if disconnectedDomain != "" {
					fmt.Println("domain disconnected ", disconnectedDomain)
					delete(s.Connections, disconnectedDomain)
				}

				conn.Close()
				break
			} else {
				panic(err)
			}
		}

		// fmt.Println("data received ", string(buf[:n]))

		words := strings.Fields(string(buf[:n]))

		if words[0] == "domain" {
			fmt.Println("domain request received ", words[1])

			// check if the domain is available
			if s.Connections[words[1]] != nil {
				conn.Write([]byte("false"))
			} else {
				// store the connection
				s.Connections[words[1]] = conn
				conn.Write([]byte("true"))
			}

		} else {
			// this is a http request from the some server
			// find the host and forward the request to the host
			// fmt.Println("http request received ", string(buf[:n]))
			req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(buf[:n]))))

			if err != nil {
				if err.Error() == "EOF" {
					fmt.Println("EOF error")
				} else {
					panic(err)
				}
			}

			host := req.Host
			fmt.Println("outside server to client ", host)
			// fmt.Println("host ", host)

			targetConn := s.Connections[host]
			fmt.Println("target connection ", targetConn)

			if targetConn == nil {
				fmt.Println("target host not found")
				continue
			}

			// join the target connection with the client connection
			go func() {

				fmt.Println("copying data from server to client")
				copied, err := io.Copy(targetConn, conn)
				fmt.Println("copied data ", copied)
				if err != nil {
					fmt.Println("error in copying data ", err)
				}
			}()

			// go func() {
			// 	fmt.Println("copying data from client to server")
			// 	_, err := io.Copy(conn, targetConn)
			// 	if err != nil {
			// 		fmt.Println("error in copying data ", err)
			// 	}
			// }()

			if err != nil {
				fmt.Println("error in copying data ", err)
			}

		}
	}

}

func StartNewServer(addr string) {

	server := &Server{
		Addr:        addr,
		Connections: make(map[string]net.Conn),
	}

	// start a tcp server
	hostConn, err := net.Listen("tcp", server.Addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := hostConn.Accept()
		if err != nil {
			panic(err)
		}

		go server.handleConnection(conn)

	}

}
