package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// Listen on port 4221
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server listening on port 4221")

	for {
		// Accept connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Buffer to read the request
	buffer := make([]byte, 1024)
	
	// Read the incoming request (we're ignoring it for now as specified)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	// Respond with HTTP/1.1 200 OK with proper CRLF line endings
	response := "HTTP/1.1 200 OK\r\n\r\n"
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err.Error())
	}
}