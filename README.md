# gostman [![Go Reference](https://pkg.go.dev/badge/github.com/minizilla/gostman.svg)](https://pkg.go.dev/github.com/minizilla/gostman) [![main](https://github.com/minizilla/gostman/actions/workflows/main.yaml/badge.svg)](https://github.com/minizilla/gostman/actions/workflows/main.yaml)

[Postman](https://www.postman.com/) nuances in [Go](https://golang.org/) Test.

## Install

Just import gostman to your test package.

```go
import github.com/minizilla/gostman
```

## Usage

Always run the runtime in the `TestMain`.

```go
func TestMain(m *testing.M) {
    os.Exit(gostman.Run(m))
}
```

*Optional*. Create gostman environment file `.gostman.env.yml` if using variable.

```yml
myenv:
  var1: "variable 1"
  var2: "variable 2"
another_env:
  var1: "another variable 1"
  var2: "another variable 2"
```

Create usual Test function `TestXxx`.

```go
func TestRequest(t *testing.T) {
    gm := gostman.New(t)

    // every request run in the subtests
    gm.GET("Request", "url", func(r *gostman.Request) {
        // just code usual go code for Pre-request Script

        value := gm.V("var1") // get variable
        gm.SetV("var2", value) // set variable

        r.Params( /*sets Params here*/ )

        r.Authorization( /*sets Authorization here*/ )

        r.Headers( /*sets Headers here*/ )

        r.Body( /*sets Body here*/ )

        r.Send( /*send the request and tests the result here*/ )
    })

    // create another request
    gm.POST("Another Request", "url", func(r *gostman.Request) {})
```

Leverage `go test` command to run the above requests.

```sh
go test # run all gostman requests in the package
go test -run Request # run all collection in the TestRequest
go test -run Request/AnotherRequest # run only AnotherRequest
go test -run Request -env myenv # run request and use "myenv" environment
go test -run Request -setenv myenv # run request, use "myenv" environment and set it for the future request
```

## The Runtime

Gostman generate file `.gostman.runtime.yml` to store runtime config and variable and change often after each request.
It is **recommended** to ignore it in the `.gitignore`.

```gitignore
*.gostman.runtime.yml
```

## Flags

- All `go test` flags
- `-env {env}` Select environment define in `.gostman.env.yml`
- `-setenv {env}` Select environment define in `.gostman.env.yml` and set it for the future request
- `-reset` Reset `.gostman.runtime.yml`
- `-debug` Run gostman in debug mode

## Example

More in the examples folder.

```go
package gostman_test

import (
    "encoding/json"
    "net/http"
    "net/url"
    "os"
    "testing"

    "github.com/minizilla/gostman"
    "github.com/minizilla/testr"
)

func TestMain(m *testing.M) {
    os.Exit(gostman.Run(m))
}

func TestRequest(t *testing.T) {
    gm := gostman.New(t)

    gm.GET("Params", "https://postman-echo.com/get", func(r *gostman.Request) {
        r.Params(func(v url.Values) {
            v.Set("foo", "bar")
        })

        r.Send(func(t *testing.T, req *http.Request, res *http.Response) {
            defer res.Body.Close()

            assert := testr.New(t)
            assert.Equal(res.StatusCode, http.StatusOK)

            var resp = struct {
                Args map[string]string `json:"args"`
            }{}
            err := json.NewDecoder(res.Body).Decode(&resp)
            assert.ErrorIs(err, nil)
            assert.Equal(resp.Args["foo"], "bar")
        })
    })
}
```
