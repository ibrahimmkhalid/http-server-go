package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

const ENCODINGS string = "gzip"

var filesPath string
var portNumber int

func main() {
	flag.StringVar(&filesPath, "directory", "", "Path to /files/ serve endpoint")
	flag.Parse()

	flag.IntVar(&portNumber, "port", 4221, "Port to bind to")

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", portNumber))
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to bind to port %v", portNumber))
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

	var stringReader = strings.NewReader(string(readArray))
	var bufReader = bufio.NewReader(stringReader)
	req, err := http.ReadRequest(bufReader)
	if err != nil {
		fmt.Println(err.Error())
	}

	if req.URL.Path == "/" {
		var res http.Response
		res.StatusCode = 200
		returnResponse(*req, res, conn)
	} else if strings.HasPrefix(req.URL.Path, "/echo/") {
		handleEchoPath(*req, conn)
	} else if req.URL.Path == "/user-agent" {
		handleUserAgentPath(*req, conn)
	} else if strings.HasPrefix(req.URL.Path, "/files/") {
		handleFilesEndpoint(*req, conn)
	} else {
		return404(conn)
	}

}

func returnResponse(req http.Request, res http.Response, conn net.Conn) {
	encodingsString := req.Header.Get("Accept-Encoding")
	if encodingsString != "" {
		var err error
		res, err = handleResponseEncoding(res, encodingsString)
		if err != nil {
			fmt.Println(err.Error())
			return500(conn)
		}
	}

	res.ProtoMajor = 1
	res.ProtoMinor = 1
	var buf bytes.Buffer
	res.Write(&buf)
	conn.Write(buf.Bytes())
}

func handleResponseEncoding(res http.Response, encodingsString string) (http.Response, error) {
	encodings := strings.Split(encodingsString, ",")
	for _, v := range encodings {
		encoding := strings.Trim(v, " ")
		if strings.Contains(ENCODINGS, encoding) {
			var data string
			var chunk []byte = make([]byte, 1024)
			for {
				n, err := res.Body.Read(chunk)
				if err != nil {
					break
				}
				data += string(chunk[:n])
			}
			compressedData, err := compressData([]byte(data), encoding)
			if err != nil {
				return res, err
			}
			res.Body = io.NopCloser(bytes.NewBuffer(compressedData))
			res.ContentLength = int64(len(compressedData))
			res.Header.Add("content-encoding", encoding)
		}
	}
	return res, nil
}

func handleEchoPath(request http.Request, conn net.Conn) {
	var data string = strings.TrimPrefix(request.URL.Path, "/echo/")
	var size int = len(data)

	var res http.Response
	res.StatusCode = 200
	res.Body = io.NopCloser(bytes.NewBufferString(data))
	res.ContentLength = int64(size)
	res.Header = make(http.Header)
	res.Header.Add("content-type", "text/plain")

	returnResponse(request, res, conn)
}

func handleUserAgentPath(request http.Request, conn net.Conn) {
	var data string = request.UserAgent()
	var size int = len(data)

	var res http.Response
	res.StatusCode = 200
	res.Body = io.NopCloser(bytes.NewBufferString(data))
	res.ContentLength = int64(size)
	res.Header = make(http.Header)
	res.Header.Add("content-type", "text/plain")

	returnResponse(request, res, conn)
}

func handleFilesEndpoint(request http.Request, conn net.Conn) {
	if filesPath == "" {
		fmt.Println("Root for /files/ endpoint not defined")
		return500(conn)
		os.Exit(1)
	}

	if request.Method == "POST" {
		handleFileUploadEndpoint(request, conn)
	} else if request.Method == "GET" {
		handleFileServeEndpoint(request, conn)
	} else {
		return404(conn)
	}
}

func handleFileUploadEndpoint(request http.Request, conn net.Conn) {
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
		return500(conn)
	}
	defer fd.Close()

	_, err = fd.Write([]byte(data))
	if err != nil {
		fmt.Println(err.Error())
		return500(conn)
	}

	fd.Sync()
	var res http.Response
	res.StatusCode = 201
	returnResponse(request, res, conn)
}

func handleFileServeEndpoint(request http.Request, conn net.Conn) {
	var filename string = strings.TrimPrefix(request.URL.Path, "/files/")
	var filePath string = filesPath + filename
	fd, err := os.Open(filePath)
	if err != nil {
		return404(conn)
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

	var res http.Response
	res.StatusCode = 200
	res.Body = io.NopCloser(bytes.NewBufferString(data))
	res.ContentLength = int64(size)
	res.Header = make(http.Header)
	res.Header.Add("content-type", "application/octet-stream")
	returnResponse(request, res, conn)
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
		return b.Bytes(), nil
	}
	return data, nil
}

func return404(conn net.Conn) {
	var res http.Response
	res.StatusCode = 404
	req := http.Request{}
	returnResponse(req, res, conn)
}

func return500(conn net.Conn) {
	var res http.Response
	res.StatusCode = 500
	req := http.Request{}
	returnResponse(req, res, conn)
}
