package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
)

// HOST	   = The host where the server will run
// LOGFILE = The file name where the logs will be saved to. The extension will be .log
// UDPPORT = The UDP port where the UDP listener will run on
// TCPPORT = The TCP port where the TCP listener will run on
const (
	HOST    = "localhost"
	LOGFILE = "syslog"
	UDPPORT = "514"
	TCPPORT = "6514"
)

// Server variables are global to allow for graceful shutdown upon key interrupt.
var globalUdpServer *net.PacketConn
var globalTcpServer *net.Listener

var m sync.Mutex

var c = make(chan os.Signal)

// This function gets called when the user interrupts code execution with Ctrl+C
// This allows for files that are being written to to not be corrupted.
func cleanup() {
	<-c
	fmt.Println()
	if globalUdpServer != nil {
		fmt.Println("Stopping UDP logging...")
		(*globalUdpServer).Close()
		globalUdpServer = nil
	}
	if globalTcpServer != nil {
		fmt.Println("Stopping TCP logging...")
		(*globalTcpServer).Close()
		globalTcpServer = nil
	}
	m.Lock()
	fmt.Println("Stopped logging successfully.")
}

// Starts the UDP listener
func udpListener(wg *sync.WaitGroup) {
	defer wg.Done()
	udpServer, err := net.ListenPacket("udp", fmt.Sprintf("%s:%s", HOST, UDPPORT))
	if err != nil {
		fmt.Println("Failed to start UDP listener.")
		return
	}

	globalUdpServer = &udpServer
	fmt.Printf("Started UDP logging on port %s.\n", UDPPORT)

	// defer udpServer.Close()
	for globalUdpServer != nil {
		buf := make([]byte, 1024)
		_, addr, err := udpServer.ReadFrom(buf)
		if err == nil {
			fmt.Printf("Received UDP log from %s\n", addr.String())
			go processLog(wg, buf)
		}
	}
}

// Starts the TCP listener
func tcpListener(wg *sync.WaitGroup) {
	defer wg.Done()
	tcpServer, err := net.Listen("tcp", fmt.Sprintf("%s:%s", HOST, TCPPORT))
	if err != nil {
		fmt.Println("Failed to start TCP listener.")
		return
	}

	globalTcpServer = &tcpServer
	fmt.Printf("Started TCP logging on port %s.\n", TCPPORT)

	// defer tcpServer.Close()
	for globalTcpServer != nil {
		conn, err := tcpServer.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				fmt.Println("Failed to establish connection")
			}
		} else {
			buf := make([]byte, 1024)
			_, err := conn.Read(buf)
			conn.Close()
			if err == nil {
				fmt.Printf("Received TCP log from %s\n", conn.RemoteAddr().String())
				go processLog(wg, buf)
			}
		}
	}
}

// Processes log files received according to the IETF format
func processLog(wg *sync.WaitGroup, buf []byte) {
	wg.Add(1)
	defer wg.Done()
	// Removes null characters from byte slice
	buf = bytes.Trim(buf, "\x00")
	// This regular expression matches a log in the IETF format.
	// If the expression is not matched, the log is still saved in another file with an appended "-invalid".
	matched, _ := regexp.Match(`^<[0-9]{1,2}>[0-9]{1} [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{0,9}|)Z [a-zA-Z0-9\.]+ [a-zA-Z0-9\.]+ - [a-zA-Z0-9\.]+ - ([^\n])+$`, buf)
	m.Lock()
	if matched {
		f, err := os.OpenFile(fmt.Sprintf("%s.log", LOGFILE), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			fmt.Println("Could not create/open log file.")
		} else {
			_, err := f.Write(buf)
			if err != nil {
				fmt.Println("Could not create the log entry.")
			}
			_, _ = f.WriteString("\n")
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
			_, _ = f.WriteString("\n")
			f.Close()
		}
	}
	m.Unlock()
}

func main() {
	// Command to bind Ctrl+C to graceful shutdown of the program
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go cleanup()
	fmt.Println("Starting logging...")
	var wg sync.WaitGroup

	wg.Add(2)
	go udpListener(&wg)
	go tcpListener(&wg)

	wg.Wait()
	fmt.Println("Finished operations.")
}
