package client

import (
	"testing"

	"strings"

	"github.com/manveru/faker"
)

const (
	Password = "secret"
)

var (
	Builder = UserdBuilder{}
)

func init() {
	fake, err := faker.New("en")
	if err != nil {
		panic(err)
	}
	Builder.Fake = fake
}
func TestIntegrationApiCreateUser__SuiteAll(t *testing.T) {
	Builder.givenNewUser(t)
}

func TestIntegrationReadUser__SuiteAll(t *testing.T) {
	userResult := Builder.givenNewUser(t)

	user, err := ApiGetUser(userResult.userID)
	if err != nil {
		t.Fatalf("Failed to read user from service: %v", err)
	}
	if user.ProfileName != userResult.UserName {
		t.Fatalf("Failed to read profile name from service: %s", user.ProfileName)
	}
	if user.Email != userResult.Email {
		t.Fatalf("Email differs")
	}
	if user.LoginName != userResult.LoginName {
		t.Fatalf("LoginName differs")
	}
}

func TestIntegrationAuth__SuiteAll(t *testing.T) {
	userResult := Builder.givenNewVerifiedUser(t)

	userId, err := ApiAuthenticate(userResult.LoginName, Password)
	if err != nil {
		t.Fatalf("Failed to perform auth: %v", err)
	}
	if userId != userResult.userID {
		t.Fatalf("Authentication succeeded, but wrong userid was received: %s instead of %s ", userId, userResult.userID)
	}
}

func TestIntegrationChangeProfileName__SuiteAll(t *testing.T) {
	userResult := Builder.givenNewUser(t)
	newName := Builder.Fake.Name()

	if err := ApiChangeProfileName(userResult.userID, newName); err != nil {
		t.Fatalf("Failed to change profile name: %v", err)
	}

	user, err := ApiGetUser(userResult.userID)
	if err != nil {
		t.Fatalf("Failed to read user: %v", err)
	}
	if user.ProfileName != newName {
		t.Fatalf("Profile Name was not changed!")
	}
}

func TestIntegrationChangeLoginCredentials__SuiteAll(t *testing.T) {
	userResult := Builder.givenNewVerifiedUser(t)
	newName := Builder.Fake.UserName()

	if err := ApiChangeLoginCredentials(userResult.userID, newName, "new_secret"); err != nil {
		t.Fatalf("Failed to change profile name: %v", err)
	}

	userID, err := ApiAuthenticate(newName, "new_secret")
	if err != nil {
		t.Fatalf("Failed to read user: %v", err)
	}
	if userID != userResult.userID {
		t.Fatalf("Logged into the wrong user!")
	}
}

func TestIntegrationChangeEmail__SuiteAll(t *testing.T) {
	userResult := Builder.givenNewVerifiedUser(t)
	newEmail := Builder.Fake.Email()

	if err := ApiChangeEmail(userResult.userID, newEmail); err != nil {
		t.Fatalf("Failed to change email: %v", err)
	}

	user, err := ApiGetUser(userResult.userID)
	if err != nil {
		t.Fatalf("Failed to read user: %v", err)
	}
	if user.Email != newEmail {
		t.Fatalf("Expected new email '%s', but got '%s'", newEmail, user.Email)
	}
}

func TestIntegrationAuthSucceedsUnverified__SuiteAuthEmailFalse(t *testing.T) {
	userResult := Builder.givenNewUser(t)

	userID, err := ApiAuthenticate(userResult.LoginName, Password)
	if err != nil {
		t.Fatalf("Failed to auth: %v", err)
	}
	if userID != userResult.userID {
		t.Fatalf("Authenticated as wrong user, got '%s', expected '%s'", userID, userResult.userID)
	}
}

func TestIntegrationAuthFailsUnverified__SuiteAuthEmailTrue(t *testing.T) {
	user := Builder.givenNewUser(t)

	_, err := ApiAuthenticate(user.LoginName, Password)
	if err == nil {
		t.Fatalf("Expected email-not-verified error, got nil")
	}

	expected := "{\"msg\":\"Email must be verified to authenticate.\"}"
	if strings.TrimSpace(err.Error()) != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, err.Error())
	}
}

// ----------------------------------------------------
type UserdBuilder struct {
	Fake *faker.Faker
}

type ApiCreateUserResult struct {
	userID    string
	Email     string
	UserName  string
	LoginName string
}

func (b *UserdBuilder) givenNewUser(t *testing.T) ApiCreateUserResult {
	user := ApiCreateUserResult{
		userID:    "",
		Email:     b.Fake.FreeEmail(),
		UserName:  b.Fake.UserName(),
		LoginName: b.Fake.UserName(),
	}
	var err error

	user.userID, err = ApiCreateUser(user.UserName, user.Email, user.LoginName, Password)
	if err != nil {
		t.Fatal(err)
	}
	if user.userID == "" {
		t.Fatalf("Received empty string instead of userID.")
	}

	return user
}

func (b *UserdBuilder) givenNewVerifiedUser(t *testing.T) ApiCreateUserResult {
	user := b.givenNewUser(t)

	err := ApiVerifyEmail(user.userID)
	if err != nil {
		t.Fatalf("Failed to verify user %s: %v", user.userID, err)
	}
	return user
}
