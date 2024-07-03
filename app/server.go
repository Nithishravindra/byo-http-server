package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

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
	httpMethod, httpPath, httpVersion := fields[0], fields[1], fields[2]
	fmt.Printf("httpMethod: %v\n", httpMethod)
	fmt.Printf("httpVersion: %v\n", httpVersion)

	if httpPath == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	connection, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	handleConnection(connection)
}
