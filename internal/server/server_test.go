package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/maximkir777/word_of_wisdom/pkg/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dummyHandler is a minimal implementation of the Handler interface.
type dummyHandler struct{}

// ProcessRequest returns a fixed response or an error if the message equals "error\n".
func (d *dummyHandler) ProcessRequest(ctx context.Context, msgStr string, clientInfo string) (*protocol.Message, error) {
	if msgStr == "error\n" {
		return nil, fmt.Errorf("dummy error")
	}
	return &protocol.Message{Header: 123, Payload: "OK"}, nil
}

func TestServerBasicConnection(t *testing.T) {
	dh := &dummyHandler{}
	srv := NewServer("127.0.0.1:0", dh, nil)
	err := srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	// Get the actual listening address.
	addr := srv.ln.Addr().String()

	// Connect to the server.
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	// Send a valid message.
	msg := &protocol.Message{Header: 1, Payload: "test"}
	_, err = conn.Write([]byte(msg.Stringify() + "\n"))
	require.NoError(t, err)

	// Read the response.
	reader := bufio.NewReader(conn)
	respStr, err := reader.ReadString('\n')
	require.NoError(t, err)
	assert.Equal(t, "123|OK\n", respStr)
}

func TestServerErrorResponse(t *testing.T) {
	dh := &dummyHandler{}
	srv := NewServer("127.0.0.1:0", dh, nil)
	err := srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	addr := srv.ln.Addr().String()
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	// Send a message that triggers an error.
	_, err = conn.Write([]byte("error\n"))
	require.NoError(t, err)

	reader := bufio.NewReader(conn)
	respStr, err := reader.ReadString('\n')
	require.NoError(t, err)
	// The server sends error messages with header ResponseResource.
	expected := fmt.Sprintf("%d|Error: dummy error\n", protocol.ResponseResource)
	assert.Equal(t, expected, respStr)
}

func TestServerGracefulShutdown(t *testing.T) {
	dh := &dummyHandler{}
	srv := NewServer("127.0.0.1:0", dh, nil)
	err := srv.Start()
	require.NoError(t, err)

	// Give some time for the server to start.
	time.Sleep(50 * time.Millisecond)
	srv.Stop()

	// After shutdown, new connections should fail.
	addr := srv.ln.Addr().String()
	_, err = net.Dial("tcp", addr)
	require.Error(t, err)
}
