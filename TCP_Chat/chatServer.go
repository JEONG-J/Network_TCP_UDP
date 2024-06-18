/* 20195914 */
/* Jeong eui chan */

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	port = "8080" // Server port configuration
)

var (
	clients   = make(map[string]net.Conn)
	nicknames = make(map[string]string)
	mu        sync.Mutex
)

func main() {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Listening on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			continue
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	reader := bufio.NewReader(conn)
	nick, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading nickname:", err.Error())
		conn.Close()
		return
	}
	nickname := strings.TrimSpace(nick)

	mu.Lock()
	// Check if the chat room is full or the nickname is already in use
	if len(clients) >= 8 {
		conn.Write([]byte("Chatting room full. Cannot connect.\n"))
		conn.Close()
		mu.Unlock()
		return
	}
	if _, ok := nicknames[nickname]; ok {
		conn.Write([]byte("Nickname already used by another user. Cannot connect.\n"))
		conn.Close()
		mu.Unlock()
		return
	}
	clients[nickname] = conn
	nicknames[nickname] = nickname
	mu.Unlock()

	welcomeMessage := fmt.Sprintf("Welcome %s to CAU net-class chat room at %s. There are %d users in the room.\n", nickname, conn.LocalAddr().String(), len(clients))
	conn.Write([]byte(welcomeMessage))

	// Log the join message to the server console
	fmt.Printf("%s joined from %s. There are %d users in the room.\n", nickname, conn.RemoteAddr().String(), len(clients))

	go handleMessages(nickname, conn)
}

func handleMessages(nickname string, conn net.Conn) {
	// Disconnect the client if the message contains the banned phrase
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(strings.ToLower(text), "i hate professor") {
			disconnectWithMessage(nickname)
			return
		}
		handleCommands(nickname, text)
	}
	disconnect(nickname)
}

func handleCommands(nickname, text string) {
	// Broadcast the message to all clients except the sender
	if strings.HasPrefix(text, "\\") {
		command := strings.Fields(text)
		switch command[0] {
		case "\\ls":
			listClients(nickname)
		case "\\secret":
			if len(command) < 3 || !clientExists(command[1]) {
				clients[nickname].Write([]byte("Error: Nickname does not exist.\n"))
				return
			}
			sendSecret(nickname, command[1], strings.Join(command[2:], " "))
		case "\\except":
			if len(command) < 3 {
				clients[nickname].Write([]byte("Usage: \\except <nickname> <message>\n"))
				return
			}
			if !clientExists(command[1]) {
				clients[nickname].Write([]byte("Error: Nickname does not exist.\n"))
				return
			}
			broadcastExcept(nickname, command[1], nickname+"> "+strings.Join(command[2:], " ")+"\n")
		case "\\ping":
			sendPing(nickname)
		case "\\quit":
			disconnect(nickname)
		default:
			clients[nickname].Write([]byte("Invalid command\n"))
			fmt.Println("Invalid command: " + text)
		}
	} else {
		broadcastExcept(nickname, "", nickname+"> "+text+"\n")
	}
}

// Check if a client with the given nickname exists
func clientExists(nickname string) bool {
	_, exists := clients[nickname]
	return exists
}

// Broadcast a message to all clients except the sender and the excluded client
func broadcastExcept(nickname, exclude, message string) {
	mu.Lock()
	defer mu.Unlock()
	for nick, conn := range clients {
		if nick != nickname && nick != exclude {
			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println("Error broadcasting to", nick, ":", err.Error())
			}
		}
	}
}

// Send the list of connected clients to the requesting client
func listClients(nickname string) {
	mu.Lock()
	defer mu.Unlock()
	clientList := "Connected users:\n"
	for nick, conn := range clients {
		clientList += fmt.Sprintf("%s - %s\n", nick, conn.RemoteAddr().String())
	}
	clients[nickname].Write([]byte(clientList))
}

// Send a secret message from the sender to the receiver
func sendSecret(sender, receiver, message string) {
	mu.Lock()
	defer mu.Unlock()
	if conn, ok := clients[receiver]; ok {
		_, err := conn.Write([]byte("from " + sender + " > " + message + "\n"))
		if err != nil {
			fmt.Println("Error sending secret message to", receiver, ":", err.Error())
		}
	}
}

// Send a ping response with the round-trip time (RTT)
func sendPing(nickname string) {
	start := time.Now()
	elapsed := time.Since(start)
	elapsedMS := float64(elapsed.Nanoseconds()) / 1e6
	clients[nickname].Write([]byte(fmt.Sprintf("RTT: %.3f ms\n", elapsedMS)))
}

// Disconnect a client with a message to all other clients
func disconnectWithMessage(nickname string) {
	mu.Lock()
	conn, ok := clients[nickname]
	if ok {
		conn.Close()
		delete(clients, nickname)
		delete(nicknames, nickname)
		message := fmt.Sprintf("[%s is disconnected. There are %d users in the chat room.]\n", nickname, len(clients))
		mu.Unlock()

		// Log the message to the server console and broadcast it to all clients
		logAndBroadcast(message)
	} else {
		mu.Unlock()
	}
}

func disconnect(nickname string) {
	mu.Lock()
	conn, ok := clients[nickname]
	if ok {
		conn.Write([]byte("gg~\n"))
		conn.Close()
		delete(clients, nickname)
		delete(nicknames, nickname)
		message := fmt.Sprintf("[%s left the room. There are %d users now.]\n", nickname, len(clients))
		mu.Unlock()

		// Log the message to the server console and broadcast it to all clients
		logAndBroadcast(message)
	} else {
		mu.Unlock()
	}
}

func logAndBroadcast(message string) {
	// Log the message to the server console
	fmt.Print(message)

	// Broadcast the message to all clients
	broadcast(message)
}

func broadcast(message string) {
	mu.Lock()
	defer mu.Unlock()
	for nick, conn := range clients {
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error broadcasting to", nick, ":", err.Error())
		}
	}
}
