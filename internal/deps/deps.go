// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package deps pins third-party modules used by the ingestion packages so that
// `go mod tidy` keeps them while those packages are being written.
package deps

import (
	_ "github.com/mmcdole/gofeed"
	_ "github.com/temoto/robotstxt"
)
