# MyIA
 **My** **I**P **A**ddress
 
## Description
Simple http server which returns requester IP address.

## Parameters
* `-l` - listen address, default `127.0.0.1:1234`
* `-p` - listen path, default `/`
* `-n` - returns IP address if only it contained 
in provided network prefix (i.e. `10.0.0.0/8`)
* `-o` - set provided value to `Access-Control-Allow-Origin` header
* `-r` - get value as client IP from provided HTTP header 
instead of HTTP remote address. May be useful if program runs 
behind reverse-proxy. _Note_: only first (nearest) address will be 
used as client address. If header value not set, HTTP remote address used.

