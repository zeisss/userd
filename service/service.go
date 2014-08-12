package service

import (
	"../user"

	"errors"
	"log"
)

var (
	InvalidCredentials       = errors.New("Invalid credentials.")
	InvalidVerificationEmail = errors.New("Email adress does not match current email for user.")
	UserEmailMustBeVerified  = errors.New("Email must be verified to authenticate.")
)

type Dependencies struct {
	IdFactory   IdFactory
	Hasher      PasswordHasher
	UserStorage UserStorage
}

type Config struct {
	AuthEmailMustBeVerified bool
}

type UserService struct {
	Dependencies
	Config
}

func (us *UserService) CreateUser(profileName, email, loginName, loginPassword string) (string, error) {
	log.Printf("call CreateUser('%s', '%s', ..)\n", profileName, email)

	passwordHash := us.Hasher.Hash(loginPassword)
	newUserId := us.IdFactory.NewUserID()

	theUser := user.User{
		ID:          newUserId,
		ProfileName: profileName,

		Email:         email,
		EmailVerified: false,

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

	return us.readModifyWrite(userID, func(user user.User) error {
		user.LoginPasswordHash = us.Hasher.Hash(newPassword)
		return nil
	})
}

func (us *UserService) ChangeProfileName(userID, profileName string) error {
	log.Printf("call ChangeProfileName('%s', '%s')\n", userID, profileName)

	return us.readModifyWrite(userID, func(user user.User) error {
		user.ProfileName = profileName
		return nil
	})
}

func (us *UserService) ChangeEmail(userID, email string) error {
	log.Printf("call ChangeEmail('%s', '%s')\n", userID, email)

	return us.readModifyWrite(userID, func(user user.User) error {
		user.Email = email
		return nil
	})
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

	if us.AuthEmailMustBeVerified {
		if !theUser.EmailVerified {
			return "", UserEmailMustBeVerified
		}
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

func (us *UserService) SetEmailVerified(userID string) error {
	log.Printf("call SetEmailVerified('%s')\n", userID)

	return us.readModifyWrite(userID, func(user user.User) error {
		user.EmailVerified = true
		return nil
	})
}

func (us *UserService) CheckAndSetEmailVerified(userID, email string) error {
	log.Printf("call CheckAndSetEmailVerified('%s', '%s')\n", userID, email)

	return us.readModifyWrite(userID, func(user user.User) error {
		if user.Email != email {
			return InvalidVerificationEmail
		}
		user.EmailVerified = true
		return nil
	})
}

func (us *UserService) readModifyWrite(userID string, modifier func(user user.User) error) error {
	user, err := us.UserStorage.Get(userID)
	if err != nil {
		return err
	}

	err = modifier(user)
	if err != nil {
		return err
	}

	return us.UserStorage.Save(user)
}
