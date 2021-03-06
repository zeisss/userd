package idfactory

import (
	"code.google.com/p/go-uuid/uuid"
)

type UUIDFactory struct{}

func (factory *UUIDFactory) NewUserID() string {
	return uuid.New()
}

func (factory *UUIDFactory) NewResetPasswordToken() string {
	return uuid.New()
}
