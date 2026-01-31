package publicid

import (
	"strings"

	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pkg/errors"
)

// Ref: https://planetscale.com/blog/why-we-chose-nanoids-for-planetscales-api
// Fixed nanoid parameters used in the Rails application.
const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	length   = 24
)

// New generates a unique public ID.
func New() (string, error) { return nanoid.Generate(alphabet, length) }

// Must is the same as New, but panics on error.
func Must() string { return "a" + nanoid.MustGenerate(alphabet, length) }

func Must12() string { return "a" + nanoid.MustGenerate(alphabet, 12) }

func Must14() string { return "a" + nanoid.MustGenerate(alphabet, 14) }

func Must7() string { return "a" + nanoid.MustGenerate(alphabet, 7) }

// Validate checks if a given field name's public ID value is valid according to
// the constraints defined by package publicid.
func Validate(fieldName, id string) error {
	if id == "" {
		return errors.Errorf("%s cannot be blank", fieldName)
	}

	if len(id) != length {
		return errors.Errorf("%s should be %d characters long", fieldName, length)
	}

	if strings.Trim(id, alphabet) != "" {
		return errors.Errorf("%s has invalid characters", fieldName)
	}

	return nil
}

func GetAlphabet() string {
	return alphabet
}
