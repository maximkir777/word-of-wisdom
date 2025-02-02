package client

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/maximkir777/word-of-wisdom/pkg/protocol"
)

// Client represents the TCP client that continuously sends requests.
type Client struct {
	ServerAddress string
	Timeout       time.Duration
}

// NewClient creates a new Client instance.
func NewClient(addr string, timeout time.Duration) *Client {
	return &Client{
		ServerAddress: addr,
		Timeout:       timeout,
	}
}

// Run starts the client loop with dynamic worker pools and graceful shutdown.
func (c *Client) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Received shutdown signal, stopping client...")
		cancel()
	}()

	// Main loop: repeatedly spawn batches of workers.
	for {
		select {
		case <-ctx.Done():
			log.Println("Client shutting down.")
			return
		default:
		}

		// Use crypto/rand to determine number of workers (range 10 to 24).
		n, err := rand.Int(rand.Reader, big.NewInt(15))
		var numWorkers int
		if err != nil {
			numWorkers = 5 // fallback
		} else {
			numWorkers = int(n.Int64()) + 10
		}
		log.Printf("Starting batch with %d workers", numWorkers)

		var wg sync.WaitGroup
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				if err := c.singleRequest(ctx, workerID); err != nil {
					log.Printf("Worker %d error: %v", workerID, err)
				}
			}(i)
		}
		wg.Wait()

		// Pause before starting the next batch.
		select {
		case <-ctx.Done():
			log.Println("Client shutting down.")
			return
		case <-time.After(2 * time.Second):
		}
	}
}

// singleRequest performs one cycle of connection, POW challenge request, solution and resource retrieval.
func (c *Client) singleRequest(ctx context.Context, workerID int) error {
	// Dial with a timeout.
	conn, err := net.DialTimeout("tcp", c.ServerAddress, c.Timeout)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Set a deadline for reading the challenge response.
	if err := conn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	// Send RequestChallenge message.
	challengeReq := &protocol.Message{
		Header:  protocol.RequestChallenge,
		Payload: "",
	}
	_, err = fmt.Fprintf(conn, "%s\n", challengeReq.Stringify())
	if err != nil {
		return fmt.Errorf("failed to send challenge request: %w", err)
	}

	// Read challenge response.
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read challenge response: %w", err)
	}
	challengeResp, err := protocol.ParseMessage(line)
	if err != nil {
		return fmt.Errorf("failed to parse challenge response: %w", err)
	}
	if challengeResp.Header != protocol.ResponseChallenge {
		return fmt.Errorf("unexpected response header: %d", challengeResp.Header)
	}

	// Split payload into two parts: fullSeed and challenge.
	// Expected format: "<difficulty,randomNumber>|<challenge>"
	parts := strings.SplitN(challengeResp.Payload, "|", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid challenge payload: %s", challengeResp.Payload)
	}
	fullSeed := parts[0]  // e.g., "3,174644643503504791"
	challenge := parts[1] // e.g., "000"
	log.Printf("Worker %d: Received challenge (seed=%s, challenge=%s)", workerID, fullSeed, challenge)

	// Remove the deadline so that POW solving is not affected.
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return fmt.Errorf("failed to reset deadline: %w", err)
	}

	// Create a timeout context for solving POW (e.g., 10 seconds).
	powCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	proof, err := solvePoW(powCtx, fullSeed, challenge)
	if err != nil {
		return fmt.Errorf("failed to solve POW: %w", err)
	}
	log.Printf("Worker %d: Proof found: %s", workerID, proof)

	// Set a new deadline for reading the resource response.
	if err := conn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return fmt.Errorf("failed to set deadline for resource response: %w", err)
	}

	// Send RequestResource message.
	resourcePayload := fmt.Sprintf("%s|%s", fullSeed, proof)
	resourceReq := &protocol.Message{
		Header:  protocol.RequestResource,
		Payload: resourcePayload,
	}
	_, err = fmt.Fprintf(conn, "%s\n", resourceReq.Stringify())
	if err != nil {
		return fmt.Errorf("failed to send resource request: %w", err)
	}

	// Read resource response.
	line, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read resource response: %w", err)
	}
	resourceResp, err := protocol.ParseMessage(line)
	if err != nil {
		return fmt.Errorf("failed to parse resource response: %w", err)
	}
	if resourceResp.Header != protocol.ResponseResource {
		return fmt.Errorf("unexpected resource response header: %d", resourceResp.Header)
	}
	log.Printf("Worker %d: Wise word: %s", workerID, resourceResp.Payload)
	return nil
}

// solvePoW finds a proof such that sha256(seed + "|" + proof) starts with the challenge prefix.
// It respects the provided context timeout.
func solvePoW(ctx context.Context, seed, challenge string) (string, error) {
	var proof int64 = 0
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			proofStr := strconv.FormatInt(proof, 10)
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(seed+"|"+proofStr)))
			if strings.HasPrefix(hash, challenge) {
				return proofStr, nil
			}
			proof++
		}
	}
}
