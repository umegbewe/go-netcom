package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Connection closed")
			return
		}
		fmt.Printf("Received %d bytes from client: %s\n", n, string(buf[:n]))
		_, err = conn.Write(buf[:n])
		if err != nil {
			fmt.Println("Error writing to client:", err)
			return
		}
	}
}

func main() {
	fmt.Println("Starting server...")
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("New client connected:", conn.RemoteAddr())
		go handleConnection(conn)
	}
}
