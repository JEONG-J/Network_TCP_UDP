/* 20195914 */
/* Jeong eui chan */

package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

const (
	server1Address = "nsl2.cau.ac.kr:45914"
	server2Address = "nsl5.cau.ac.kr:55914"
)

func main() {
	// Check for the correct number of command-line arguments
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run SplitFileClient.go <put/get> <filename>")
		return
	}

	command := os.Args[1]
	filename := os.Args[2]

	// Execute the appropriate function based on the command
	switch command {
	case "put":
		putFile(filename)
	case "get":
		getFile(filename)
	default:
		fmt.Println("Unknown command:", command)
	}
}

// Function to handle the 'put' command
func putFile(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Split the file data into two parts
	part1, part2 := splitFile(data)
	part1Filename := strings.TrimSuffix(filename, ".txt") + "-part1.txt"
	part2Filename := strings.TrimSuffix(filename, ".txt") + "-part2.txt"

	// Send each part to the respective server
	if !sendToServer(server1Address, "put "+part1Filename, part1) {
		fmt.Println("Failed to send part1 to server 1")
		return
	}
	if !sendToServer(server2Address, "put "+part2Filename, part2) {
		fmt.Println("Failed to send part2 to server 2")
		return
	}
}

// Function to handle the 'get' command
func getFile(filename string) {
	part1Filename := strings.TrimSuffix(filename, ".txt") + "-part1.txt"
	part2Filename := strings.TrimSuffix(filename, ".txt") + "-part2.txt"

	// Receive each part from the respective server
	part1 := receiveFromServer(server1Address, "get "+part1Filename)
	if part1 == nil {
		fmt.Println("Failed to receive part1 from server 1")
		return
	}

	part2 := receiveFromServer(server2Address, "get "+part2Filename)
	if part2 == nil {
		fmt.Println("Failed to receive part2 from server 2")
		return
	}

	// Merge the two parts and save to a new file
	merged := mergeFile(part1, part2)
	mergedFilename := strings.TrimSuffix(filename, ".txt") + "-merged.txt"
	err := ioutil.WriteFile(mergedFilename, merged, 0644)
	if err != nil {
		fmt.Println("Error writing merged file:", err)
		return
	}
	fmt.Println("Merged file saved as", mergedFilename)
}

// Function to split the data into two parts
func splitFile(data []byte) ([]byte, []byte) {
	var part1, part2 []byte
	for i := 0; i < len(data); i++ {
		if i%2 == 0 {
			part1 = append(part1, data[i])
		} else {
			part2 = append(part2, data[i])
		}
	}
	return part1, part2
}

// Function to merge two parts into one
func mergeFile(part1, part2 []byte) []byte {
	var merged []byte
	i, j := 0, 0
	for i < len(part1) || j < len(part2) {
		if i < len(part1) {
			merged = append(merged, part1[i])
			i++
		}
		if j < len(part2) {
			merged = append(merged, part2[j])
			j++
		}
	}
	return merged
}

// Function to send data to a server
func sendToServer(address, command string, data []byte) bool {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return false
	}
	defer conn.Close()

	// Send the command to the server
	_, err = fmt.Fprintln(conn, command)
	if err != nil {
		fmt.Println("Error sending command:", err)
		return false
	}

	// Send the data to the server
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error sending data:", err)
		return false
	}
	return true
}

// Function to receive data from a server
func receiveFromServer(address, command string) []byte {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return nil
	}
	defer conn.Close()

	// Send the command to the server
	_, err = fmt.Fprintln(conn, command)
	if err != nil {
		fmt.Println("Error sending command:", err)
		return nil
	}

	// Read the response from the server
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Println("Error receiving data:", err)
		return nil
	}

	// Check for an error message from the server
	if strings.HasPrefix(string(data), "Error:") {
		fmt.Println(string(data))
		return nil
	}

	return data
}
