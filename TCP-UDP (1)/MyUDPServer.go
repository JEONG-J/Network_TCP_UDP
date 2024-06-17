/* MyUDPServer.go */
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

func main() {
	serverPort := "25914"   // Setting the port number for the server
	startTime := time.Now() // Recording the server start time
	requestCount := 0       // Counter for counting request

	/* Creating a PacketConn for UDP connection, error handling omitted */
	pconn, err := net.ListenPacket("udp", ":"+serverPort)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen on port %s: %v\n", serverPort, err)
		os.Exit(1)
	}
	defer pconn.Close() // Closing pconn when function exitst

	fmt.Printf("Server is ready to receive on port %s\n", serverPort)

	/* Creating and setting up a channel for handling interrupt singals */
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Handling interrupt signals in a separate goroutine
	go func() {
		<-c
		fmt.Println("\nBye bye~")
		os.Exit(0)
	}()

	buffer := make([]byte, 1024) // Buffer for reading UDP messages

	for {
		// Receiving a message frome the client
		count, r_addr, err := pconn.ReadFrom(buffer)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from UDP: %v\n", err)
			continue // Skip this iteration on error
		}

		fmt.Printf("UDP message from %s\n", r_addr.String())
		requestCount++

		// Converting the received message to a string and processing it
		receiveMsg := strings.TrimSpace(string(buffer[:count]))
		parts := strings.Split(receiveMsg, ":")

		// Extracting and converting the command number
		command, err := strconv.Atoi(parts[0])

		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid command received: %s\n", parts[0])
			continue // Skip this iteration on error
		}

		fmt.Printf("Command %d\n", command)

		var response string
		// Performing differnet operations on the command
		switch command {
		case 1:
			// If the command is 1, convert the received message to uppercase
			if len(parts) > 1 {
				response = strings.ToUpper(parts[1])
			}
		case 2:
			// If the command is 2, send server uptime in HH:MM:SS format
			elapsedTime := time.Since(startTime)
			hours := int(elapsedTime.Hours())
			minutes := int(elapsedTime.Minutes()) % 60
			seconds := int(elapsedTime.Seconds()) % 60
			response = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		case 3:
			// If the command is 3, send the client't address and port
			clientIP, clientPort, _ := net.SplitHostPort(r_addr.String())
			response = fmt.Sprintf("%s:%s", clientIP, clientPort)
		case 4:
			// If the command is 4, send the request count
			response = strconv.Itoa(requestCount)
		default:
			// Default: do not send a message
			break
		}

		// Sending a response to the client only if response is not empty
		if response != "" {
			pconn.WriteTo([]byte(response), r_addr)
		}
	}
}
