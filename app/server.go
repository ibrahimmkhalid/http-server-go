package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

const RESPONSE_200 string = "HTTP/1.1 200 OK\r\n\r\n"
const RESPONSE_404 string = "HTTP/1.1 404 Not Found\r\n\r\n"

var filesPath string

func main() {

	flag.StringVar(&filesPath, "directory", "", "Path to /files/ serve endpoint")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go router(conn)
	}
}

func router(conn net.Conn) {
	defer conn.Close()
	var readArray = make([]byte, 1024) // to make preallocated array, use the "make" function
	_, err := conn.Read(readArray)
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
	} else if strings.HasPrefix(urlPath, "/files/") {
		conn.Write([]byte(handleFileServeEndpoint(urlPath)))
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

func handleFileServeEndpoint(urlPath string) string {
	if filesPath == "" {
		fmt.Println("Root for /files/ endpoint not defined")
		os.Exit(1)
	}

	var filename string = strings.TrimPrefix(urlPath, "/files/")
	var filePath string = filesPath + filename
	fd, err := os.Open(filePath)
	if err != nil {
		return RESPONSE_404
	}
	defer fd.Close()

	var chunk []byte = make([]byte, 1024)

	var n int
	var size int
	var data string
	for {
		n, err = fd.Read(chunk)
		if err != nil {
			break
		}
		size += n
		data += string(chunk[:n])
	}

	return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %v\r\n\r\n%s", size, data)
}
