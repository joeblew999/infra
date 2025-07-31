# pocketbase

https://pocketbase.io/docs/

https://github.com/pocketbase/pocketbase

NOTE: I do not know yet IF having this as part of the standard built will bloat thngs. Lets see 


We use the hooks and subscriptions that Pocketbase has, along with DataStar SSE. So then All our code is in golang, and we do not have any JS. 

Uses pkg/config to get Default data paths and ports, so we never have a clash.

Use pkg/log to tell us it has started !

```sh

go run . service

# Or disable it:
go run . service --no-pocketbase

``` 

## Admin

We will need a way to provision Admin. 

## Routes 

The started web server has the following default routes:

http://127.0.0.1:8090 - if pb_public directory exists, serves the static content from it (html, css, images, etc.)
http://127.0.0.1:8090/_/ - superusers dashboard
http://127.0.0.1:8090/api/ - REST-ish API

## Storage

Is can use the FS or S3 for File storage.

The prebuilt PocketBase executable will create and manage 2 new directories alongside the executable:

pb_data - stores your application data, uploaded files, etc. (usually should be added in .gitignore).

pb_migrations - contains JS migration files with your collection changes (can be safely committed in your repository).


## auth

https://github.com/spinspire/pocketbase-sveltekit-starter/blob/master/pb/webauthn/webauthn.go looks good for later

## stripe

https://github.com/mrwyndham/pocketbase-stripe

Looks like it might need to be forked and updated for the latest PB and STRIPE.

github.com/stripe/stripe-go/v76

## sms

github.com/pocketbase/pocketbase/tools/sms is proposed.

