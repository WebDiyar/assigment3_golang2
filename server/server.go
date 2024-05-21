package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Constants for connection configuration
const (
	CONN_PORT = ":3335" // Connection port
	CONN_TYPE = "tcp"   // Connection type
)

// Variables for managing clients and message logging
var (
	clientNames   = make(map[net.Conn]string) // Map to store client names
	historyMutex  sync.Mutex                  // Mutex for synchronizing history log access
	clientList    sync.Map                    // Concurrent map for storing active clients
	activeClients = make(chan net.Conn)       // Channel for handling active clients
)

// Function to log messages to the history.log file
func logMessage(message string) {
	historyMutex.Lock()         // Lock mutex for safe access
	defer historyMutex.Unlock() // Ensure mutex is unlocked after function execution

	// Open history log file in append mode, create if not exists
	file, err := os.OpenFile("history.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening history log:", err)
		return
	}
	defer file.Close() // Ensure file is closed after function execution

	_, err = file.WriteString(message) // Write message to the log file
	if err != nil {
		log.Println("Error writing to history log:", err)
	}
}

// Function to broadcast messages to all clients
func broadcastMessage(msg string) {
	clientList.Range(func(key, value interface{}) bool {
		conn, ok := key.(net.Conn)
		if !ok {
			return true // Continue to next client if type assertion fails
		}

		_, err := conn.Write([]byte(msg)) // Write message to client
		if err != nil {
			log.Println("Broadcast error:", err)
			conn.Close()            // Close connection on error
			clientList.Delete(conn) // Remove client from list
		}
		logMessage(msg) // Log the broadcast message
		return true
	})
}

// Function to handle individual client connections
func handleConnection(conn net.Conn, reader *bufio.Reader) {
	defer func() {
		conn.Close()              // Close connection on function exit
		clientList.Delete(conn)   // Remove client from list
		delete(clientNames, conn) // Remove client name
	}()

	activeClients <- conn // Send connection to activeClients channel

	for {
		msg, err := reader.ReadString('\n') // Read message from client
		if err != nil {
			if err == io.EOF {
				// Handle client disconnection
				leaveMessage := fmt.Sprintf("%s Notice: \"%s\" left the chat\n", time.Now().Format("15:04"), clientNames[conn])
				broadcastMessage(leaveMessage)
				break
			}
			log.Printf("Client disconnected: %v\n", conn.RemoteAddr())
			leaveMessage := fmt.Sprintf("%s Notice: \"%s\" left the chat\n", time.Now().Format("15:04"), clientNames[conn])
			broadcastMessage(leaveMessage)
			break
		}

		handleMessage(conn, strings.TrimSpace(msg)) // Handle the received message
	}
}

// Function to handle specific client messages
func handleMessage(conn net.Conn, msg string) {
	if strings.HasPrefix(msg, "/join") {
		// Handle client joining with nickname
		nickname := strings.TrimSpace(strings.TrimPrefix(msg, "/join"))
		joinMessage := fmt.Sprintf("%s Notice: \"%s\" joined the chat\n", time.Now().Format("15:04"), nickname)
		clientNames[conn] = nickname
		clientList.Store(conn, nickname)
		broadcastMessage(joinMessage)
	} else if strings.HasPrefix(msg, "/users") {
		// Handle request for number of clients
		message := fmt.Sprintf("%s Notice: \"%s\" clients in chat\n", time.Now().Format("15:04"), strconv.Itoa(len(clientNames)))
		broadcastMessage(message)
	} else {
		// Handle regular chat messages
		nickname, err := clientNames[conn]
		if !err {
			conn.Write([]byte("Write: /join `nickname`.\n"))
			return
		}
		chatMessage := fmt.Sprintf("%s - %s: %s\n", time.Now().Format("15:04"), nickname, msg)
		broadcastMessage(chatMessage)
	}
}

// Main function to start the server
func main() {
	ln, err := net.Listen(CONN_TYPE, CONN_PORT) // Start TCP server
	if err != nil {
		log.Fatal("Error starting TCP server:", err)
	}
	defer ln.Close() // Ensure listener is closed on function exit

	log.Println("Chat server started on " + CONN_PORT)

	// Goroutine to log new active clients
	go func() {
		for conn := range activeClients {
			log.Printf("New client connected: %v\n", conn.RemoteAddr())
		}
	}()

	// Main loop to accept client connections
	for {
		conn, err := ln.Accept() // Accept new client connection
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		reader := bufio.NewReader(conn)   // Create buffered reader for client connection
		go handleConnection(conn, reader) // Handle client connection in new goroutine
	}
}
