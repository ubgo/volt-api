package util

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cohesivestack/valgo"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Response ...
type Response struct {
	Message    string      `json:"message,omitempty"`
	StatusCode int         `json:"statusCode,omitempty"`
	Code       string      `json:"code,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

// ResponseError ...
type ResponseError struct {
	Message    string      `json:"message,omitempty"`
	StatusCode int         `json:"statusCode,omitempty"`
	Code       string      `json:"code,omitempty"`
	Errors     interface{} `json:"errors,omitempty"`
}

// Send ResponseError to HTTP in JSON format
// func (r ResponseError) Send(c *gin.Context, code int) {
// 	c.JSON(code, r)
// }

func (r ResponseError) Send(c *gin.Context, code int) {
	c.JSON(code, gin.H{
		"error": r,
	})
}

func SendResponseError(c *gin.Context, message string, code string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": &ResponseError{
			Message: message,
			Code:    code,
		},
	})
}

// convert the validation errors into the error list so we can return multiple errors on frontend
func ErrorToGqlError(err error, ctx context.Context) error {
	errList := gqlerror.List{}

	switch err := err.(type) {
	case *valgo.Error:
		for _, e := range err.Errors() {
			for _, msg := range e.Messages() {
				errList = append(errList, &gqlerror.Error{
					Path:    graphql.GetPath(ctx),
					Message: msg,
					Extensions: map[string]interface{}{
						"field": e.Name(),
					},
				})
			}
		}
	default:
		errList = append(errList, &gqlerror.Error{
			Message: err.Error(),
		})
	}

	return errList
}
