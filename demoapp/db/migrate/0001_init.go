package migrate

/*
|--------------------------------------------------------------------------
| 0001 Init
|--------------------------------------------------------------------------
|
| Implements 0001 Init as part of the migrate package in the Gofreight
| framework.
| 
*/

import (
	"context"

	"github.com/lsgser/gofreight/database"
)

func init() {
	database.RegisterMigration("0001_init", up0001Init, down0001Init)
}

// up0001Init is a bootstrap migration — confirms the migration runner is wired.
func up0001Init(ctx context.Context) error {
	return nil
}

func down0001Init(ctx context.Context) error {
	return nil
}
