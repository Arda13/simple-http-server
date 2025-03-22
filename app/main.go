package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
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

	// Read and discard headers - not needed for this implementation
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading headers:", err.Error())
			return
		}
		// Headers end with an empty line (just \r\n)
		if line == "\r\n" {
			break
		}
	}

	// Determine response based on path
	if path == "/" {
		// Create an empty map for headers
		emptyHeaders := make(map[string]string)
		sendResponse(conn, "200 OK", emptyHeaders, "")
	} else if strings.HasPrefix(path, "/echo/") {
		// Extract the string after "/echo/"
		echoStr := path[len("/echo/"):]
		
		// Send response with Content-Type and Content-Length headers
		headers := make(map[string]string)
		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = strconv.Itoa(len(echoStr))
		
		sendResponse(conn, "200 OK", headers, echoStr)
	} else {
		// Create an empty map for headers
		emptyHeaders := make(map[string]string)
		sendResponse(conn, "404 Not Found", emptyHeaders, "")
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

// Send an HTTP response with optional headers and body
func sendResponse(conn net.Conn, status string, headers map[string]string, body string) {
	// Build the response
	response := fmt.Sprintf("HTTP/1.1 %s\r\n", status)
	
	// Add headers
	for key, value := range headers {
		response += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	
	// Add empty line to separate headers from body
	response += "\r\n"
	
	// Add body if provided
	response += body
	
	// Send response
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err.Error())
	}
}