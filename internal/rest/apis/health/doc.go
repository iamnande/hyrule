// Package health implements an opinionated health API. Service, deployment and
// runtime data are included and expected as options for proper use. It is
// designed for use in an orchestrated environment. Liveness and readiness
// probes should be provided rather explicitly. Both soft and hard dependencies
// can be provided as options.
//
// Soft - If this goes down, the service should still be able to function
// without reliance on this dependency (e.g. feature flag client with health
// check polling).
//
// Hard - If this goes down, the service will noticeably degrade. Customers
// will experience errors (e.g. database backend goes down).
//
// NOTE: https://youtu.be/1EBfxjSFAxQ?si=dgDF0936vosVJGmw
package health
