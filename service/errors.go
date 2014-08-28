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
	return errgo.Cause(err) == storage.UserNotFound
}

func IsEmailAlreadyTakenError(err error) bool {
	return errgo.Cause(err) == storage.EmailAlreadyTaken
}

func IsLoginNameAlreadyTakenError(err error) bool {
	return errgo.Cause(err) == storage.LoginNameAlreadyTaken
}

func IsUserEmailMustBeVerifiedError(err error) bool {
	return errgo.Cause(err) == UserEmailMustBeVerified
}

func IsInvalidCredentials(err error) bool {
	return errgo.Cause(err) == InvalidCredentials
}

func IsServiceError(err error) bool {
	err = errgo.Cause(err)
	return err == InvalidArguments || IsInvalidCredentials(err) || err == InvalidVerificationEmail
}
