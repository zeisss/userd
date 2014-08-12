package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Call interface {
}

type QueryParams interface {
	QueryParams() url.Values
}

type PostForm interface {
	// Body returns the content-type and stream for the POST body to for this call.
	Body() (string, io.Reader)
}

type PostCallForm interface {
	PostForm() url.Values
}

// ----------------------------------------
type OKHandler interface {
	ResponseOK(resp *http.Response) (interface{}, error)
}

type NoContentHandler interface {
	ResponseNoContent(resp *http.Response) (interface{}, error)
}

type CreatedHandler interface {
	ResponseCreated(resp *http.Response) (interface{}, error)
}

type NotFoundHandler interface {
	ResponseNotFound(resp *http.Response) (interface{}, error)
}

type BadRequestHandler interface {
	ResponseBadRequest(resp *http.Response) (interface{}, error)
}

type FallbackHandler interface {
	// HandleFallback is called when the Call object does not implement the appropriate response handler.
	HandleFallback(response *http.Response) (interface{}, error)
}

type RequestErrorHandler interface {
	HandleRequestError(err error) (interface{}, error)
}

// -----------------------------------------

func Execute(url string, call Call) (interface{}, error) {
	var method string
	var response *http.Response
	var httpErr error

	if c, ok := call.(QueryParams); ok {
		url = url + "?" + c.QueryParams().Encode()
	}

	if c, ok := call.(PostCallForm); ok {
		method = "POST"
		response, httpErr = http.PostForm(url, c.PostForm())
	} else if c, ok := call.(PostForm); ok {
		method = "POST"
		contentType, body := c.Body()
		response, httpErr = http.Post(url, contentType, body)
	} else {
		method = "GET"
		response, httpErr = http.Get(url)
	}

	// If the call itself failed, call the RequestErrorHandler or just return directly
	if httpErr != nil {
		if c, ok := call.(RequestErrorHandler); ok {
			return c.HandleRequestError(httpErr)
		} else {
			return nil, httpErr
		}
	}
	defer response.Body.Close()

	// Now, call supported handlers

	switch response.StatusCode {
	case http.StatusOK:
		if c, ok := call.(OKHandler); ok {
			return c.ResponseOK(response)
		}
	case http.StatusCreated:
		if c, ok := call.(CreatedHandler); ok {
			return c.ResponseCreated(response)
		}
	case http.StatusNoContent:
		if c, ok := call.(NoContentHandler); ok {
			return c.ResponseNoContent(response)
		}
	case http.StatusNotFound:
		if c, ok := call.(NotFoundHandler); ok {
			return c.ResponseNotFound(response)
		}
	case http.StatusBadRequest:
		if c, ok := call.(BadRequestHandler); ok {
			return c.ResponseBadRequest(response)
		}
	default:
	}

	if c, ok := call.(FallbackHandler); ok {
		return c.HandleFallback(response)
	}
	return nil, fmt.Errorf("No handler found for status code %d for URL %s %s", response.StatusCode, method, url)

}

// ------------------
