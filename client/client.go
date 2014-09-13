package client

import (
	"github.com/juju/errgo"
)

//
// endpoint = "http://userd.acme.com/"
func Dial(endpoint string) *Client {
	return &Client{endpoint}
}

type Client struct {
	Endpoint string
}

func (c *Client) endpoint(action string) string {
	return c.Endpoint + action
}

func (c *Client) GetUser(userID string) (ApiUser, error) {
	var result ApiUser

	_, err := Execute(Endpoint("get"), GetUserCall{JsonCall{&result}, userID})
	if err != nil {
		return result, errgo.Mask(err)
	}
	return result, nil
}

// Creates a user and returns the user_id.
func (c *Client) CreateUser(profileName, email, loginName, loginPassword string) (string, error) {
	userID, err := Execute(c.endpoint("v1/user/create"), CreateUserCall{profileName, email, loginName, loginPassword})
	if err != nil {
		return "", errgo.Mask(err)
	}
	return userID.(string), nil
}

func (c *Client) Authenticate(loginName, loginPassword string) (string, error) {
	userID, err := Execute(c.endpoint("v1/user/authenticate"), AuthenticateCall{Name: loginName, Password: loginPassword})
	if err != nil {
		return "", errgo.Mask(err)
	}
	return userID.(string), nil
}

func (c *Client) VerifyEmail(userID string) error {
	_, err := Execute(c.endpoint("v1/user/verify_email"), VerifyEmailCall{UserID: userID})
	return errgo.Mask(err)
}

func (c *Client) ChangeProfileName(userID, profileName string) error {
	_, err := Execute(c.endpoint("v1/user/change_profile_name"), ChangeProfileNameCall{ID: userID, ProfileName: profileName})
	return errgo.Mask(err)
}

func (c *Client) ChangeLoginCredentials(userID, name, password string) error {
	_, err := Execute(c.endpoint("v1/user/change_login_credentials"), ChangeLoginCredentialsCall{ID: userID, Login: name, Password: password})
	return errgo.Mask(err)
}

func (c *Client) ChangeEmail(userID, newEmail string) error {
	_, err := Execute(c.endpoint("v1/user/change_email"), ChangeEmailCall{ID: userID, Email: newEmail})
	return errgo.Mask(err)
}

func (c *Client) NewResetPasswordToken(email string) (string, error) {
	token, err := Execute(c.endpoint("v1/user/new_reset_login_credentials_token"), NewResetPasswordToken{email})
	if err != nil {
		return "", errgo.Mask(err)
	}
	return token.(string), nil
}

func (c *Client) ResetLoginCredentials(token, login_name, login_password string) error {
	_, err := Execute(Endpoint("v1/user/reset_login_credentials"), ResetLoginCredentials{token, login_name, login_password})
	return errgo.Mask(err)
}
