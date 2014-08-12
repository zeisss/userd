package service

import (
	"./storage"
)

func IsNotFoundError(err error) bool {
	return err == storage.UserNotFound
}
