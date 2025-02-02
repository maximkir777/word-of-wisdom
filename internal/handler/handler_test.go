package handler

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/maximkir777/word-of-wisdom/pkg/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dummyWoW is a stub for the WoW interface.
type dummyWoW struct{}

func (d *dummyWoW) GetRandomWiseWord() string {
	return "dummy wise word"
}

// dummyPoW is a stub for the PoW interface.
type dummyPoW struct {
	validProof        bool
	trackRequestCount int
}

func (d *dummyPoW) GenerateChallenge() (string, string) {
	return "dummySeed", "dummyChallenge"
}

func (d *dummyPoW) VerifyPoW(seed, proof string) bool {
	return d.validProof
}

func (d *dummyPoW) TrackRequest() {
	d.trackRequestCount++
}

func TestProcessRequestQuit(t *testing.T) {
	ctx := context.Background()
	h := NewWowHandler(&dummyWoW{}, &dummyPoW{validProof: true})
	message := &protocol.Message{Header: protocol.Quit, Payload: ""}
	msgStr := message.Stringify()
	_, err := h.ProcessRequest(ctx, msgStr, "testClient")
	require.Error(t, err)
	require.Equal(t, ErrQuit, err)
}

func TestProcessRequestInvalidMessage(t *testing.T) {
	ctx := context.Background()
	h := NewWowHandler(&dummyWoW{}, &dummyPoW{validProof: true})
	msgStr := "invalid_message_without_separator"
	_, err := h.ProcessRequest(ctx, msgStr, "testClient")
	require.Error(t, err)
	// Accept error containing either "cannot parse header" or "message doesn't match protocol"
	assert.True(t, strings.Contains(err.Error(), "cannot parse header") ||
		strings.Contains(err.Error(), "message doesn't match protocol"))
}

func TestRequestChallenge(t *testing.T) {
	ctx := context.Background()
	h := NewWowHandler(&dummyWoW{}, &dummyPoW{validProof: true})
	message := &protocol.Message{Header: protocol.RequestChallenge, Payload: ""}
	msgStr := message.Stringify()
	resp, err := h.ProcessRequest(ctx, msgStr, "testClient")
	require.NoError(t, err)
	require.Equal(t, protocol.ResponseChallenge, resp.Header)
	expectedPayload := "dummySeed|dummyChallenge"
	assert.Equal(t, expectedPayload, resp.Payload)
}

func TestRequestResourceValid(t *testing.T) {
	ctx := context.Background()
	dp := &dummyPoW{validProof: true}
	h := NewWowHandler(&dummyWoW{}, dp)
	payload := "dummySeed|anyProof"
	message := &protocol.Message{Header: protocol.RequestResource, Payload: payload}
	msgStr := message.Stringify()
	resp, err := h.ProcessRequest(ctx, msgStr, "testClient")
	require.NoError(t, err)
	require.Equal(t, protocol.ResponseResource, resp.Header)
	assert.Equal(t, "dummy wise word", resp.Payload)
	assert.Equal(t, 1, dp.trackRequestCount)
}

func TestRequestResourceInvalidPayload(t *testing.T) {
	ctx := context.Background()
	h := NewWowHandler(&dummyWoW{}, &dummyPoW{validProof: true})
	// Payload without the required separator inside (should have two parts)
	message := &protocol.Message{Header: protocol.RequestResource, Payload: "onlyOnePart"}
	msgStr := message.Stringify()
	_, err := h.ProcessRequest(ctx, msgStr, "testClient")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payload format")
}

func TestRequestResourceInvalidProof(t *testing.T) {
	ctx := context.Background()
	h := NewWowHandler(&dummyWoW{}, &dummyPoW{validProof: false})
	payload := "dummySeed|wrongProof"
	message := &protocol.Message{Header: protocol.RequestResource, Payload: payload}
	msgStr := message.Stringify()
	_, err := h.ProcessRequest(ctx, msgStr, "testClient")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid proof")
}

func TestProcessRequestContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := NewWowHandler(&dummyWoW{}, &dummyPoW{validProof: true})
	message := &protocol.Message{Header: protocol.RequestChallenge, Payload: ""}
	msgStr := message.Stringify()
	_, err := h.ProcessRequest(ctx, msgStr, "testClient")
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestProcessRequestUnknownHeader(t *testing.T) {
	ctx := context.Background()
	h := NewWowHandler(&dummyWoW{}, &dummyPoW{validProof: true})
	msgStr := "99|"
	_, err := h.ProcessRequest(ctx, msgStr, "testClient")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown message type")
}

func TestProcessRequestSlowContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	time.Sleep(20 * time.Millisecond)
	h := NewWowHandler(&dummyWoW{}, &dummyPoW{validProof: true})
	message := &protocol.Message{Header: protocol.RequestChallenge, Payload: ""}
	msgStr := message.Stringify()
	_, err := h.ProcessRequest(ctx, msgStr, "testClient")
	require.Error(t, err)
}
