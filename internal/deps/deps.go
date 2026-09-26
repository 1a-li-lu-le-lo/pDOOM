// Copyright NU Cybernetics. PDUM — research prototype.
// Package deps pins third-party modules used across the Go services so that
// `go mod tidy` keeps them while individual packages are still being written.
package deps

import (
	_ "github.com/mmcdole/gofeed"
	_ "github.com/santhosh-tekuri/jsonschema/v6"
	_ "github.com/temoto/robotstxt"
)
