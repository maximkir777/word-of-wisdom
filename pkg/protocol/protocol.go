package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

// Header of TCP-message in protocol, means type of message
const (
	Quit              = iota // on quit each side (server or client) should close connection
	RequestChallenge         // from client to server - request new challenge from server
	ResponseChallenge        // from server to client - message with challenge for client
	RequestResource          // from client to server - message with solved challenge
	ResponseResource         // from server to client - message with useful info if solution is correct, or with error if not
)

// Message is the message struct for both server and client.
type Message struct {
	Header  int
	Payload string
}

// Stringify serializes the message to send it over a TCP connection.
// The divider between header and payload is "|".
func (m *Message) Stringify() string {
	return fmt.Sprintf("%d|%s", m.Header, m.Payload)
}

// ParseMessage parses a Message from the input string, checking header and payload.
func ParseMessage(str string) (*Message, error) {
	str = strings.TrimSpace(str)
	// If the message starts with "|" then header is missing.
	if len(str) > 0 && str[0] == '|' {
		return nil, fmt.Errorf("message doesn't match protocol")
	}

	parts := strings.SplitN(str, "|", 2) // limit to 2 parts
	if len(parts) < 1 || len(parts) > 2 {
		return nil, fmt.Errorf("message doesn't match protocol")
	}

	// Parse header.
	msgType, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse header")
	}

	msg := Message{
		Header: msgType,
	}

	if len(parts) == 2 {
		msg.Payload = parts[1]
	}
	return &msg, nil
}
