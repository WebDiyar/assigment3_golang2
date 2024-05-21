package main

import (
	"bufio"   
	"fmt"      
	"net"      
	"os"     
	"strings"
	"sync"     
)

// Constants for server connection configuration
const (
	SERVER_PORT = ":3335" // Server port
	SERVER_TYPE = "tcp"   // Server type
)

var wg sync.WaitGroup // WaitGroup to manage goroutines

// Function to read messages from the server and print them
func readMessages(conn net.Conn) {
	defer wg.Done() // Decrement WaitGroup counter on function exit
	reader := bufio.NewReader(conn) // Create buffered reader for server connection

	for {
		message, err := reader.ReadString('\n') // Read message from server
		if err != nil {
			fmt.Println("Error reading from server:", err)
			break
		}
		fmt.Print(message) // Print received message to console
	}
}

// Function to write messages to the server
func writeMessages(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin) // Create buffered reader for standard input
	writer := bufio.NewWriter(conn)     // Create buffered writer for server connection

	for {
		message, err := reader.ReadString('\n') // Read message from standard input
		if err != nil {
			fmt.Println("Error reading from stdin:", err)
			continue
		}

		_, err = writer.WriteString(message) // Write message to server
		if err != nil {
			fmt.Println("Error writing to server:", err)
			continue
		}
		writer.Flush() // Flush the writer to ensure message is sent

		// Exit if user types "/exit"
		if strings.TrimSpace(message) == "/exit" {
			fmt.Println("Exiting chat...")
			break
		}
	}
}

// Main function to start the client
func main() {
	conn, err := net.Dial(SERVER_TYPE, SERVER_PORT) // Connect to the server
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close() // Ensure connection is closed on function exit

	wg.Add(2) // Add two goroutines to the WaitGroup

	go readMessages(conn) // Start goroutine to read messages from server
	go writeMessages(conn) // Start goroutine to write messages to server

	wg.Wait() // Wait for all goroutines to finish
}
