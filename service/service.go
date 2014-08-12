package service

import (
	"../user"

	"errors"
	"log"
)

var (
	InvalidCredentials = errors.New("Invalid credentials")
)

type UserService struct {
	IdFactory   IdFactory
	Hasher      PasswordHasher
	UserStorage UserStorage
}

func (us *UserService) CreateUser(profileName, email, loginName, loginPassword string) (string, error) {
	log.Printf("call CreateUser('%s', '%s', ..)\n", profileName, email)

	passwordHash := us.Hasher.Hash(loginPassword)
	newUserId := us.IdFactory.NewUserID()

	theUser := user.User{
		ID:          newUserId,
		ProfileName: profileName,
		Email:       email,

		LoginName:         loginName,
		LoginPasswordHash: passwordHash,
	}

	err := us.UserStorage.Save(theUser)
	return newUserId, err
}

func (us *UserService) GetUser(id string) (user.User, error) {
	log.Printf("call GetUser('%s')\n", id)
	return us.UserStorage.Get(id)
}

func (us *UserService) ChangePassword(userID, newPassword string) error {
	log.Printf("call ChangePassword('%s', ..)\n", userID)

	theUser, err := us.UserStorage.Get(userID)
	if err != nil {
		return err
	}

	theUser.LoginPasswordHash = us.Hasher.Hash(newPassword)

	return us.UserStorage.Save(theUser)
}

func (us *UserService) ChangeProfileName(userID, profileName string) error {
	log.Printf("call ChangeProfileName('%s', '%s')\n", userID, profileName)

	theUser, err := us.UserStorage.Get(userID)
	if err != nil {
		return err
	}

	theUser.ProfileName = profileName

	return us.UserStorage.Save(theUser)
}

func (us *UserService) ChangeEmail(userID, email string) error {
	log.Printf("call ChangeEmail('%s', '%s')\n", userID, email)

	theUser, err := us.UserStorage.Get(userID)
	if err != nil {
		return err
	}

	theUser.Email = email

	return us.UserStorage.Save(theUser)
}

// Authenticate checks whether a user with the given login credentials exists.
// Returns an error if the credentials are incorrect or the user cannot be authorized.
//
// Error Helpers
//
func (us *UserService) Authenticate(loginName, loginPassword string) (string, error) {
	theUser, err := us.UserStorage.FindByLoginName(loginName)
	if err != nil {
		return "", err
	}

	passwordMatch := us.Hasher.Verify(loginPassword, theUser.LoginPasswordHash)
	if !passwordMatch {
		return "", InvalidCredentials
	}

	needsRehash := us.Hasher.NeedsRehash(theUser.LoginPasswordHash)
	if needsRehash {
		theUser.LoginPasswordHash = us.Hasher.Hash(loginPassword)

		// NOTE: we ignore any error here. Main intent of this function is to provide authentication
		us.UserStorage.Save(theUser)
	}

	return theUser.ID, nil
}
