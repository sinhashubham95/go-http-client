package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"github.com/gojek/heimdall"
	"github.com/gojek/heimdall/httpclient"
	"github.com/gojek/heimdall/hystrix"
	"github.com/google/uuid"
	"golang.org/x/net/publicsuffix"
)

// Logger is the logger used to log the http traces
type Logger func(ctx context.Context, msg string)

// Metric is the information used to tracking the performance
type Metric struct {
	Status          int   `json:"status"`
	LatencyInMillis int64 `json:"latency"`
}

// Metrics provides the basic information for status and latency
type Metrics func(ctx context.Context, name string, m Metric)

// Client is the http client
type Client struct {
	httpClients map[string]ClientRequestMapping

	ol sync.Once
	l  Logger

	om sync.Once
	m  Metrics
}

// ClientRequestMapping provides a container for heimdall client and associated RequestConfig.
type ClientRequestMapping struct {
	heimdallClient heimdall.Client
	requestConfig  *RequestConfig
}

// ConfigureHTTPClient receives RequestConfigs and initializes one http client per RequestConfig.
// It creates heimdall http or hystrix client based on the configuration provided in RequestConfig.
// Returns the instance of Client
func ConfigureHTTPClient(requestConfigs ...*RequestConfig) *Client {
	httpClients := make(map[string]ClientRequestMapping)

	for _, requestConfig := range requestConfigs {
		if requestConfig != nil {
			clientRequestMapping :=
				ClientRequestMapping{
					heimdallClient: buildHTTPClient(requestConfig),
					requestConfig:  requestConfig,
				}
			httpClients[requestConfig.name] = clientRequestMapping
		}
	}

	client := Client{
		httpClients: httpClients,
	}

	return &client
}

// WithLogger is used to provide the logger instance for the http client created
func (c *Client) WithLogger(l Logger) *Client {
	if l != nil {
		c.ol.Do(func() {
			c.l = l
		})
	}
	return c
}

// WithMetrics is used to provide the metrics instance for the http client created
func (c *Client) WithMetrics(m Metrics) *Client {
	if m != nil {
		c.om.Do(func() {
			c.m = m
		})
	}
	return c
}

// Request receives Request param to execute. It will fetch the right http client for given Request name
// and use it to execute based on attributes provided in Request
// It returns http.Response and error
func (c *Client) Request(request *Request) (*http.Response, error) {
	client := c.httpClients[request.name]

	// set the method and url using the initial config
	if request.method == "" {
		request.method = client.requestConfig.method
	}
	if request.url == "" {
		request.url = client.requestConfig.url
	}
	// append static headers if exists
	if request.headerParams == nil {
		request.headerParams = client.requestConfig.headers
	}

	// fill the request-id header for log tracing
	request.SetHeaderParam(requestIDHeader, getRequestID(request.ctx))

	// start the timer
	start := time.Now()

	// get the http request
	req, err := getRequest(request.ctx, request.method, request.url, request.queryParams,
		request.headerParams, request.body)
	if err != nil {
		return nil, err
	}

	// now perform the request
	response, err := client.heimdallClient.Do(req)
	if err == nil && response == nil {
		return nil, errors.New("unable to fetch response")
	}
	if err == nil {
		// end the timer and log latency and status code
		c.logLatencyAndStatusCode(request, start, response.StatusCode)
		c.metricLatencyAndStatusCode(request, start, response.StatusCode)
	}

	return response, err
}

func (c *Client) logLatencyAndStatusCode(request *Request, start time.Time, statusCode int) {
	if c.l != nil {
		c.l(request.ctx, fmt.Sprintf("Fulfilled http request %s with status %d in duration %d ms",
			request.name, statusCode, time.Now().Sub(start).Milliseconds()))
	}
}

func (c *Client) metricLatencyAndStatusCode(request *Request, start time.Time, statusCode int) {
	if c.m != nil {
		c.m(request.ctx, request.name, Metric{Status: statusCode, LatencyInMillis: time.Now().Sub(start).Milliseconds()})
	}
}

func getRequestID(ctx context.Context) string {
	if ctx != nil {
		id := ctx.Value(idParam)
		if id != nil {
			if s, ok := id.(string); ok {
				return s
			}
		}
	}
	return uuid.NewString()
}

// This is an internal method to form the http.Request based on various parameters.
func getRequest(ctx context.Context, method string, url string, queryParams map[string]string,
	headerParams map[string]string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if ctx != nil {
		request = request.WithContext(ctx)
	}

	if queryParams != nil {
		q := request.URL.Query()
		for k, v := range queryParams {
			q.Add(k, v)
		}
		request.URL.RawQuery = q.Encode()
	}

	for k, v := range headerParams {
		request.Header.Add(k, v)
	}

	return request, err
}

// Internal method to build http or hystrix client based on settings provided in RequestConfig.
// It will create hystrix client if hystrixConfig is provided else it will provide httpclient.
func buildHTTPClient(requestConfig *RequestConfig) heimdall.Client {
	if requestConfig.hystrixConfig == nil {
		httpClient := httpclient.NewClient(
			httpclient.WithHTTPClient(getClient(requestConfig)),
			httpclient.WithHTTPTimeout(requestConfig.timeout),
			httpclient.WithRetryCount(requestConfig.retryCount),
			httpclient.WithRetrier(getRetrier(requestConfig)),
		)
		return httpClient
	} else {
		hystixClient := hystrix.NewClient(
			hystrix.WithHTTPClient(getClient(requestConfig)),
			hystrix.WithCommandName(requestConfig.name),
			hystrix.WithHTTPTimeout(requestConfig.timeout),
			hystrix.WithRetryCount(requestConfig.retryCount),
			hystrix.WithRetrier(getRetrier(requestConfig)),
			hystrix.WithHystrixTimeout(requestConfig.hystrixConfig.hystrixTimeout),
			hystrix.WithMaxConcurrentRequests(requestConfig.hystrixConfig.maxConcurrentRequests),
			hystrix.WithErrorPercentThreshold(requestConfig.hystrixConfig.errorPercentThreshold),
			hystrix.WithSleepWindow(requestConfig.hystrixConfig.sleepWindowInMillis),
			hystrix.WithRequestVolumeThreshold(requestConfig.hystrixConfig.requestVolumeThreshold),
			hystrix.WithFallbackFunc(requestConfig.hystrixConfig.fallback),
		)

		return hystixClient
	}
}

// This creates http client and setup transport based on RequestConfig settings.
// Following are default transport settings:
// ForceAttemptHTTP2 : true
// MaxIdleConnsPerHost : runtime.GOMAXPROCS(0) + 1
func getClient(requestConfig *RequestConfig) heimdall.Doer {
	// get the default cookie jar
	cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Jar:           cookieJar,
		Timeout:       requestConfig.timeout,
		Transport:     requestConfig.transport,
		CheckRedirect: requestConfig.checkRedirect,
	}

	client = setProxy(requestConfig, client)

	return client
}

// This sets the proxy to transport using proxy provided in RequestConfig
func setProxy(requestConfig *RequestConfig, client *http.Client) *http.Client {
	if requestConfig.proxyURL != "" {
		transport, err := transport(client)
		if err != nil {
			log.Printf("%v", err)
			return client
		}

		pURL, err := url.Parse(requestConfig.proxyURL)
		if err != nil {
			log.Printf("%v", err)
			return client
		}

		transport.Proxy = http.ProxyURL(pURL)

	}

	return client

}

// Transport method returns `*http.Transport` currently in use or error
// in case currently used `transport` is not a `*http.Transport`.
func transport(c *http.Client) (*http.Transport, error) {
	if t, ok := c.Transport.(*http.Transport); ok {
		return t, nil
	}
	return nil, errors.New("current transport is not an *http.Transport instance")
}

// This constructs the retry function (ConstantBackoff, ExponentialBackoff or NoRetrier) based on
// BackoffPolicy settings provided in RequestConfig
// NoRetry is used if no BackoffPolicy setting are provided
func getRetrier(requestConfig *RequestConfig) heimdall.Retriable {
	if requestConfig.backoffPolicy != nil && requestConfig.backoffPolicy.constantBackoff != nil {
		return heimdall.NewRetrier(heimdall.NewConstantBackoff(requestConfig.backoffPolicy.constantBackoff.interval,
			requestConfig.backoffPolicy.constantBackoff.maximumJitterInterval))
	} else if requestConfig.backoffPolicy != nil && requestConfig.backoffPolicy.exponentialBackoff != nil {
		return heimdall.NewRetrier(heimdall.NewExponentialBackoff(
			requestConfig.backoffPolicy.exponentialBackoff.initialTimeout,
			requestConfig.backoffPolicy.exponentialBackoff.maxTimeout,
			requestConfig.backoffPolicy.exponentialBackoff.exponentFactor,
			requestConfig.backoffPolicy.exponentialBackoff.maximumJitterInterval))
	}

	return heimdall.NewNoRetrier()
}
