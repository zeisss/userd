package v2

import (
	"encoding/json"
	"net/http"

	"github.com/juju/errgo"

	servicePkg "../../service"
	apiSchemaPkg "github.com/catalyst-zero/api-schema"
	middlewarePkg "github.com/catalyst-zero/middleware-server"
)

const (
	MSG_INVALID_USER_OR_MAIL = "Cannot log in. Invalid user or mail given. Aborting..."
	MSG_INVALID_PASSWORD     = "Cannot log in. Invalid password given. Aborting..."
)

func respondErrorInvalidCredentials(ctx *middlewarePkg.Context) error {
	return ctx.Response.Json(apiSchemaPkg.StatusResourceInvalidCredentials(), http.StatusInternalServerError)
}

func (v2 *V2) CreateUser(res http.ResponseWriter, req *http.Request, ctx *middlewarePkg.Context) error {
	var payload map[string]string
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return errgo.Newf("Invalid request body. Expecting json. Aborting...")
	}
	loginName := payload["username"]
	email := payload["email"]
	loginPassword := payload["password"]
	profileName := loginName

	userID, err := v2.UserService.CreateUser(profileName, email, loginName, loginPassword)
	if err != nil {
		return errgo.Mask(err)
	}

	return ctx.Response.Json(apiSchemaPkg.StatusData(userID), http.StatusOK)
}

func (v2 *V2) LoginUser(res http.ResponseWriter, req *http.Request, ctx *middlewarePkg.Context) error {
	// Get user or mail.
	userOrMail := ctx.MuxVars["userOrMail"]

	// Get password from req body.
	var payload map[string]string
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return errgo.Newf("Invalid request body. Expecting json. Aborting...")
	}
	password := payload["password"]

	// Authenticate.
	userID, err := v2.UserService.Authenticate(userOrMail, password)
	if servicePkg.IsNotFoundError(err) {
		return ctx.Response.Json(apiSchemaPkg.StatusResourceNotFound(), http.StatusInternalServerError)
	} else if servicePkg.IsInvalidArguments(err) {
		return respondErrorInvalidCredentials(ctx)
	} else if servicePkg.IsInvalidCredentials(err) {
		return respondErrorInvalidCredentials(ctx)
	} else if err != nil {
		return errgo.Mask(err)
	}

	// Respond user id.
	return ctx.Response.Json(apiSchemaPkg.StatusData(userID), http.StatusOK)
}
