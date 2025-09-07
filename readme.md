# Nikola Telemetry

This project is currently a work in progress.

This project contains a gRPC proxy for using the Tesla Vehicle Command API and
Telemetry services. This proxy can be thought of as a one-stop-shop for
interacting with your Tesla vehicle, with the added benefit of automatic code
generation for API clients.

In addition, this service will automate much of the setup and maintenance of
the telemetry service (aka "Streaming Signals") so you can focus on your
application or integration.


## Configuration

The path to the configuration file defaults to `./config/config.json`, but can
be changed using the `CONFIG_PATH` environment variable. Here is an example
configuration file:

```json
{
	"PublicAddress": ":8080",
	"PrivateAddress": ":8081",
	"Domain": "nikola-telemetry-test.shaunkeys.com",
	"TeslaRegistrationUrl": "https://fleet-auth.prd.vn.cloud.tesla.com/oauth2/v3/token",
	"TeslaBaseUrl": "https://fleet-api.prd.na.vn.cloud.tesla.com",
	"TeslaPrivkeyFile": "./secrets/private-key.pem",
	"TeslaPubkeyFile": "./secrets/public-key.pem",
	"TeslaClientIdFile": "./secrets/client-id.txt",
	"TeslaClientSecretFile": "./secrets/client-secret.txt"
}
```


### Relative path behavior

Note that by default all relative file paths are in relation to the config file
itself, not the working directory of the application. This can be overridden
using the `BasePath` option. For example, setting `BasePath` to
`/path/to/my/files` will set an absolute path as the reference point for all
files referenced in the config. However, relative paths in `BasePath` will be
respected, so setting `BasePath` to `.` will make all paths relative to the
current working directory of the application, rather than the config file.


## Why the name?

Since this is a personal project, and I am not offering a service of any kind,
I'm not all that concerned about trademarks. To be honest, I just got
frustrated with the API sign-up process, and wanted to come up with something
snarky to get past their validation.


### Small rant about the application setup process

Tesla's API application setup wizard is trash, and makes a whole bunch of dumb
decisions that are just bad UX.

1. Session timeouts that expire without any kind of prompt or warning, and
   trash your whole form.
  * This wouldn't be so bad if it weren't for the following points, which all
    waste time during sign-up.
2. Prohibition of the name tesla in the domain
  * Okay, I kind of get this. They're trying to prevent trademark infringement,
    but let's be real. The way they have this set up, you kind of have to
    dedicate a whole subdomain for this purpose, so now people have to come up
    with some stupid secret code-name that means Tesla instead of just putting
    it in the name. As long as they're not naming their actual company Tesla or
    anything, who cares?
3. Full setup with TLS, web server, and keys required prior to even filling out
   the form.
  * What do we do if we're developing... I don't know... a *new* product? What
    are we supposed to do? We don't know how the API works yet, and we can't
    test anything without an application credential, which we can't get without
    a full working product? WTH?!
4. Some kind of allergy to reverse proxies?
  * I really don't get this one. Why does Tesla do this? Hardly anybody
    realistically runs their site without a reverse proxy nowadays, especially
    if you're using a microservices architecture, which Tesla seems to be
    advocating for. Many reverse proxies don't support SNI routing, so why do
    they not support them? It's all HTTP... I guess you could make the argument
    that end-to-end encryption is a good thing to have between Tesla and the
    rest of your application, and Tesla can ensure their security better if
    they own the whole encryption pipeline, but I don't see this fitting well
    into every scenario. They should have *something*, like some recommended
    configurations to make sure your proxy is secure.

So now that I've wasted all this time cobbling together a barely-functioning
system built out of twigs to get past form validation, the session timeout
activates and erases everything I've filled out so far... They really need to
hire a UX designer to actually test this stuff...
