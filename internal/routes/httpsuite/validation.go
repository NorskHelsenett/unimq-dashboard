package httpsuite

import (
	"github.com/go-playground/validator/v10"
)

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
