# README #

Simple Http

Simple http server that provides 3 endpoints:

1. `POST /hash` accepts a password; returns a job identifier immediately; then waits 5 seconds and compute a SHA512 
encoding of the plain text password.
2. `GET /hash` accepts a job identifier; returns the base64 encoded password hash.
3. `GET /stats` returns a JSON object for the total hash requests since the server started along with the average time 
of a hash request in milliseconds.

## Purpose ##

To experiment with a simple Go http server that provides some basic handlers and graceful shutdown. 

## Technology Stack ##

* Golang


### Setup ###

No special setup required.  To run simply type `go run simple_http.go`


### Notes ###

* Currently, it artificially takes 5 seconds to convert a plain text password into a SHA512.  If the user requests the
SHA512 password before the 5 seconds, there will be no return value.  This isn't an ideal behaviour/customer experience.
It would be better to return a more complex json object what contains a `password` and `status` field.  The `status`
field would indicate if the password is in a `PENDING` state for the case when the password hasn't been encoded yet.

For example:

```
{
  "password": "",
  "status": "PENDING"
}
```

* Statistical calculations for the `/stats` endpoint only includes `/hash/` calls that were successful.
* One possible optimisation would be to not re-generate new SHA512 for previous passwords that are the same.  Instead, 
just use pointers and create a one-to-many relationship between `jobIs` and SHA512 passwords.
* Currently, I'm doing a simple hard wait of 5 seconds for all the background jobs to complete.  This should be enough
time given that the background `func` takes a hardcoded time of 5 second.  That being said, this is still not a very
robust approach.  A couple of approaches would be to:
    1. Use a `chan` to indicate any outstanding background jobs.  Once the `chan` is empty then exit.
    2. Use a thread pool, and then wait for outstanding threads to finish before exiting.