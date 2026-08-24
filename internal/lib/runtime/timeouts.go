package runtime

import (
	"fmt"
	"time"
)

const (
	StartTimeout = 5 * time.Second
	DrainTimeout = 5 * time.Second

	// must clear DrainTimeout with margin, or fx races the drain sleep - see
	// the init() check below, which catches it the moment that stops being true.
	StopTimeout = DrainTimeout + 5*time.Second
)

func init() {
	if StopTimeout <= DrainTimeout {
		panic(fmt.Sprintf(
			"runtime: StopTimeout (%s) must exceed DrainTimeout (%s) - otherwise fx kills "+
				"the shutdown hook mid-drain instead of ever waiting for it to finish",
			StopTimeout, DrainTimeout,
		))
	}
}
