User Service
============

Microservice to create/read/update users and organize them into groups. Supports email verification and password resets.

## Email Verification

When creating a new user, the email is considered 'unverified'. Based on the `--auth-email` command line arguments,
this might be required for authentication. To verify an email, a separate call to `/verify_email` is needed.


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
