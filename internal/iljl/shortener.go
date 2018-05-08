package iljl

import (
	gonanoid "github.com/matoous/go-nanoid"
	"gitlab.com/lowgroundandbigshoes/iljl/internal"
)

// generateID generates a new id
func generateID() (shortID string) {
	a := internal.Config.ShortID.Alphabet
	l := internal.Config.ShortID.Length
	// a and l are validated before
	shortID, _ = gonanoid.Generate(a, l)
	return
}
