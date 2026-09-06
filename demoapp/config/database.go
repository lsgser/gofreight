package config

/*
|--------------------------------------------------------------------------
| Database Configuration
|--------------------------------------------------------------------------
|
| Optional programmatic migrations for apps that prefer Go over SQL files.
| Most apps use db/migrate/*.sql and the gofreight migrate command instead.
|
*/

import "github.com/lsgser/gofreight/database"

/*
|--------------------------------------------------------------------------
| RunMigrations
|--------------------------------------------------------------------------
|
| Runs inline SQL migrations (optional — prefer db/migrate/).
|
*/
func RunMigrations() error {
	return database.Migrate(
		/* Add migration SQL here */
	)
}
