package base62

import (
	"errors"
	"math"
	"slices"
	"strings"
)

const codecAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var (
	ErrEmpty            = errors.New("base62: empty code")
	ErrInvalidCharacter = errors.New("base62: invalid character")
	ErrOverflow         = errors.New("base62: code is too long")
)

// Encode renders a number with no padding.
func Encode(v uint64) string {
	var res []byte

	for {
		res = append(res, codecAlphabet[v%62])
		v /= 62
		if v == 0 {
			break
		}
	}

	// Remainders are collected least-significant first, so reverse them before returning the code.
	slices.Reverse(res)

	return string(res)
}

func encodeRecursiveHelper(v uint64, b *strings.Builder) {
	if v >= 62 {
		encodeRecursiveHelper(v/62, b)
	}

	b.WriteByte(codecAlphabet[v%62])
}

func EncodeRecursive(v uint64) string {
	var bld strings.Builder
	bld.Grow(12)
	encodeRecursiveHelper(v, &bld)
	return bld.String()
}

// EncodeWidth does the same but left-pads with zeroes to at least width characters. The value itself is never truncated.
func EncodeWidth(v uint64, width int) string {
	s := Encode(v)

	len := len(s)

	if width <= len {
		return s
	}

	var pad = width - len

	s = strings.Repeat("0", pad) + s

	return s
}

// Decode parses a code back into a number.
func Decode(s string) (uint64, error) {
	if s == "" {
		return 0, ErrEmpty
	}

	var efflen uint64 = 0
	for _, r := range s {
		if !strings.ContainsRune(codecAlphabet, r) {
			return 0, ErrInvalidCharacter
		}
		efflen++
	}

	if efflen > 11 {
		return 0, ErrOverflow
	}

	var v uint64 = 0
	for _, r := range s {
		idx := strings.IndexRune(codecAlphabet, r)

		digit := uint64(idx)
		if efflen == 11 {
			const maxPrefix = math.MaxUint64 / 62
			const maxDigit = math.MaxUint64 % 62

			if v > maxPrefix || (v == maxPrefix && digit > maxDigit) {
				return 0, ErrOverflow
			}
		}
		v = v*62 + digit
	}

	return v, nil
}
