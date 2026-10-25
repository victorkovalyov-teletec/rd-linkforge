package link

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/victorkovalyov-teletec/linkforge/internal/base62"
)

type Service struct {
	id    uint64
	store Store
	now   func() time.Time
}

type Option func(*Service)

var _ Shortener = (*Service)(nil)

func WithClock(t func() time.Time) Option {
	return func(s *Service) {
		s.now = t
	}
}

func NewService(st Store, opts ...Option) *Service {
	var s = &Service{id: 1, store: st, now: time.Now}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *Service) Shorten(ctx context.Context, sh ShortenRequest) (Link, error) {
	if sh.OwnerID == "" {
		return Link{}, ErrEmptyOwner
	}
	if sh.Target == "" {
		return Link{}, ErrEmptyTarget
	}

	var _, err = New(s.id, sh.Target)
	if err != nil {
		return Link{}, err
	}

	var now = s.now().UTC()

	if !sh.ExpiresAt.IsZero() && !sh.ExpiresAt.After(now) {
		return Link{}, ErrInvalidExpiry
	}

	var len = len(sh.Alias)
	if len > 0 {
		if len < 3 || len > 32 {
			return Link{}, &AliasError{Alias: sh.Alias, Err: ErrInvalidAlias}
		}

		// check for invalid symbols. we did it this way coz using of regexp is too expensive
		for _, c := range sh.Alias {
			if (c >= 'a' && c <= 'z') ||
				(c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') ||
				c == '_' || c == '-' {
				continue
			}

			return Link{}, &AliasError{Alias: sh.Alias, Err: ErrInvalidAlias}
		}

		switch strings.ToLower(sh.Alias) {
		case "api", "ws", "healthz", "readyz", "metrics", "admin":
			return Link{}, &AliasError{Alias: sh.Alias, Err: ErrReservedAlias}
		}
	}

	var link Link
	for try := 3; try > 0; {
		var e error
		var id = s.id
		var code = sh.Alias
		if code == "" {
			// empty code means we need to generate it
			code = base62.EncodeWidth(id, 6)
		}
		link, e = s.store.Create(ctx, Link{ID: id, Target: sh.Target, OwnerID: sh.OwnerID, ExpiresAt: sh.ExpiresAt.UTC(), CreatedAt: now, Code: code})
		if e != nil {
			if errors.Is(e, ErrAliasTaken) {
				if sh.Alias != "" || try == 0 {
					return Link{}, &AliasError{Alias: sh.Alias, Err: e}
				}
				s.id++
				try--
				if try > 0 {
					continue
				}
				return Link{}, &AliasError{Alias: code, Err: e}
			}
			return Link{}, fmt.Errorf("link: create: %w", e)
		}
		s.id++
		break
	}

	return link, nil
}

func (s *Service) Resolve(ctx context.Context, code string) (Link, error) {
	var link, err = s.store.ByCode(ctx, code)
	if !link.ExpiresAt.IsZero() {
		if !s.now().Before(link.ExpiresAt) {
			return link, ErrExpired
		}
	}
	return link, err
}

func (s *Service) List(ctx context.Context, owner string, page Page) ([]Link, error) {
	if owner == "" {
		return nil, ErrEmptyOwner
	}

	return s.store.ListByOwner(ctx, owner, page.Normalize())
}

func (s *Service) Delete(ctx context.Context, owner string, code string) error {
	if owner == "" {
		return ErrEmptyOwner
	}

	var link, err = s.store.ByCode(ctx, code)
	if err != nil {
		return ErrNotFound
	}

	if link.OwnerID != owner {
		return ErrForbidden
	}

	err = s.store.Delete(ctx, link.ID)

	return err
}
