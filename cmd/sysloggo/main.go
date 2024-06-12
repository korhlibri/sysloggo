package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
)

const (
	HOST    = "localhost"
	LOGFILE = "syslog"
)

func cleanupListener(serverUdp *net.PacketConn, serverTcp *net.Listener) {
	if serverUdp != nil {
		(*serverUdp).Close()
	} else {
		(*serverTcp).Close()
	}
}

func udpListener(wg *sync.WaitGroup, c *chan os.Signal) {
	defer wg.Done()
	udpServer, err := net.ListenPacket("udp", HOST+":514")
	if err != nil {
		fmt.Println("Failed to start UDP listener.")
		return
	}
	fmt.Println("Started UDP logging.")

	go func() {
		<-(*c)
		cleanupListener(&udpServer, nil)
		fmt.Println("Stopped UDP logging.")
	}()

	// defer udpServer.Close()
	for {
		buf := make([]byte, 1024)
		_, _, err := udpServer.ReadFrom(buf)
		if err == nil {
			go processLog(buf)
		}
	}
}

func tcpListener(wg *sync.WaitGroup, c *chan os.Signal) {
	defer wg.Done()
	tcpServer, err := net.Listen("tcp", HOST+":6514")
	if err != nil {
		fmt.Println("Failed to start TCP listener.")
		return
	}
	fmt.Println("Started TCP logging.")

	go func() {
		<-(*c)
		cleanupListener(nil, &tcpServer)
		fmt.Println("Stopped TCP logging.")
	}()

	// defer tcpServer.Close()
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
		} else {
			_, err := f.Write(buf)
			if err != nil {
				fmt.Println("Could not create the log entry.")
			}
			f.Close()
		}
	} else {
		f, err := os.OpenFile(fmt.Sprintf("%s-invalid.log", LOGFILE), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			fmt.Println("Could not create/open log file.")
		} else {
			_, err := f.Write(buf)
			if err != nil {
				fmt.Println("Could not create the log entry.")
			}
			f.Close()
		}
	}
}

func main() {
	fmt.Println("Starting logging...")
	var wg sync.WaitGroup
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	wg.Add(2)
	go udpListener(&wg, &c)
	go tcpListener(&wg, &c)

	wg.Wait()
}
