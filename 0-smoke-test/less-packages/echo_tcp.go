package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

func server(port int) {
	socketFD, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_IP)

	if err != nil {
		log.Fatal("Socket: ", err)
	}
	listen_on := [4]byte{0, 0, 0, 0}
	socketAddr := unix.SockaddrInet4{
		Port: port,
		Addr: listen_on,
	}

	err = unix.Bind(socketFD, &socketAddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}
	err = unix.Listen(socketFD, 10)
	if err != nil {
		log.Fatal("Listen: ", err)
	}
	log.Printf("Listening on port %d", port)
	var activeFDSet, readFDSet unix.FdSet
	FDZero(&activeFDSet)
	FDSet(socketFD, &activeFDSet)
	fdAddrMap := make(map[int]unix.Sockaddr, unix.FD_SETSIZE)
	for {
		readFDSet = activeFDSet
		_, err := unix.Select(unix.FD_SETSIZE, &readFDSet, nil, nil, nil)
		if err != nil {
			log.Fatal("Select: ", err)
		}
		for i := 0; i < unix.FD_SETSIZE; i++ {
			if isPosSet(i, &readFDSet) {
				if i == socketFD {
					newConn, newConnAddr, err := unix.Accept(socketFD)
					if err != nil {
						log.Fatal("Accept: ", err)
					}
					FDSet(newConn, &activeFDSet)
					fdAddrMap[newConn] = newConnAddr
				} else {
					msg := make([]byte, 4096)
					msgSize, err := unix.Read(i, msg)
					if err != nil {
						log.Print("Recvfrom: ", err)
						FDClr(i, &activeFDSet)
						delete(fdAddrMap, i)
						unix.Close(i)
						continue
					}
					_, err = unix.Write(i, msg[:msgSize])
					if err != nil {
						log.Print("Sendmsg: ", err)
					}
					if msgSize == 0 {
						FDClr(i, &activeFDSet)
						delete(fdAddrMap, i)
						unix.Close(i)
					}
				}
			}
		}
	}
}

func FDZero(fdSet *unix.FdSet) {
	for i := 0; i < len(fdSet.Bits); i++ {
		fdSet.Bits[i] = 0
	}
}

func FDSet(pos int, fdSet *unix.FdSet) {
	if pos < 0 || pos >= unix.FD_SETSIZE {
		return
	}
	index := pos / 64
	bit := pos % 64
	if index < len(fdSet.Bits) {
		fdSet.Bits[index] |= (1 << bit)
	}
}

func FDClr(pos int, fdSet *unix.FdSet) {
	if pos < 0 || pos >= unix.FD_SETSIZE {
		return
	}
	index := pos / 64
	bit := pos % 64
	if index < len(fdSet.Bits) {
		fdSet.Bits[index] &^= (1 << bit)
	}
}

func isPosSet(pos int, fdSet *unix.FdSet) bool {
	if pos < 0 || pos >= unix.FD_SETSIZE {
		return false
	}
	index := pos / 64
	bit := pos % 64
	if index < len(fdSet.Bits) {
		return fdSet.Bits[index]&(1<<bit) != 0
	}
	return false
}

func main() {
	fmt.Println("Hello World!")
	server(8080)
}
