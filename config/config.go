package config

import (
	"fmt"
	"os"
)

const (
	Port            = "PORT"
	ClientID        = "CLIENT_ID"
	ClientSecret    = "CLIENT_SECRET"
	ServiceAccount1 = "SERVICE_ACCOUNT_1"
	ServiceAccount2 = "SERVICE_ACCOUNT_2"
)

type Environment map[string]string

func New() (Environment, error) {
	env := make(Environment)
	for _, key := range []string{
		Port,
		ClientID,
		ClientSecret,
		ServiceAccount1,
		ServiceAccount2,
	} {
		v, ok := os.LookupEnv(key)
		if !ok {
			return nil, fmt.Errorf("can't find '%s' in environment", key)
		}
		env[key] = v
	}
	return env, nil
}
