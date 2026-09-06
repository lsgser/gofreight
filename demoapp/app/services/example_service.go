package services

/*
|--------------------------------------------------------------------------
| Example Service
|--------------------------------------------------------------------------
|
| Services hold business logic and keep controllers thin. Register services
| in bootstrap/app.go and resolve them via the container:
|
|   app.Singleton("example", func() any { return services.NewExampleService() })
|
| Generate a new service:
|   gofreight make:service OrderProcessing
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
