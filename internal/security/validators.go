package security

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ТРЕБОВАНИЕ 1: Белая валидация
var (
	urlRegex    = regexp.MustCompile(`^https?://[^\s]+$`)
	tagRegex    = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,30}$`)
	titleRegex  = regexp.MustCompile(`^[a-zA-Z0-9а-яА-Я\s\-\_\.\,\:\;\!\?]{1,200}$`)
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// Валидация URL
func (v *Validator) ValidateURL(url string) error {
	if utf8.RuneCountInString(url) > 2048 {
		return ErrURLTooLong
	}
	if !urlRegex.MatchString(url) {
		return ErrInvalidURL
	}
	return nil
}

// Валидация email
func (v *Validator) ValidateEmail(email string) error {
	if utf8.RuneCountInString(email) > 255 {
		return ErrEmailTooLong
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}

// Валидация пароля
func (v *Validator) ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return ErrPasswordTooShort
	}
	if utf8.RuneCountInString(password) > 128 {
		return ErrPasswordTooLong
	}
	// Минимум: буква, цифра, спецсимвол
	hasLetter := false
	hasDigit := false
	for _, r := range password {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
		}
		if r >= '0' && r <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return ErrPasswordWeak
	}
	return nil
}

// Валидация тегов
func (v *Validator) ValidateTags(tags []string) error {
	if len(tags) > 20 {
		return ErrTooManyTags
	}
	seen := make(map[string]bool)
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag == "" {
			continue
		}
		if !tagRegex.MatchString(tag) {
			return ErrInvalidTag
		}
		if seen[tag] {
			return ErrDuplicateTag
		}
		seen[tag] = true
	}
	return nil
}

// Валидация заголовка
func (v *Validator) ValidateTitle(title string) error {
	if utf8.RuneCountInString(title) > 200 {
		return ErrTitleTooLong
	}
	if !titleRegex.MatchString(title) {
		return ErrInvalidTitle
	}
	return nil
}

// Ошибки валидации
var (
	ErrURLTooLong      = ValidationError{"URL слишком длинный (макс. 2048 символов)"}
	ErrInvalidURL      = ValidationError{"Неверный формат URL"}
	ErrEmailTooLong    = ValidationError{"Email слишком длинный"}
	ErrInvalidEmail    = ValidationError{"Неверный формат email"}
	ErrPasswordTooShort = ValidationError{"Пароль должен содержать минимум 8 символов"}
	ErrPasswordTooLong  = ValidationError{"Пароль слишком длинный (макс. 128 символов)"}
	ErrPasswordWeak     = ValidationError{"Пароль должен содержать буквы и цифры"}
	ErrTooManyTags     = ValidationError{"Слишком много тегов (макс. 20)"}
	ErrInvalidTag      = ValidationError{"Неверный формат тега"}
	ErrDuplicateTag    = ValidationError{"Дублирующийся тег"}
	ErrTitleTooLong    = ValidationError{"Заголовок слишком длинный"}
	ErrInvalidTitle    = ValidationError{"Заголовок содержит недопустимые символы"}
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
