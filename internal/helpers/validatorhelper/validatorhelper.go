package validatorhelper

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

var Validator *validator.Validate

func init() {
	Validator = validator.New(validator.WithRequiredStructEnabled())
	RegisterOptionalValidation(Validator)
	RegisterNonNullValidation(Validator)
	RegisterAlarmTypeValidation(Validator)
}

// ValidationErrors represents a collection of validation errors for an HTTP request.
type ValidationErrors struct {
	Errors map[string][]string `json:"errors,omitempty"`
}

func (ve *ValidationErrors) Error() string {
	if ve == nil || len(ve.Errors) == 0 {
		return ""
	}

	var errMsg string
	for field, errs := range ve.Errors {
		for _, err := range errs {
			errMsg += field + ": " + err + "; "
		}
	}
	return errMsg
}

func FormatValidationErrors(errs validator.ValidationErrors) string {
	var messages []string

	for _, err := range errs {
		field := strings.ToLower(err.Field())

		switch err.Tag() {
		case "opt":
			messages = append(messages,
				fmt.Sprintf("%s: %s", field, err.Error()))
		case "email":
			messages = append(messages,
				fmt.Sprintf("%s must be a valid email", field))
		case "min":
			messages = append(messages,
				fmt.Sprintf("%s must be at least %s characters", field, err.Param()))
		case "max":
			messages = append(messages,
				fmt.Sprintf("%s must be at most %s characters", field, err.Param()))
		case "gte":
			messages = append(messages,
				fmt.Sprintf("%s must be >= %s", field, err.Param()))
		case "lte":
			messages = append(messages,
				fmt.Sprintf("%s must be <= %s", field, err.Param()))
		case "oneof":
			messages = append(messages,
				fmt.Sprintf("%s must be one of: %s", field, err.Param()))
		case "nonull":
			messages = append(messages,
				fmt.Sprintf("%s cannot be null", field))
		default:
			messages = append(messages,
				fmt.Sprintf("%s failed validation: %s", field, err.Tag()))
		}
	}

	return strings.Join(messages, "; ")
}

// IsRequestValid validates the provided request struct using the go-playground/validator package.
// It returns a ValidationErrors instance if validation fails, or nil if the request is valid.
func IsRequestValid(request any) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(request)
	if err != nil {
		return err
	}
	return nil
}

// Interface that all Optional types implement
type OptionalAny interface {
	IsSet() bool
	IsNull() bool
	Any() (interface{}, bool) // Returns the value as interface{}
}

func RegisterNonNullValidation(v *validator.Validate) {
	err := v.RegisterValidation("nonull", func(fl validator.FieldLevel) bool {
		optAny, ok := fl.Field().Interface().(OptionalAny)
		if !ok {
			return true
		}

		// If the field is set and null, that's invalid
		return !optAny.IsNull()
	})

	if err != nil {
		slog.Error("failed to register nonull validation", "error", err)
	}
}

func RegisterOptionalValidation(v *validator.Validate) {
	err := v.RegisterValidation("opt", func(fl validator.FieldLevel) bool {
		// Check if this field implements OptionalAny
		optAny, ok := fl.Field().Interface().(OptionalAny)
		if !ok {
			// Not an Optional type, skip validation
			return true
		}

		// If unset or null, validation passes
		// (use "required" tag if you want to forbid unset)
		if !optAny.IsSet() || optAny.IsNull() {
			return true
		}
		// Get the actual value
		val, ok := optAny.Any()
		if !ok {
			return true // No value to validate
		}
		// Get the validation rules from the tag parameter
		param := fl.Param()
		if param == "" {
			return true // No rules specified
		}
		// The param contains rules separated by semicolons
		// e.g., "opt=min=2;max=100" → "min=2,max=100"
		rules := strings.ReplaceAll(param, ";", ",")
		// Create a temporary validator and validate the inner value
		tmpValidator := validator.New()
		return tmpValidator.Var(val, rules) == nil
	})

	if err != nil {
		panic("failed to register opt validation: " + err.Error())
	}
}

func RegisterAlarmTypeValidation(v *validator.Validate) {
	err := v.RegisterValidation("alarmtype", func(fl validator.FieldLevel) bool {
		// Check if this field implements OptionalAny
		optAny, ok := fl.Field().Interface().(OptionalAny)
		if !ok {
			// Not an Optional type, skip validation
			return true
		}

		// If unset or null, validation passes
		if !optAny.IsSet() || optAny.IsNull() {
			return true
		}

		val, ok := optAny.Any()
		if !ok {
			return true // No value to validate
		}

		strVal, ok := val.(string)
		if !ok {
			return false // Not a string, invalid
		}

		return models.IsValidAlarmType(strVal)
	})

	if err != nil {
		panic("failed to register alarmtype validation: " + err.Error())
	}
}
