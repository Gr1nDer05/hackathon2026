package api

import (
	"encoding/json"
	"errors"
	"net/mail"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	fullNamePartPattern = regexp.MustCompile(`^[А-ЯЁа-яё-]+$`)
	cityPattern         = regexp.MustCompile(`^[А-ЯЁа-яё -]+$`)
	phonePattern        = regexp.MustCompile(`^\+7 \d{3} \d{3}-\d{2}-\d{2}$`)
)

func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		message, fieldErrors := describeBindingError(err, dst)
		writeError(c, 400, message, fieldErrors)
		return false
	}

	return true
}

func describeBindingError(err error, dst any) (string, map[string]string) {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		fieldErrors := make(map[string]string, len(validationErrs))
		for _, item := range validationErrs {
			fieldName := jsonFieldName(dst, item.StructField())
			switch item.Tag() {
			case "required":
				fieldErrors[fieldName] = "This field is required"
			case "email":
				fieldErrors[fieldName] = "Enter a valid email"
			default:
				fieldErrors[fieldName] = "Invalid value"
			}
		}
		return "Validation failed", fieldErrors
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return "Invalid JSON body", nil
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		fieldName := typeErr.Field
		if fieldName == "" {
			fieldName = typeErr.Struct
		}
		if fieldName != "" {
			return "Validation failed", map[string]string{
				jsonFieldName(dst, fieldName): "Invalid value type",
			}
		}
		return "Invalid JSON body", nil
	}

	if strings.HasPrefix(err.Error(), "json: unknown field ") {
		fieldName := strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), `"`)
		return "Validation failed", map[string]string{
			fieldName: "This field is not allowed",
		}
	}

	return "Validation failed", nil
}

func jsonFieldName(dst any, structField string) string {
	if dst == nil || structField == "" {
		return structField
	}

	typ := reflect.TypeOf(dst)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	field, ok := typ.FieldByName(structField)
	if !ok {
		return structField
	}

	tag := strings.TrimSpace(field.Tag.Get("json"))
	if tag == "" {
		return structField
	}

	name := strings.Split(tag, ",")[0]
	if name == "" || name == "-" {
		return structField
	}

	return name
}

func validateCreatePsychologistInput(input domain.CreatePsychologistInput) map[string]string {
	fieldErrors := make(map[string]string)
	validateEmail(fieldErrors, "email", input.Email, true)
	validateFullName(fieldErrors, "full_name", input.FullName, true)
	validatePassword(fieldErrors, "password", input.Password, true)
	return emptyFieldErrors(fieldErrors)
}

func validatePsychologistAccountInput(input domain.UpdatePsychologistAccountInput) map[string]string {
	fieldErrors := make(map[string]string)
	validateEmail(fieldErrors, "email", input.Email, true)
	validateFullName(fieldErrors, "full_name", input.FullName, true)
	return emptyFieldErrors(fieldErrors)
}

func validatePsychologistProfileInput(input domain.UpdatePsychologistProfileInput, requireCoreFields bool) map[string]string {
	fieldErrors := make(map[string]string)
	validateCity(fieldErrors, "city", input.City, requireCoreFields)
	validateSpecialization(fieldErrors, "specialization", input.Specialization, requireCoreFields)
	return emptyFieldErrors(fieldErrors)
}

func validatePsychologistCardInput(input domain.UpdatePsychologistCardInput, requirePhone bool) map[string]string {
	fieldErrors := make(map[string]string)
	validateEmail(fieldErrors, "contact_email", input.ContactEmail, false)
	validatePhone(fieldErrors, "contact_phone", input.ContactPhone, requirePhone)
	return emptyFieldErrors(fieldErrors)
}

func validateEmail(fieldErrors map[string]string, fieldName string, value string, required bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			fieldErrors[fieldName] = "This field is required"
		}
		return
	}

	if _, err := mail.ParseAddress(value); err != nil {
		fieldErrors[fieldName] = "Enter a valid email"
	}
}

func validateFullName(fieldErrors map[string]string, fieldName string, value string, required bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			fieldErrors[fieldName] = "This field is required"
		}
		return
	}

	parts := strings.Fields(value)
	if len(parts) != 3 {
		fieldErrors[fieldName] = "Full name must contain exactly 3 words"
		return
	}

	for _, part := range parts {
		if !fullNamePartPattern.MatchString(part) {
			fieldErrors[fieldName] = "Use only Cyrillic letters, spaces, and hyphens"
			return
		}
	}
}

func validatePassword(fieldErrors map[string]string, fieldName string, value string, required bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			fieldErrors[fieldName] = "This field is required"
		}
		return
	}

	if utf8.RuneCountInString(value) < 8 {
		fieldErrors[fieldName] = "Password must be at least 8 characters long"
		return
	}

	hasLetter := false
	hasDigit := false
	for _, r := range value {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		fieldErrors[fieldName] = "Password must contain at least 1 letter and 1 digit"
	}
}

func validatePhone(fieldErrors map[string]string, fieldName string, value string, required bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			fieldErrors[fieldName] = "This field is required"
		}
		return
	}

	if !phonePattern.MatchString(value) {
		fieldErrors[fieldName] = "Use format +7 999 123-45-67"
	}
}

func validateCity(fieldErrors map[string]string, fieldName string, value string, required bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			fieldErrors[fieldName] = "This field is required"
		}
		return
	}

	if !cityPattern.MatchString(value) {
		fieldErrors[fieldName] = "Use only Russian letters, spaces, and hyphens"
	}
}

func validateSpecialization(fieldErrors map[string]string, fieldName string, value string, required bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			fieldErrors[fieldName] = "This field is required"
		}
		return
	}

	if utf8.RuneCountInString(value) < 4 {
		fieldErrors[fieldName] = "Must be at least 4 characters long"
	}
}

func emptyFieldErrors(fieldErrors map[string]string) map[string]string {
	if len(fieldErrors) == 0 {
		return nil
	}

	return fieldErrors
}
