package jobs

/*
|--------------------------------------------------------------------------
| Queueable Jobs
|--------------------------------------------------------------------------
|
| Jobs handle asynchronous work — sending email, processing uploads, calling
| external APIs. Dispatch from controllers; process with queue:work.
|
| Generate a job:
|   gofreight make:job SendNewsletter
|
| Run the worker (requires REDIS_URL when QUEUE_DRIVER=redis):
|   gofreight queue:work
|
*/
