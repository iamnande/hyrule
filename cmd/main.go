package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	signupapi "github.com/iamnande/hyrule/cmd/signup-api"
	"github.com/iamnande/hyrule/internal/version"
)

type serviceCMD = string

const (
	defaultEntrypoint = "version"

	cmdSignUpAPI serviceCMD = "signup-api"
	cmdVersion   serviceCMD = "version"

	serviceNameFormat    = "%s-%s"
	serviceVersionFormat = "%s %s %s"
)

var availableCMDs = []serviceCMD{
	cmdSignUpAPI,
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
	case cmdSignUpAPI:
		version.ServiceName = fmt.Sprintf(serviceNameFormat, version.ServicePrefix, cmdSignUpAPI)
		signupapi.Run()
	case cmdVersion:
		fmt.Printf(serviceVersionFormat, version.ServicePrefix, cmdVersion, version.ServiceVersion)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "error: unknown entrypoint")
	}
}
