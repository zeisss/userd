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

	// EventStream.Publish() is called for every succesfull event in the UserService. Should also forward to EventCollector.
	EventStream EventStream
}

type Config struct {
	AuthEmailMustBeVerified bool
	MaxItems                int
}

func NewUserService(config Config, deps Dependencies) *UserService {
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

func (us *UserService) CreateUser(profileName, email, loginName, loginPassword string) (string, error) {
	if profileName == "" || email == "" || loginName == "" || loginPassword == "" {
		return "", InvalidArguments
	}
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
		return "", Mask(err)
	}

	us.logEvent("user.created", map[string]interface{}{
		"user_id":      newUserID,
		"profile_name": profileName,
	})

	return newUserID, nil
}

func (us *UserService) GetUser(id string) (user.User, error) {
	log.Printf("call GetUser('%s')\n", id)
	user, err := us.UserStorage.Get(id)
	return user, Mask(err)
}

func (us *UserService) ChangeLoginCredentials(userID, newLogin, newPassword string) error {
	if userID == "" || newLogin == "" || newPassword == "" {
		return InvalidArguments
	}
	log.Printf("call ChangeLoginCredentials('%s', ..)\n", userID)

	return us.readModifyWrite(userID, func(user *user.User) error {
		user.LoginName = newLogin
		user.LoginPasswordHash = us.Hasher.Hash(newPassword)
		return nil
	}, func(user *user.User) {
		us.logEvent("user.change_login_credentials", map[string]interface{}{
			"user_id": userID,
		})
	})
}

func (us *UserService) ChangeProfileName(userID, profileName string) error {
	if userID == "" || profileName == "" {
		return InvalidArguments
	}
	log.Printf("call ChangeProfileName('%s', '%s')\n", userID, profileName)

	return us.readModifyWrite(userID, func(user *user.User) error {
		user.ProfileName = profileName
		return nil
	}, func(user *user.User) {
		us.logEvent("user.change_profile_name", map[string]interface{}{
			"user_id":      userID,
			"profile_name": profileName,
		})
	})
}

func (us *UserService) ChangeEmail(userID, email string) error {
	if userID == "" || email == "" {
		return InvalidArguments
	}
	log.Printf("call ChangeEmail('%s', '%s')\n", userID, email)

	return us.readModifyWrite(userID, func(user *user.User) error {
		user.Email = email
		return nil
	}, func(user *user.User) {
		us.logEvent("user.change_email", map[string]interface{}{
			"user_id": userID,
			"email":   email,
		})
	})
}

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

	theUser, err := us.UserStorage.FindByLoginName(loginName)
	if err != nil {
		return "", Mask(err)
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

	us.logEvent("user.authenticated", map[string]interface{}{
		"user_id": theUser.ID,
	})

	return theUser.ID, nil
}

func (us *UserService) SetEmailVerified(userID string) error {
	if userID == "" {
		return InvalidArguments
	}
	log.Printf("call SetEmailVerified('%s')\n", userID)

	return us.readModifyWrite(userID, func(user *user.User) error {
		user.EmailVerified = true
		return nil
	}, func(user *user.User) {
		us.logEvent("user.email_verified", map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		})
	})
}

func (us *UserService) CheckAndSetEmailVerified(userID, email string) error {
	if userID == "" || email == "" {
		return InvalidArguments
	}
	log.Printf("call CheckAndSetEmailVerified('%s', '%s')\n", userID, email)

	return us.readModifyWrite(userID, func(user *user.User) error {
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

// readModifyWrite reads the user with the given userID, applies modifier to it, saves the result
// and calls all success function if no error occured.
func (us *UserService) readModifyWrite(userID string, modifier func(user *user.User) error, success ...func(user *user.User)) error {
	user, err := us.UserStorage.Get(userID)
	if err != nil {
		return Mask(err)
	}

	// log.Printf("READ %v\n", user)

	err = modifier(&user)
	if err != nil {
		return Mask(err)
	}

	// log.Printf("WRITE %v\n", user)

	err = us.UserStorage.Save(user)
	if err != nil {
		return Mask(err)
	}

	for _, f := range success {
		f(&user)
	}
	return nil
}

// logEvent serializes the entry with `encoding/json` and writes it to the us.EventLog
func (us *UserService) logEvent(tag string, entry interface{}) {
	data, err := json.Marshal(entry)
	if err != nil {
		// Our own data structs should always be jsonizable - if not we have a bug
		panic(err)
	}
	go us.EventStream.Publish(tag, data)
	go us.EventCollector.publish(tag, data)
}
