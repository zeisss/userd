package service

import (
	"./user"

	"github.com/juju/errgo"

	"encoding/json"
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
	EventStream EventStream
}

type Config struct {
	AuthEmailMustBeVerified bool
}

type UserService struct {
	Dependencies
	Config
}

var (
	counterCreateUser = NewSuccessFailureCounter("service.CreateUser")
)

func (us *UserService) CreateUser(profileName, email, loginName, loginPassword string) (string, error) {
	log.Printf("call CreateUser('%s', '%s', ..)\n", profileName, email)

	passwordHash := us.Hasher.Hash(loginPassword)
	newUserID := us.IdFactory.NewUserID()

	theUser := user.User{
		ID:          newUserID,
		ProfileName: profileName,

		Email:         email,
		EmailVerified: false,

		LoginName:         loginName,
		LoginPasswordHash: passwordHash,
	}

	err := us.UserStorage.Save(theUser)

	if err != nil {
		counterCreateUser.Failure()
		return "", errgo.Mask(err)
	}

	us.logEvent("user.created", struct {
		UserID      string `json:"user_id"`
		ProfileName string `json:"profile_name"`
	}{newUserID, profileName})

	counterCreateUser.Success()

	return newUserID, nil
}

var (
	counterGetUser = NewSuccessFailureCounter("service.ChangeLoginCredentials")
)

func (us *UserService) GetUser(id string) (user.User, error) {
	log.Printf("call GetUser('%s')\n", id)
	user, err := us.UserStorage.Get(id)
	counterGetUser.CountError(err)
	return user, errgo.Mask(err)
}

var (
	counterChangeLoginCredentials = NewSuccessFailureCounter("service.ChangeLoginCredentials")
)

func (us *UserService) ChangeLoginCredentials(userID, newLogin, newPassword string) error {
	log.Printf("call ChangePassword('%s', ..)\n", userID)

	return counterChangeLoginCredentials.CountError(us.readModifyWrite(userID, func(user *user.User) error {
		user.LoginName = newLogin
		user.LoginPasswordHash = us.Hasher.Hash(newPassword)
		return nil
	}, func(user *user.User) {
		us.logEvent("user.change_login_credentials", struct {
			UserID string `json:"user_id"`
		}{userID})
	}))
}

var (
	counterChangeProfileName = NewSuccessFailureCounter("service.ChangeProfileName")
)

func (us *UserService) ChangeProfileName(userID, profileName string) error {
	log.Printf("call ChangeProfileName('%s', '%s')\n", userID, profileName)

	return counterChangeProfileName.CountError(us.readModifyWrite(userID, func(user *user.User) error {
		user.ProfileName = profileName
		return nil
	}, func(user *user.User) {
		us.logEvent("user.change_profile_name", struct {
			UserID      string `json:"user_id"`
			ProfileName string `json:"profile_name"`
		}{userID, profileName})
	}))
}

var (
	counterChangeEmail = NewSuccessFailureCounter("service.ChangeEmail")
)

func (us *UserService) ChangeEmail(userID, email string) error {
	log.Printf("call ChangeEmail('%s', '%s')\n", userID, email)

	return counterChangeEmail.CountError(us.readModifyWrite(userID, func(user *user.User) error {
		user.Email = email
		return nil
	}, func(user *user.User) {
		us.logEvent("user.change_email", struct {
			UserID string `json:"user_id"`
			Email  string `json:"email"`
		}{userID, email})
	}))
}

var (
	counterAuthenticate = NewSuccessFailureCounter("service.Authenticate")
)

// Authenticate checks whether a user with the given login credentials exists.
// Returns an error if the credentials are incorrect or the user cannot be authorized.
//
// Error Helpers
//
func (us *UserService) Authenticate(loginName, loginPassword string) (string, error) {
	log.Printf("call Authenticate('%s', ...)\n", loginName)

	theUser, err := us.UserStorage.FindByLoginName(loginName)
	if err != nil {
		counterAuthenticate.Failure()
		return "", errgo.Mask(err)
	}

	if us.AuthEmailMustBeVerified {
		if !theUser.EmailVerified {
			counterAuthenticate.Failure()
			return "", UserEmailMustBeVerified
		}
	}

	passwordMatch := us.Hasher.Verify(loginPassword, theUser.LoginPasswordHash)
	if !passwordMatch {
		counterAuthenticate.Failure()
		return "", InvalidCredentials
	}

	needsRehash := us.Hasher.NeedsRehash(theUser.LoginPasswordHash)
	if needsRehash {
		theUser.LoginPasswordHash = us.Hasher.Hash(loginPassword)

		// NOTE: we ignore any error here. Main intent of this function is to provide authentication
		us.UserStorage.Save(theUser)
	}

	us.logEvent("user.authenticated", struct {
		UserID string `json:"user_id"`
	}{theUser.ID})

	counterAuthenticate.Success()

	return theUser.ID, nil
}

var (
	counterSetEmailVerified = NewSuccessFailureCounter("service.SetEmailVerified")
)

func (us *UserService) SetEmailVerified(userID string) error {
	log.Printf("call SetEmailVerified('%s')\n", userID)

	return counterSetEmailVerified.CountError(us.readModifyWrite(userID, func(user *user.User) error {
		user.EmailVerified = true
		return nil
	}, func(user *user.User) {
		us.logEvent("user.email_verified", struct {
			UserID string `json:"user_id"`
			Email  string `json:"email"`
		}{userID, user.Email})
	}))
}

var (
	counterCheckAndSetEmailVerified = NewSuccessFailureCounter("service.CheckAndSetEmailVerified")
)

func (us *UserService) CheckAndSetEmailVerified(userID, email string) error {
	log.Printf("call CheckAndSetEmailVerified('%s', '%s')\n", userID, email)

	return counterCheckAndSetEmailVerified.CountError(us.readModifyWrite(userID, func(user *user.User) error {
		if user.Email != email {
			return InvalidVerificationEmail
		}
		user.EmailVerified = true
		return nil
	}, func(user *user.User) {
		us.logEvent("user.email_verified", struct {
			UserID string `json:"user_id"`
			Email  string `json:"email"`
		}{userID, email})
	}))
}

// readModifyWrite reads the user with the given userID, applies modifier to it, saves the result
// and calls all success function if no error occured.
func (us *UserService) readModifyWrite(userID string, modifier func(user *user.User) error, success ...func(user *user.User)) error {
	user, err := us.UserStorage.Get(userID)
	if err != nil {
		return errgo.Mask(err)
	}

	err = modifier(&user)
	if err != nil {
		return errgo.Mask(err)
	}

	err = us.UserStorage.Save(user)
	if err != nil {
		return errgo.Mask(err)
	}

	for _, f := range success {
		f(&user)
	}
	return nil
}

// logEvent serializes the entry with `encoding/json` and writes it to the us.EventStream
func (us *UserService) logEvent(tag string, entry interface{}) {
	data, err := json.Marshal(entry)
	if err != nil {
		// Our own data structs should always be jsonizable - if not we have a bug
		panic(err)
	}
	us.EventStream.Publish(tag, data)
}
