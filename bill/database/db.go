package database

import (
	"encore.dev/storage/sqldb"
)

var BillDB = sqldb.NewDatabase("bill", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// encore db reset by the service name to reset db
