/* 20195914 */
/* Jeong eui chan */

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	serverAddress = "nsl2.cau.ac.kr:25914" // Server address and port settings
)

func main() {
	// Receive nickname from command line argument
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run ChatClient.go <nickname>")
		os.Exit(1)
	}
	nickname := os.Args[1]

	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// Send nickname to server
	conn.Write([]byte(nickname + "\n"))

	// Run a goroutine to read messages from the server
	go readFromServer(conn)

	// Set up a signal handler to detect Ctrl+C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println()
		fmt.Printf("gg~\n")
		conn.Write([]byte("\\quit\n"))
		os.Exit(0)
	}()

	// Send user input to the server
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "\\quit" {
			conn.Write([]byte("\\quit\n"))
			fmt.Printf("gg~\n")
			break
		}
		conn.Write([]byte(text + "\n"))
	}
}

func readFromServer(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected from server.")
			os.Exit(0)
		}
		fmt.Print(message)
	}
}
