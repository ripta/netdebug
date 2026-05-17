package bench

import "time"

type Result struct {
	Start                     time.Time
	End                       time.Time
	TotalDuration             time.Duration
	ServerDurationNs          int64
	BytesSentUncompressed     int64
	BytesSentWire             int64
	BytesReceivedUncompressed int64
	BytesReceivedWire         int64
	PodName                   string
	PodHostname               string
	PeerAddr                  string
	Err                       error
}
