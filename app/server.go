package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

const RESPONSE_200 string = "HTTP/1.1 200 OK\r\n\r\n"
const RESPONSE_404 string = "HTTP/1.1 404 Not Found\r\n\r\n"

func main() {
	fmt.Println("Logs from your program will appear here!")

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

	/* // task 4, extract url path
	if urlPath == "/" {
		conn.Write([]byte(RESPONSE_200))
	} else {
		conn.Write([]byte(RESPONSE_404))
	} */

	if strings.HasPrefix(urlPath, "/echo/") {
		response, err := handleEchoPath(urlPath)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		conn.Write([]byte(response))
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

func handleEchoPath(urlPath string) (string, error) {
	var data string = strings.TrimPrefix(urlPath, "/echo/")
	var size int = len(data)

	var buffer bytes.Buffer
	buffer.WriteString(RESPONSE_200)
	buffer.WriteString(fmt.Sprintf("Content-Type: text/plain\r\nContent-Length: %v\r\n\r\n%s", size, data))
	return buffer.String(), nil
}
