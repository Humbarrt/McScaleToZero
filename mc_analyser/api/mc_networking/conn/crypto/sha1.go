package crypto

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"strings"
)

type SHA1 struct {
	hex string
}

func (s *SHA1) Update(input string) {
	h := sha1.New()
	io.WriteString(h, input)
	hash := h.Sum(nil)

	// Check for negative hashes
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = twosComplement(hash)
	}

	// Trim away zeroes
	res := strings.TrimLeft(hex.EncodeToString(hash), "0")
	if negative {
		res = "-" + res
	}

	s.hex = res
}

func twosComplement(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = byte(^p[i])
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
	return p
}

func NewSHA1() *SHA1 {
	return &SHA1{}
}
