package config

/*
|--------------------------------------------------------------------------
| Database
|--------------------------------------------------------------------------
|
| Implements Database as part of the config package in the Gofreight
| framework. Key symbols: RunMigrations.
| 
| Symbols defined here include: RunMigrations (/*
| |--------------------------------------------------------------------------
| | RunMigrations
| |--------------------------------------------------------------------------
| | | Runs inline SQL migrations (optional — prefer db/migrate/). | */).
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
