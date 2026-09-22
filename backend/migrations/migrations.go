package migrations

import _ "embed"

// InitSQL contains the idempotent schema needed by the service.
//
//go:embed 001_init.sql
var InitSQL string
