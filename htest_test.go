// SPDX-License-Identifier: AGPL-3.0-only
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>

package htest_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/trim21/htest"
)

type res struct {
	Q string `json:"q"`
	I int    `json:"i"`
}

func TestClientFullExample(t *testing.T) {
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

func TestPathWithQuery(t *testing.T) {
	t.Parallel()
	app := echo.New()

	app.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, res{I: 5, Q: c.QueryParam("q")})
	})

	res := htest.New(t, app).Query("a", "2").Query("b", "3").Get("/test?a=1")

	require.Equal(t, "/test?a=1&a=2&b=3", res.Req.RequestURI)
}

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
