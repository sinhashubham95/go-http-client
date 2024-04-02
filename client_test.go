package httpclient

import (
	"net/http"
	"testing"

	"github.com/go-playground/assert"
	"github.com/stretchr/testify/require"
)

func TestDemo(t *testing.T) {
	configMap := map[string]interface{}{
		"method":          "GET",
		"url":             "https://www.google.co.in",
		"timeoutinmillis": 5000,
		"retrycount":      3,
		"backoffpolicy": map[string]interface{}{
			"constantbackoff": map[string]interface{}{
				"intervalinmillis":          2,
				"maxjitterintervalinmillis": 5,
			},
			"exponentialbackoff": map[string]interface{}{
				"initialtimeoutinmillis":    2,
				"maxtimeoutinmillis":        10,
				"exponentfactor":            2.0,
				"maxjitterintervalinmillis": 2,
			},
		},
		"hystrixconfig": map[string]interface{}{
			"maxconcurrentrequests":  10,
			"errorpercentthreshold":  20,
			"sleepwindowinmillis":    10,
			"requestvolumethreshold": 10,
		},
		"headers": map[string]interface{}{
			"limit":  10,
			"flag":   true,
			"source": "internal",
		},
	}

	//setup activity
	requestConfig := NewRequestConfig("test", configMap)

	client := ConfigureHTTPClient(requestConfig)

	// request
	req := NewRequest("test")
	res, err := client.Request(req)

	require.NoError(t, err, "should not have failed to make a GET request")

	assert.Equal(t, req.headerParams["flag"], "true")
	assert.Equal(t, req.headerParams["source"], "internal")
	assert.Equal(t, req.headerParams["limit"], "10")

	assert.Equal(t, http.StatusOK, res.StatusCode)
}
