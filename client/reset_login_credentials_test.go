package client

import (
	"testing"
)

func TestIntegrationUserLoginCredentialsForgotten__SuiteAll(t *testing.T) {
	user := Builder.givenNewVerifiedUser(t)

	newLoginName := Builder.Fake.UserName()

	token, err := ApiNewResetPasswordToken(user.Email)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := ApiResetLoginCredentials(token, newLoginName, Password); err != nil {
		t.Fatalf("%v", err)
	}

	if userID, err := ApiAuthenticate(newLoginName, Password); err != nil {
		t.Fatalf("error %v", err)
	} else if userID != user.userID {
		t.Fatalf("Authenticated as wrong user. expected '%s' != actual '%s'", user.userID, userID)
	}
}
