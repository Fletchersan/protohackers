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
	listen_on := [4]byte{127, 0, 0, 1}
	socketAddr := unix.SockaddrInet4{
		Port: 8000,
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
	var activeFDSet, readFDSet unix.FdSet
	activeFDSet.Bits = [32]int32{}
	activeFDSet.Bits[socketFD/32] |= (1 << uint(socketFD) % 32)
	fdAddrMap := make(map[int]unix.Sockaddr, unix.FD_SETSIZE)
	for {
		readFDSet = activeFDSet
		_, err := unix.Select(int(1024), &readFDSet, nil, nil, nil)
		if err != nil {
			log.Fatal("Select: ", err)
		}
		for i := 0; i < unix.FD_SETSIZE+1; i++ {
			if readFDSet.Bits[i/32]&(1<<uint(i)%32) != 0 {
				if i == socketFD {
					newConn, newConnAddr, err := unix.Accept(socketFD)
					if err != nil {
						log.Fatal("Accept: ", err)
					}
					activeFDSet.Bits[newConn/32] |= (1 << uint(i) % 32)
					fdAddrMap[newConn] = newConnAddr
				}
			} else {
				msg := make([]byte, 8000)
				_, _, err := unix.Recvfrom(i, msg, 0)
				if err != nil {
					log.Print("Recvfrom: ", err)
					activeFDSet.Bits[i/32] &^= (1 << (uint(i) % 32))
					unix.Close(i)
					delete(fdAddrMap, i)
					continue
				}
				clientAddr := fdAddrMap[i].(*unix.SockaddrInet4)
				err = unix.Sendmsg(
					i, msg, nil, clientAddr, unix.MSG_DONTWAIT,
				)
				if err != nil {
					log.Print("Sendmsg: ", err)
				}
				activeFDSet.Bits[i/32] &^= (1 << (uint(i) % 32))
				delete(fdAddrMap, i)
				unix.Close(i)
			}
		}
	}
}

func main() {
	fmt.Println("Hello World!")
	server(8080)
}
