package validate

import (
	"errors"

	"github.com/asaskevich/govalidator"
)

const slugPattern = "^[a-z][a-z0-9\\-\\_]{0,63}$"
const namePattern = "^[a-zA-Z\\s,\\.'\\-pL]{1,64}$"

// Slug validates whether the input meets slug requirements
func Slug(input, fieldName string, errorMessage *string) error {
	if govalidator.StringMatches(input, slugPattern) {
		return nil
	}
	var msg string
	if errorMessage != nil {
		msg = *errorMessage
	} else {
		msg = fieldName + " names can only use a-z, 0-9, hyphens and underscores"
	}
	return errors.New(msg)
}

// Email validates whether the input is a valid email format
func Email(input string) error {
	if govalidator.IsEmail(input) {
		return nil
	}
	return errors.New("Please enter a valid email address")
}

// Name validates whether the input meets first name last name requirements
func Name(input string) error {
	if govalidator.StringMatches(input, namePattern) {
		return nil
	}
	return errors.New("Please enter a valid name")
}

// Username validates whether the input meets username requirements
func Username(input string) error {
	message := "Please enter a valid username"
	return Slug(input, "username", &message)
}

// Password ensures the input meets password requirements
func Password(input string) error {
	length := len(input)
	if length >= 8 {
		return nil
	}
	if length > 0 {
		return errors.New("Passwords must be at least 8 characters")
	}

	return errors.New("Please enter your password")
}
