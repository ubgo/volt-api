package natshandlers

type Code string

func (e Code) ToString() string {
	return string(e)
}

// Common error codes
var (
	ErrCodeJSONUnmarshal            Code = "json/unmarshal"
	ErrCodeDatabaseUnexpected       Code = "db/unexpected"
	ErrCodeUnauthorized             Code = "auth/unauthorized"
	ErrCodeValidationFailed         Code = "validation/failed"
	ErrCodeInternalServerUnexpected Code = "server/unexpected"
	ErrCodeDatabaseNotFound         Code = "db/record(s)-not-found"
	ErrCodeURIParse                 Code = "url/uri-parse"
)
