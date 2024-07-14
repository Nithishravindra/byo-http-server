package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type request struct {
	httpMethod    string
	httpPath      string
	httpVersion   string
	headers       map[string]string
	status        string
	contentType   string
	contentLength int
}

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

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var buff = make([]byte, 1024)
	_, err := conn.Read(buff)
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		return
	}
	lines := strings.Split(string(buff), "\n")
	httpSign := lines[0]

	fields := strings.Fields(httpSign)
	reqBody := &request{
		httpMethod:  fields[0],
		httpPath:    fields[1],
		httpVersion: fields[2],
		headers:     extractHeaders(fields[0], lines),
	}
	fmt.Printf("reqBody: %v\n", reqBody)
	fmt.Printf("httpMethods: %v\n", reqBody.httpMethod)
	fmt.Printf("httpVersion: %v\n", reqBody.httpVersion)
	fmt.Printf("httpPath: %v\n", reqBody.httpPath)

	var responseBody string
	if strings.HasPrefix(reqBody.httpPath, "/files") {
		parts := strings.Split(reqBody.httpPath, "/")
		dir := os.Args[2]
		file := parts[2]
		data, err := os.ReadFile(dir + file)

		if err != nil {
			fmt.Println("File not found:", err)
			reqBody.status = "HTTP/1.1 404 Not Found"
			reqBody.contentType = ""
			responseBody = ""
		} else {
			fmt.Println("File exists.")
			reqBody.status = "HTTP/1.1 200 OK"
			reqBody.contentType = "application/octet-stream"
			responseBody = string(data)
			reqBody.contentLength = len(data)
		}
	} else if strings.HasPrefix(reqBody.httpPath, "/user-agent") {
		responseBody = reqBody.headers["User-Agent"]
		reqBody.status = "HTTP/1.1 200 OK"
		reqBody.contentType = "text/plain"
		reqBody.contentLength = len(responseBody)
	} else if strings.HasPrefix(reqBody.httpPath, "/echo/") {
		responseBody = strings.TrimPrefix(reqBody.httpPath, "/echo/")
		reqBody.status = "HTTP/1.1 200 OK"
		reqBody.contentType = "text/plain"
		reqBody.contentLength = len(responseBody)
	} else if reqBody.httpPath == "/" {
		responseBody = ""
		reqBody.status = "HTTP/1.1 200 OK"
		reqBody.contentType = "text/plain"
		reqBody.contentLength = len(responseBody)
	} else {
		reqBody.status = "HTTP/1.1 404 Not Found"
		responseBody = "404 Not Found"
		reqBody.contentType = "text/plain"
		reqBody.contentLength = len(responseBody)
	}

	response := fmt.Sprintf("%s\r\n", reqBody.status)
	if reqBody.contentType != "" {
		response += fmt.Sprintf("Content-Type: %s\r\n", reqBody.contentType)
	}
	if reqBody.contentLength != 0 {
		response += fmt.Sprintf("Content-Length: %d\r\n", len(responseBody))
	}
	response += "\r\n" + responseBody
	fmt.Printf("response: %v\n", response)
	conn.Write([]byte(response))
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(connection)
	}
}
