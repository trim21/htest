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

package htest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type Request struct {
	t           *testing.T
	r           *require.Assertions
	headers     http.Header
	urlQuery    url.Values
	formData    url.Values
	cookies     map[string]string
	httpVerb    string
	contentType string
	endpoint    string
	httpBody    []byte
	srv         http.Handler
}

func New(t *testing.T, server http.Handler) *Request {
	t.Helper()

	return &Request{
		srv:      server,
		r:        require.New(t),
		t:        t,
		urlQuery: url.Values{},
		cookies:  make(map[string]string),
		formData: url.Values{},
		headers:  http.Header{http.CanonicalHeaderKey("user-agent"): {"chii-test-client"}},
	}
}

func (r *Request) newRequest(httpVerb string, endpoint string) *Response {
	r.t.Helper()
	r.httpVerb = httpVerb
	r.endpoint = endpoint

	return r.execute()
}

func (r *Request) Get(entrypoint string) *Response {
	r.t.Helper()
	return r.newRequest(http.MethodGet, entrypoint)
}

func (r *Request) Post(entrypoint string) *Response {
	r.t.Helper()
	return r.newRequest(http.MethodPost, entrypoint)
}

func (r *Request) Put(entrypoint string) *Response {
	r.t.Helper()
	return r.newRequest(http.MethodPut, entrypoint)
}

func (r *Request) Patch(path string) *Response {
	r.t.Helper()
	return r.newRequest(http.MethodPatch, path)
}

func (r *Request) Delete(entrypoint string) *Response {
	r.t.Helper()
	return r.newRequest(http.MethodDelete, entrypoint)
}

func (r *Request) Cookie(key, value string) *Request {
	r.t.Helper()

	r.cookies[key] = value

	return r
}

func (r *Request) Query(key, value string) *Request {
	r.t.Helper()
	r.urlQuery.Set(key, value)
	return r
}

func (r *Request) Header(key, value string) *Request {
	r.t.Helper()
	r.headers.Set(key, value)

	return r
}

func (r *Request) Form(key, value string) *Request {
	r.t.Helper()
	if r.contentType == "" {
		r.contentType = echo.MIMEApplicationForm
	}

	r.r.Equal(r.contentType, echo.MIMEApplicationForm,
		"content-type should be empty or 'application/x-www-form-urlencoded',"+
			" can't mix .Form(...) with .BodyJSON(...)")

	r.formData.Set(key, value)
	r.httpBody = []byte(r.formData.Encode())

	return r
}

func (r *Request) BodyJSON(v any) *Request {
	r.t.Helper()
	require.Empty(r.t, r.contentType, "content-type should not be empty")

	var err error
	r.httpBody, err = json.Marshal(v)
	require.NoError(r.t, err)

	r.contentType = echo.MIMEApplicationJSON

	return r
}

func (r *Request) StdRequest() *http.Request {
	r.t.Helper()
	var body io.ReadCloser = http.NoBody
	if r.httpBody != nil {
		r.headers.Set(echo.HeaderContentLength, strconv.Itoa(len(r.httpBody)))
		if r.headers.Get(echo.HeaderContentType) == "" {
			r.headers.Set(echo.HeaderContentType, r.contentType)
		}

		body = io.NopCloser(bytes.NewBuffer(r.httpBody))
	}

	path := r.endpoint
	if len(r.urlQuery) > 0 {
		u, err := url.ParseRequestURI(r.endpoint)
		require.NoError(r.t, err)

		q, err := url.ParseQuery(u.RawQuery)
		require.NoError(r.t, err)

		for key, values := range r.urlQuery {
			for _, value := range values {
				q.Add(key, value)
			}
		}

		u.RawQuery = q.Encode()

		path = u.Path + "?" + q.Encode()
	}

	req := httptest.NewRequest(r.httpVerb, path, body)
	req.Header = r.headers
	for name, value := range r.cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	req.RemoteAddr = "0.0.0.0:3000"

	return req
}

func (r *Request) execute() *Response {
	r.t.Helper()

	resp := httptest.NewRecorder()

	req := r.StdRequest()
	r.srv.ServeHTTP(resp, req)

	return &Response{
		Req:        req,
		t:          r.t,
		StatusCode: resp.Code,
		Header:     resp.Header(),
		Body:       resp.Body.Bytes(),
		cookies:    parseCookies(r.t, resp.Header().Get(echo.HeaderSetCookie)),
	}
}

func parseCookies(t *testing.T, rawCookies string) []*http.Cookie {
	t.Helper()

	if rawCookies == "" {
		return nil
	}

	r := http.Response{
		Header: http.Header{echo.HeaderSetCookie: {rawCookies}},
	}

	return r.Cookies()
}

type Response struct {
	t          *testing.T
	Header     http.Header
	Body       []byte
	cookies    []*http.Cookie
	StatusCode int
	Req        *http.Request
}

func (r *Response) JSON(v any) *Response {
	r.t.Helper()

	if strings.HasPrefix(r.Header.Get(echo.HeaderContentType), echo.MIMEApplicationJSON) {
		require.NoError(r.t, json.Unmarshal(r.Body, v))
	}

	return r
}

func (r *Response) BodyString() string {
	return string(r.Body)
}

func (r *Response) ExpectCode(t int) *Response {
	r.t.Helper()

	require.Equalf(r.t, t, r.StatusCode, "expecting http response status code %d, body: %s", t, r.BodyString())

	return r
}

func (r *Response) Cookies() []*http.Cookie {
	r.t.Helper()

	return r.cookies
}
