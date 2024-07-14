package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Request struct to hold HTTP request details
type Request struct {
	httpMethod    string
	httpPath      string
	httpVersion   string
	headers       map[string]string
	status        string
	contentType   string
	contentLength int
}

// extractHeaders extracts headers from the HTTP request
func extractHeaders(httpVerb string, requestRows []string) map[string]string {
	headers := make(map[string]string)
	var lastHeader int
	if httpVerb == "GET" {
		lastHeader = len(requestRows)
	} else {
		lastHeader = len(requestRows) - 1
	}
	for _, str := range requestRows[1:lastHeader] {
		parts := strings.SplitN(str, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}
	return headers
}

// handleConnection handles the incoming connections
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read data from the connection
	var buff = make([]byte, 1024)
	_, err := conn.Read(buff)
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		return
	}

	// Parse the HTTP request
	lines := strings.Split(string(buff), "\n")
	httpSign := lines[0]
	fields := strings.Fields(httpSign)
	reqBody := &Request{
		httpMethod:  fields[0],
		httpPath:    fields[1],
		httpVersion: fields[2],
		headers:     extractHeaders(fields[0], lines),
	}
	fmt.Printf("Request Body: %v\n", reqBody)

	var responseBody string

	// Handle different paths
	switch {
	case strings.HasPrefix(reqBody.httpPath, "/files"):
		handleFileRequest(reqBody, &responseBody, lines)
	case strings.HasPrefix(reqBody.httpPath, "/user-agent"):
		reqBody.status = "HTTP/1.1 200 OK"
		reqBody.contentType = "text/plain"
		responseBody = reqBody.headers["User-Agent"]
	case strings.HasPrefix(reqBody.httpPath, "/echo/"):
		reqBody.status = "HTTP/1.1 200 OK"
		reqBody.contentType = "text/plain"
		responseBody = strings.TrimPrefix(reqBody.httpPath, "/echo/")
	case reqBody.httpPath == "/":
		reqBody.status = "HTTP/1.1 200 OK"
		reqBody.contentType = "text/plain"
		responseBody = ""
	default:
		reqBody.status = "HTTP/1.1 404 Not Found"
		reqBody.contentType = "text/plain"
		responseBody = "404 Not Found"
	}

	// Build and send the response
	buildAndSendResponse(conn, reqBody, responseBody)
}

// handleFileRequest handles file GET and POST requests
func handleFileRequest(reqBody *Request, responseBody *string, lines []string) {
	parts := strings.Split(reqBody.httpPath, "/")
	dir := os.Args[2]
	file := parts[2]

	if reqBody.httpMethod == "GET" {
		data, err := os.ReadFile(dir + file)
		if err != nil {
			fmt.Println("File not found:", err)
			reqBody.status = "HTTP/1.1 404 Not Found"
			reqBody.contentType = ""
			*responseBody = ""
		} else {
			fmt.Println("File exists.")
			reqBody.status = "HTTP/1.1 200 OK"
			reqBody.contentType = "application/octet-stream"
			*responseBody = string(data)
			reqBody.contentLength = len(data)
		}
	} else if reqBody.httpMethod == "POST" {
		content := strings.Trim(lines[len(lines)-1], "\x00")
		if err := os.WriteFile(dir+file, []byte(content), 0644); err == nil {
			fmt.Println("Wrote file")
			reqBody.status = "HTTP/1.1 201 Created"
		} else {
			reqBody.status = "HTTP/1.1 404 Not Found"
		}
	}
}

// buildAndSendResponse builds the HTTP response and sends it to the client
func buildAndSendResponse(conn net.Conn, reqBody *Request, responseBody string) {
	response := fmt.Sprintf("%s\r\n", reqBody.status)
	if reqBody.contentType != "" {
		response += fmt.Sprintf("Content-Type: %s\r\n", reqBody.contentType)
	}
	if reqBody.contentLength != 0 {
		response += fmt.Sprintf("Content-Length: %d\r\n", reqBody.contentLength)
	}
	response += "\r\n" + responseBody
	fmt.Printf("Response: %v\n", response)
	conn.Write([]byte(response))
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	// Start listening on the specified port
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	// Accept and handle incoming connections
	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(connection)
	}
}
