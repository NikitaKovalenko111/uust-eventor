package user_provider

type UserProvider interface {
	GetUser(id int) error
	CreateUser()
}
