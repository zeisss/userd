package user

type User struct {
	ID string

	ProfileName string

	LoginName         string
	LoginPasswordHash string

	Email         string
	EmailVerified bool
}
