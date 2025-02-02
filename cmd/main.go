package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	inviteapi "github.com/iamnande/hyrule/cmd/invite-api"
	registrationapi "github.com/iamnande/hyrule/cmd/registration-api"
	"github.com/iamnande/hyrule/internal/version"
)

type serviceCMD = string

const (
	defaultEntrypoint = "version"

	cmdVersion         serviceCMD = "version"
	cmdRegistrationAPI serviceCMD = "registration-api"
	cmdInviteAPI       serviceCMD = "invite-api"

	serviceNameFormat    = "%s-%s"
	serviceVersionFormat = "%s %s %s"
)

var availableCMDs = []serviceCMD{
	cmdVersion,
	cmdRegistrationAPI,
	cmdInviteAPI,
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
	case cmdRegistrationAPI:
		version.ServiceName = fmt.Sprintf(serviceNameFormat, version.ServicePrefix, cmdRegistrationAPI)
		registrationapi.Run()
	case cmdInviteAPI:
		version.ServiceName = fmt.Sprintf(serviceNameFormat, version.ServicePrefix, cmdInviteAPI)
		inviteapi.Run()
	case cmdVersion:
		fmt.Printf(serviceVersionFormat, version.ServicePrefix, cmdVersion, version.ServiceVersion)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "error: unknown entrypoint")
	}
}
