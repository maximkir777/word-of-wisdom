package wow

import (
	"testing"
)

func TestService_GetRandomWiseWord(t *testing.T) {
	s := NewService()
	t.Log(s.GetRandomWiseWord())
}
