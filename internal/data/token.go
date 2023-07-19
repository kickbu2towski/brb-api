package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const ScopeAuthentication = "authentication"

type Token struct {
	PlainText string `json:"token"`
	Hash []byte `json:"-"`
	Scope string `json:"-"`
	UserID string `json:"-"`
	ExpiryTime time.Time `json:"expiry_time"`
}

func NewToken(userID string, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		ExpiryTime: time.Now().Add(ttl),
		UserID: userID,
		Scope: scope,
	}

	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}

type TokenModel struct {
	Pool *pgxpool.Pool
}

func (m *TokenModel) Insert(ctx context.Context, t *Token) error {
	stmt := `INSERT INTO tokens(hash, user_id, scope, expiry_time) VALUES($1, $2, $3, $4)`
	args := []any{t.Hash, t.UserID, t.Scope, t.ExpiryTime}
	_, err := m.Pool.Exec(ctx, stmt, args...)
	return err
}

func (m *TokenModel) DeleteForUser(ctx context.Context, id, scope string) error {
	// this will logout the user from all the devices
	stmt := `DELETE FROM tokens WHERE user_id = $1 AND scope = $2`
	_, err := m.Pool.Exec(ctx, stmt, id, scope)
	return err
}
