package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
)

const (
	HOST    = "localhost"
	LOGFILE = "syslog"
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
		_, _, err := udpServer.ReadFrom(buf)
		if err == nil {
			go processLog(buf)
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
			buf := make([]byte, 1024)
			_, err := conn.Read(buf)
			conn.Close()
			if err == nil {
				go processLog(buf)
			}
		}
	}
}

func processLog(buf []byte) {
	// This regular expression matches a log in the IETF format.
	matched, _ := regexp.Match(`^<[0-9]{1,2}>[0-9]{1} [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{0,9}|)Z [a-zA-Z0-9\.]+ [a-zA-Z0-9\.]+ - [a-zA-Z0-9\.]+ - ([^\n])+$`, buf)
	if matched {
		f, err := os.OpenFile(fmt.Sprintf("%s.log", LOGFILE), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			fmt.Println("Could not create/open log file.")
		}

	}
}

func main() {
	fmt.Println("Starting logging...")
	go udpListener()
	go tcpListener()
}
