package service

import (
	"./user"

	"log"
)

const logUserStorageCalls = false

type UserStorageWrapper struct {
	UserStorage UserStorage
}

func (w *UserStorageWrapper) Save(user user.User) error {
	err := w.UserStorage.Save(user)
	if logUserStorageCalls {
		log.Printf("UserStorage.Save(%#v) =>\n\t%#v", user, err)
	}
	return err
}
func (w *UserStorageWrapper) Get(userId string) (user.User, error) {
	user, err := w.UserStorage.Get(userId)
	if logUserStorageCalls {
		log.Printf("UserStorage.Get(%#v) =>\n\t(%#v, %#v)", userId, user, err)
	}
	return user, err
}
func (w *UserStorageWrapper) FindByLoginName(loginName string) (user.User, error) {
	user, err := w.UserStorage.FindByLoginName(loginName)
	if logUserStorageCalls {
		log.Printf("UserStorage.FindByLoginName(%#v) =>\n\t(%#v, %#v)", loginName, user, err)
	}
	return user, err
}
func (w *UserStorageWrapper) FindByEmail(email string) (user.User, error) {
	user, err := w.UserStorage.FindByEmail(email)
	if logUserStorageCalls {
		log.Printf("UserStorage.FindByEmail(%#v) =>\n\t(%#v, %#v)", email, user, err)
	}
	return user, err
}
func (w *UserStorageWrapper) FindByResetPasswordToken(token string) (user.User, error) {
	user, err := w.UserStorage.FindByResetPasswordToken(token)
	if logUserStorageCalls {
		log.Printf("UserStorage.FindByResetPasswordToken(%#v) =>\n\t(%#v, %#v)", token, user, err)
	}
	return user, err
}
