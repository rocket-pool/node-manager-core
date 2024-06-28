package client

import "time"

type StandardHttpClient struct {
	*StandardClient
}

// Create a new client instance.
// Most calls will use the fast timeout, but queries to validator status will use the slow timeout since they can be very large.
// Set a timeout of 0 to disable it.
func NewStandardHttpClient(providerAddress string, fastTimeout time.Duration, slowTimeout time.Duration) *StandardHttpClient {
	provider := NewBeaconHttpProvider(providerAddress, fastTimeout, slowTimeout)
	return &StandardHttpClient{
		StandardClient: NewStandardClient(provider),
	}
}
