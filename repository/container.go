package repository

import "github.com/rs/zerolog"

type Container struct {
	UserRepository *UserRepository
}

func NewContainer(logger zerolog.Logger) *Container {
	return &Container{
		UserRepository: NewUserRepository(logger),
	}
}
