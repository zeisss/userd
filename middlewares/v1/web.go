package v1

import (
	httputil "../../http"
	"../../service"
	"../../service/user"

	"github.com/gorilla/mux"
	"github.com/juju/errgo"
	metrics "github.com/rcrowley/go-metrics"

	"log"
	"net/http"
)

var (
	MaskError = errgo.MaskFunc(
		service.IsServiceError,
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

	mux.Methods("POST").Path("/v1/user/new_reset_login_credentials_token").Handler(&NewResetLoginCredentialsHandler{base})
	mux.Methods("POST").Path("/v1/user/reset_login_credentials").Handler(&ResetCredentialsTokenHandler{base})

	mux.Methods("GET").Path("/v1/feed").Handler(&FeedWriter{base})
	mux.Methods("GET").Path("/v1/metrics").Handler(&MetricsWriter{metrics.DefaultRegistry})

	return mux
}

// --------------------------------------------------------------------------------------------

type BaseHandler struct {
	UserService *service.UserService
}

func (base *BaseHandler) writeProcessingError(resp http.ResponseWriter, err error) {
	httputil.WriteJSONErrorPage(resp, http.StatusInternalServerError, "An Internal Error occured. Please try again later.")

	log.Printf("Internal error: %#v\n", err)
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
	} else if service.IsEmailAlreadyTakenError(err) || service.IsLoginNameAlreadyTakenError(err) || service.IsServiceError(err) {
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

	var response service.CreateUserResponse

	if err := h.UserService.CreateUser(service.CreateUserRequest{profileName, email, loginName, loginPassword}, &response); err != nil {
		h.handleProcessingError(resp, req, MaskError(err))
	} else {
		resp.Header().Add("location", "/v1/user/get?id="+response.UserID)
		resp.WriteHeader(http.StatusCreated)
		resp.Write([]byte(response.UserID))
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

	var response service.GetUserResponse

	if err := h.UserService.GetUser(service.GetUserRequest{userId}, &response); err != nil {
		h.handleProcessingError(resp, req, MaskError(err))
	} else {
		h.writeUser(resp, &response.User)
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

	if err := h.UserService.ChangeLoginCredentials(service.ChangeLoginCredentialsRequest{userID, newLogin, newPassword}); err != nil {
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

	if err := h.UserService.ChangeProfileName(service.ChangeProfileNameRequest{userID, newProfileName}); err != nil {
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

	if err := h.UserService.ChangeEmail(service.ChangeEmailRequest{userID, newEmail}); err != nil {
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

	var response service.AuthenticateResponse
	err := h.UserService.Authenticate(service.AuthenticateRequest{loginName, loginPassword}, &response)
	if err != nil {
		h.handleProcessingError(resp, req, MaskError(err))
	} else {
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte(response.UserID))
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
		err = h.UserService.CheckAndSetEmailVerified(service.CheckAndSetEmailVerifiedRequest{userID, email})
	} else {
		err = h.UserService.SetEmailVerified(service.SetEmailVerifiedRequest{userID})
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

// ----------------------------------------------

type NewResetLoginCredentialsHandler struct{ BaseHandler }

func (r *NewResetLoginCredentialsHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	email := req.FormValue("email")

	var response service.NewResetCredentialsTokenResponse

	err := r.UserService.NewResetLoginCredentialsToken(service.NewResetCredentialsTokenRequest{email}, &response)
	if err != nil {
		r.handleProcessingError(resp, req, err)
	} else {
		httputil.WriteJSONResponse(resp, http.StatusOK, map[string]interface{}{
			"token": response.Token,
		})
	}
}

// ----------------------------------------------

type ResetCredentialsTokenHandler struct{ BaseHandler }

func (r *ResetCredentialsTokenHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	token := req.FormValue("token")
	login_name := req.FormValue("login_name")
	login_password := req.FormValue("login_password")

	var response service.ResetCredentialsResponse

	if err := r.UserService.ResetCredentialsWithToken(service.ResetCredentialsRequest{token, login_name, login_password}, &response); err != nil {
		r.handleProcessingError(resp, req, err)
	} else {
		httputil.WriteNoContent(resp)
	}
}

// ----------------------------------------------

type FeedWriter struct{ BaseHandler }

func (h *FeedWriter) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")
	h.UserService.EventCollector.WriteJSONOnce(resp)
}

// ----------------------------------------------

type MetricsWriter struct {
	Registry metrics.Registry
}

func (m *MetricsWriter) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")
	metrics.WriteJSONOnce(m.Registry, resp)
}
