package gqlmodel

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

func MarshalUidScalar(b uuid.UUID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(b.String()))
	})
}

// It will receive the value from the user and we can validate it before sending to the func
// This function will be execute only if type is defnied NON NULL i.e UID!
func UnmarshalUidScalar(v interface{}) (uuid.UUID, error) {
	val, err := uuid.Parse(v.(string))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("wrong uuid format")
	}

	return val, nil
}
