package client

import (
	"fmt"
	"testing"
)

const (
	Password = "secret"
)

func ExampleApiCreateAndReadUser() {
	userID, err := ApiCreateUser("CEO", "ceo@acme.com", "CEO", "secret-passphrase")
	if err != nil {
		panic(err)
	}

	user, err := ApiGetUser(userID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Profile Name: %s\n", user.ProfileName)
	fmt.Printf("Email: %s\n", user.Email)

	// Output:
	// Profile Name: CEO
	// Email: ceo@acme.com
}

func TestIntegrationApiCreateUser__SuiteAll(t *testing.T) {
	RunApiCreateUser(t, "CreateUserTest")
}

func TestIntegrationReadUser__SuiteAll(t *testing.T) {
	userResult := RunApiCreateUser(t, "ReadUserTest")

	user, err := ApiGetUser(userResult.userID)
	if err != nil {
		t.Fatalf("Failed to read user from service: %v", err)
	}
	if user.ProfileName != "ReadUserTest" {
		t.Fatalf("Failed to read profile name from service: %s", user.ProfileName)
	}
}

func TestIntegrationAuth__SuiteAll(t *testing.T) {
	userResult := RunApiCreateAndVerifyUser(t, "TestAuth")

	userId, err := ApiAuthenticate("TestAuth", Password)
	if err != nil {
		t.Fatalf("Failed to perform auth: %v", err)
	}
	if userId != userResult.userID {
		t.Fatalf("Authentication succeeded, but wrong userid was received: %s instead of %s ", userId, userResult.userID)
	}
}

func TestIntegrationChangeProfileName__SuiteAll(t *testing.T) {
	userResult := RunApiCreateUser(t, "ChangeProfileName")

	if err := ApiChangeProfileName(userResult.userID, "ChangeProfileNameChanged"); err != nil {
		t.Fatalf("Failed to change profile name: %v", err)
	}

	user, err := ApiGetUser(userResult.userID)
	if err != nil {
		t.Fatalf("Failed to read user: %v", err)
	}
	if user.ProfileName != "ChangeProfileNameChanged" {
		t.Fatalf("Profile Name was not changed!")
	}
}

func TestIntegrationChangeLoginCredentials__SuiteAll(t *testing.T) {
	userResult := RunApiCreateAndVerifyUser(t, "TestChangeLoginCredentials")

	if err := ApiChangeLoginCredentials(userResult.userID, "TestChangeLoginCredentialsChanged", "new_secret"); err != nil {
		t.Fatalf("Failed to change profile name: %v", err)
	}

	userID, err := ApiAuthenticate("TestChangeLoginCredentialsChanged", "new_secret")
	if err != nil {
		t.Fatalf("Failed to read user: %v", err)
	}
	if userID != userResult.userID {
		t.Fatalf("Logged into the wrong user!")
	}
}

func TestIntegrationAuthSucceedsUnverified__SuiteAuthEmailFalse(t *testing.T) {
	userResult := RunApiCreateUser(t, "AuthFailsUnauthenticated")

	userID, err := ApiAuthenticate("AuthFailsUnauthenticated", Password)
	if err != nil {
		t.Fatalf("Failed to auth: %v", err)
	}
	if userID != userResult.userID {
		t.Fatalf("Authenticated as wrong user, got '%s', expected '%s'", userID, userResult.userID)
	}
}

func TestIntegrationAuthFailsUnverified__SuiteAuthEmailTrue(t *testing.T) {
	_ = RunApiCreateUser(t, "AuthFailsUnauthenticated")

	_, err := ApiAuthenticate("AuthFailsUnauthenticated", Password)
	if err == nil {
		t.Fatalf("Expected email-not-verified error, got nil")
	}
	if err.Error() != "Email must be verified to authenticate." {
		t.Fatalf("Expected 'Email must be verified to authenticate.', got '%s'", err.Error())
	}
}

// ----------------------------------------------------

type ApiCreateUserResult struct {
	userID string
}

func RunApiCreateUser(t *testing.T, username string) ApiCreateUserResult {
	userID, err := ApiCreateUser(username, username+"@moinz.de", username, Password)
	if err != nil {
		t.Fatal(err)
	}
	if userID == "" {
		t.Fatalf("Received empty string instead of userID.")
	}

	return ApiCreateUserResult{userID}
}

func RunApiCreateAndVerifyUser(t *testing.T, username string) ApiCreateUserResult {
	user := RunApiCreateUser(t, username)

	err := ApiVerifyEmail(user.userID)
	if err != nil {
		t.Fatalf("Failed to verify user %s: %v", user.userID, err)
	}
	return user
}
