package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/maximkir777/word-of-wisdom/pkg/protocol"
)

// ErrQuit is returned when the client requests to close the connection.
var ErrQuit = errors.New("client requests to close connection")

// WoW defines the interface for the Word-of-Wisdom service.
type WoW interface {
	GetRandomWiseWord() string
}

// PoW defines the interface for Proof-of-Work operations.
type PoW interface {
	GenerateChallenge() (string, string)
	VerifyPoW(seed, proof string) bool
	TrackRequest()
}

// Impl implements the request handler.
type Impl struct {
	wow WoW
	pow PoW
}

// NewWowHandler creates a new instance of the request handler.
func NewWowHandler(wow WoW, pow PoW) *Impl {
	return &Impl{
		wow: wow,
		pow: pow,
	}
}

// ProcessRequest parses and processes the incoming message.
func (i *Impl) ProcessRequest(ctx context.Context, msgStr string, clientInfo string) (*protocol.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		msg, err := protocol.ParseMessage(msgStr)
		if err != nil {
			return nil, err
		}
		switch msg.Header {
		case protocol.Quit:
			return nil, ErrQuit
		case protocol.RequestChallenge:
			return i.requestChallenge(clientInfo)
		case protocol.RequestResource:
			return i.requestResource(msg.Payload, clientInfo)
		default:
			return nil, fmt.Errorf("unknown message type")
		}
	}
}

// requestChallenge generates a PoW challenge and returns it to the client.
func (i *Impl) requestChallenge(clientInfo string) (*protocol.Message, error) {
	slog.Info("Client requests challenge", "client", clientInfo)
	seed, challenge := i.pow.GenerateChallenge()
	msg := protocol.Message{
		Header:  protocol.ResponseChallenge,
		Payload: seed + "|" + challenge,
	}
	return &msg, nil
}

// requestResource verifies the PoW solution and returns a wise word if correct.
func (i *Impl) requestResource(payload, clientInfo string) (*protocol.Message, error) {
	slog.Info("Client requests resource", "client", clientInfo)
	parts := strings.Split(payload, "|")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid payload format")
	}
	seed, proof := parts[0], parts[1]
	if !i.pow.VerifyPoW(seed, proof) {
		return nil, fmt.Errorf("invalid proof")
	}
	// Track the successful request for dynamic difficulty adjustment
	i.pow.TrackRequest()

	// Return a random wise word
	wiseWord := i.wow.GetRandomWiseWord()
	msg := protocol.Message{
		Header:  protocol.ResponseResource,
		Payload: wiseWord,
	}
	return &msg, nil
}
