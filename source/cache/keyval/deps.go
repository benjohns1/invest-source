package keyval

import (
	"time"
)

var (
	// Now function for retrieving the current timestamp. Override this for unit tests.
	Now = time.Now
)
