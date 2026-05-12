package user_service

import user_storage "eventor/internal/user/storage/repositories/user"

type UserService struct {
	UserRepo *user_storage.UserRepo
	// Here are repositories associated with this service
}

func Init(UserRepo *user_storage.UserRepo) *UserService {
	return &UserService{
		UserRepo: UserRepo,
	}
}
