package client

import (
	"github.com/juju/errgo"

	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var UnexpectedStatusCode = errors.New("Service returned unexpected status code.")

var endpoint = "http://localhost:8080/v1/user/"

func SetEndpoint(url string) {
	endpoint = url
}

func Endpoint(action string) string {
	return endpoint + action
}

func ApiCreateUser(profileName, email, loginName, loginPassword string) (string, error) {
	params := url.Values{}
	params.Add("profile_name", profileName)
	params.Add("email", email)
	params.Add("login_name", loginName)
	params.Add("login_password", loginPassword)

	return postFormAndExpectAndReturnBodyString("create", params, http.StatusCreated)
}

type ApiUser struct {
	ProfileName   string `json:"profile_name"`
	LoginName     string `json:"login_name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func ApiGetUser(userID string) (ApiUser, error) {
	var result ApiUser

	params := url.Values{}
	params.Add("id", userID)

	resp, err := getAndExpect("get", params, http.StatusOK)
	if err != nil {
		return result, errgo.Mask(err)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func getAndExpect(action string, params url.Values, expectedStatusCode int) (*http.Response, error) {
	resp, err := http.Get(Endpoint(action) + "?" + params.Encode())
	if err != nil {
		return resp, errgo.Mask(err)
	}
	if resp.StatusCode != expectedStatusCode {
		log.Printf("URL 'GET %s' returned code %d, expected %d", Endpoint(action), resp.StatusCode, expectedStatusCode)
		return resp, UnexpectedStatusCode
	}
	return resp, nil
}

func postFormAndExpectAndReturnBodyString(action string, params url.Values, expectedStatusCode int) (string, error) {
	resp, err := http.PostForm(Endpoint(action), params)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != expectedStatusCode {
		log.Printf("URL 'POST %s' returned code %d, expected %d", Endpoint(action), resp.StatusCode, expectedStatusCode)
		return "", UnexpectedStatusCode
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errgo.Mask(err)
	}
	return string(data), nil
}

// --------------------

type JsonCall struct {
	Target interface{}
}

func (c JsonCall) ResponseOK(resp *http.Response) (interface{}, error) {
	if err := json.NewDecoder(resp.Body).Decode(c.Target); err != nil {
		return c.Target, errgo.Mask(err)
	}
	return c.Target, nil
}

// ------------------------

type BodyReader struct{}

func (c BodyReader) ResponseOK(resp *http.Response) (interface{}, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errgo.Mask(err)
	}
	return string(data), nil
}

func (call BodyReader) ResponseBadRequest(resp *http.Response) (interface{}, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errgo.Mask(err)
	}
	return "", errors.New(string(data))
}

// ------------------------

func ApiVerifyEmail(userID string) error {
	_, err := Execute(Endpoint("verify_email"), VerifyEmailCall{UserID: userID})
	return errgo.Mask(err)
}

type VerifyEmailCall struct {
	UserID string
	Email  string
}

func (call VerifyEmailCall) PostForm() url.Values {
	p := url.Values{}
	p.Set("id", call.UserID)
	if call.Email != "" {
		p.Set("email", call.Email)
	}
	return p
}

func (call VerifyEmailCall) ResponseNoContent(resp *http.Response) (interface{}, error) {
	return nil, nil
}

// ------------------------

func ApiAuthenticate(loginName, loginPassword string) (string, error) {
	userID, err := Execute(Endpoint("authenticate"), AuthenticateCall{Name: loginName, Password: loginPassword})
	if err != nil {
		return "", errgo.Mask(err)
	}
	return userID.(string), nil
}

type AuthenticateCall struct {
	BodyReader
	Name     string
	Password string
}

func (call AuthenticateCall) PostForm() url.Values {
	p := url.Values{}
	p.Set("name", call.Name)
	p.Set("password", call.Password)
	return p
}

// ------------------------

func ApiChangeProfileName(userID, profileName string) error {
	_, err := Execute(Endpoint("change_profile_name"), ChangeProfileNameCall{ID: userID, ProfileName: profileName})
	return errgo.Mask(err)
}

type ChangeProfileNameCall struct {
	ID          string
	ProfileName string
}

func (call ChangeProfileNameCall) PostForm() url.Values {
	p := url.Values{}
	p.Set("id", call.ID)
	p.Set("profile_name", call.ProfileName)
	return p
}

func (call ChangeProfileNameCall) ResponseNoContent(resp *http.Response) (interface{}, error) {
	return nil, nil
}

// ------------------------

func ApiChangeEmail(userID, newEmail string) error {
	_, err := Execute(Endpoint("change_email"), ChangeEmailCall{ID: userID, Email: newEmail})
	return errgo.Mask(err)
}

type ChangeEmailCall struct {
	ID    string
	Email string
}

func (call ChangeEmailCall) PostForm() url.Values {
	p := url.Values{}
	p.Set("id", call.ID)
	p.Set("email", call.Email)
	return p
}

func (call ChangeEmailCall) ResponseNoContent(resp *http.Response) (interface{}, error) {
	return nil, nil
}

// ------------------------

func ApiChangeLoginCredentials(userID, name, password string) error {
	_, err := Execute(Endpoint("change_login_credentials"), ChangeLoginCredentialsCall{ID: userID, Login: name, Password: password})
	return errgo.Mask(err)
}

type ChangeLoginCredentialsCall struct {
	ID       string
	Login    string
	Password string
}

func (call ChangeLoginCredentialsCall) PostForm() url.Values {
	p := url.Values{}
	p.Set("id", call.ID)
	p.Set("name", call.Login)
	p.Set("password", call.Password)
	return p
}

func (call ChangeLoginCredentialsCall) ResponseNoContent(resp *http.Response) (interface{}, error) {
	return nil, nil
}

// ------------------------

func ApiNewResetPasswordToken(email, login_name string) (string, error) {
	token, err := Execute(Endpoint("new_reset_login_credentials_token"), NewResetPasswordToken{email, login_name})
	if err != nil {
		return "", errgo.Mask(err)
	}
	return token.(string), nil
}

type NewResetPasswordToken struct {
	Email     string
	LoginName string
}

func (call NewResetPasswordToken) PostForm() url.Values {
	p := url.Values{}
	p.Set("login_name", call.LoginName)
	p.Set("email", call.Email)
	return p
}

func (call NewResetPasswordToken) ResponseOK(resp *http.Response) (interface{}, error) {
	target := map[string]interface{}{}

	if err := json.NewDecoder(resp.Body).Decode(&target); err != nil {
		return "", errgo.Mask(err)
	}
	return target["token"], nil
}

// ------------------------

func ApiResetLoginCredentials(token, login_name, login_password string) error {
	_, err := Execute(Endpoint("reset_login_credentials"), ResetLoginCredentials{token, login_name, login_password})
	return errgo.Mask(err)
}

type ResetLoginCredentials struct {
	Token         string
	LoginName     string
	LoginPassword string
}

func (call ResetLoginCredentials) PostForm() url.Values {
	p := url.Values{}
	p.Set("login_name", call.LoginName)
	p.Set("token", call.Token)
	p.Set("login_password", call.LoginPassword)
	return p
}

func (call ResetLoginCredentials) ResponseNoContent(resp *http.Response) (interface{}, error) {
	return nil, nil
}

// ------------------------
