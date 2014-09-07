package service

import (
	"./user"

	"encoding/json"
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

		MetricExecutor: NewMetricExecutor(),
	}
}

type UserService struct {
	Dependencies
	Config
	*MetricExecutor

	// EventCollector is used by any consumer of the UserService which needs access to the previous events.
	EventCollector *EventCollector
}

type CreateUserRequest struct {
	ProfileName   string
	Email         string
	LoginName     string
	LoginPassword string
}

type CreateUserResponse struct {
	UserID string
}

func (us *UserService) CreateUser(req CreateUserRequest, resp *CreateUserResponse) error {
	return us.execute("CreateUser", req, resp, func() error {
		if req.ProfileName == "" || req.Email == "" || req.LoginName == "" || req.LoginPassword == "" {
			return Mask(InvalidArguments)
		}

		passwordHash := us.Hasher.Hash(req.LoginPassword)
		newUserID := us.IdFactory.NewUserID()

		theUser := user.User{
			ID:          newUserID,
			ProfileName: req.ProfileName,

			Email:         req.Email,
			EmailVerified: false,

			LoginName:         req.LoginName,
			LoginPasswordHash: passwordHash,
		}

		if err := us.UserStorage.Save(theUser); err != nil {
			return Mask(err)
		}

		us.logEvent("user.created", map[string]interface{}{
			"user_id":      newUserID,
			"profile_name": req.ProfileName,
			"email":        req.Email,
		})

		resp.UserID = newUserID
		return nil
	})

}

type GetUserRequest struct {
	UserID string
}
type GetUserResponse struct {
	User user.User
}

func (us *UserService) GetUser(req GetUserRequest, response *GetUserResponse) error {
	return us.execute("GetUser", req, response, func() error {
		u, err := us.UserStorage.Get(req.UserID)
		response.User = u
		return Mask(err)
	})
}

type ChangeLoginCredentialsRequest struct {
	UserID        string
	LoginName     string
	LoginPassword string
}

func (us *UserService) ChangeLoginCredentials(req ChangeLoginCredentialsRequest) error {
	return us.execute("ChangeLoginCredentials", req, nil, func() error {
		if req.UserID == "" || req.LoginName == "" || req.LoginPassword == "" {
			return Mask(InvalidArguments)
		}

		return us.readModifyWrite(req.UserID, func(user *user.User) error {
			user.LoginName = req.LoginName
			user.LoginPasswordHash = us.Hasher.Hash(req.LoginPassword)
			return nil
		}, func(user *user.User) {
			us.logEvent("user.change_login_credentials", map[string]interface{}{
				"user_id": req.UserID,
			})
		})
	})

}

type ChangeProfileNameRequest struct {
	UserID      string
	ProfileName string
}

func (us *UserService) ChangeProfileName(request ChangeProfileNameRequest) error {
	return us.execute("ChangeProfileName", request, nil, func() error {
		if request.UserID == "" || request.ProfileName == "" {
			return Mask(InvalidArguments)
		}

		return us.readModifyWrite(request.UserID, func(user *user.User) error {
			user.ProfileName = request.ProfileName
			return nil
		}, func(user *user.User) {
			us.logEvent("user.change_profile_name", map[string]interface{}{
				"user_id":      request.UserID,
				"profile_name": request.ProfileName,
			})
		})
	})

}

type ChangeEmailRequest struct {
	UserID string
	Email  string
}

func (us *UserService) ChangeEmail(request ChangeEmailRequest) error {
	return us.execute("ChangeEmail", request, nil, func() error {
		if request.UserID == "" || request.Email == "" {
			return Mask(InvalidArguments)
		}
		return us.readModifyWrite(request.UserID, func(user *user.User) error {
			user.Email = request.Email
			return nil
		}, func(user *user.User) {
			us.logEvent("user.change_email", map[string]interface{}{
				"user_id": request.UserID,
				"email":   request.Email,
			})
		})
	})

}

type AuthenticateRequest struct {
	LoginName     string
	LoginPassword string
}

type AuthenticateResponse struct {
	UserID string
}

// Authenticate checks whether a user with the given login credentials exists.
// Returns an error if the credentials are incorrect or the user cannot be authorized.
//
// Error Helpers
//
func (us *UserService) Authenticate(request AuthenticateRequest, response *AuthenticateResponse) error {
	return us.execute("Authenticate", request, response, func() error {
		if request.LoginName == "" || request.LoginPassword == "" {
			return Mask(InvalidArguments)
		}

		theUser, err := us.UserStorage.FindByLoginName(request.LoginName)
		if err != nil {
			return Mask(err)
		}

		if us.AuthEmailMustBeVerified {
			if !theUser.EmailVerified {
				return UserEmailMustBeVerified
			}
		}

		passwordMatch := us.Hasher.Verify(request.LoginPassword, theUser.LoginPasswordHash)
		if !passwordMatch {
			return InvalidCredentials
		}

		needsRehash := us.Hasher.NeedsRehash(theUser.LoginPasswordHash)
		if needsRehash {
			theUser.LoginPasswordHash = us.Hasher.Hash(request.LoginPassword)

			// NOTE: we ignore any error here. Main intent of this function is to provide authentication
			us.UserStorage.Save(theUser)
		}

		us.logEvent("user.authenticated", map[string]interface{}{
			"user_id": theUser.ID,
		})

		// Finish
		response.UserID = theUser.ID
		return nil
	})

}

type SetEmailVerifiedRequest struct {
	UserID string
}

func (us *UserService) SetEmailVerified(request SetEmailVerifiedRequest) error {
	return us.execute("SetEmailVerified", request, nil, func() error {
		if request.UserID == "" {
			return InvalidArguments
		}

		return us.readModifyWrite(request.UserID, func(user *user.User) error {
			user.EmailVerified = true
			return nil
		}, func(user *user.User) {
			us.logEvent("user.email_verified", map[string]interface{}{
				"user_id": user.ID,
				"email":   user.Email,
			})
		})
	})

}

type CheckAndSetEmailVerifiedRequest struct {
	UserID string
	Email  string
}

func (us *UserService) CheckAndSetEmailVerified(req CheckAndSetEmailVerifiedRequest) error {
	return us.execute("CheckAndSetEmailVerifiedRequest", req, nil, func() error {
		if req.UserID == "" || req.Email == "" {
			return InvalidArguments
		}

		return us.readModifyWrite(req.UserID, func(user *user.User) error {
			if user.Email != req.Email {
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
	})
}

type NewResetCredentialsTokenRequest struct {
	Email string
}

type NewResetCredentialsTokenResponse struct {
	Token string
}

// NewResetLoginCredentialsToken creates a new reset password token, associates it with the user and returns it. The
// consumer should forward this token to the user's email (or via another communication medium which is known
// to reach the real user) to verify that the initiator is the real user.
//
// Event: user.new_reset_password_token(user_id, email, token)
//
// Returs the new token to reset the password with or an error if no user could be found.
func (us *UserService) NewResetLoginCredentialsToken(request NewResetCredentialsTokenRequest, response *NewResetCredentialsTokenResponse) error {
	return us.execute("NewResetLoginCredentialsToken", request, response, func() error {
		if request.Email == "" {
			// Only one may be empty
			return Mask(InvalidArguments)
		}

		u, err := us.UserStorage.FindByEmail(request.Email)
		if err != nil {
			return Mask(err)
		}

		now := time.Now()
		u.ResetPasswordToken = us.IdFactory.NewResetPasswordToken()
		u.ResetPasswordTokenIssued = &now

		if err := us.UserStorage.Save(u); err != nil {
			return Mask(err)
		}

		us.logEvent("user.new_reset_login_credentials_token", map[string]interface{}{
			"user_id":   u.ID,
			"email":     u.Email,
			"token":     u.ResetPasswordToken,
			"timestamp": u.ResetPasswordTokenIssued,
		})
		response.Token = u.ResetPasswordToken
		return nil

	})
}

type ResetCredentialsRequest struct {
	Token         string
	LoginName     string
	LoginPassword string
}

type ResetCredentialsResponse struct {
	UserID string
}

// ResetCredentialsWithToken checks for users with the given token and resets their login credentials to given values.
//
// Event: user.
func (us *UserService) ResetCredentialsWithToken(req ResetCredentialsRequest, resp *ResetCredentialsResponse) error {
	return us.execute("ResetCredentialsWithToken", req, resp, func() error {
		if req.Token == "" || req.LoginName == "" || req.LoginPassword == "" {
			return Mask(InvalidArguments)
		}

		user, err := us.UserStorage.FindByResetPasswordToken(req.Token)
		if err != nil {
			if IsNotFoundError(err) {
				return Mask(InvalidArguments)
			}
			return Mask(err)
		}

		if time.Now().After(user.ResetPasswordTokenIssued.Add(us.ResetPasswordExpireTime)) {
			return Mask(ResetPasswordTokenExpired)
		}

		user.LoginName = req.LoginName
		user.LoginPasswordHash = us.Hasher.Hash(req.LoginPassword)
		user.ResetPasswordToken = ""
		user.ResetPasswordTokenIssued = nil

		if err := us.UserStorage.Save(user); err != nil {
			return Mask(err)
		}

		us.logEvent("user.login_credentials_resetted", map[string]interface{}{
			"user_id": user.ID,
		})
		resp.UserID = user.ID
		return nil
	})

}

// readModifyWrite reads the user with the given userID, applies modifier to it, saves the result
// and calls all success function if no error occured.
func (us *UserService) readModifyWrite(userID string, modifier func(user *user.User) error, success ...func(user *user.User)) error {
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
