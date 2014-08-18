package main

import (
	"./service"
	"./user"

	httputil "./http"

	"encoding/json"
	"log"
	"net/http"
)

func NewUserAPIHandler(userService *service.UserService) http.Handler {
	base := BaseHandler{userService}

	mux := http.NewServeMux()
	mux.Handle("/v1/user/create", httputil.EnforeMethod("POST", &CreateUserHandler{base}))
	mux.Handle("/v1/user/get", httputil.EnforeMethod("GET", &GetUserHandler{base}))
	mux.Handle("/v1/user/change_login_credentials", httputil.EnforeMethod("POST", &ChangeLoginCredentialsHandler{base}))
	mux.Handle("/v1/user/change_email", httputil.EnforeMethod("POST", &ChangeEmailHandler{base}))
	mux.Handle("/v1/user/change_profile_name", httputil.EnforeMethod("POST", &ChangeProfileNameHandler{base}))

	mux.Handle("/v1/user/authenticate", httputil.EnforeMethod("POST", &AuthenticationHandler{base}))

	mux.Handle("/v1/user/verify_email", httputil.EnforeMethod("POST", &VerifyEmailHandler{base}))

	return mux
}

// --------------------------------------------------------------------------------------------

type BaseHandler struct {
	UserService *service.UserService
}

func (base *BaseHandler) writeNotFoundError(resp http.ResponseWriter) {
	resp.WriteHeader(http.StatusNotFound)
}

func (base *BaseHandler) writeProcessingError(resp http.ResponseWriter, err error) {
	resp.WriteHeader(http.StatusInternalServerError)

	log.Printf("Internal error: %v\n", err)
}

func (base *BaseHandler) writeBadRequest(resp http.ResponseWriter, message ...string) {
	resp.WriteHeader(http.StatusBadRequest)

	for _, msg := range message {
		resp.Write([]byte(msg))
	}
}

func (base *BaseHandler) UserID(req *http.Request) (string, bool) {
	userID := req.FormValue("id")
	if userID == "" {
		return "", false
	}
	return userID, true
}

func (base *BaseHandler) handleProcessingError(resp http.ResponseWriter, req *http.Request, err error) {
	if service.IsNotFoundError(err) {
		base.writeNotFoundError(resp)
	} else if service.IsEmailAlreadyTakenError(err) || service.IsLoginNameAlreadyTakenError(err) {
		base.writeBadRequest(resp, err.Error())
	} else if err == service.InvalidCredentials {
		base.writeBadRequest(resp)
	} else if service.IsUserEmailMustBeVerifiedError(err) {
		base.writeBadRequest(resp, err.Error())
	} else {
		base.writeProcessingError(resp, err)
	}
}

// --------------------------------------------------------------------------------------------

type CreateUserHandler struct {
	BaseHandler
}

func (h *CreateUserHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	profileName := req.PostFormValue("profile_name")
	email := req.PostFormValue("email")

	loginName := req.PostFormValue("login_name")
	loginPassword := req.PostFormValue("login_password")

	userID, err := h.UserService.CreateUser(profileName, email, loginName, loginPassword)

	if err != nil {
		h.handleProcessingError(resp, req, err)
	} else {
		resp.Header().Add("location", "/v1/user/get?id="+userID)
		resp.WriteHeader(http.StatusCreated)
		resp.Write([]byte(userID))
	}
}

// -------------------------------------------

type GetUserHandler struct {
	BaseHandler
}

func (h *GetUserHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	userId, ok := h.UserID(req)
	if !ok {
		h.writeBadRequest(resp)
		return
	}

	user, err := h.UserService.GetUser(userId)
	if err != nil {
		h.handleProcessingError(resp, req, err)
	} else {
		if err := h.writeUser(resp, &user); err != nil {
			panic(err)
		}
	}
}

func (h *GetUserHandler) writeUser(resp http.ResponseWriter, theUser *user.User) error {
	result := map[string]interface{}{}
	result["profile_name"] = theUser.ProfileName
	result["email"] = theUser.Email
	result["login_name"] = theUser.LoginName
	result["email_verified"] = theUser.EmailVerified

	return json.NewEncoder(resp).Encode(result)
}

/// ----------------------------------------------

type ChangeLoginCredentialsHandler struct{ BaseHandler }

func (h *ChangeLoginCredentialsHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	userID, ok := h.UserID(req)
	if !ok {
		h.writeBadRequest(resp)
		return
	}

	newLogin := req.FormValue("name")
	if newLogin == "" {
		h.writeBadRequest(resp, "Parameter 'name' is required.")
		return
	}

	newPassword := req.FormValue("password")
	if newPassword == "" {
		h.writeBadRequest(resp)
		return
	}

	if err := h.UserService.ChangeLoginCredentials(userID, newLogin, newPassword); err != nil {
		h.handleProcessingError(resp, req, err)
	} else {
		resp.WriteHeader(http.StatusNoContent)
	}
}

// -----------------------------------------------

type ChangeProfileNameHandler struct{ BaseHandler }

func (h *ChangeProfileNameHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	userID, ok := h.UserID(req)
	if !ok {
		h.writeBadRequest(resp)
		return
	}

	newProfileName := req.FormValue("profile_name")
	if newProfileName == "" {
		h.writeBadRequest(resp)
		return
	}

	if err := h.UserService.ChangeProfileName(userID, newProfileName); err != nil {
		h.handleProcessingError(resp, req, err)
	} else {
		resp.WriteHeader(http.StatusNoContent)
	}
}

// -----------------------------------------------

type ChangeEmailHandler struct{ BaseHandler }

func (h *ChangeEmailHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	userID, ok := h.UserID(req)
	if !ok {
		h.writeBadRequest(resp)
		return
	}

	newEmail := req.FormValue("email")
	if newEmail == "" {
		h.writeBadRequest(resp)
		return
	}

	if err := h.UserService.ChangeEmail(userID, newEmail); err != nil {
		h.handleProcessingError(resp, req, err)
	} else {
		resp.WriteHeader(http.StatusNoContent)
	}
}

// -----------------------------------------------

type AuthenticationHandler struct{ BaseHandler }

func (h *AuthenticationHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	loginName := req.PostFormValue("name")
	loginPassword := req.PostFormValue("password")

	if loginName == "" || loginPassword == "" {
		h.writeBadRequest(resp)
		return
	}

	userID, err := h.UserService.Authenticate(loginName, loginPassword)
	if err != nil {
		h.handleProcessingError(resp, req, err)
	} else {
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte(userID))
	}

}

// ----------------------------------------------
type VerifyEmailHandler struct{ BaseHandler }

func (h *VerifyEmailHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	userID, ok := h.UserID(req)
	if !ok {
		h.writeBadRequest(resp, "No id parameter given.")
		return
	}

	email, emailGiven := h.Email(req)

	var err error
	if emailGiven {
		err = h.UserService.CheckAndSetEmailVerified(userID, email)
	} else {
		err = h.UserService.SetEmailVerified(userID)
	}

	if err != nil {
		h.handleProcessingError(resp, req, err)
	} else {
		resp.WriteHeader(http.StatusNoContent)
	}
}

func (h *VerifyEmailHandler) Email(req *http.Request) (string, bool) {
	email, ok := req.Form["email"]
	if !ok || len(email) == 0 || email[0] == "" {
		return "", false
	}
	return email[0], true
}
