package service

import (
	"./user"

	"encoding/json"
	"log"
	"time"
)

type Dependencies struct {
	IdFactory   IdFactory
	Hasher      PasswordHasher
	UserStorage UserStorage

	// EventStream.Publish() is called for every succesfull event in the UserService. Should also forward to EventCollector.
	EventStream EventStream
}

type Config struct {
	AuthEmailMustBeVerified bool
	MaxItems                int

	// How long can a ResetPasswordToken be used?
	ResetPasswordExpireTime time.Duration
}

func (c Config) ValidateValues() error {
	if c.MaxItems <= 0 {
		return newInvalidConfig("MaxItems", c.MaxItems)
	}
	if c.ResetPasswordExpireTime <= 0 {
		return newInvalidConfig("ResetPasswordExpireTime", c.ResetPasswordExpireTime)
	}
	return nil
}

// Validate checks all values and panics if any one is invalid.
func (c Config) Validate() {
	if err := c.ValidateValues(); err != nil {
		panic(err)
	}
}

func NewUserService(config Config, deps Dependencies) *UserService {
	config.Validate()

	deps.UserStorage = &UserStorageWrapper{deps.UserStorage}
	return &UserService{
		Dependencies: deps,
		Config:       config,

		EventCollector: NewEventCollector(config.MaxItems),
	}
}

type UserService struct {
	Dependencies
	Config

	// EventCollector is used by any consumer of the UserService which needs access to the previous events.
	EventCollector *EventCollector
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

		us.logEvent("user.created", map[string]interface{}{
			"user_id":      newUserID,
			"profile_name": profileName,
			"email":        email,
		})

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
		us.logEvent("user.change_login_credentials", map[string]interface{}{
			"user_id": userID,
		})
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
		us.logEvent("user.change_profile_name", map[string]interface{}{
			"user_id":      userID,
			"profile_name": profileName,
		})
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
		us.logEvent("user.change_email", map[string]interface{}{
			"user_id": userID,
			"email":   email,
		})
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

		us.logEvent("user.authenticated", map[string]interface{}{
			"user_id": theUser.ID,
		})

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
		us.logEvent("user.email_verified", map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		})
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
		us.logEvent("user.email_verified", map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		})
	})
}

// NewResetLoginCredentialsToken creates a new reset password token, associates it with the user and returns it. The
// consumer should forward this token to the user's email (or via another communication medium which is known
// to reach the real user) to verify that the initiator is the real user.
//
// Event: user.new_reset_password_token(user_id, email, token)
//
// Returs the new token to reset the password with or an error if no user could be found.
func (us *UserService) NewResetLoginCredentialsToken(email string) (string, error) {
	if email == "" {
		// Only one may be empty
		return "", InvalidArguments
	}
	log.Printf("call NewResetLoginCredentialsToken('%s')", email)

	u, err := us.UserStorage.FindByEmail(email)
	if err != nil {
		return "", Mask(err)
	}

	now := time.Now()
	u.ResetPasswordToken = us.IdFactory.NewResetPasswordToken()
	u.ResetPasswordTokenIssued = &now

	if err := us.UserStorage.Save(u); err != nil {
		return "", Mask(err)
	}

	us.logEvent("user.new_reset_login_credentials_token", map[string]interface{}{
		"user_id":   u.ID,
		"email":     u.Email,
		"token":     u.ResetPasswordToken,
		"timestamp": u.ResetPasswordTokenIssued,
	})
	return u.ResetPasswordToken, nil
}

// ResetCredentialsWithToken checks for users with the given token and resets their login credentials to given values.
//
// Event: user.
func (us *UserService) ResetCredentialsWithToken(resetPasswordToken, new_login_name, new_login_password string) (string, error) {
	if resetPasswordToken == "" || new_login_name == "" || new_login_password == "" {
		return "", Mask(InvalidArguments)
	}

	user, err := us.UserStorage.FindByResetPasswordToken(resetPasswordToken)
	if err != nil {
		if IsNotFoundError(err) {
			return "", Mask(InvalidArguments)
		}
		return "", Mask(err)
	}

	if time.Now().After(user.ResetPasswordTokenIssued.Add(us.ResetPasswordExpireTime)) {
		return "", Mask(ResetPasswordTokenExpired)
	}

	user.LoginName = new_login_name
	user.LoginPasswordHash = us.Hasher.Hash(new_login_password)
	user.ResetPasswordToken = ""
	user.ResetPasswordTokenIssued = nil

	if err := us.UserStorage.Save(user); err != nil {
		return "", Mask(err)
	}

	us.logEvent("user.login_credentials_resetted", map[string]interface{}{
		"user_id": user.ID,
	})

	return user.ID, nil
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
	go us.EventStream.Publish(tag, data)
	go us.EventCollector.publish(tag, data)
}
