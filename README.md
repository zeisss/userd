User Service
============

Microservice to create/read/update users and organize them into groups. Supports email verification and password resets.


# API

## POST /v1/user/create

+ Request 

	login_name=mr.example@acme.com&login_password=TopSecret&profile_name=Mr.%20Example&email=mr.example@acme.com

+ Response 200

	+ Headers

			Location: /v1/user/get?id=1

	+ Body

			1


## GET /v1/user/get?id={userid}

+ Response 200

		{
			"profile_name": "ZeissS",
			"email": "stephan@moinz.de"
		}