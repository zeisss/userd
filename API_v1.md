## API

### POST /v1/user/create

Creates a new user.

Event: user.created (user_id, profile_name, email)

+ Request 

	login_name=mr.example@acme.com&login_password=TopSecret&profile_name=Mr.%20Example&email=mr.example@acme.com

+ Response 200

	+ Headers

			Location: /v1/user/get?id=1

	+ Body

			1


### GET /v1/user/get?id={userid}

+ Response 200

		{
			"profile_name": "ZeissS",
			"email": "stephan@moinz.de",
			"email_verified": false
		}

+ Response 404

### POST /v1/user/verify_email?id={userid}&email={email}

Flags the email of the user as verified. The `email` parameter is optional and can be used to ensure that the correct email
gets verified - maybe the user changed the email after the original verification email was sent out.

Event: user.email_verified (user_id, email)

+ Response 204
+ Response 400
+ Response 404

### POST /v1/user/change_email?id={userid}&email={email}

Updates the email of the user identified by `userid`.

Event: user.change_email (user_id, email)

### POST /v1/user/change_profile_name?id={userid}&profile_name={name}

Changes the profile name of the user.

Event: user.change_profile_name (user_id, profile_name)

### POST /v1/user/change_login_credentials?id={userid}&name={name}&password={password}

Updates the credentials to be used with `/authenticate`.

Event: user.change_login_credentials (user_id)

### POST /v1/user/authenticate?name={login_name}&password={login_password}

Performs an authentication with given credentials. If the credentials are valid and the user can be authenticated (e.g. is not locked), the userid will be returned.

Event: user.authenticated (user_id)

+ Response 204

		{userid}

+ Response 400
+ Response 404

### POST /v1/user/new_reset_login_credentials_token?email={email}&login_name={login_name}

Creates a new reset password token, associates it with the user and returns it. The consumer should forward this token to the user's email (or via another communication medium which is known to reach the real user) to verify that the initiator is the real user.

One of the arguments must be given, the second if optional. If both are given, a user must match both values.

Event: user.new_reset_login_credentials_token(user_id, token)

+ Response 200

		{
			"token": "{token}"
		}

+ Response 404

		Unknown user id.

### POST /v1/user/reset_login_credentials?token={token}&login_name={login_name}&login_password={login_password}

Resets the user's credentials.

+ Response 204
+ Response 400

		Invalid token.


### GET /v1/feed

Returns all collected events.

+ Response 200 (application/json)

		{"message": {"userid": "userid()", "email:" "email()"}, "timestamp": "2014-09-01T23:50:50Z+02:00", "tag": "user.created"}
		{"message": {"userid": "userid()", "email:" "email()"}, "timestamp": "2014-09-01T23:55:50Z+02:00", "tag": "user.created"}
		... more events ...


