package main

import (
	"./service"
	"./service/hasher"
	"./service/idfactory"
	"./service/storage"

	"flag"
	"net/http"
)

var (
	listenAddress = flag.String("listen", "localhost:8080", "The address to listen on.")

	hasherBcryptCost = flag.Int("hasher-bcrypt-cost", hasher.BcryptDefaultCost, "The cost to apply when hashing new passwords.")
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

func main() {
	flag.Parse()

	idFactory := IdFactory()
	hasher := PasswordHasher()
	userStorage := UserStorage()

	userService := service.UserService{idFactory, hasher, userStorage}

	base := BaseHandler{&userService}
	http.Handle("/v1/user/create", EnforeMethod("POST", &CreateUserHandler{base}))
	http.Handle("/v1/user/get", EnforeMethod("GET", &GetUserHandler{base}))
	http.Handle("/v1/user/change_password", EnforeMethod("POST", &ChangePasswordHandler{base}))
	http.Handle("/v1/user/change_email", EnforeMethod("POST", &ChangeEmailHandler{base}))
	http.Handle("/v1/user/change_profile_name", EnforeMethod("POST", &ChangeProfileNameHandler{base}))

	http.Handle("/v1/user/authenticate", EnforeMethod("POST", &AuthenticationHandler{base}))

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		panic(err)
	}
}
