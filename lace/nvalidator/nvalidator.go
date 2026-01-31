package nvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	Config
}

type Config struct {
	playgroundValidate *validator.Validate
}

func isAlphaNumSpace(fl validator.FieldLevel) bool {
	reg := regexp.MustCompile("^[a-zA-Z0-9 ]+$")
	return reg.MatchString(fl.Field().String())
}

func New() Validator {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Using the names which have been specified for JSON representations of structs, rather than normal Go field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	validate.RegisterValidation("alphanum_space", isAlphaNumSpace)

	return Validator{Config{playgroundValidate: validate}}
}

// ValidateStruct is customized implementation of playgroundValidator.
// It collects errors and consolidates them into a single error.
func (pkg Validator) ValidateStruct(a interface{}) error {
	err := pkg.playgroundValidate.Struct(a)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		// err is not of type playgroundValidator.ValidationErrors, return original error
		return err
	}

	var errorList []error

	for _, err := range validationErrors {
		switch err.Tag() {

		case "required", "required_with", "required_without_all":
			errorList = append(errorList,
				fmt.Errorf("%s is a required field", err.Field()))

		case "max":
			errorList = append(errorList,
				fmt.Errorf("%s must be a maximum of %s in length", err.Field(), err.Param()))

		case "min":
			errorList = append(errorList,
				fmt.Errorf("%s must be a minimum of %s in length", err.Field(), err.Param()))

		case "len":
			errorList = append(errorList,
				fmt.Errorf("%s must be a %s in length", err.Field(), err.Param()))

		case "url":
			errorList = append(errorList,
				fmt.Errorf("%s must be a valid URL", err.Field()))

		case "alphanum_space":
			errorList = append(errorList,
				fmt.Errorf("%s can only contain alphabetic, numeric and space characters", err.Field()))

		case "datetime":
			if err.Param() == "2006-01-02" {
				errorList = append(errorList,
					fmt.Errorf("%s must be a valid date", err.Field()))
			} else {
				errorList = append(errorList,
					fmt.Errorf("%s must follow %s format", err.Field(), err.Param()))
			}

		case "uppercase":
			errorList = append(errorList,
				fmt.Errorf("%s must be uppercase", err.Field()))

		default:
			errorList = append(errorList,
				fmt.Errorf("something wrong on %s; %s", err.Field(), err.Tag()))
		}
	}

	return errors.Join(errorList...)
}
