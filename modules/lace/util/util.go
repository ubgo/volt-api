// Package jutil provide some common sets of shared function
package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lace/jsontype"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func Ptr[T any](v T) *T {
	return &v
}

func ToJSON(v any) *jsontype.JSON {
	json, err := jsontype.ToJson(v)
	if err != nil {
		return nil
	}
	return json
}

// Throw Malfunction error for Gin Request
func GinErrorMalfunction(c *gin.Context) {
	err := &ResponseError{
		Message: "Malfunctioned request.",
	}

	err.Send(c, http.StatusUnauthorized)
	// c.JSON(http.StatusBadRequest, gin.H{"error": error})
}

// Throw Bad Request error for Gin Request
func GinErrorBadRequest(c *gin.Context) {
	err := &ResponseError{
		Message: "Bad request.",
	}

	err.Send(c, http.StatusUnauthorized)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, x-captcha-res, X-CAPTCHA-BYPASS, x-auth-token, token, Website, x-app, openai-organization, x-workspace-id, x-internal-key, x-aiuser-token, x-guest-id")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
func IsCli() bool {
	return os.Getenv("appmode") == "cli"
}

func SetCliAppMode() {
	err := os.Setenv("appmode", "cli")
	fmt.Println(err)
}

func InterfaceToStruct(in, out interface{}) error {
	buf := new(bytes.Buffer)

	err := json.NewEncoder(buf).Encode(in)
	if err != nil {
		return err
	}

	err = json.NewDecoder(buf).Decode(out)
	if err != nil {
		return err
	}

	return nil
}

// Convert map[string]interface{} to bytes for JSON.rawmessage used in dblogs table data
func MustMarshalData(value interface{}) []byte {
	s, _ := json.Marshal(value)
	return s
}

func InterfaceUnmarshal(in, out interface{}) error {
	dbByte, err := json.Marshal(in)
	if err != nil {
		return err
	}

	err = json.Unmarshal(dbByte, out)
	if err != nil {
		return err
	}

	return nil
}

func GetTotalPages(totalCount, limit int) int {
	return int(math.Ceil(float64(totalCount) / float64(limit)))
}

type Name struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func ParseName(fullName string) Name {
	nameParts := strings.Split(strings.TrimSpace(fullName), " ")

	var name Name
	if len(nameParts) > 0 {
		name.FirstName = nameParts[0]
	}

	if len(nameParts) > 1 {
		name.LastName = nameParts[1]
	}

	return name
}

// todo_5lwjnp2kas will return todo used in graphql detect Noder type
func GetNodeType(s string) string {
	// Find the index of the first
	index := strings.Index(s, "_")

	// If there's no, return the original string
	if index == -1 {
		return ""
	}

	// Return the substring before the first underscore
	return s[:index]
}

func SizeUnitToHumanReadable(sizeInBytes int64, unit string) string {
	if unit == "kb" {
		return fmt.Sprintf("%.2f KB", float64(sizeInBytes)/1000)
	}

	if unit == "mb" {
		return fmt.Sprintf("%.2f MB", float64(sizeInBytes)/1000/1000)
	}

	return fmt.Sprintf("%.2f GB", float64(sizeInBytes)/1000/1000/1000)
}

// FileNameWithoutExtension returns filename without extension
func FileNameWithoutExtension(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}
