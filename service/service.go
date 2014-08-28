package service

import (
	"./user"

	"encoding/json"
	"log"
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
	metricCreateUser = NewSuccessFailureMetric("service.CreateUser")
)

func (us *UserService) CreateUser(profileName, email, loginName, loginPassword string) (string, error) {
	if profileName == "" || email == "" || loginName == "" || loginPassword == "" {
		return "", InvalidArguments
	}
	log.Printf("call CreateUser('%s', '%s', ..)\n", profileName, email)

	var result string = ""
	err := metricCreateUser.Run(func() error {
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
			return Mask(err)
		}

		us.logEvent("user.created", struct {
			UserID      string `json:"user_id"`
			ProfileName string `json:"profile_name"`
		}{newUserID, profileName})

		result = newUserID
		return nil
	})
	return result, err

}

var (
	metricGetUser = NewSuccessFailureMetric("service.ChangeLoginCredentials")
)

func (us *UserService) GetUser(id string) (user.User, error) {
	log.Printf("call GetUser('%s')\n", id)
	user, err := us.UserStorage.Get(id)
	metricGetUser.CountError(err)
	return user, Mask(err)
}

var (
	metricChangeLoginCredentials = NewSuccessFailureMetric("service.ChangeLoginCredentials")
)

func (us *UserService) ChangeLoginCredentials(userID, newLogin, newPassword string) error {
	if userID == "" || newLogin == "" || newPassword == "" {
		return InvalidArguments
	}
	log.Printf("call ChangeLoginCredentials('%s', ..)\n", userID)

	return us.readModifyWrite(metricChangeLoginCredentials, userID, func(user *user.User) error {
		user.LoginName = newLogin
		user.LoginPasswordHash = us.Hasher.Hash(newPassword)
		return nil
	}, func(user *user.User) {
		us.logEvent("user.change_login_credentials", struct {
			UserID string `json:"user_id"`
		}{userID})
	})
}

var (
	metricChangeProfileName = NewSuccessFailureMetric("service.ChangeProfileName")
)

func (us *UserService) ChangeProfileName(userID, profileName string) error {
	if userID == "" || profileName == "" {
		return InvalidArguments
	}
	log.Printf("call ChangeProfileName('%s', '%s')\n", userID, profileName)

	return us.readModifyWrite(metricChangeProfileName, userID, func(user *user.User) error {
		user.ProfileName = profileName
		return nil
	}, func(user *user.User) {
		us.logEvent("user.change_profile_name", struct {
			UserID      string `json:"user_id"`
			ProfileName string `json:"profile_name"`
		}{userID, profileName})
	})
}

var (
	metricChangeEmail = NewSuccessFailureMetric("service.ChangeEmail")
)

func (us *UserService) ChangeEmail(userID, email string) error {
	if userID == "" || email == "" {
		return InvalidArguments
	}
	log.Printf("call ChangeEmail('%s', '%s')\n", userID, email)

	return us.readModifyWrite(metricChangeEmail, userID, func(user *user.User) error {
		user.Email = email
		return nil
	}, func(user *user.User) {
		us.logEvent("user.change_email", struct {
			UserID string `json:"user_id"`
			Email  string `json:"email"`
		}{userID, email})
	})
}

var (
	metricAuthenticate = NewSuccessFailureMetric("service.Authenticate")
)

// Authenticate checks whether a user with the given login credentials exists.
// Returns an error if the credentials are incorrect or the user cannot be authorized.
//
// Error Helpers
//
func (us *UserService) Authenticate(loginName, loginPassword string) (string, error) {
	if loginName == "" || loginPassword == "" {
		return "", InvalidArguments
	}
	log.Printf("call Authenticate('%s', ...)\n", loginName)

	var result string
	err := metricAuthenticate.Run(func() error {
		theUser, err := us.UserStorage.FindByLoginName(loginName)
		if err != nil {
			metricAuthenticate.Failure()
			return Mask(err)
		}

		if us.AuthEmailMustBeVerified {
			if !theUser.EmailVerified {
				metricAuthenticate.Failure()
				return UserEmailMustBeVerified
			}
		}

		passwordMatch := us.Hasher.Verify(loginPassword, theUser.LoginPasswordHash)
		if !passwordMatch {
			metricAuthenticate.Failure()
			return InvalidCredentials
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

		metricAuthenticate.Success()

		// Finish
		result = theUser.ID
		return nil
	})
	return result, err

}

var (
	metricSetEmailVerified = NewSuccessFailureMetric("service.SetEmailVerified")
)

func (us *UserService) SetEmailVerified(userID string) error {
	if userID == "" {
		return InvalidArguments
	}
	log.Printf("call SetEmailVerified('%s')\n", userID)

	return us.readModifyWrite(metricSetEmailVerified, userID, func(user *user.User) error {
		user.EmailVerified = true
		return nil
	}, func(user *user.User) {
		us.logEvent("user.email_verified", struct {
			UserID string `json:"user_id"`
			Email  string `json:"email"`
		}{userID, user.Email})
	})
}

var (
	metricCheckAndSetEmailVerified = NewSuccessFailureMetric("service.CheckAndSetEmailVerified")
)

func (us *UserService) CheckAndSetEmailVerified(userID, email string) error {
	if userID == "" || email == "" {
		return InvalidArguments
	}
	log.Printf("call CheckAndSetEmailVerified('%s', '%s')\n", userID, email)

	return us.readModifyWrite(metricCheckAndSetEmailVerified, userID, func(user *user.User) error {
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
	})
}

// readModifyWrite reads the user with the given userID, applies modifier to it, saves the result
// and calls all success function if no error occured.
func (us *UserService) readModifyWrite(m SuccessFailureMetric, userID string, modifier func(user *user.User) error, success ...func(user *user.User)) error {
	return m.Run(func() error {
		user, err := us.UserStorage.Get(userID)
		if err != nil {
			return Mask(err)
		}

		err = modifier(&user)
		if err != nil {
			return Mask(err)
		}

		err = us.UserStorage.Save(user)
		if err != nil {
			return Mask(err)
		}

		for _, f := range success {
			f(&user)
		}
		return nil
	})
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
