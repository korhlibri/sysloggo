# sysloggo
A simple syslog server listener written in Golang. Includes a Python testing script

## Run
The server can be ran or compiled using the specific go commands. Make sure to modify the global constant variables in `main.go` that declare the ports and the host address.
```
go run ./cmd/sysloggo/main.go
```
```
go build ./cmd/sysloggo/main.go
``` 