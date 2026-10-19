package store

import (
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

func New() Store {
	return Store{
		byID: make(map[uint64]link.Link),
	}
}

func Add(s Store, l link.Link) error {
	if _, exists := s.byID[l.ID]; exists {
		return ErrDuplicateID
	}

	s.byID[l.ID] = l

	return nil
}

func Get(s Store, id uint64) (link.Link, bool) {
	l, ok := s.byID[id]

	return l, ok
}

func All(s Store) []link.Link {
	links := make([]link.Link, 0, len(s.byID))

	for _, l := range s.byID {
		links = append(links, l)
	}

	sort.Slice(links, func(i, j int) bool {
		return links[i].ID < links[j].ID
	})

	return links
}

func Count(s Store) int {
	return len(s.byID)
}
