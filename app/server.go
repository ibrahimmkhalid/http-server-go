package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
)

const RESPONSE_200 string = "HTTP/1.1 200 OK\r\n\r\n"
const RESPONSE_201 string = "HTTP/1.1 201 Created\r\n\r\n"
const RESPONSE_404 string = "HTTP/1.1 404 Not Found\r\n\r\n"
const RESPONSE_500 string = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
const ENCODINGS string = "gzip"

var filesPath string

func main() {
	fmt.Println("Server started...")

	flag.StringVar(&filesPath, "directory", "", "Path to /files/ serve endpoint")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()

	tmp, _ := os.Create("run")
	tmp.WriteString(fmt.Sprintf("%d", rand.Intn(100)))
	tmp.Sync()
	tmp.Close()

	fmt.Println("Connection accepted")
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

	var stringReader = strings.NewReader(string(readArray))
	var bufReader = bufio.NewReader(stringReader)
	req, err := http.ReadRequest(bufReader)
	if err != nil {
		fmt.Println(err.Error())
	}

	if req.URL.Path == "/" {
		conn.Write([]byte(RESPONSE_200))
	} else if strings.HasPrefix(req.URL.Path, "/echo/") {
		conn.Write([]byte(handleEchoPath(*req)))
	} else if req.URL.Path == "/user-agent" {
		conn.Write([]byte(handleUserAgentPath(*req)))
	} else if strings.HasPrefix(req.URL.Path, "/files/") {
		conn.Write([]byte(handleFilesEndpoint(*req)))
	} else {
		conn.Write([]byte(RESPONSE_404))
	}

}

func handleEchoPath(request http.Request) string {
	var data string = strings.TrimPrefix(request.URL.Path, "/echo/")
	var size int = len(data)

	encodingsString := request.Header.Get("Accept-Encoding")
	if encodingsString != "" {
		encodings := strings.Split(encodingsString, ",")
		for _, v := range encodings {
			encoding := strings.Trim(v, " ")
			if strings.Contains(ENCODINGS, encoding) {
				compressedData, err := compressData([]byte(data), encoding)
				if err != nil {
					fmt.Println(err.Error())
					return RESPONSE_500
				}
				size = len(compressedData)
				return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: %s\r\nContent-Length: %v\r\n\r\n%s", encoding, size, compressedData)
			}
		}
	}

	return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%s", size, data)
}

func handleUserAgentPath(request http.Request) string {
	var data string = request.UserAgent()
	var size int = len(data)
	return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%s", size, data)
}

func handleFilesEndpoint(request http.Request) string {
	if filesPath == "" {
		fmt.Println("Root for /files/ endpoint not defined")
		os.Exit(1)
	}

	if request.Method == "POST" {
		return handleFileUploadEndpoint(request)
	} else if request.Method == "GET" {
		return handleFileServeEndpoint(request.URL.Path)
	} else {
		return RESPONSE_404
	}
}

func handleFileUploadEndpoint(request http.Request) string {
	body := request.Body
	var chunk []byte = make([]byte, 1024)

	var size int
	var data string
	for {
		n, err := body.Read(chunk)
		if err != nil && !strings.Contains(err.Error(), "EOF") {
			break
		}
		size += n
		data += string(chunk[:n])
		if err != nil {
			break
		}
	}

	var filename string = strings.TrimPrefix(request.URL.Path, "/files/")
	var filePath string = filesPath + filename
	fd, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err.Error())
		return RESPONSE_500
	}
	defer fd.Close()

	_, err = fd.Write([]byte(data))
	if err != nil {
		fmt.Println(err.Error())
		return RESPONSE_500
	}

	fd.Sync()
	return RESPONSE_201
}

func handleFileServeEndpoint(urlPath string) string {
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

func compressData(data []byte, algorithm string) ([]byte, error) {
	if algorithm == "gzip" {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		if _, err := gz.Write(data); err != nil {
			return data, err
		}
		if err := gz.Close(); err != nil {
			return data, err
		}
		fmt.Println(b.Bytes())
		return b.Bytes(), nil
	}
	return data, nil
}
