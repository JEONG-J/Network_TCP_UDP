/**
 * TCPServer.go
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
	"syscall"
	"time"
)

/* handleConnection handles an individual client connection */
func handleConnection(conn net.Conn, requestCount *int) {
	defer conn.Close()           // Ensure the connection is closed when this function returns
	buffer := make([]byte, 1024) // Buffer for storing incoming data

	for {
		count, err := conn.Read(buffer) // Read data from the connection
		if err != nil {
			break // If there is an error, exit the loop
		}

		(*requestCount)++ // Increment the request count

		// Process the received message
		receiveMsg := strings.TrimSpace(string(buffer[:count]))
		parts := strings.Split(receiveMsg, ":")
		command, err := strconv.Atoi(parts[0])

		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid command received: %v\n", err)
			continue // Skip this iteration on error
		}

		fmt.Printf("Connection request from %s\n", conn.RemoteAddr().String())
		fmt.Printf("Command %d\n", command)

		var response string
		// Switch on the command received
		switch command {
		case 1: // Convert text to uppcase
			if len(parts) > 1 {
				response = strings.ToUpper(parts[1])
			}
		case 2: // Get server running time
			elapsedTime := time.Since(startTime)
			hours := int(elapsedTime.Hours())
			minutes := int(elapsedTime.Minutes()) % 60
			seconds := int(elapsedTime.Seconds()) % 60
			response = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		case 3: // Get client IP address and port number
			response = conn.RemoteAddr().String()
		case 4: // Get server request count
			response = strconv.Itoa(*requestCount)
		}

		// Send the response back to the client
		_, err = conn.Write([]byte(response + "\n"))
		if err != nil {
			break // If ther is and error sending, exit the loop
		}
	}
}

var startTime time.Time // Store the starte tune if the server
var requestCount int    // Count the number of requests

func main() {
	startTime = time.Now() // Record the start time when the server starts
	serverPort := "8080"  // Define the server port

	/* Start listening for incomming TCP connections*/
	listener, err := net.Listen("tcp", ":"+serverPort)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listening: %v\n", err)
		os.Exit(1)
	}

	defer listener.Close()

	/* Setup channel and signal handling for graceful shutdown*/
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nBye bye~")
		os.Exit(0)
	}()

	fmt.Printf("Server is ready to receive on port %s\n", serverPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handleConnection(conn, &requestCount)
	}
}
