package pow

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PoW struct {
	baseDifficulty    int
	maxDifficulty     int
	currentDifficulty int
	mu                sync.RWMutex
	loadWindow        []time.Time
	windowSize        int
	windowDuration    time.Duration
}

func NewPoW(base, max int, windowSize int, windowDuration time.Duration) *PoW {
	return &PoW{
		baseDifficulty:    base,
		maxDifficulty:     max,
		currentDifficulty: base,
		windowSize:        windowSize,
		windowDuration:    windowDuration,
		loadWindow:        make([]time.Time, 0, windowSize),
	}
}

func (p *PoW) StartDifficultyAdjuster() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			p.adjustDifficulty()
		}
	}()
}

func (p *PoW) adjustDifficulty() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// clean up old load data
	now := time.Now().UTC()
	threshold := now.Add(-p.windowDuration)

	start := 0
	for ; start < len(p.loadWindow); start++ {
		if p.loadWindow[start].After(threshold) {
			break
		}
	}
	p.loadWindow = p.loadWindow[start:]

	// calculate new difficulty
	requestRate := len(p.loadWindow) * int(time.Hour/p.windowDuration)
	newDiff := p.baseDifficulty + requestRate/10

	if newDiff > p.maxDifficulty {
		newDiff = p.maxDifficulty
	} else if newDiff < p.baseDifficulty {
		newDiff = p.baseDifficulty
	}

	p.currentDifficulty = newDiff
}

func (p *PoW) TrackRequest() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.loadWindow) >= p.windowSize {
		p.loadWindow = p.loadWindow[1:]
	}
	p.loadWindow = append(p.loadWindow, time.Now().UTC())
}

func (p *PoW) GenerateChallenge() (string, string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	max := new(big.Int).Lsh(big.NewInt(1), 63) // max = 2^63
	n, err := rand.Int(rand.Reader, max)
	var randVal int64
	if err != nil {
		randVal = 0
	} else {
		randVal = n.Int64()
	}

	seed := fmt.Sprintf("%d|%d", p.currentDifficulty, randVal)

	challenge := fmt.Sprintf("%0*d", p.currentDifficulty, 0)
	return seed, challenge
}

func (p *PoW) VerifyPoW(seed, proof string) bool {
	parts := strings.Split(seed, "|")
	if len(parts) != 2 {
		return false
	}
	difficulty, err := strconv.Atoi(parts[0])
	if err != nil || difficulty < 1 {
		return false
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(seed+proof)))
	target := fmt.Sprintf("%0*d", difficulty, 0)
	return strings.HasPrefix(hash, target)
}
