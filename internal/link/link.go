package link

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/victorkovalyov-teletec/linkforge/internal/base62"
)

type Link struct {
	ID        uint64
	Code      string
	Target    string
	OwnerID   string
	CreatedAt time.Time
	ExpiresAt time.Time // нульове значення: ніколи не спливає
}

type ShortenRequest struct {
	Target    string
	Alias     string // порожнє означає: згенерувати код
	OwnerID   string
	ExpiresAt time.Time
}

type Page struct {
	Cursor string
	Limit  int
}

type AliasError struct {
	Alias string
	Err   error
}

func (e *AliasError) Error() string {
	return fmt.Errorf("alias %q: %v", e.Alias, e.Err).Error()
}

func (e *AliasError) Unwrap() error {
	return e.Err
}

var (
	ErrEmptyTarget       = errors.New("link: empty target")
	ErrUnsupportedScheme = errors.New("link: unsupported scheme")
)

var (
	ErrNotFound      = errors.New("link: not found")
	ErrAliasTaken    = errors.New("link: alias already taken")
	ErrInvalidAlias  = errors.New("link: invalid alias")
	ErrReservedAlias = errors.New("link: reserved alias")
	ErrForbidden     = errors.New("link: forbidden")
	ErrExpired       = errors.New("link: expired")
	ErrInvalidExpiry = errors.New("link: invalid expiry")
	ErrEmptyOwner    = errors.New("link: owner is required")
	ErrInvalidCursor = errors.New("link: invalid cursor")
)

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

func (p Page) Normalize() Page {
	var cp = p
	if p.Limit > MaxPageLimit {
		cp.Limit = MaxPageLimit
	}
	if p.Limit <= 0 {
		cp.Limit = DefaultPageLimit
	}

	return cp
}

func (p Page) AfterID() (uint64, error) {
	if p.Cursor == "" {
		return 0, nil
	}

	var v, e = base62.Decode(p.Cursor)
	if e != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidCursor, e)
	}

	return v, nil
}

// New builds a link, validating the target address.
func New(id uint64, target string) (Link, error) {
	if target == "" {
		return Link{}, ErrEmptyTarget
	}

	if !strings.HasPrefix(target, "http://") {
		if !strings.HasPrefix(target, "https://") {
			return Link{}, ErrUnsupportedScheme
		}
	}

	return Link{
		ID:        id,
		Code:      base62.EncodeWidth(id, 6),
		Target:    target,
		CreatedAt: time.Now().UTC(),
	}, nil
}
