/**
 * MultiTCPServer.go
 **/

/* 20195914 Jeong eui chan */

package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	startTime     time.Time //server start time
	requestCount  int       // Variable to count the number of requests
	clientIDCount int       // counter for client ID
	clientMap     sync.Map  // sync.Map to store client information
)

// ClientInfo stores information about each client.
type ClientInfo struct {
	id     int
	socket net.Conn
}

// printClientCount prints the number of connected clients every 10 seconds.
func printClientCount() {
	for {
		time.Sleep(10 * time.Second)
		var count int
		clientMap.Range(func(_, _ interface{}) bool {
			count++
			return true
		})
		fmt.Printf("[Time: %s] Number of clients connected = %d\n", time.Now().Format("15:04:05"), count)
	}
}

// handleConnection handles individual client connections.
func handleConnection(conn net.Conn, clientInfo *ClientInfo) {
	defer func() {
		conn.Close()
		clientMap.Delete(clientInfo.id)
		var currentClientCount int
		clientMap.Range(func(_, _ interface{}) bool {
			currentClientCount++
			return true
		})
		fmt.Printf("[Time: %s] Client %d disconnected. Number of clients connected = %d\n", time.Now().Format("15:04:05"), clientInfo.id, currentClientCount)
	}()

	buffer := make([]byte, 1024) // Buffer for receiving data

	for {
		count, err := conn.Read(buffer) // read data from connection
		if err != nil {
			break // End loop when error occurs
		}

		requestCount++ // increase number of requests

		// Process received message
		receiveMsg := strings.TrimSpace(string(buffer[:count]))
		parts := strings.Split(receiveMsg, ":")
		command, err := strconv.Atoi(parts[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid command received: %v\n", err)
			continue // 에러 시 이번 반복 스킵
		}

		var response string
		switch command {
		case 1: // Convert text to uppercase
			if len(parts) > 1 {
				response = strings.ToUpper(parts[1])
			}
		case 2: // Get server execution time
			elapsedTime := time.Since(startTime)
			hours := int(elapsedTime.Hours())
			minutes := int(elapsedTime.Minutes()) % 60
			seconds := int(elapsedTime.Seconds()) % 60
			response = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		case 3: // Get client IP address and port number
			response = conn.RemoteAddr().String()
		case 4: // Get number of server requests
			response = strconv.Itoa(requestCount)
		}

		// Send response to client
		_, err = conn.Write([]byte(response))
		if err != nil {
			break // End loop when an error occurs during transmission
		}
	}
}

func main() {
	startTime = time.Now()
	serverPort := "25914"

	// Start listening for TCP connections
	listener, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listening: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	//Signal processing settings for graceful shutdown
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nBye bye~")
		os.Exit(0)
	}()

	fmt.Printf("Server is ready to receive on port %s\n", serverPort)
	go printClientCount() // Start a goroutine that periodically prints the number of clients

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		clientIDCount++
		clientInfo := &ClientInfo{
			id:     clientIDCount,
			socket: conn,
		}
		clientMap.Store(clientIDCount, clientInfo)
		var currentClientCount int
		clientMap.Range(func(_, _ interface{}) bool {
			currentClientCount++
			return true
		})
		fmt.Printf("[Time: %s] Client %d connected. Number of clients connected = %d\n", time.Now().Format("15:04:05"), clientIDCount, currentClientCount)
		go handleConnection(conn, clientInfo) // Start a goroutine to handle each client
	}
}
