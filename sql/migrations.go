package sql

import "embed"

//go:embed schema/*
var MigrationsFS embed.FS
