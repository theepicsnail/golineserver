package main

import "flag"
import "time"
import "fmt"
import "net"
import "bufio"

var port = flag.Int("port", 1234, "port to listen on")
var delimiter = flag.String("delimiter", "\n", "delimiter to split messages on")
var timeout = flag.Int("timeout", 3000, "How long to wait on a client before killing it (Seconds)")

type Server struct {
	clients      []net.Conn
	listener     net.Listener
	delim        byte
	writeTimeout time.Duration
}

func NewServer() (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		return nil, err
	}

	return &Server{
		[]net.Conn{},
		ln,
		'\n',
		time.Duration(*timeout) * time.Second,
	}, nil
}

func (server *Server) Serve() {
	for {
		fmt.Println("Waiting for connection")
		conn, err := server.listener.Accept()
		fmt.Println(conn.RemoteAddr(), "Connected")

		if err != nil {
			fmt.Println(err)
			continue
		}

		server.addClient(conn)
	}
}

func (server *Server) addClient(client net.Conn) {
	go func() {
		defer server.removeClient(client)

		fmt.Println("ADD", client.RemoteAddr())
		server.clients = append(server.clients, client)

		reader := bufio.NewReader(client)
		for {
			line, err := reader.ReadString(server.delim)
			if err != nil {
				fmt.Println("ERR", client.RemoteAddr(), err)
				return
			}
			server.broadcast(line)
		}
	}()
}

func (server *Server) removeClient(client net.Conn) {
	fmt.Println("REM", client.RemoteAddr())

	numClients := len(server.clients)

	for p, v := range server.clients {
		if v == client {
			fmt.Println("Removed")
			server.clients[p] = server.clients[numClients-1]
			server.clients[numClients-1] = nil
			server.clients = server.clients[:numClients-1]
			return
		}
	}
	fmt.Println("Couldn't find client?")
}

func (server *Server) broadcast(data string) {
	dataLen := len(data)
	fmt.Println("Broadcasting", data)
	for _, client := range server.clients {
		fmt.Println("...to", client)
		client.SetWriteDeadline(time.Now().Add(server.writeTimeout))
		written, err := client.Write([]byte(data))
		if err != nil || dataLen != written {
			defer server.removeClient(client)
		}
	}
	fmt.Println("Broadcast finished")
}

func main() {
	flag.Parse()
	fmt.Printf("Port     =%v\n", *port)
	fmt.Printf("Delimiter=%#v\n", *delimiter)
	fmt.Printf("Timeout  =%v\n", *timeout)

	server, err := NewServer()
	if err != nil {
		panic(err)
	}

	server.Serve()
}
