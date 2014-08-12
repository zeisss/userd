package service

import (
	"./storage"
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
