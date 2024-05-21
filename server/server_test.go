package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// Helper type to create in-memory net.Conn setup
type mockConn struct {
	net.Conn
	r *bufio.Reader
	w *bytes.Buffer
}

// Override ReadString method for mockConn
func (mc *mockConn) ReadString(delim byte) (string, error) {
	return mc.r.ReadString(delim)
}

// Override Write method for mockConn
func (mc *mockConn) Write(b []byte) (int, error) {
	return mc.w.Write(b)
}

// Override Close method for mockConn
func (mc *mockConn) Close() error {
	return nil
}

// Helper type to mock net.Addr
type mockAddr struct{}

// Override RemoteAddr method for mockConn
func (mc *mockConn) RemoteAddr() net.Addr {
	return &mockAddr{}
}

// Override Network method for mockAddr
func (m *mockAddr) Network() string { return "tcp" }

// Override String method for mockAddr
func (m *mockAddr) String() string { return "127.0.0.1:0" }

// Helper function to create new mockConn
func newMockConn(input string) *mockConn {
	return &mockConn{
		r: bufio.NewReader(strings.NewReader(input)),
		w: new(bytes.Buffer),
	}
}

// Test function for logMessage
func TestLogMessage(t *testing.T) {
	expected := "Test Log\n"
	logMessage(expected) // Write test log message

	// Read the log file to verify its contents
	content, err := os.ReadFile("history.log")
	if err != nil {
		t.Fatalf("Failed to read history log: %s", err)
	}
	if !strings.Contains(string(content), expected) {
		t.Errorf("Expected file content %q, got %q", expected, string(content))
	}

	// Clean up: Remove the test log entry
	cleanUpLog("Test Log\n")
}

// Helper function to clean up test log entry
func cleanUpLog(testEntry string) {
	content, err := os.ReadFile("history.log")
	if err != nil {
		return
	}

	// Remove the test entry and write back the cleaned content
	updatedContent := strings.Replace(string(content), testEntry, "", -1)
	err = os.WriteFile("history.log", []byte(updatedContent), 0644)
	if err != nil {
	}
}

// Test function for broadcastMessage
func TestBroadcastMessage(t *testing.T) {
	client := newMockConn("")
	clientList.Store(client, "tester")
	defer clientList.Delete(client)

	msg := "Hello Everyone!!! (TEST)\n"
	broadcastMessage(msg)

	if got := client.w.String(); !strings.Contains(got, msg) {
		t.Errorf("Expected broadcast message %q, got %q", msg, got)
	}
}

// Test function for handleMessage
func TestHandleMessage(t *testing.T) {
	client := newMockConn("")
	defer clientList.Delete(client)

	currentTime := time.Now().Format("15:04")
	handleMessage(client, "/join Diyar")
	expected := fmt.Sprintf("%s Notice: \"Diyar\" joined the chat\n", currentTime)

	if got := client.w.String(); !strings.Contains(got, expected) {
		t.Errorf("Expected join message %q, got %q", expected, got)
	}
}

// Test function for handleConnection
func TestHandleConnection(t *testing.T) {
	input := "/join Dominic\nHello\n"
	client := newMockConn(input)
	clientList.Store(client, "Dominic")
	defer clientList.Delete(client)

	done := make(chan bool)
	go func() {
		// Drain the activeClients channel to avoid blocking
		for range activeClients {
		}
	}()

	// Start handleConnection in a goroutine
	go func() {
		handleConnection(client, client.r)
		done <- true
	}()

	<-done

	currentTime := time.Now().Format("15:04")
	expected := fmt.Sprintf("%s - Dominic: Hello\n", currentTime)
	got := client.w.String()
	if !strings.Contains(got, expected) {
		t.Errorf("Expected chat message %q, got %q", expected, got)
	}
}

// New test function for handling invalid commands
func TestHandleInvalidCommand(t *testing.T) {
	client := newMockConn("")
	defer clientList.Delete(client)

	handleMessage(client, "/invalid")
	expected := "Write: /join `nickname`.\n"

	if got := client.w.String(); !strings.Contains(got, expected) {
		t.Errorf("Expected error message %q, got %q", expected, got)
	}
}

// New test function for handling /users command
func TestHandleUsersCommand(t *testing.T) {
	client := newMockConn("")
	clientList.Store(client, "tester")
	defer clientList.Delete(client)

	handleMessage(client, "/users")
	expected := "clients in chat\n"

	if got := client.w.String(); !strings.Contains(got, expected) {
		t.Errorf("Expected users message %q, got %q", expected, got)
	}
}

// New test function to ensure logMessage creates file if it doesn't exist
func TestLogMessageCreatesFile(t *testing.T) {
	_ = os.Remove("history.log") // Ensure file does not exist

	logMessage("Test Log\n")
	if _, err := os.Stat("history.log"); os.IsNotExist(err) {
		t.Fatalf("Expected history.log file to be created, but it doesn't exist")
	}
	cleanUpLog("Test Log\n")
}

// New test function for handling client disconnect
func TestHandleClientDisconnect(t *testing.T) {
	client := newMockConn("")
	clientList.Store(client, "tester")
	defer clientList.Delete(client)

	done := make(chan bool)
	go func() {
		handleConnection(client, client.r)
		done <- true
	}()

	time.Sleep(time.Second) // Allow some time for the goroutine to run

	client.Close()
	<-done // Wait for handleConnection to complete

	_, ok := clientList.Load(client)
	if ok {
		t.Errorf("Expected client to be removed from clientList after disconnect")
	}
}
