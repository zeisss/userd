package client

import (
	"fmt"
	"testing"
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

func TestApiCreateUser(t *testing.T) {
	RunApiCreateUser(t, "CreateUserTest")
}

func TestReadUser(t *testing.T) {
	userResult := RunApiCreateUser(t, "ReadUserTest")

	user, err := ApiGetUser(userResult.userID)
	if err != nil {
		t.Fatalf("Failed to read user from service: %v", err)
	}
	if user.ProfileName != "ReadUserTest" {
		t.Fatalf("Failed to read profile name from service: %s", user.ProfileName)
	}
}

// ----------------------------------------------------

type ApiCreateUserResult struct {
	userID string
}

func RunApiCreateUser(t *testing.T, username string) ApiCreateUserResult {
	userID, err := ApiCreateUser(username, username+"@moinz.de", username, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if userID == "" {
		t.Fatalf("Received empty string instead of userID.")
	}

	return ApiCreateUserResult{userID}
}
