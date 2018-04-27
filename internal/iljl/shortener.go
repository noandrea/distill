package iljl

import (
	"github.com/jbrodriguez/mlog"
	gonanoid "github.com/matoous/go-nanoid"
	"gitlab.com/lowgroundandbigshoes/iljl/internal"
)

// GenerateID generates a new id
func GenerateID() (shortID string) {
	a := internal.Config.ShortID.Alphabet
	l := internal.Config.ShortID.Length
	shortID, err := gonanoid.Generate(a, l)
	if err != nil {
		mlog.Error(err)
	}
	return
}
