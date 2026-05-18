package bench

import "time"

// Result records one completed RPC. TotalDuration is wall-clock as
// observed by the worker; ServerDurationNs is the value the echo
// server reports in its response, so TotalDuration - ServerDurationNs
// approximates network plus framing time. UpstreamDurationNs comes
// from the x-envoy-upstream-service-time response header and is only
// meaningful when HasUpstreamTime is true; a sidecar that does not
// emit the header leaves UpstreamDurationNs at zero and the flag
// false, distinguishing "no measurement" from a genuine zero.
// BytesSent* and BytesReceived* are filled by the wire-length stats
// handler in compression.go: the Uncompressed counters are the
// marshaled message length, the Wire counters include any
// compression and gRPC framing. PodName, PodHostname, and PeerAddr
// form the backend-identification fallback chain consumed by
// aggregator.backendKey, in that order of preference. Err is non-nil
// when the RPC failed; it is the raw grpc status error so
// status.FromError on it recovers the code.
type Result struct {
	Start                     time.Time
	End                       time.Time
	TotalDuration             time.Duration
	ServerDurationNs          int64
	UpstreamDurationNs        int64
	HasUpstreamTime           bool
	BytesSentUncompressed     int64
	BytesSentWire             int64
	BytesReceivedUncompressed int64
	BytesReceivedWire         int64
	PodName                   string
	PodHostname               string
	PeerAddr                  string
	Err                       error
}
