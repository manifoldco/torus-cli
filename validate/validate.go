package validate

import (
	"fmt"

	"github.com/asaskevich/govalidator"

	"github.com/manifoldco/promptui"
)

const slugPattern = "^[a-z][a-z0-9\\-\\_]{0,63}$"
const namePattern = "^[a-zA-Z\\s,\\.'\\-pL]{3,64}$"
const inviteCodePattern = "(?i)^[0-9a-ht-zjkmnpqr]{10}$"
const verifyCodePattern = "(?i)^[0-9a-ht-zjkmnpqr]{9}$"

const slugErrorPattern = "%s must be between 1 and 64 characters in length and only contain alphabetical letters, numbers, hyphens, and underscores."
const nameErrorPattern = "%s must be between 3 and 64 characters in length and only contain letters, commas, periods, apostraphes, and hyphens."

// ProjectName validates whether the input meets the project name requirements
var ProjectName Func

// OrgName validates whether the input meets the org name requirements
var OrgName Func

// TeamName validates whether the input meets the team name requirements
var TeamName Func

// RoleName validates whether the input meets the role name requirements
var RoleName Func

// PolicyName validates whether the input meets the policy name requirements
var PolicyName Func

// Username validates whether the input meets the username requirements
var Username Func

func init() {
	ProjectName = SlugValidator("Project names")
	OrgName = SlugValidator("Org names")
	TeamName = SlugValidator("Team names")
	RoleName = SlugValidator("Machine role names")
	Username = SlugValidator("Usernames")
	PolicyName = SlugValidator("Policy names")
}

// ValidationError represents an error encountered when validating a field
type ValidationError struct {
	msg string
}

// Error returns the error message completing the Error interface
func (e *ValidationError) Error() string {
	return e.msg
}

// NewValidationError returns a validation error
func NewValidationError(msg string) error {
	return &ValidationError{msg: msg}
}

// Func represents a validation function
type Func promptui.ValidateFunc

// Confirmer returns a confirmation validation function which validates the
// input depending on whether or not this prompt is default Yes or No.
func Confirmer(defaultYes bool) Func {
	return func(input string) error {
		if defaultYes {
			if input == "" || input == "y" || input == "N" {
				return nil
			}

			return NewValidationError("You must specify either 'y' for yes or 'N' for no.")
		}

		if input == "" || input == "Y" || input == "n" {
			return nil
		}

		return NewValidationError("You must specify either 'Y' for yes or 'n' for no.")
	}
}

// OneOf returns a validation function which validates whether or not the input
// matches one of the given options.
func OneOf(choices []string) Func {
	return func(input string) error {
		for _, v := range choices {
			if input == v {
				return nil
			}
		}

		return NewValidationError(fmt.Sprintf("%s is not a valid option.", input))
	}
}

// SlugValidator returns a validation function
func SlugValidator(fieldName string) Func {
	return func(input string) error {
		if govalidator.StringMatches(input, slugPattern) {
			return nil
		}

		return NewValidationError(fmt.Sprintf(slugErrorPattern, fieldName))
	}
}

// InviteCode validates whether the input meets the invite code requirements
func InviteCode(input string) error {
	if govalidator.StringMatches(input, inviteCodePattern) {
		return nil
	}

	return NewValidationError("Please enter a valid invite code. Make sure to copy the entire code from the email!")
}

// VerificationCode validates whether the input meets the verification code requirements
func VerificationCode(input string) error {
	if govalidator.StringMatches(input, verifyCodePattern) {
		return nil
	}

	return NewValidationError("Please enter a valid verification code. Make sure to copy the entire code from the email!")
}

// Description validates whether the input meets the descriptin requirements
func Description(input, fieldName string) error {
	if len(input) <= 500 {
		return nil
	}

	return NewValidationError(fieldName + " descriptions must be less than 500 characters")
}

// Email validates whether the input is a valid email format
func Email(input string) error {
	if govalidator.IsEmail(input) {
		return nil
	}

	return NewValidationError("Please enter a valid email address")
}

// Name validates whether the input meets first name last name requirements
func Name(input string) error {
	if govalidator.StringMatches(input, namePattern) {
		return nil
	}

	return NewValidationError(fmt.Sprintf(nameErrorPattern, "Names"))
}

// Password ensures the input meets password requirements
func Password(input string) error {
	length := len(input)
	if length >= 8 {
		return nil
	}
	if length > 0 {
		return NewValidationError("Passwords must be at least 8 characters")
	}

	return NewValidationError("Please enter a password")
}

// ConfirmPassword ensures the input meets the password requirements and
// matches the previously provided password
func ConfirmPassword(previous string) Func {
	return func(input string) error {
		err := Password(input)
		if err != nil {
			return err
		}

		if input != previous {
			return NewValidationError("The password you provided does not match the previous password you provided!")
		}

		return nil
	}
}
