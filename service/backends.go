package service

import (
	"./user"
)

type IdFactory interface {
	NewUserID() string

	// Generates a new password token to be used for the password reset feature.
	// The result must never be empty.
	NewResetPasswordToken() string
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
	FindByEmail(email string) (user.User, error)
	FindByResetPasswordToken(token string) (user.User, error)
}

// EventLog abstracts any eventlog for store the business events of the UserService.
// Could write to RabbitMQ, Apache Kafka or just plain files.
// This is a write-only interface, errors are propagted to stderr or similar..
type EventStream interface {
	// Log forwards the given entry to the eventlog.
	Publish(tag string, entry []byte)
}
