package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
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

	// Create a buffered reader for the connection
	reader := bufio.NewReader(conn)
	
	// Read the request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request line:", err.Error())
		return
	}

	// Parse the request line to extract the path
	path := parsePath(requestLine)

	// Determine response based on path
	var response string
	if path == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	// Send response
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err.Error())
	}
}

// Parse the request line to extract the path
func parsePath(requestLine string) string {
	// Split the request line by spaces
	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	
	// Check if we have enough parts for a valid request line
	if len(parts) < 2 {
		return ""
	}
	
	// The path is the second part of the request line
	return parts[1]
}