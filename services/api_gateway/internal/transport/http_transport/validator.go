package httptransport

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	validate     = validator.New()
	versionRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)
)

func init() {
	validate.RegisterValidation("is_version", func(fl validator.FieldLevel) bool {
		return versionRegex.MatchString(fl.Field().String())
	})
}
