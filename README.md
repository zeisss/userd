User Service
============

Microservice to create/read/update and authenticate users.

## Running
### Installation

This repository is `go get`able. 

```
GOPATH=$(pwd) go get github.com/ZeissS/userd
./bin/userd
```

### Configuration

Userd is completly configurable via command line arguments. Call `userd --help` to see a list of options or checkout `main.go`.

## Usage

### About Users

The user object currently consists of only a few fields:

 * `profile_name` - A name that the user should be able to give himself. 
 * `email` - The email of the user
 * `email_verified` - Has the email of the user already been verified to work?
 * `login_name` and `login_password_hash` - The login credentials needed for `Authenticate()`

The `email` and `login_name` must each be unique among all users. 

If the consumer wants to use the email as the login_name, it must be provided separately for each field. The consumer is responsible for updating both fields (see the API), if the email changes.

### Email Verification

When creating a new user, the email is considered 'unverified'. Based on the `--auth-email` command line arguments,
this might be required for authentication to work. To verify an email, a separate call to `/verify_email` is needed.

### Password Encryption

Passwords are hashed using the `code.google.com/p/go.crypto/bcrypt` library before storing.

## API

See `API_v1.md` for the current old-school interface. For V2 we will make this a bit more REST like. Comming soon.