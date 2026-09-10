package integrations

/*
|--------------------------------------------------------------------------
| Errors
|--------------------------------------------------------------------------
|
| Implements Errors as part of the integrations package in the Gofreight
| framework.
| 
| Integrations register pluggable drivers for mail, storage, cache, queue,
| and custom third-party APIs.
| 
| Active() resolves the configured implementation from environment
| variables; wire Application in ConfigureIntegrations.
| 
| Built-in connectors cover SMTP, SendGrid, S3-compatible storage, and
| Redis without vendor-specific SDKs in app code.
| 
| Symbols defined here include: ErrNotConfigured (exported value).
| 
*/

import "errors"

var ErrNotConfigured = errors.New("integration not configured")
