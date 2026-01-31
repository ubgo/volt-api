package utils

import (
	"errors"
	"lace/jsontype"
	"unicode/utf8"

	"github.com/cohesivestack/valgo"
)

func Ptr[T any](v T) *T {
	return &v
}

func ValgoFirstError(val *valgo.Validation) error {
	valErrors := val.Error().(*valgo.Error).Errors()
	for _, value := range valErrors {
		return errors.New(value.Messages()[0])
	}

	return nil
}

func ShortText(s string, i int) string {
	if len(s) < i {
		return s
	}

	if utf8.ValidString(s[:i]) {
		return s[:i]
	}
	return s[:i+1]

}

func ToJSON(v any) *jsontype.JSON {
	json, err := jsontype.ToJson(v)
	if err != nil {
		return nil
	}
	return json
}

// StrPtr returns a pointer to the string, or nil if the string is empty.
func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
