package preview

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/fastygo/cms/internal/domain/content"
)

type Token string

type Access struct {
	Token       Token
	EntryID     content.ID
	PrincipalID string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

func NewToken() (Token, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return Token(hex.EncodeToString(raw[:])), nil
}

func (a Access) ValidAt(now time.Time) bool {
	return a.Token != "" && now.Before(a.ExpiresAt)
}
