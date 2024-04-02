package httpclient

import "time"

var (
	defaultKeepAlive             = time.Second * 30
	defaultIdleConnectionTimeout = time.Second * 90
	defaultSleepWindowInMillis   = 5000
	requestIDHeader              = "X-requestId"
	idParam                      = "id"
)
