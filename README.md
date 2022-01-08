![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/rovergulf/utils)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# busybox

Simple Golang HTTP REST Server debug tool

### HTTP Server API
Handles three paths:
- `/metrics` - Prometheus metrics handler
- `/health` - Can be used health check
- `/` - Debug logging of incoming request headers

### How to run

From source:
```shell
# build binary
go build -o busybox

# get app description and help
./busybox --help

# run server
./busybox --listen-addr=:8081 --log_json=true
```

Docker image:
```shell
docker build --no-cache -t busybox

docker run busybox -p 8081:8081
```

Helm Chart installation
```
TBD as public chart would be available
```
