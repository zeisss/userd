package user

import (
	"time"
)

type User struct {
	ID string

	ProfileName string

	LoginName         string
	LoginPasswordHash string

	Email         string
	EmailVerified bool

	ResetPasswordToken       string
	ResetPasswordTokenIssued *time.Time
}
