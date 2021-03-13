package main

import "github.com/kelseyhightower/envconfig"

const appID = "orderservice"

type config struct {
	ServeRESTAddress string `envconfig:"servr_rest_address" default:":8000"`
	DBType string `envconfig:"database_type" default:"mysql"`
	DBName string `envconfig:"database_name" default:"/cafe_test"`
	DBUsername string `envconfig:"database_username" default:"root"`
	DBPassword string `envconfig:"database_password" default:"Future1994!)"`
}

func parseEnv() (*config, error) {
	c := new(config)
	if err := envconfig.Process(appID, c); err != nil {
		return nil, err
	}
	return c, nil
}
