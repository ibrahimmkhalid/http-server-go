package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

const RESPONSE_200 string = "HTTP/1.1 200 OK\r\n\r\n"
const RESPONSE_404 string = "HTTP/1.1 404 Not Found\r\n\r\n"

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	var readArray = make([]byte, 1024) // to make preallocated array, use the "make" function
	_, err = conn.Read(readArray)
	if err != nil {
		fmt.Println("Failed to read request")
		os.Exit(1)
	}

	urlPath, err := parseURLPath(string(readArray))
	if err != nil {
		fmt.Println(err.Error())
	}

	if urlPath == "/" {
		conn.Write([]byte(RESPONSE_200))
	} else if strings.HasPrefix(urlPath, "/echo/") {
		conn.Write([]byte(handleEchoPath(urlPath)))
	} else if urlPath == "/user-agent" {
		conn.Write([]byte(handleUserAgentPath(string(readArray))))
	} else {
		conn.Write([]byte(RESPONSE_404))
	}

}

func parseURLPath(requestString string) (string, error) {
	if len(requestString) < 14 {
		return "", errors.New("Invalid request")
	}

	requestParts := strings.SplitAfter(requestString, " ")

	if len(requestParts) < 3 {
		return "", errors.New("Invalid request")
	}

	return strings.Trim(requestParts[1], " \t\n\r"), nil
}

func handleEchoPath(urlPath string) string {
	var data string = strings.TrimPrefix(urlPath, "/echo/")
	var size int = len(data)

	return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%s", size, data)
}

func handleUserAgentPath(requestString string) string {
	var stringReader = strings.NewReader(requestString)
	var bufReader = bufio.NewReader(stringReader)
	req, err := http.ReadRequest(bufReader)
	if err != nil {
		fmt.Println(err.Error())
	}
	var data string = req.UserAgent()
	var size int = len(data)
	return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%s", size, data)
}
