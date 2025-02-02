package wow

import (
	"crypto/rand"
	"math/big"
)

// Service provides random wise words.
type Service struct {
	words []string
}

func NewService() *Service {
	wiseWords := []string{
		"Every cloud has a silver lining",
		"Actions speak louder than words",
		"Turn over a new leaf",
		"The early bird catches the worm",
		"Rome wasn't built in a day",
		"Kill two birds with one stone",
		"Where there's smoke, there's fire",
		"Don't put all your eggs in one basket",
		"Burn the midnight oil",
		"When in Rome, do as the Romans do",
		"Bite the bullet",
		"Break the ice",
		"Let bygones be bygones",
		"Think outside the box",
		"A picture is worth a thousand words",
		"Keep your chin up",
		"The ball is in your court",
		"Cross that bridge when you come to it",
		"Better late than never",
		"You reap what you sow",
	}
	return &Service{words: wiseWords}
}

// GetRandomWiseWord returns a random wise word using crypto/rand.
func (s *Service) GetRandomWiseWord() string {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(s.words))))
	if err != nil {
		return s.words[0]
	}
	return s.words[n.Int64()]
}
