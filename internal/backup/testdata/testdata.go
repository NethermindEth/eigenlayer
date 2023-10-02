package testdata

import "embed"

//go:embed all:*
var TestData embed.FS
