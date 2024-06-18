/**
 * MyTCPClient.go
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

/* startMenu displays the menu and returns the user choice */
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
	// Server address and port
	serverName := "localhost"
	serverPort := "8080"

	// Establish a TCP connection to the server
	conn, err := net.Dial("tcp", serverName+":"+serverPort)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error dialing TCP server: %v\n", err)
		os.Exit(1)
	}

	defer conn.Close()

	/* Set up channel and signal handling for graceful shutdown */
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nBye bye~")
		os.Exit(0)
	}()

	for {
		// DIsplay menu and get user choice
		choice := startMenu()

		if choice == 5 {
			fmt.Println("Bye bye~")
			break
		}

		var message string

		// Formulate the message based on the choice
		if choice == 1 {
			fmt.Print("Input sentence: ")
			input, err := bufio.NewReader(os.Stdin).ReadString('\n')

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
				continue
			}

			message = fmt.Sprintf("1:%s", strings.TrimSpace(input))
		} else {
			message = fmt.Sprintf("%d", choice)
		}

		// Record start time for RTT calculation send message to server
		startTime := time.Now()
		conn.Write([]byte(message + "\n"))

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from TCP server: %v\n", err)
			continue
		}

		response := strings.TrimSpace(string(buffer[:n]))
		endTime := time.Since(startTime).Seconds() * 1000

		/* Handle the response based on the choice*/
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
