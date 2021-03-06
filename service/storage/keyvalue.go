package storage

import (
	"../user"

	"github.com/juju/errgo"

	"encoding/json"
)

const debugKeyValue = true

type keyValueIndex interface {
	Put(key, userID string) error
	Remove(key string) error
	Lookup(key string) (string, bool, error)
}

type keyValueStorageDriver interface {
	// Set writes the json with data
	Set(userID, json string) error

	// Lookup returns the json previously written with Set().
	Lookup(userID string) (string, bool, error)

	// Index is called initially to create a helper for accessing an index
	Index(name string) keyValueIndex
}

type keyValueStorage struct {
	LoginNames         keyValueIndex
	Emails             keyValueIndex
	ResetPasswordToken keyValueIndex

	Driver keyValueStorageDriver
}

func newKeyValueStorage(driver keyValueStorageDriver) *keyValueStorage {
	loginNames := driver.Index("login_name")
	emails := driver.Index("emails")
	resedPasswordToken := driver.Index("reset_password_token")

	return &keyValueStorage{
		Driver:             driver,
		LoginNames:         loginNames,
		Emails:             emails,
		ResetPasswordToken: resedPasswordToken,
	}
}

func (s *keyValueStorage) Save(user user.User) error {
	if user.ID == "" {
		return errgo.Mask(InvalidUserObject)
	}
	if user.Email == "" {
		return errgo.Mask(InvalidUserObject)

	}

	if user.LoginName == "" {
		return errgo.Mask(InvalidUserObject)
	}

	// Unique Index Validation
	if taken, err := s.checkTakenByOtherUser(s.Emails, user.Email, user.ID); err != nil {
		return errgo.Mask(err)
	} else if taken {
		return EmailAlreadyTaken
	}

	if taken, err := s.checkTakenByOtherUser(s.LoginNames, user.LoginName, user.ID); err != nil {
		return errgo.Mask(err)
	} else if taken {
		return LoginNameAlreadyTaken
	}

	if user.ResetPasswordToken != "" {
		if taken, err := s.checkTakenByOtherUser(s.ResetPasswordToken, user.ResetPasswordToken, user.ID); err != nil {
			return errgo.Mask(err)
		} else if taken {
			return LoginNameAlreadyTaken
		}
	}

	// Write
	oldUser, err := s.noLockLookup(user.ID)
	if err != nil && err != UserNotFound {
		return errgo.Mask(err)
	}

	if oldUser.Email != "" {
		s.Emails.Remove(oldUser.Email)
	}

	if oldUser.LoginName != "" {
		s.LoginNames.Remove(oldUser.LoginName)
	}

	if oldUser.ResetPasswordToken != "" {
		s.ResetPasswordToken.Remove(oldUser.ResetPasswordToken)
	}

	s.Emails.Put(user.Email, user.ID)
	s.LoginNames.Put(user.LoginName, user.ID)
	if user.ResetPasswordToken != "" {
		s.ResetPasswordToken.Put(user.ResetPasswordToken, user.ID)
	}

	data, err := json.Marshal(user)
	if err != nil {
		return errgo.Mask(err)
	}
	s.Driver.Set(user.ID, string(data))

	return nil
}

func (s *keyValueStorage) Get(userID string) (user.User, error) {
	if userID == "" {
		panic("Invalid parameter: userID is empty.")
	}

	return s.noLockLookup(userID)
}

func (s *keyValueStorage) FindByLoginName(loginName string) (user.User, error) {
	userID, ok, err := s.LoginNames.Lookup(loginName)
	if err != nil {
		return user.User{}, errgo.Mask(err)
	}
	if !ok {
		return user.User{}, UserNotFound
	}
	return s.noLockLookup(userID)
}
func (s *keyValueStorage) FindByEmail(email string) (user.User, error) {
	userID, ok, err := s.Emails.Lookup(email)
	if err != nil {
		return user.User{}, errgo.Mask(err)
	}
	if !ok {
		return user.User{}, UserNotFound
	}
	return s.noLockLookup(userID)
}
func (s *keyValueStorage) FindByResetPasswordToken(token string) (user.User, error) {
	userID, ok, err := s.ResetPasswordToken.Lookup(token)
	if err != nil {
		return user.User{}, errgo.Mask(err)
	}
	if !ok {
		return user.User{}, UserNotFound
	}
	return s.noLockLookup(userID)
}

// -------------------------------------------------

func (s *keyValueStorage) checkTakenByOtherUser(index keyValueIndex, key, userID string) (bool, error) {
	otherUserID, taken, err := index.Lookup(key)
	if err != nil {
		return false, errgo.Mask(err)
	}
	if taken && otherUserID != userID {
		return true, nil
	}
	return false, nil
}

func (s *keyValueStorage) noLockLookup(userID string) (user.User, error) {
	userJson, ok, err := s.Driver.Lookup(userID)
	if err != nil {
		return user.User{}, errgo.Mask(err)
	}

	if !ok {
		return user.User{}, UserNotFound
	} else {
		var user user.User
		if err := json.Unmarshal([]byte(userJson), &user); err != nil {
			return user, errgo.Mask(err)
		}
		return user, nil
	}
}
