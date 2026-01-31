package jsonslice

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

type JsonSlice []interface{}

// need for gqlgen
func MarshalJsonSlice(f JsonSlice) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		byt, err := json.Marshal(f)
		if err != nil {
			panic(err)
		}
		w.Write(byt)
	})
}

// need for gqlgen
func UnmarshalJsonSlice(v any) (JsonSlice, error) {
	byt, err := json.Marshal(v)

	if err != nil {
		return JsonSlice{}, fmt.Errorf("%T is not an JsonSlice", v)
	}

	fmt.Println(byt)

	var result JsonSlice
	err = json.Unmarshal(byt, &result)

	if err != nil {
		return JsonSlice{}, fmt.Errorf("%T cannot unmarshal", v)
	}

	return result, nil
}
