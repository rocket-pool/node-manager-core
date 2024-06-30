package client

import "time"

// Options for the standard HTTP client
type StandardHttpClientOpts struct {
	// The time to wait for a request that is expected to return quickly
	FastTimeout time.Duration

	// The time to wait for a request that is expected to take a lot of processing on the BN and return slowly
	SlowTimeout time.Duration
}

// Standard high-level client for interacting with a Beacon Node over HTTP
type StandardHttpClient struct {
	*StandardClient
}

// Create a new client instance.
func NewStandardHttpClient(providerAddress string, opts *StandardHttpClientOpts) (*StandardHttpClient, error) {
	var provider *BeaconHttpProvider
	var err error
	if opts != nil {
		provider, err = NewBeaconHttpProvider(providerAddress, &BeaconHttpProviderOpts{
			DefaultFastTimeout: opts.FastTimeout,
			DefaultSlowTimeout: opts.SlowTimeout,
		})
	} else {
		provider, err = NewBeaconHttpProvider(providerAddress, nil)
	}
	if err != nil {
		return nil, err
	}

	return &StandardHttpClient{
		StandardClient: NewStandardClient(provider),
	}, nil
}
