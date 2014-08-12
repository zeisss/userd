package service

import (
	"../user"
)

type IdFactory interface {
	NewUserID() string
}

// This follows the design of the PHP password_* functions. The client don't need to know anything about the user algorithms.
type PasswordHasher interface {
	NeedsRehash(password_hash string) bool
	Hash(password string) string
	Verify(password, passwordHash string) bool
}

type UserStorage interface {
	Save(user user.User) error
	Get(userId string) (user.User, error)

	FindByLoginName(loginName string) (user.User, error)
}
