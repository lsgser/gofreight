package services

/*
|--------------------------------------------------------------------------
| Example Service
|--------------------------------------------------------------------------
|
| Implements Example Service as part of the services package in the
| Gofreight framework. Key symbols: ExampleService, NewExampleService.
| 
| Symbols defined here include: ExampleService (exported type);
| NewExampleService (/*
| |--------------------------------------------------------------------------
| | NewExampleService
| |--------------------------------------------------------------------------
| | | Creates a new ExampleService instance. | */).
| 
*/

/*
|--------------------------------------------------------------------------
| ExampleService
|--------------------------------------------------------------------------
|
| Demonstrates the service layer pattern.
|
*/
type ExampleService struct{}

/*
|--------------------------------------------------------------------------
| NewExampleService
|--------------------------------------------------------------------------
|
| Creates a new ExampleService instance.
|
*/
func NewExampleService() *ExampleService { return &ExampleService{} }
