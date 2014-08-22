package storage

import "errors"

var (
	InvalidUserObject = errors.New("Invalid user object")
	UserNotFound      = errors.New("No user found.")

	LoginNameAlreadyTaken = errors.New("The given loginName is already taken.")
	EmailAlreadyTaken     = errors.New("The given email address is already taken.")
)
