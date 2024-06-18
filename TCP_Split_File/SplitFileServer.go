/* 20195914 */
/* Jeong eui chan */

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

// The main function is the entry point of the program.
func main() {
	// Check if the number of command-line arguments is 2. If not, print the usage message.
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run SplitFileServer.go <port>")
		return
	}

	// Get the port number and start a TCP listener.
	port := os.Args[1]
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port", port)

	// Infinite loop to accept connections.
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		// Handle each connection in a separate goroutine.
		go handleConnection(conn)
	}
}

// handleConnection handles each client connection.
func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	command, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// Trim and split the command into parts.
	command = strings.TrimSpace(command)
	parts := strings.Split(command, " ")
	if len(parts) < 2 {
		fmt.Println("Invalid command:", command)
		return
	}
	commandType := parts[0]
	filename := parts[1]

	// Execute the appropriate function based on the command type.
	switch commandType {
	case "put":
		receiveFile(reader, filename)
	case "get":
		sendFile(conn, filename)
	default:
		fmt.Println("Unknown command:", commandType)
	}
}

// receiveFile receives a file from the client and saves it.
func receiveFile(reader *bufio.Reader, filename string) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println("Error reading file data:", err)
		return
	}
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	fmt.Println("File received and saved as", filename)
}

// sendFile reads a file and sends it to the client.
func sendFile(conn net.Conn, filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(conn, "Error: file %s does not exist\n", filename)
		} else {
			fmt.Println("Error reading file:", err)
		}
		return
	}
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error sending file:", err)
	}
	fmt.Println("File sent:", filename)
}
