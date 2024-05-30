package main

import (
	"fmt"
	"net"
)

const (
	HOST = "localhost"
)

func udpListener() {
	udpServer, err := net.ListenPacket("udp", "HOST"+":514")
	if err != nil {
		fmt.Println("Failed to start UDP listener.")
		return
	}
	fmt.Println("Started UDP logging.")

	defer udpServer.Close()
	for {
		buf := make([]byte, 1024)
		_, addr, err := udpServer.ReadFrom(buf)
		if err == nil {
			go processLog(buf, addr)
		}
	}
}

func tcpListener() {
	tcpServer, err := net.Listen("tcp", "HOST"+":6514")
	if err != nil {
		fmt.Println("Failed to start TCP listener.")
		return
	}
	fmt.Println("Started TCP logging.")

	defer tcpServer.Close()
	for {
		conn, err := tcpServer.Accept()
		if err != nil {
			fmt.Println("Failed to establish connection")
		} else {
			buffer := make([]byte, 1024)
			_, err := conn.Read(buffer)
			conn.Close()
			if err == nil {

			}
		}
	}
}

func processLog(buf []byte, addr net.Addr) {

}

func main() {
	fmt.Println("Starting logging...")
	go udpListener()
	go tcpListener()
}
