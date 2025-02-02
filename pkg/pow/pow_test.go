package pow

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoW_ChallengeGeneration(t *testing.T) {
	p := NewPoW(2, 5, 100, 5*time.Minute)
	seed, challenge := p.GenerateChallenge()

	parts := strings.Split(seed, ",")
	require.Len(t, parts, 2, "Seed should contain two parts separated by a comma")
	_, err := strconv.Atoi(parts[0])
	assert.NoError(t, err, "First part of seed should be numeric")

	_, err = strconv.Atoi(challenge)
	assert.NoError(t, err, "Challenge should be numeric")
	assert.Len(t, challenge, 2, "Challenge length should match base difficulty")
}

func TestPoW_VerificationProcess(t *testing.T) {
	p := NewPoW(3, 5, 100, 5*time.Minute)
	seed, _ := p.GenerateChallenge()

	t.Run("valid proof", func(t *testing.T) {
		proof := solvePoW(seed)
		assert.True(t, p.VerifyPoW(seed, proof), "Valid proof should pass verification")
	})

	t.Run("invalid proof", func(t *testing.T) {
		assert.False(t, p.VerifyPoW(seed, "wrong"), "Invalid proof should fail verification")
	})

	t.Run("malformed seed", func(t *testing.T) {
		assert.False(t, p.VerifyPoW("invalid-seed", "123"), "Malformed seed should fail verification")
	})
}

func TestPoW_DifficultyAdjustment(t *testing.T) {
	p := NewPoW(2, 5, 10, 1*time.Minute)
	p.StartDifficultyAdjuster()

	// Initial state
	assert.Equal(t, 2, p.currentDifficulty, "Initial difficulty should match base")

	// Simulate load
	for i := 0; i < 15; i++ {
		p.TrackRequest()
	}

	// Force immediate adjustment
	p.adjustDifficulty()

	t.Run("difficulty increases under load", func(t *testing.T) {
		assert.GreaterOrEqual(t, p.currentDifficulty, 3, "Difficulty should increase under load")
	})

	t.Run("max difficulty limit", func(t *testing.T) {
		p.currentDifficulty = 5
		for i := 0; i < 20; i++ {
			p.TrackRequest()
		}
		p.adjustDifficulty()
		assert.Equal(t, 5, p.currentDifficulty, "Should not exceed max difficulty")
	})
}

func TestPoW_ConcurrentSafety(t *testing.T) {
	p := NewPoW(5, 10, 100, 5*time.Minute)

	for i := 0; i < 100; i++ {
		go func() {
			p.GenerateChallenge()
			p.TrackRequest()
			p.adjustDifficulty()
		}()
	}
}

func solvePoW(seed string) string {
	parts := strings.Split(seed, ",")
	if len(parts) != 2 {
		return ""
	}
	difficulty, err := strconv.Atoi(parts[0])
	if err != nil {
		return ""
	}
	var proof int
	for {
		proofStr := strconv.Itoa(proof)
		hash := sha256.Sum256([]byte(seed + "|" + proofStr))
		if fmt.Sprintf("%x", hash)[:difficulty] == fmt.Sprintf("%0*d", difficulty, 0) {
			return proofStr
		}
		proof++
	}
}
