package version

const ServicePrefix = "hyrule"

var (
	ServiceVersion string
	ServiceCommit  string
	RepositoryURL  string
)

// ServiceInfo bundles a service's identity - the "who am I" a service needs
// in more than one place (logging, tracing, discovery) so it flows through
// fx as one value instead of half-injected, half-ambient-global.
type ServiceInfo struct {
	Name    string
	Version string
	Commit  string
}

func NewServiceInfo(serviceName string) ServiceInfo {
	return ServiceInfo{
		Name:    ServicePrefix + "-" + serviceName,
		Version: ServiceVersion,
		Commit:  ServiceCommit,
	}
}
