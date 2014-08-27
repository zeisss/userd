package cli

import (
	httputil ".."

	"log"
	"net/http"
)

func LogRecord(req *http.Request, recorder httputil.ResponseRecorder) {
	log.Printf("%s %s \"%s\" %d %d\n",
		req.Method,
		req.URL,
		req.UserAgent(),
		recorder.StatusCode,
		recorder.TimeTaken,
	)
}

type HttpServerStarter struct {
	ListenAddress string
	LogRequests   bool

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
		ListenAddress: "localhost:8080",
	}
}

func NewStarterFromFlagSet(flagSet FlagSet) *HttpServerStarter {
	starter := NewHttpServerStarter()
	flagSet.StringVar(&starter.ListenAddress, "listen", "localhost:8080", "The address to listen on.")
	flagSet.BoolVar(&starter.LogRequests, "log-requests", false, "Should requests be logged to stdout")

	flagSet.BoolVar(&starter.UseHttps, "https-enable", false, "Enable HTTPS listening in favor of HTTP.")
	flagSet.StringVar(&starter.HttpsCertificateFile, "https-certificate", "server.cert", "The certificate to use for SSL.")
	flagSet.StringVar(&starter.HttpsKeyFile, "https-key", "server.key", "The keyfile to use for SSL.")
	return starter
}

func (starter *HttpServerStarter) StartHttpInterface(handler http.Handler) {
	if starter.LogRequests {
		handler = &httputil.RequestLogger{handler, LogRecord}
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
