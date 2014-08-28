package service

import (
	"./storage"

	"github.com/juju/errgo"
)

var (
	Mask = errgo.MaskFunc(IsServiceError, IsNotFoundError, IsEmailAlreadyTakenError, IsLoginNameAlreadyTakenError, IsUserEmailMustBeVerifiedError)
)

var (
	InvalidArguments         = errgo.New("Invalid arguments.")
	InvalidCredentials       = errgo.New("Invalid credentials.")
	InvalidVerificationEmail = errgo.New("Email adress does not match current email for user.")
	UserEmailMustBeVerified  = errgo.New("Email must be verified to authenticate.")
)

func IsNotFoundError(err error) bool {
	return err == storage.UserNotFound
}

func IsEmailAlreadyTakenError(err error) bool {
	return err == storage.EmailAlreadyTaken
}

func IsLoginNameAlreadyTakenError(err error) bool {
	return err == storage.LoginNameAlreadyTaken
}

func IsUserEmailMustBeVerifiedError(err error) bool {
	return err == UserEmailMustBeVerified
}

func IsServiceError(err error) bool {
	return err == InvalidArguments || err == InvalidCredentials || err == InvalidVerificationEmail
}
