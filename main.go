package main

import (
	"./service"
	"./service/events"
	"./service/hasher"
	"./service/idfactory"
	"./service/storage"

	"flag"
	"net/http"
	"os"
)

var (
	// Service/Logic
	listenAddress = flag.String("listen", "localhost:8080", "The address to listen on.")
	authEmail     = flag.Bool("auth-email", true, "Must the email adress be verified for an authentication to succeed.")

	// Frontend - HTTP
	httpsUse             = flag.Bool("https-enable", false, "Enable HTTPS listening in favor of HTTP.")
	httpsCertificateFile = flag.String("https-certificate", "server.cert", "The certificate to use for SSL.")
	httpsKeyFile         = flag.String("https-key", "server.key", "The keyfile to use for SSL.")

	// Backend Switches
	backendEventLog = flag.String("eventlog", "none", "Should events be logger? log or none")

	// Backend - Hasher
	hasherBcryptCost = flag.Int("hasher-bcrypt-cost", hasher.BcryptDefaultCost, "The cost to apply when hashing new passwords.")

	// Backend - EventLog
	/// Log
	eventLogFile = flag.String("event-log-file", "-", "Where to write the eventlog. - for stdout.")
	eventLogMode = flag.Uint("event-log-mode", 0700, "Mode to create logfile with - defaults to 0700.")

	/// Cores
	eventCoresUrl = flag.String("event-cores-url", "amqp://guest:guest@localhost", "An amqp url to connect to.")
)

func UserStorage() service.UserStorage {
	return storage.NewLocalStorage()
}

func IdFactory() service.IdFactory {
	return &idfactory.UUIDFactory{}
}

func PasswordHasher() service.PasswordHasher {
	return hasher.NewBcryptHasher(*hasherBcryptCost)
}

func EventLog() service.EventLog {
	switch *backendEventLog {
	case "cores":
		return events.NewCoresAmqpEventLog(*eventCoresUrl)
	case "log":
		var err error
		out := os.Stdout
		if *eventLogFile != "-" {
			out, err = os.OpenFile(*eventLogFile, os.O_WRONLY, os.FileMode(*eventLogMode))
			if err != nil {
				panic(err)
			}
		}
		return events.NewLogStreamEventLog(out)
	case "none":
		return events.NewNoneEventLog()
	default:
		panic("Unknown -eventlog value: " + *backendEventLog)
	}
}

func StartHttpInterface(userService *service.UserService) {
	base := BaseHandler{userService}
	http.Handle("/v1/user/create", EnforeMethod("POST", &CreateUserHandler{base}))
	http.Handle("/v1/user/get", EnforeMethod("GET", &GetUserHandler{base}))
	http.Handle("/v1/user/change_login_credentials", EnforeMethod("POST", &ChangeLoginCredentialsHandler{base}))
	http.Handle("/v1/user/change_email", EnforeMethod("POST", &ChangeEmailHandler{base}))
	http.Handle("/v1/user/change_profile_name", EnforeMethod("POST", &ChangeProfileNameHandler{base}))

	http.Handle("/v1/user/authenticate", EnforeMethod("POST", &AuthenticationHandler{base}))

	http.Handle("/v1/user/verify_email", EnforeMethod("POST", &VerifyEmailHandler{base}))

	if *httpsUse {
		if err := http.ListenAndServeTLS(*listenAddress, *httpsCertificateFile, *httpsKeyFile, nil); err != nil {
			panic(err)
		}
	} else {
		if err := http.ListenAndServe(*listenAddress, nil); err != nil {
			panic(err)
		}
	}
}

func main() {
	flag.Parse()

	idFactory := IdFactory()
	hasher := PasswordHasher()
	userStorage := UserStorage()
	eventLog := EventLog()

	userService := service.UserService{
		service.Dependencies{idFactory, hasher, userStorage, eventLog},
		service.Config{*authEmail},
	}

	StartHttpInterface(&userService)
}
