// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/weavatar/weavatar/internal/app"
	"github.com/weavatar/weavatar/internal/bootstrap"
	"github.com/weavatar/weavatar/internal/route"
	"github.com/weavatar/weavatar/internal/service"
)

import (
	_ "time/tzdata"
)

// Injectors from wire.go:

// initCli init command line.
func initCli() (*app.Cli, error) {
	cliService := service.NewCliService()
	cli := route.NewCli(cliService)
	command := bootstrap.NewCli(cli)
	appCli := app.NewCli(command)
	return appCli, nil
}
