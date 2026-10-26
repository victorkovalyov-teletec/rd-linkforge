package main

import (
	"context"
	"fmt"
	"os"

	"github.com/victorkovalyov-teletec/linkforge/internal/link"
	"github.com/victorkovalyov-teletec/linkforge/internal/store"
)

func main() {
	ctx := context.Background()

	s := store.New()

	srv := link.NewService(s)

	for _, target := range os.Args[1:] {
		_, err := srv.Shorten(ctx, link.ShortenRequest{
			Target:  target,
			OwnerID: "demo",
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", target, err)
			continue
		}
	}

	links := s.All()
	for _, l := range links {
		fmt.Printf("%s %s\n", l.Code, l.Target)
	}

	fmt.Printf("\nTotal: %d\n", s.Count())
}
