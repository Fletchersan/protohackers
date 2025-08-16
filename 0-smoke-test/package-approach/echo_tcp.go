package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

func server(port int) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer listener.Close()
	fmt.Printf("Listening on :%d", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			return
		}
		go func(c net.Conn) {
			io.Copy(c, c)
			c.Close()
		}(conn)
	}
}

func main() {
	fmt.Println("Hello World!")
	server(8080)
}
