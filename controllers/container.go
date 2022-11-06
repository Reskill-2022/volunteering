package controllers

import "github.com/rs/zerolog"

type Container struct {
	UserController *UserController
}

func NewContainer(logger zerolog.Logger) *Container {
	return &Container{
		UserController: NewUserController(logger),
	}
}
