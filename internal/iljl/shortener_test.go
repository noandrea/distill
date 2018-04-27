package iljl

import (
	"regexp"
	"testing"

	"gitlab.com/lowgroundandbigshoes/iljl/internal"
)

func TestGenerateID(t *testing.T) {

	tests := []struct {
		Alphabet      string
		Length        int
		InvalidRegexp string
	}{
		{"1234567890", 7, "[qwertyuiopasdfghjkl]"},
		{"1234567890", 5, "[qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM]"},
		{"1234567890", 4, "[qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM]"},
		{"1234567890", 20, "[qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM]"},
		{"qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM", 6, "[1234567890]"},
		{"abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789", 30, "[iIl1o0O]"},
	}
	for _, tt := range tests {
		internal.Config = internal.ConfigSchema{
			ShortID: internal.ShortIDConfig{
				Alphabet: tt.Alphabet,
				Length:   tt.Length,
			},
		}
		t.Run(tt.Alphabet, func(t *testing.T) {
			gotShortID := GenerateID()
			if len(gotShortID) != tt.Length {
				t.Errorf("GenerateID() = %v, len = %v, want %v", gotShortID, len(gotShortID), tt.Length)
			}
			m, _ := regexp.MatchString(tt.InvalidRegexp, gotShortID)
			if m {
				t.Errorf("GenerateID() = %v, match = %v, expected no match", gotShortID, tt.InvalidRegexp)
			}
		})
	}
}
