package bill 

import (
	"encore.dev/storage/sqldb"
)


var BillDB = sqldb.NewDatabase("bill", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})