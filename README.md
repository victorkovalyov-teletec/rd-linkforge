# Linkforge

A Go URL shortener with click analytics.

The domain `Shortener` creates, resolves, lists, and deletes links. Every
transport goes through it, and storage is hidden behind the `link.Store`
interface.

Build: `go build ./...`

Run: `go run ./cmd/linkforge https://example.com`

Test: `go test ./...`
