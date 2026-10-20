package store

import (
	"context"
	"errors"
	"sort"

	"github.com/victorkovalyov-teletec/linkforge/internal/link"
)

var (
	ErrDuplicateID = errors.New("store: duplicate id")
)

type Store struct {
	byID map[uint64]link.Link
}

var _ link.Store = (*Store)(nil)

func New() *Store {
	return &Store{
		byID: make(map[uint64]link.Link),
	}
}

func (s *Store) Create(ctx context.Context, lnk link.Link) (link.Link, error) {
	var _, found = s.byID[lnk.ID]
	if found {
		return link.Link{}, ErrDuplicateID
	}

	var _, err = s.ByCode(ctx, lnk.Code)
	if err == nil {
		return link.Link{}, link.ErrAliasTaken
	}

	s.byID[lnk.ID] = lnk

	return lnk, nil
}

func (s *Store) ByCode(ctx context.Context, code string) (link.Link, error) {
	for _, lnk := range s.byID {
		if code == lnk.Code {
			return lnk, nil
		}
	}
	return link.Link{}, link.ErrNotFound
}

func (s *Store) ListByOwner(ctx context.Context, owner string, page link.Page) ([]link.Link, error) {
	page = page.Normalize()

	afterID, err := page.AfterID()
	if err != nil {
		return nil, err
	}

	links := make([]link.Link, 0, page.Limit)

	for _, lnk := range s.byID {
		if lnk.OwnerID != owner || lnk.ID <= afterID {
			continue
		}

		links = append(links, lnk)
	}

	sort.Slice(links, func(i, j int) bool {
		return links[i].ID < links[j].ID
	})

	if len(links) > page.Limit {
		links = links[:page.Limit]
	}

	return links, nil
}

func (s *Store) Delete(ctx context.Context, id uint64) error {
	var _, found = s.byID[id]
	if !found {
		return link.ErrNotFound
	}

	delete(s.byID, id)

	return nil
}

func (s *Store) Get(id uint64) (link.Link, bool) {
	l, ok := s.byID[id]

	return l, ok
}

func (s *Store) All() []link.Link {
	links := make([]link.Link, 0, len(s.byID))

	for _, l := range s.byID {
		links = append(links, l)
	}

	sort.Slice(links, func(i, j int) bool {
		return links[i].ID < links[j].ID
	})

	return links
}

func (s *Store) Count() int {
	return len(s.byID)
}
