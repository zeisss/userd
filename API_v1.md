## API

### POST /v1/user/create

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

+ Response 204
+ Response 400
+ Response 404

### POST /v1/user/change_email?id={userid}&email={email}

Updates the email of the user identified by `userid`.

### POST /v1/user/change_profile_name?id={userid}&profile_name={name}

### POST /v1/user/change_login_credentials?id={userid}&name={name}&password={password}

Updates the credentials to be used with `/authenticate`.

### POST /v1/user/authenticate?name={login_name}&password={login_password}

Performs an authentication with given credentials. If the credentials are valid and the user can be authenticated (e.g. is not locked), the userid will be returned.

+ Response 204

		{userid}

+ Response 400
+ Response 404


### GET /v1/metrics

Returns the internally collected metrics as JSON.

+ Response 200

		{
			"metricname": {
				"value": 1
			},
			"nextmetric...":...

		}
	