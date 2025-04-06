package model

import (
	"fmt"
	"strings"
)

type UUID struct {
	raw [16]byte
}

func (u *UUID) LongString() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.raw[0:4], u.raw[4:6], u.raw[6:8], u.raw[8:10], u.raw[10:])
}

func (u *UUID) ShortString() string {
	return strings.ReplaceAll(u.LongString(), "-", "")
}

func (u *UUID) Bytes() []byte {
	return u.raw[:]
}

func NewUUIDFromString(s string) (*UUID, error) {
	var u UUID
	//replace all dashes with empty string
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return nil, fmt.Errorf("UUID must be 32 characters long")
	}

	for i := 0; i < 16; i++ {
		//parse 2 characters from the string
		b, err := fmt.Sscanf(s[i*2:i*2+2], "%x", &u.raw[i])
		if err != nil {
			return nil, fmt.Errorf("UUID must be 32 characters long")
		}
		if b != 1 {
			return nil, fmt.Errorf("UUID must be 32 characters long")
		}
	}
	return &u, nil
}

func NewUUIDFromBytes(b []byte) (*UUID, error) {
	if len(b) != 16 {
		return nil, fmt.Errorf("UUID must be 16 bytes long")
	}
	var u UUID
	copy(u.raw[:], b)
	return &u, nil
}
