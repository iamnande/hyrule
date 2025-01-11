package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	adminapi "github.com/iamnande/hyrule/cmd/admin-api"
	"github.com/iamnande/hyrule/internal/version"
)

type serviceCMD = string

const (
	defaultEntrypoint = "version"

	cmdAdminAPI serviceCMD = "admin-api"
	cmdVersion  serviceCMD = "version"

	serviceNameFormat    = "%s-%s"
	serviceVersionFormat = "%s %s %s"
)

var availableCMDs = []serviceCMD{
	cmdAdminAPI,
	cmdVersion,
}

func main() {
	var flagCMD string
	flag.StringVar(&flagCMD,
		"cmd",
		defaultEntrypoint,
		"available commands: "+strings.Join(availableCMDs, ","),
	)
	flag.Parse()

	switch flagCMD {
	case cmdAdminAPI:
		version.ServiceName = fmt.Sprintf(serviceNameFormat, version.ServicePrefix, cmdAdminAPI)
		adminapi.Run()
	case cmdVersion:
		fmt.Printf(serviceVersionFormat, version.ServicePrefix, cmdVersion, version.ServiceVersion)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "error: unknown entrypoint")
	}
}
