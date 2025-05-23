// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonhttptest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/calmw/bee-tron/pkg/jsonhttp"
)

// Request is a testing helper function that makes an HTTP request using
// provided client with provided method and url. It performs a validation on
// expected response code and additional options. It returns response headers if
// the request and all validation are successful. In case of any error, testing
// Errorf or Fatal functions will be called.
func Request(tb testing.TB, client *http.Client, method, url string, responseCode int, opts ...Option) http.Header {
	tb.Helper()

	o := new(options)
	for _, opt := range opts {
		if err := opt.apply(o); err != nil {
			tb.Fatal(err)
		}
	}

	req, err := http.NewRequest(method, url, o.requestBody)
	if err != nil {
		tb.Fatal(err)
	}
	req.Header = o.requestHeaders
	if o.ctx != nil {
		req = req.WithContext(o.ctx)
	}
	resp, err := client.Do(req)
	if err != nil {
		tb.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != responseCode {
		tb.Errorf("got response status %s, want %v %s", resp.Status, responseCode, http.StatusText(responseCode))
	}

	for _, key := range o.nonEmptyResponseHeaders {
		if val := resp.Header.Get(key); val == "" {
			tb.Errorf("header key=[%s] should be set", key)
		}
	}

	if headers := o.expectedResponseHeaders; headers != nil {
		for key, values := range headers {
			got := sort.StringSlice(resp.Header.Values(key))
			want := sort.StringSlice(values)
			if !reflect.DeepEqual(got, want) {
				tb.Errorf("header values for key=[%s] not as expected, got: %v, want %v", key, got, want)
			}
		}

		// When "Content-Length" header is set additionally assert
		// that resp.ContentLength has the same value.
		if want := headers.Get("Content-Length"); want != "" {
			got := strconv.FormatInt(resp.ContentLength, 10)
			if want != got {
				tb.Errorf("http.Response.ContentLength not as expected, got %v, want %v", got, want)
			}
		}
	}

	if o.expectedResponse != nil {
		got, err := io.ReadAll(resp.Body)
		if err != nil {
			tb.Fatal(err)
		}

		if !bytes.Equal(got, o.expectedResponse) {
			tb.Errorf("got response %q, want %q", string(got), string(o.expectedResponse))
		}
		return resp.Header
	}

	if o.expectedJSONResponse != nil {
		if v := resp.Header.Get("Content-Type"); v != jsonhttp.DefaultContentTypeHeader {
			tb.Errorf("got content type %q, want %q", v, jsonhttp.DefaultContentTypeHeader)
		}
		got, err := io.ReadAll(resp.Body)
		if err != nil {
			tb.Fatal(err)
		}
		got = bytes.TrimSpace(got)

		want, err := json.Marshal(o.expectedJSONResponse)
		if err != nil {
			tb.Fatal(err)
		}

		if !bytes.Equal(got, want) {
			tb.Errorf("got json response %q, want %q", string(got), string(want))
		}
		return resp.Header
	}

	if o.unmarshalResponse != nil {
		if err := json.NewDecoder(resp.Body).Decode(&o.unmarshalResponse); err != nil {
			tb.Fatal(err)
		}
		return resp.Header
	}
	if o.responseBody != nil {
		got, err := io.ReadAll(resp.Body)
		if err != nil {
			tb.Fatal(err)
		}
		*o.responseBody = got
	}
	if o.noResponseBody {
		got, err := io.ReadAll(resp.Body)
		if err != nil {
			tb.Fatal(err)
		}
		if len(got) > 0 {
			tb.Errorf("got response body %q, want none", string(got))
		}
	}
	return resp.Header
}

// WithContext sets a context to the request made by the Request function.
func WithContext(ctx context.Context) Option {
	return optionFunc(func(o *options) error {
		o.ctx = ctx
		return nil
	})
}

// WithRequestBody writes a request body to the request made by the Request
// function.
func WithRequestBody(body io.Reader) Option {
	return optionFunc(func(o *options) error {
		o.requestBody = body
		return nil
	})
}

// WithJSONRequestBody writes a request JSON-encoded body to the request made by
// the Request function.
func WithJSONRequestBody(r interface{}) Option {
	return optionFunc(func(o *options) error {
		b, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("json encode request body: %w", err)
		}
		o.requestBody = bytes.NewReader(b)
		return nil
	})
}

// WithMultipartRequest writes a multipart request with a single file in it to
// the request made by the Request function.
func WithMultipartRequest(body io.Reader, length int, filename, contentType string) Option {
	return optionFunc(func(o *options) error {
		buf := bytes.NewBuffer(nil)
		mw := multipart.NewWriter(buf)
		hdr := make(textproto.MIMEHeader)
		if filename != "" {
			hdr.Set("Content-Disposition", fmt.Sprintf("form-data; name=%q", filename))
		}
		if contentType != "" {
			hdr.Set("Content-Type", contentType)
		}
		if length > 0 {
			hdr.Set("Content-Length", strconv.Itoa(length))
		}
		part, err := mw.CreatePart(hdr)
		if err != nil {
			return fmt.Errorf("create multipart part: %w", err)
		}
		if _, err = io.Copy(part, body); err != nil {
			return fmt.Errorf("copy file data to multipart part: %w", err)
		}
		if err := mw.Close(); err != nil {
			return fmt.Errorf("close multipart writer: %w", err)
		}
		o.requestBody = buf
		if o.requestHeaders == nil {
			o.requestHeaders = make(http.Header)
		}
		o.requestHeaders.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%q", mw.Boundary()))
		return nil
	})
}

// WithRequestHeader adds a single header to the request made by the Request
// function. To add multiple headers call multiple times this option when as
// arguments to the Request function.
func WithRequestHeader(key, value string) Option {
	return optionFunc(func(o *options) error {
		if o.requestHeaders == nil {
			o.requestHeaders = make(http.Header)
		}
		o.requestHeaders.Add(key, value)
		return nil
	})
}

// WithExpectedResponse validates that the response from the request in the
// Request function matches completely bytes provided here.
func WithExpectedResponse(response []byte) Option {
	return optionFunc(func(o *options) error {
		o.expectedResponse = response
		return nil
	})
}

// WithExpectedResponseHeader validates that the response from the request
// has header with specified value
func WithExpectedResponseHeader(key, value string) Option {
	return optionFunc(func(o *options) error {
		if o.expectedResponseHeaders == nil {
			o.expectedResponseHeaders = make(http.Header)
		}
		o.expectedResponseHeaders.Add(key, value)
		return nil
	})
}

// WithExpectedContentLength is shorthand for creating "Content-Length" header check.
func WithExpectedContentLength(value int) Option {
	return WithExpectedResponseHeader("Content-Length", strconv.Itoa(value))
}

// WithNonEmptyResponseHeader validates that the response from the request
// has header with non empty value.
func WithNonEmptyResponseHeader(key string) Option {
	return optionFunc(func(o *options) error {
		if o.nonEmptyResponseHeaders == nil {
			o.nonEmptyResponseHeaders = make([]string, 0, 1)
		}
		o.nonEmptyResponseHeaders = append(o.nonEmptyResponseHeaders, key)
		return nil
	})
}

// WithExpectedJSONResponse validates that the response from the request in the
// Request function matches JSON-encoded body provided here.
func WithExpectedJSONResponse(response interface{}) Option {
	return optionFunc(func(o *options) error {
		o.expectedJSONResponse = response
		return nil
	})
}

// WithUnmarshalJSONResponse unmarshals response body from the request in the
// Request function to the provided response. Response must be a pointer.
func WithUnmarshalJSONResponse(response interface{}) Option {
	return optionFunc(func(o *options) error {
		o.unmarshalResponse = response
		return nil
	})
}

// WithPutResponseBody replaces the data in the provided byte slice with the
// data from the response body of the request in the Request function.
//
// Example:
//
//	var respBytes []byte
//	options := []jsonhttptest.Option{
//		jsonhttptest.WithPutResponseBody(&respBytes),
//	}
func WithPutResponseBody(b *[]byte) Option {
	return optionFunc(func(o *options) error {
		o.responseBody = b
		return nil
	})
}

// WithNoResponseBody ensures that there is no data sent by the response of the
// request in the Request function.
func WithNoResponseBody() Option {
	return optionFunc(func(o *options) error {
		o.noResponseBody = true
		return nil
	})
}

type options struct {
	ctx                     context.Context
	requestBody             io.Reader
	requestHeaders          http.Header
	expectedResponseHeaders http.Header
	nonEmptyResponseHeaders []string
	expectedResponse        []byte
	expectedJSONResponse    interface{}
	unmarshalResponse       interface{}
	responseBody            *[]byte
	noResponseBody          bool
}

type Option interface {
	apply(*options) error
}
type optionFunc func(*options) error

func (f optionFunc) apply(r *options) error { return f(r) }
