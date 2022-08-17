package httpbin

import (
	// log http requests

	"io/ioutil"

	"github.com/motemen/go-loghttp"

	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

// apiVersion is the API version in use by this client.
const apiVersion = "2.27"

// uaVersion is the userAgent version sent to your API so they can track usage
// of this library.
const uaVersion = "1.0.0"

// defaultAuthorizationTokenPrefix provides a default fallback for jwt style authentication
// but can be overridden with custom values.
const defaultAuthorizationTokenPrefix = "Bearer"

// HTTPDoer is used for making HTTP requests. This implementation is generally
// a *http.Client.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client manages communication with the Example API.
type Client struct {
	// userAgent sets the User-Agent header for requests so you can
	// track usage of the client.
	userAgent string

	// Client is the HTTP Client used to communicate with the API.
	// By default this uses http.DefaultClient, so there are no timeouts
	// configured. It's recommended you set your own HTTP client with
	// reasonable timeouts for your application.
	Client HTTPDoer

	// Client options allow the caller to configure various parts of the
	// request.
	Options Opts

	// Services used for talking with different parts of the API
	HttpMethods HttpMethodsService
}

type serviceImpl struct {
	client *Client
}

type Opts struct {
	// Host is the base url for requests.
	Host *url.URL
	// Set debug level for logging http calls in the client http sdk. Defaults to false
	Debug bool
	// Version (optional) can be set if a versioned api is preferred, ex: /v1/items vs /v2/items
	Version string
	// Authorization holds the auth token and prefix to build up the header
	Authorization
}

type Authorization struct {
	// Prefix can override the default Bearer token prefix (Optional)
	Prefix string
	// Contains the authentication token, usually a JWT (Required)
	Token string
}

// NewClient returns a new instance of *Client.
func NewClient(opts *Opts) (*Client, error) {
	// Setup a sensible default http client
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	// Overwrite the transport with logging ability
	// TODO: expand this part more to aid in effective debug ability
	// TODO: Option to pass in a custom transport that logs using a passed in logger interface
	if opts.Debug {
		httpClient.Transport = loghttp.DefaultTransport
	}

	// We default to JWT style Bearer authentication tokens
	if opts.Authorization.Prefix == "" {
		opts.Authorization.Prefix = defaultAuthorizationTokenPrefix
	}

	// Setup default client
	client := &Client{
		Client:  httpClient,
		Options: *opts,
		userAgent: fmt.Sprintf(
			"sgen/HttpBin %s; Go (%s) [%s-%s]",
			uaVersion,
			runtime.Version(),
			runtime.GOARCH,
			runtime.GOOS,
		),
	}

	// Setup client service implementations
	client.HttpMethods = &httpMethodsImpl{client: client}

	return client, nil
}

// newRequest creates an authenticated API request that is ready to send.
func (c *Client) newRequest(ctx context.Context, method string, path string, body interface{}) (*http.Request, error) {
	switch {
	case c.Options.Version != "":
		path = fmt.Sprintf("/%s/%s", c.Options.Version, strings.TrimPrefix(path, "/"))
	default:
		path = strings.TrimPrefix(path, "/")
	}

	u := c.Options.Host
	u.Path = path

	// Request body
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	// We default to JWT Bearer: <token> types and only set the header if the token has been set on the client
	if c.Options.Authorization.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.Options.Authorization.Prefix, c.Options.Authorization.Token))
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-Api-Version", apiVersion)

	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	return req, err
}

// do takes a prepared API request and makes the API call to Recurly.
// It will decode the XML into a destination struct you provide as well
// as parse any validation errors that may have occurred.
// It returns a Response object that provides a wrapper around http.Response
// with some convenience methods.
func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*response, error) {
	req = req.WithContext(ctx)

	resp, err := c.Client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return nil, err
	}
	defer resp.Body.Close()

	response := newResponse(resp)
	if resp.StatusCode == http.StatusNoContent {
		return response, nil
		// } else if resp.StatusCode == http.StatusTooManyRequests {
		// 	return nil, &RateLimitError{
		// 		Response: resp,
		// 		Rate:     response.rate,
		// 	}
	} else if v != nil && resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		// TODO: expand this part more to aid in effective debug ability
		// TODO: Option to pass in a custom transport that logs using a passed in logger interface
		if c.Options.Debug {
			body, _ := ioutil.ReadAll(resp.Body)
			prettyString(string(body))
		}

		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else if err := json.NewDecoder(resp.Body).Decode(&v); err != nil && err != io.EOF {
			return response, err
		}

		return response, nil
	}
	// } else if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
	// 	return response, response.parseClientError(v)
	// } else if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
	// 	return nil, &ServerError{Response: resp}
	// }

	return response, nil
}

func prettyString(str string) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		panic(fmt.Errorf("unable to pretty print response: %w", err))
	}

	fmt.Println(prettyJSON.String())
}

// response is a Recurly API response. This wraps the standard http.Response
// returned from Recurly and provides access to pagination cursors and rate
// limits.
type response struct {
	*http.Response

	// // The next cursor (if available) when paginating results.
	// cursor string

	// // Rate limits.
	// rate Rate
}

// NewResponse creates a new Response for the provided http.Response.
func newResponse(r *http.Response) *response {
	resp := &response{Response: r}
	// resp.populatePageCursor()
	// resp.populateRateLimit()
	return resp
}
