package embed

import "embed"

//go:embed all:default/*
var DefaultFS embed.FS

//go:embed all:font/*
var FontFS embed.FS
