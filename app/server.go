package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

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

	var readarr = make([]byte, 1024) // to make preallocated array, use the "make" function
	_, err = conn.Read(readarr)
	if err != nil {
		fmt.Println("Failed to read request")
		os.Exit(1)
	}

	var resp200 string = "HTTP/1.1 200 OK\r\n\r\n"
	var resp404 string = "HTTP/1.1 404 Not Found\r\n\r\n"

	if strings.HasPrefix(string(readarr), "GET / HTTP/1.1\r\n") {
		conn.Write([]byte(resp200))
	} else {
		conn.Write([]byte(resp404))
	}

}
