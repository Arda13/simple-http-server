package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	// Parse command line flags
	directoryFlag := flag.String("directory", "", "Directory to serve files from")
	flag.Parse()

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
		go handleConnection(conn, *directoryFlag)
	}
}

func handleConnection(conn net.Conn, directory string) {
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

	// Read headers
	headers := make(map[string]string)
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
		
		// Parse header and add to map
		parseHeader(line, headers)
	}

	// Determine response based on path
	if path == "/" {
		emptyHeaders := make(map[string]string)
		sendResponse(conn, "200 OK", emptyHeaders, "")
	} else if strings.HasPrefix(path, "/echo/") {
		// Extract the string after "/echo/"
		echoStr := path[len("/echo/"):]
		
		// Send response with Content-Type and Content-Length headers
		respHeaders := make(map[string]string)
		respHeaders["Content-Type"] = "text/plain"
		respHeaders["Content-Length"] = strconv.Itoa(len(echoStr))
		
		sendResponse(conn, "200 OK", respHeaders, echoStr)
	} else if path == "/user-agent" {
		// Get the User-Agent header
		userAgent := headers["user-agent"]
		
		// Send response with Content-Type and Content-Length headers
		respHeaders := make(map[string]string)
		respHeaders["Content-Type"] = "text/plain"
		respHeaders["Content-Length"] = strconv.Itoa(len(userAgent))
		
		sendResponse(conn, "200 OK", respHeaders, userAgent)
	} else if strings.HasPrefix(path, "/files/") {
		// Handle file request
		if directory == "" {
			// No directory specified
			emptyHeaders := make(map[string]string)
			sendResponse(conn, "404 Not Found", emptyHeaders, "")
			return
		}

		// Extract the filename from the path
		filename := path[len("/files/"):]
		
		// Build the full file path
		filePath := filepath.Join(directory, filename)
		
		// Try to open the file
		file, err := os.Open(filePath)
		if err != nil {
			// File doesn't exist or can't be opened
			emptyHeaders := make(map[string]string)
			sendResponse(conn, "404 Not Found", emptyHeaders, "")
			return
		}
		defer file.Close()
		
		// Get file info to determine size
		fileInfo, err := file.Stat()
		if err != nil {
			emptyHeaders := make(map[string]string)
			sendResponse(conn, "500 Internal Server Error", emptyHeaders, "")
			return
		}
		
		// Read file contents
		fileContents := make([]byte, fileInfo.Size())
		_, err = io.ReadFull(file, fileContents)
		if err != nil {
			emptyHeaders := make(map[string]string)
			sendResponse(conn, "500 Internal Server Error", emptyHeaders, "")
			return
		}
		
		// Send response with file contents
		respHeaders := make(map[string]string)
		respHeaders["Content-Type"] = "application/octet-stream"
		respHeaders["Content-Length"] = strconv.FormatInt(fileInfo.Size(), 10)
		
		sendResponse(conn, "200 OK", respHeaders, string(fileContents))
	} else {
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

// Parse a header line and add it to the headers map
func parseHeader(line string, headers map[string]string) {
	line = strings.TrimSpace(line)
	
	// Find the colon that separates header name from value
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return // Not a valid header
	}
	
	// Extract name and value
	name := strings.ToLower(strings.TrimSpace(line[:colonIdx]))
	value := strings.TrimSpace(line[colonIdx+1:])
	
	// Add to headers map
	headers[name] = value
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