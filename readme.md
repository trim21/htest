# htest

[![Go Reference](https://pkg.go.dev/badge/github.com/trim21/htest.svg)](https://pkg.go.dev/github.com/trim21/htest)

Chainable light-weight http client for testing golang `http.Handler`

```golang
package main_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"
)

func TestGet(t *testing.T) {
	t.Parallel()
	app := echo.New()

	app.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, res{I: 5, Q: c.QueryParam("q")})
	})

	var r res
	htest.New(t, app).
		Query("q", "v").
		Get("/test").
		JSON(&r).
		ExpectCode(http.StatusOK)

	require.Equal(t, 5, r.I)
	require.Equal(t, "v", r.Q)
}

```


JSON

```golang
package main_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"
)

func TestJSON(t *testing.T) {
	t.Parallel()

	app := echo.New()
	app.POST("/", func(c echo.Context) error {
		var r json.RawMessage
		err := json.NewDecoder(c.Request().Body).Decode(&r)
		require.NoError(t, err)

		return c.JSON(http.StatusOK, r)
	})

	var r struct {
		Hello int `json:"hello"`
		World int `json:"world"`
	}

	htest.New(t, app).
		BodyJSON(map[string]int{"hello": 1, "world": 2}).
		Post("/").
		ExpectCode(http.StatusOK).
		JSON(&r)

	require.Equal(t, 1, r.Hello)
	require.Equal(t, 2, r.World)
}
```


Form

```golang
package main_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"
)

func TestForm(t *testing.T) {
	t.Parallel()

	app := echo.New()
	app.POST("/", func(c echo.Context) error {
		form, err := c.FormParams()
		require.NoError(t, err)

		return c.JSON(http.StatusOK, res{Q: form.Get("q")})
	})

	var r res
	res := htest.New(t, app).
		Form("q", "form-value").
		Post("/").
		ExpectCode(http.StatusOK).
		JSON(&r)

	require.Equal(t, "form-value", r.Q, res.BodyString())
}
```
