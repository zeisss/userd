package cli

import (
	httputil ".."

	"net/http"
	"strings"
)

const (
	DEFAULT_APP_NAME = "userd"
	DEFAULT_LISTEN   = "localhost:8080"

	DEFAULT_LOG_LEVEL    = "debug"
	DEFAULT_LOG_REQUESTS = false
	DEFAULT_USE_HTTPS    = false
)

type HttpServerStarter struct {
	AppName       string
	Host          string
	Port          string
	ListenAddress string

	LogLevel    string
	LogRequests bool

	UseHttps             bool
	HttpsCertificateFile string
	HttpsKeyFile         string
}

type FlagSet interface {
	StringVar(p *string, name, defaultValue, help string)
	BoolVar(p *bool, name string, defaultValue bool, help string)
}

func NewHttpServerStarter() *HttpServerStarter {
	return &HttpServerStarter{
		AppName:       DEFAULT_APP_NAME,
		ListenAddress: DEFAULT_LISTEN,
	}
}

func NewStarterFromFlagSet(flagSet FlagSet) *HttpServerStarter {
	starter := NewHttpServerStarter()

	flagSet.StringVar(&starter.ListenAddress, "listen", DEFAULT_LISTEN, "The address to listen on.")
	flagSet.BoolVar(&starter.LogRequests, "log-requests", DEFAULT_LOG_REQUESTS, "Should requests be logged to stdout")

	flagSet.BoolVar(&starter.UseHttps, "https-enable", DEFAULT_USE_HTTPS, "Enable HTTPS listening in favor of HTTP.")
	flagSet.StringVar(&starter.HttpsCertificateFile, "https-certificate", "server.cert", "The certificate to use for SSL.")
	flagSet.StringVar(&starter.HttpsKeyFile, "https-key", "server.key", "The keyfile to use for SSL.")

	// Set host and port for convinience.
	splitted := strings.Split(starter.ListenAddress, ":")
	starter.Host = splitted[0]
	starter.Port = splitted[1]

	return starter
}

func (starter *HttpServerStarter) StartHttpInterface(handler http.Handler) {
	if starter.LogRequests {
		handler = &httputil.RequestLogger{handler}
	}

	if starter.UseHttps {
		if err := http.ListenAndServeTLS(starter.ListenAddress, starter.HttpsCertificateFile, starter.HttpsKeyFile, handler); err != nil {
			panic(err)
		}
	} else {
		if err := http.ListenAndServe(starter.ListenAddress, handler); err != nil {
			panic(err)
		}
	}
}
