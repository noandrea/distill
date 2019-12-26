package common

import (
	"bufio"
	"fmt"
	"os"
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

// DefaultIfEmptyStr set a default for a string if it is nulled
func DefaultIfEmptyStr(s *string, defaultS string) {
	if IsEmptyStr(*s) {
		*s = defaultS
	}
}

// DefaultIfEmptyInt set the value of an int to a default if it is nulled (0)
func DefaultIfEmptyInt(v *int, defaultV int) {
	if *v <= 0 {
		*v = defaultV
	}
}

// DefaultIfEmptyUint64 set the value of an int to a default if it is nulled (0)
func DefaultIfEmptyUint64(v *uint64, defaultV uint64) {
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

// GenerateSecret generate a string that can be used as secrete api key
func GenerateSecret() string {
	a := "asdfghjklqwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890#!-"
	l := 50
	secret, _ := RandomString(a, l)
	return secret
}

// AskYes prompt a yes/no question to the prompt
func AskYes(question string, defaultYes bool) (isYes bool) {
	fmt.Print(question)
	if defaultYes {
		fmt.Print(" [yes]: ")
	} else {
		fmt.Print(" [no]: ")
	}
	reader := bufio.NewReader(os.Stdin)
	reply, _ := reader.ReadString('\n')
	DefaultIfEmptyStr(&reply, "yes")
	if IsEqStr(reply, "yes") {
		return true
	}
	return
}
