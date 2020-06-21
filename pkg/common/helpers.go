package common

import (
	"encoding/binary"
	"fmt"
	"strings"

	gonanoid "github.com/matoous/go-nanoid"
)

// IsEqStr tells if two strings a and b are equals after trimming spaces and lowercasing
func IsEqStr(a, b string) bool {
	return strings.ToLower(strings.TrimSpace(a)) == strings.ToLower(strings.TrimSpace(b))
}

// IsEmptyStr tells if a string is empty or not
func IsEmptyStr(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IfEmptyThen check if a string is empty and return the default if it is
func IfEmptyThen(s, thenDefault string) string {
	if IsEmptyStr(s) {
		return thenDefault
	}
	return s
}

// DefaultIfEmptyInt set the value of an int to a default if it is nulled (0)
func DefaultIfEmptyInt(v *int, defaultV int) {
	if *v <= 0 {
		*v = defaultV
	}
}

// RandomString generate a random string of required length using alphabet
func RandomString(alphabet string, length int) (s string, err error) {
	if IsEmptyStr(alphabet) {
		err = fmt.Errorf("alphabet must not be empty")
		return
	}
	if length <= 0 {
		err = fmt.Errorf("string length must be longer than 0")
		return
	}
	return gonanoid.Generate(alphabet, length)
}

//   ____  ____  ________  _____ 4   _______  ________  _______     ______
//  |_   ||   _||_   __  ||_   _|   |_   __ \|_   __  ||_   __ \  .' ____ \
//    | |__| |    | |_ \_|  | |       | |__) | | |_ \_|  | |__) | | (___ \_|
//    |  __  |    |  _| _   | |   _   |  ___/  |  _| _   |  __ /   _.____`.
//   _| |  | |_  _| |__/ | _| |__/ | _| |_    _| |__/ | _| |  \ \_| \____) |
//  |____||____||________||________||_____|  |________||____| |___|\______.'
//

// Itoa byte array to uint64
func Itoa(i uint64) (b []byte) {
	b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return
}

// Atoi byte array to uint64
func Atoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// Min return the min betweek numbers
func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
