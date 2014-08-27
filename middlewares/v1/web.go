package v1

import (
	httputil "../../http"
	"../../service"
	"../../service/user"

	"github.com/gorilla/mux"
	"github.com/juju/errgo"

	"log"
	"net/http"
)

var (
	MaskError = errgo.MaskFunc(
		service.IsNotFoundError, service.IsEmailAlreadyTakenError,
		service.IsLoginNameAlreadyTakenError, service.IsUserEmailMustBeVerifiedError,
	)
)

func NewUserAPIHandler(userService *service.UserService) http.Handler {
	base := BaseHandler{userService}

	mux := mux.NewRouter()
	mux.Methods("POST").Path("/v1/user/create").Handler(&CreateUserHandler{base})
	mux.Methods("GET").Path("/v1/user/get").Handler(&GetUserHandler{base})
	mux.Methods("POST").Path("/v1/user/change_login_credentials").Handler(&ChangeLoginCredentialsHandler{base})
	mux.Methods("POST").Path("/v1/user/change_email").Handler(&ChangeEmailHandler{base})
	mux.Methods("POST").Path("/v1/user/change_profile_name").Handler(&ChangeProfileNameHandler{base})
	mux.Methods("POST").Path("/v1/user/verify_email").Handler(&VerifyEmailHandler{base})

	mux.Methods("POST").Path("/v1/user/authenticate").Handler(&AuthenticationHandler{base})

	return mux
}

// --------------------------------------------------------------------------------------------

type BaseHandler struct {
	UserService *service.UserService
}

func (base *BaseHandler) writeProcessingError(resp http.ResponseWriter, err error) {
	httputil.WriteJSONErrorPage(resp, http.StatusInternalServerError, "An Internal Error occured. Please try again later.")

	log.Printf("Internal error: %v\n", err)
}

func (base *BaseHandler) UserID(req *http.Request) (string, bool) {
	userID := req.FormValue("id")
	if userID == "" {
		return "", false
	}
	return userID, true
}

func (base *BaseHandler) handleProcessingError(resp http.ResponseWriter, req *http.Request, err error) {
	err = errgo.Cause(err)
	if service.IsNotFoundError(err) {
		httputil.WriteNotFound(resp)
	} else if service.IsEmailAlreadyTakenError(err) || service.IsLoginNameAlreadyTakenError(err) {
		httputil.WriteBadRequest(resp, req, err.Error())
	} else if err == service.InvalidCredentials {
		httputil.WriteBadRequest(resp, req)
	} else if service.IsUserEmailMustBeVerifiedError(err) {
		httputil.WriteBadRequest(resp, req, err.Error())
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
		h.handleProcessingError(resp, req, MaskError(err))
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
		httputil.WriteBadRequest(resp, req)
		return
	}

	user, err := h.UserService.GetUser(userId)
	if err != nil {
		h.handleProcessingError(resp, req, MaskError(err))
	} else {
		h.writeUser(resp, &user)
	}
}

func (h *GetUserHandler) writeUser(resp http.ResponseWriter, theUser *user.User) {
	result := map[string]interface{}{}
	result["profile_name"] = theUser.ProfileName
	result["email"] = theUser.Email
	result["login_name"] = theUser.LoginName
	result["email_verified"] = theUser.EmailVerified

	httputil.WriteJSONResponse(resp, http.StatusOK, result)
}

/// ----------------------------------------------

type ChangeLoginCredentialsHandler struct{ BaseHandler }

func (h *ChangeLoginCredentialsHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	userID, ok := h.UserID(req)
	if !ok {
		httputil.WriteBadRequest(resp, req)
		return
	}

	newLogin := req.FormValue("name")
	if newLogin == "" {
		httputil.WriteBadRequest(resp, req, "Parameter 'name' is required.")
		return
	}

	newPassword := req.FormValue("password")
	if newPassword == "" {
		httputil.WriteBadRequest(resp, req)
		return
	}

	if err := h.UserService.ChangeLoginCredentials(userID, newLogin, newPassword); err != nil {
		h.handleProcessingError(resp, req, MaskError(err))
	} else {
		resp.WriteHeader(http.StatusNoContent)
	}
}

// -----------------------------------------------

type ChangeProfileNameHandler struct{ BaseHandler }

func (h *ChangeProfileNameHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	userID, ok := h.UserID(req)
	if !ok {
		httputil.WriteBadRequest(resp, req)
		return
	}

	newProfileName := req.FormValue("profile_name")
	if newProfileName == "" {
		httputil.WriteBadRequest(resp, req)
		return
	}

	if err := h.UserService.ChangeProfileName(userID, newProfileName); err != nil {
		h.handleProcessingError(resp, req, MaskError(err))
	} else {
		httputil.WriteNoContent(resp)
	}
}

// -----------------------------------------------

type ChangeEmailHandler struct{ BaseHandler }

func (h *ChangeEmailHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	userID, ok := h.UserID(req)
	if !ok {
		httputil.WriteBadRequest(resp, req)
		return
	}

	newEmail := req.FormValue("email")
	if newEmail == "" {
		httputil.WriteBadRequest(resp, req)
		return
	}

	if err := h.UserService.ChangeEmail(userID, newEmail); err != nil {
		h.handleProcessingError(resp, req, MaskError(err))
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
		httputil.WriteBadRequest(resp, req)
		return
	}

	userID, err := h.UserService.Authenticate(loginName, loginPassword)
	if err != nil {
		h.handleProcessingError(resp, req, MaskError(err))
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
		httputil.WriteBadRequest(resp, req, "No id parameter given.")
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
		h.handleProcessingError(resp, req, MaskError(err))
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
