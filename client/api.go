package client

import (
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
	ProfileName string `json:"profile_name"`
	LoginName   string `json:"login_name"`
	Email       string `json:"email"`
}

func ApiGetUser(userID string) (ApiUser, error) {
	var result ApiUser

	params := url.Values{}
	params.Add("id", userID)

	resp, err := getAndExpect("get", params, http.StatusOK)
	if err != nil {
		return result, err
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func getAndExpect(action string, params url.Values, expectedStatusCode int) (*http.Response, error) {
	resp, err := http.Get(Endpoint(action) + "?" + params.Encode())
	if err != nil {
		return resp, err
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
		return "", err
	}
	return string(data), nil
}
