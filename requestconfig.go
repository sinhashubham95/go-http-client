package httpclient

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/spf13/cast"
)

// RequestConfig is the type for a request configuration
type RequestConfig struct {
	name                  string
	method                string
	url                   string
	timeout               time.Duration
	connectTimeout        time.Duration
	keepAlive             time.Duration
	maxIdleConnections    int
	idleConnectionTimeout time.Duration
	tlsHandshakeTimeout   time.Duration
	expectContinueTimeout time.Duration
	proxyURL              string
	retryCount            int
	backoffPolicy         *BackoffPolicy
	hystrixConfig         *HystrixConfig
	transport             http.RoundTripper
	headers               map[string]string
	checkRedirect         func(*http.Request, []*http.Request) error
}

// NewRequestConfig is used to create a new request configuration from a map of configurations.
func NewRequestConfig(name string, configMap map[string]interface{}) *RequestConfig {
	rc := RequestConfig{
		name: name,
	}

	if configMap != nil {
		var err error

		rc.method, _ = getConfigOptionString(configMap, "method")
		rc.url, _ = getConfigOptionString(configMap, "url")

		timeout, err := getConfigOptionInt(configMap, "timeoutinmillis")
		if err == nil {
			rc.timeout = time.Duration(timeout) * time.Millisecond
		}

		connectTimeout, err := getConfigOptionInt(configMap, "connecttimeoutinmillis")
		if err == nil {
			rc.connectTimeout = time.Duration(connectTimeout) * time.Millisecond
		} else {
			rc.connectTimeout = rc.timeout / 10
		}

		keepAlive, err := getConfigOptionInt(configMap, "keepaliveinmillis")
		if err == nil {
			rc.keepAlive = time.Duration(keepAlive) * time.Millisecond
		} else {
			rc.keepAlive = defaultKeepAlive
		}

		maxIdleConnections, err := getConfigOptionInt(configMap, "maxidleconnections")
		if err == nil {
			rc.maxIdleConnections = maxIdleConnections
		} else {
			rc.maxIdleConnections = runtime.GOMAXPROCS(0) + 1
		}

		idleConnectionTimeout, err := getConfigOptionInt(configMap, "idleconnectiontimeoutinmillis")
		if err == nil {
			rc.idleConnectionTimeout = time.Duration(idleConnectionTimeout) * time.Millisecond
		} else {
			rc.idleConnectionTimeout = defaultIdleConnectionTimeout
		}

		tlsHandshakeTimeout, err := getConfigOptionInt(configMap, "tlshandshaketimeoutinmillis")
		if err == nil {
			rc.tlsHandshakeTimeout = time.Duration(tlsHandshakeTimeout) * time.Millisecond
		}

		expectContinueTimeout, err := getConfigOptionInt(configMap, "expectcontinuetimeoutinmillis")
		if err == nil {
			rc.expectContinueTimeout = time.Duration(expectContinueTimeout) * time.Millisecond
		}

		rc.proxyURL, _ = getConfigOptionString(configMap, "proxyurl")

		rc.retryCount, err = getConfigOptionInt(configMap, "retrycount")
		if err != nil {
			rc.retryCount = 1
		}

		backoffPolicyMap, err := getConfigOptionMap(configMap, "backoffpolicy")
		if err == nil {
			rc.backoffPolicy = NewBackoffPolicy(backoffPolicyMap)
		}

		hystrixConfig, err := getConfigOptionMap(configMap, "hystrixconfig")
		if err == nil {
			rc.hystrixConfig = NewHystrixConfig(hystrixConfig)
		}

		headers, err := getConfigOptionMap(configMap, "headers")
		if err == nil {
			rc.headers = cast.ToStringMapString(headers)
		}

		tlsMinVersion, _ := getConfigOptionString(configMap, "tlsminversion")

		var tlsConfig *tls.Config
		switch tlsMinVersion {
		case "1.0":
			tlsConfig = &tls.Config{MinVersion: tls.VersionTLS10}
		case "1.1":
			tlsConfig = &tls.Config{MinVersion: tls.VersionTLS11}
		case "1.2":
			tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		case "1.3":
			tlsConfig = &tls.Config{MinVersion: tls.VersionTLS13}
		default:
			tlsConfig = nil
		}

		// Setting Default Transport.
		dialer := &net.Dialer{
			Timeout:   rc.connectTimeout,
			KeepAlive: rc.keepAlive,
		}
		rc.transport = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          rc.maxIdleConnections,
			IdleConnTimeout:       rc.idleConnectionTimeout,
			TLSHandshakeTimeout:   rc.tlsHandshakeTimeout,
			ExpectContinueTimeout: rc.expectContinueTimeout,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
			TLSClientConfig:       tlsConfig,
		}

	}
	return &rc
}

// SetName is used to set name for request
func (rc *RequestConfig) SetName(name string) *RequestConfig {
	rc.name = name
	return rc
}

// SetMethod is used to set method for request
func (rc *RequestConfig) SetMethod(method string) *RequestConfig {
	rc.method = method
	return rc
}

// SetURL is used to set the url for request
func (rc *RequestConfig) SetURL(url string) *RequestConfig {
	rc.url = url
	return rc
}

// SetProxy is used to set the proxy url for request
func (rc *RequestConfig) SetProxy(proxyURL string) *RequestConfig {
	rc.proxyURL = proxyURL
	return rc
}

// SetTimeout is used to set the timeout for request
func (rc *RequestConfig) SetTimeout(timeout time.Duration) *RequestConfig {
	rc.timeout = timeout
	return rc
}

// SetConnectTimeout is used to set connect timeout
func (rc *RequestConfig) SetConnectTimeout(connectTimeout time.Duration) *RequestConfig {
	rc.connectTimeout = connectTimeout
	return rc
}

// SetKeepAlive is used to set the keep alive for request
func (rc *RequestConfig) SetKeepAlive(keepalive time.Duration) *RequestConfig {
	rc.keepAlive = keepalive
	return rc
}

// SetMaxIdleConnections is used to set the max idle connections for request
func (rc *RequestConfig) SetMaxIdleConnections(maxIdleConnections int) *RequestConfig {
	rc.maxIdleConnections = maxIdleConnections
	return rc
}

// SetIdleConnectionTimeout is used to set the idle connection timeout for request
func (rc *RequestConfig) SetIdleConnectionTimeout(idleConnectionTimeout time.Duration) *RequestConfig {
	rc.idleConnectionTimeout = idleConnectionTimeout
	return rc
}

// SetTLSHandshakeTimeout is used to set the tls handshake timeout for request
func (rc *RequestConfig) SetTLSHandshakeTimeout(tlsHandshakeTimeout time.Duration) *RequestConfig {
	rc.tlsHandshakeTimeout = tlsHandshakeTimeout
	return rc
}

// SetExpectContinueTimeout is used to set expect continue timeout for request
func (rc *RequestConfig) SetExpectContinueTimeout(expectContinueTimeout time.Duration) *RequestConfig {
	rc.expectContinueTimeout = expectContinueTimeout
	return rc
}

// SetRetryCount is used to set the retry count for request
func (rc *RequestConfig) SetRetryCount(retryCount int) *RequestConfig {
	rc.retryCount = retryCount
	return rc
}

// SetBackoffPolicy is used to set the backoff policy for request
func (rc *RequestConfig) SetBackoffPolicy(backoffPolicy *BackoffPolicy) *RequestConfig {
	rc.backoffPolicy = backoffPolicy
	return rc
}

// SetHystrixConfig is used to set the hystrix config for the request
func (rc *RequestConfig) SetHystrixConfig(hystrixConfig *HystrixConfig) *RequestConfig {
	rc.hystrixConfig = hystrixConfig
	return rc
}

// SetHystrixFallback is used to set the hystrix fallback for the request
func (rc *RequestConfig) SetHystrixFallback(fallbackFn func(error) error) *RequestConfig {
	if rc.hystrixConfig != nil {
		rc.hystrixConfig.fallback = fallbackFn
	}
	return rc
}

// Transport is used to get the transport set for Request config.
func (rc *RequestConfig) Transport() http.RoundTripper {
	return rc.transport
}

// SetTransport is used to set the transport. Setting the transport overrides the default transport
func (rc *RequestConfig) SetTransport(transport http.RoundTripper) *RequestConfig {
	rc.transport = transport
	return rc
}

// SetCheckRedirect CheckRedirect specifies the policy for handling redirects.
func (rc *RequestConfig) SetCheckRedirect(checkRedirect func(*http.Request, []*http.Request) error) *RequestConfig {
	rc.checkRedirect = checkRedirect
	return rc
}

// SetHeaders is used to set the headers. Setting the headers overrides the default header
func (rc *RequestConfig) SetHeaderParams(headers map[string]interface{}) *RequestConfig {
	rc.headers = cast.ToStringMapString(headers)
	return rc
}

func getConfigOptionInt(options map[string]interface{}, key string) (int, error) {
	var val interface{}
	var ok bool
	var s int
	if val, ok = options[key]; ok {
		return cast.ToIntE(val)
	} else {
		return s, fmt.Errorf("missing %s", key)
	}
}

func getConfigOptionFloat(options map[string]interface{}, key string) (float64, error) {
	var val interface{}
	var ok bool
	var s float64
	if val, ok = options[key]; ok {
		return cast.ToFloat64E(val)
	} else {
		return s, fmt.Errorf("missing %s", key)
	}
}

func getConfigOptionMap(options map[string]interface{}, key string) (map[string]interface{}, error) {
	var val interface{}
	var ok bool
	var s map[string]interface{}
	if val, ok = options[key]; ok {
		return cast.ToStringMapE(val)
	} else {
		return s, fmt.Errorf("missing %s", key)
	}
}

func getConfigOptionString(options map[string]interface{}, key string) (string, error) {
	var val interface{}
	var ok bool
	var s string
	if val, ok = options[key]; ok {
		return cast.ToStringE(val)
	} else {
		return s, fmt.Errorf("missing %s", key)
	}
}
