package main

import (
	"./service"
	"./user"

	"encoding/json"
	"log"
	"net/http"
)

// --------------------------------------------------------------------------------------------

type HTTPMethodCheckWrapper struct {
	AllowedMethod string
	Next          http.Handler
}

func EnforeMethod(method string, handler http.Handler) http.Handler {
	return &HTTPMethodCheckWrapper{method, handler}
}

func (h *HTTPMethodCheckWrapper) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method != h.AllowedMethod {
		resp.WriteHeader(http.StatusMethodNotAllowed)

		resp.Write([]byte("Method not allowed.\n"))
	} else {
		h.Next.ServeHTTP(resp, req)
	}
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
		if service.IsEmailAlreadyTakenError(err) || service.IsLoginNameAlreadyTakenError(err) {
			h.writeBadRequest(resp, err.Error())
		} else {
			h.writeProcessingError(resp, err)
		}
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
		if service.IsNotFoundError(err) {
			h.writeNotFoundError(resp)
		} else {
			h.writeProcessingError(resp, err)
		}
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
		if service.IsNotFoundError(err) {
			h.writeNotFoundError(resp)
		} else {
			h.writeProcessingError(resp, err)
		}
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
		if service.IsNotFoundError(err) {
			h.writeNotFoundError(resp)
		} else {
			h.writeProcessingError(resp, err)
		}
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
		if service.IsNotFoundError(err) {
			h.writeNotFoundError(resp)
		} else {
			h.writeProcessingError(resp, err)
		}
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
		if service.IsNotFoundError(err) {
			h.writeNotFoundError(resp)
		} else if err == service.InvalidCredentials {
			h.writeBadRequest(resp)
		} else if service.IsUserEmailMustBeVerifiedError(err) {
			h.writeBadRequest(resp, err.Error())
		} else {
			h.writeProcessingError(resp, err)
		}
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
		if service.IsNotFoundError(err) {
			h.writeNotFoundError(resp)
		} else {
			h.writeProcessingError(resp, err)
		}
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
