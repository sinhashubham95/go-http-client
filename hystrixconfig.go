package httpclient

import "time"

// HystrixConfig is the configuration for hystrix
type HystrixConfig struct {
	hystrixTimeout         time.Duration
	maxConcurrentRequests  int
	errorPercentThreshold  int
	sleepWindowInMillis    int
	requestVolumeThreshold int
	fallback               func(error) error
}

// NewHystrixConfig is used to create a new configuration for hystrix
func NewHystrixConfig(configMap map[string]interface{}) *HystrixConfig {
	hystrixConfig := &HystrixConfig{}
	hystrixTimeout, err := getConfigOptionInt(configMap, "hystrixtimeoutinmillis")
	if err == nil {
		hystrixConfig.hystrixTimeout = time.Duration(hystrixTimeout) * time.Millisecond
	}
	sleepWindow, err := getConfigOptionInt(configMap, "sleepwindowinmillis")
	if err == nil {
		hystrixConfig.sleepWindowInMillis = sleepWindow
	} else {
		hystrixConfig.sleepWindowInMillis = defaultSleepWindowInMillis
	}
	hystrixConfig.maxConcurrentRequests, _ = getConfigOptionInt(configMap, "maxconcurrentrequests")
	hystrixConfig.errorPercentThreshold, _ = getConfigOptionInt(configMap, "errorpercentthreshold")
	hystrixConfig.requestVolumeThreshold, _ = getConfigOptionInt(configMap, "requestvolumethreshold")
	return hystrixConfig
}

// SetHystrixTimeout is used to set the hystrix timeout
func (hc *HystrixConfig) SetHystrixTimeout(hystrixTimeout time.Duration) *HystrixConfig {
	hc.hystrixTimeout = hystrixTimeout
	return hc
}

// SetMaxConcurrentRequests is used to set the max concurrent requests in hystrix
func (hc *HystrixConfig) SetMaxConcurrentRequests(maxConcurrentRequests int) *HystrixConfig {
	hc.maxConcurrentRequests = maxConcurrentRequests
	return hc
}

// SetErrorPercentThreshold is used to set the error percent threshold in hystrix
func (hc *HystrixConfig) SetErrorPercentThreshold(errorPercentThreshold int) *HystrixConfig {
	hc.errorPercentThreshold = errorPercentThreshold
	return hc
}

// SetSleepWindowInMillis is used to set the sleep window in hystrix
func (hc *HystrixConfig) SetSleepWindowInMillis(sleepWindowInMillis int) *HystrixConfig {
	hc.sleepWindowInMillis = sleepWindowInMillis
	return hc
}

// SetRequestVolumeThreshold is used to set the request volume threshold
func (hc *HystrixConfig) SetRequestVolumeThreshold(requestVolumeThreshold int) *HystrixConfig {
	hc.requestVolumeThreshold = requestVolumeThreshold
	return hc
}

// SetFallback is used to set the fallback
func (hc *HystrixConfig) SetFallback(fallbackFn func(error) error) *HystrixConfig {
	hc.fallback = fallbackFn
	return hc
}
