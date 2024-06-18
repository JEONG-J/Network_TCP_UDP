/**
 * UDPClient.go
 **/

/* 20195914 Jeong eui chan */

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/* startMenu function displays the menu and returns the user choice */
func startMenu() int {
	var choice int
	fmt.Println("<Menu>")
	fmt.Println("1) Convert text to UPPER-case")
	fmt.Println("2) Get server running time")
	fmt.Println("3) Get my IP address and port number")
	fmt.Println("4) Get server request count")
	fmt.Println("5) Exit")
	fmt.Print("Input option: ")
	fmt.Scanln(&choice)
	return choice
}

func main() {
	serverName := "nsl2.cau.ac.kr"             // Define the server address
	serverPort := "25914"                      // Define the server port
	pconn, err := net.ListenPacket("udp", ":") // Open a UDP connection

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening UDP connection: %v\n", err)
		os.Exit(1)
	}

	defer pconn.Close() // Ensure the connections is closed on function exit

	/* Get local address information */
	localAddr := pconn.LocalAddr().(*net.UDPAddr)
	fmt.Printf("Client is running on port %d\n", localAddr.Port)

	/* Creating and setting up a channel for handling interrupt singals */
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nBye bye~")
		os.Exit(0)
	}()

	for {
		choice := startMenu() // Display menu and get user choice

		if choice == 5 {
			fmt.Println("Bye bye~")
			break
		}

		var message string
		if choice == 1 {
			fmt.Print("Input sentence: ")
			input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			input = strings.TrimSpace(input)
			message = fmt.Sprintf("1:%s", input)
		} else {
			message = fmt.Sprintf("%d", choice)
		}

		startTime := time.Now()

		/* Resolve server address and send the message */
		serverAddr, err := net.ResolveUDPAddr("udp", serverName+":"+serverPort)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving UDP address: %v\n", err)
			continue
		}

		_, err = pconn.WriteTo([]byte(message), serverAddr)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending UDP message: %v\n", err)
			continue
		}

		buffer := make([]byte, 1024)
		n, _, err := pconn.ReadFrom(buffer) // Read response from server

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from UDP: %v\n", err)
			continue
		}

		endTime := time.Since(startTime).Seconds() * 1000 // Calculate round-trip time

		response := strings.TrimSpace(string(buffer[:n])) // Process server response

		/* Print server response based on the command */
		switch choice {
		case 2:
			parts := strings.Split(response, ":")
			if len(parts) == 3 {
				hours, hoursErr := strconv.Atoi(parts[0])
				minutes, minutesErr := strconv.Atoi(parts[1])
				seconds, secondsErr := strconv.Atoi(parts[2])

				if hoursErr == nil && minutesErr == nil && secondsErr == nil {
					formattedTime := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
					fmt.Printf("Reply from server: run time = %s\n", formattedTime)
				} else {
					fmt.Println("Invalid time format received from server")
				}
			} else {
				fmt.Println("Invalid time format received from server")
			}
		case 3:
			parts := strings.Split(response, ":")
			fmt.Printf("Reply from server: client IP = %s, port = %s\n", parts[0], parts[1])
		case 4:
			fmt.Printf("Reply from server: requests served = %s\n", response)
		default:
			fmt.Printf("Reply from server: %s\n", response)
		}

		fmt.Printf("RTT = %.3f ms\n", endTime)
	}
}
