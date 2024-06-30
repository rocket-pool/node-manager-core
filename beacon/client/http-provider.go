package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/rocket-pool/node-manager-core/log"
)

const (
	RequestContentType = "application/json"

	RequestSyncStatusPath                  = "/eth/v1/node/syncing"
	RequestEth2ConfigPath                  = "/eth/v1/config/spec"
	RequestEth2DepositContractMethod       = "/eth/v1/config/deposit_contract"
	RequestCommitteePath                   = "/eth/v1/beacon/states/%s/committees"
	RequestGenesisPath                     = "/eth/v1/beacon/genesis"
	RequestFinalityCheckpointsPath         = "/eth/v1/beacon/states/%s/finality_checkpoints"
	RequestForkPath                        = "/eth/v1/beacon/states/%s/fork"
	RequestValidatorsPath                  = "/eth/v1/beacon/states/%s/validators"
	RequestVoluntaryExitPath               = "/eth/v1/beacon/pool/voluntary_exits"
	RequestAttestationsPath                = "/eth/v1/beacon/blocks/%s/attestations"
	RequestBeaconBlockPath                 = "/eth/v2/beacon/blocks/%s"
	RequestBeaconBlockHeaderPath           = "/eth/v1/beacon/headers/%s"
	RequestValidatorSyncDuties             = "/eth/v1/validator/duties/sync/%s"
	RequestValidatorProposerDuties         = "/eth/v1/validator/duties/proposer/%s"
	RequestWithdrawalCredentialsChangePath = "/eth/v1/beacon/pool/bls_to_execution_changes"

	MaxRequestValidatorsCount = 600

	DefaultFastTimeout time.Duration = 5 * time.Second
	DefaultSlowTimeout time.Duration = 30 * time.Second
)

type BeaconHttpProviderOpts struct {
	DefaultFastTimeout time.Duration
	DefaultSlowTimeout time.Duration
}

type BeaconHttpProvider struct {
	baseUrl            *url.URL
	httpClient         http.Client
	defaultFastTimeout time.Duration
	defaultSlowTimeout time.Duration
}

// Creates a new HTTP provider for the Beacon API
func NewBeaconHttpProvider(providerAddress string, opts *BeaconHttpProviderOpts) (*BeaconHttpProvider, error) {
	baseUrl, err := url.Parse(providerAddress)
	if err != nil {
		return nil, fmt.Errorf("error parsing provider address [%s] into URL: %w", providerAddress, err)
	}
	provider := &BeaconHttpProvider{
		baseUrl:    baseUrl,
		httpClient: http.Client{},
	}

	// Specify the default timeouts to use for calls that aren't given one by the caller
	if opts != nil {
		provider.defaultFastTimeout = opts.DefaultFastTimeout
		provider.defaultSlowTimeout = opts.DefaultSlowTimeout
	} else {
		provider.defaultFastTimeout = DefaultFastTimeout
		provider.defaultSlowTimeout = DefaultSlowTimeout
	}
	return provider, nil
}

func (p *BeaconHttpProvider) Beacon_Attestations(ctx context.Context, blockId string) (AttestationsResponse, bool, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Run the request
	url := p.baseUrl.JoinPath(fmt.Sprintf(RequestAttestationsPath, blockId))
	responseBody, status, err := p.getRequest(ctx, url)
	if err != nil {
		return AttestationsResponse{}, false, fmt.Errorf("error getting attestations data for slot %s: %w", blockId, err)
	}
	if status == http.StatusNotFound {
		return AttestationsResponse{}, false, nil
	}
	if status != http.StatusOK {
		return AttestationsResponse{}, false, fmt.Errorf("error getting attestations data for slot %s: HTTP status %d; response body: '%s'", blockId, status, string(responseBody))
	}
	var attestations AttestationsResponse
	if err := json.Unmarshal(responseBody, &attestations); err != nil {
		return AttestationsResponse{}, false, fmt.Errorf("error decoding attestations data for slot %s: %w", blockId, err)
	}
	return attestations, true, nil
}

func (p *BeaconHttpProvider) Beacon_Block(ctx context.Context, blockId string) (BeaconBlockResponse, bool, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Run the request
	url := p.baseUrl.JoinPath(fmt.Sprintf(RequestBeaconBlockPath, blockId))
	responseBody, status, err := p.getRequest(ctx, url)
	if err != nil {
		return BeaconBlockResponse{}, false, fmt.Errorf("error getting beacon block data: %w", err)
	}
	if status == http.StatusNotFound {
		return BeaconBlockResponse{}, false, nil
	}
	if status != http.StatusOK {
		return BeaconBlockResponse{}, false, fmt.Errorf("error getting beacon block data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var beaconBlock BeaconBlockResponse
	if err := json.Unmarshal(responseBody, &beaconBlock); err != nil {
		return BeaconBlockResponse{}, false, fmt.Errorf("error decoding beacon block data: %w", err)
	}
	return beaconBlock, true, nil
}

func (p *BeaconHttpProvider) Beacon_BlsToExecutionChanges_Post(ctx context.Context, request BLSToExecutionChangeRequest) error {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Perform the post request
	requestArray := []BLSToExecutionChangeRequest{request} // This route must be wrapped in an array
	url := p.baseUrl.JoinPath(RequestWithdrawalCredentialsChangePath)
	responseBody, status, err := p.postRequest(ctx, url, requestArray)
	if err != nil {
		return fmt.Errorf("error broadcasting withdrawal credentials change for validator %s: %w", request.Message.ValidatorIndex, err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("error broadcasting withdrawal credentials change for validator %s: HTTP status %d; response body: '%s'", request.Message.ValidatorIndex, status, string(responseBody))
	}
	return nil
}

func (p *BeaconHttpProvider) Beacon_Committees(ctx context.Context, stateId string, epoch *uint64) (CommitteesResponse, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultSlowTimeout)
	defer cancel()

	// Create the URL
	url := p.baseUrl.JoinPath(fmt.Sprintf(RequestCommitteePath, stateId))
	if epoch != nil {
		query := url.Query()
		query.Add("epoch", strconv.FormatUint(*epoch, 10))
		url.RawQuery = query.Encode()
	}

	// Committees responses are large, so let the json decoder read it in a buffered fashion
	reader, status, err := p.getRequestReader(ctx, url)
	if err != nil {
		return CommitteesResponse{}, fmt.Errorf("error getting committees: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	if status != http.StatusOK {
		body, _ := io.ReadAll(reader)
		return CommitteesResponse{}, fmt.Errorf("error getting committees: HTTP status %d; response body: '%s'", status, string(body))
	}

	d := committeesDecoderPool.Get().(*committeesDecoder)
	defer func() {
		d.currentReader = nil
		committeesDecoderPool.Put(d)
	}()

	d.currentReader = &reader

	// Begin decoding
	var committees CommitteesResponse
	if err := d.decoder.Decode(&committees); err != nil {
		return CommitteesResponse{}, fmt.Errorf("error decoding committees: %w", err)
	}

	return committees, nil
}

func (p *BeaconHttpProvider) Beacon_FinalityCheckpoints(ctx context.Context, stateId string) (FinalityCheckpointsResponse, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Run the request
	url := p.baseUrl.JoinPath(fmt.Sprintf(RequestFinalityCheckpointsPath, stateId))
	responseBody, status, err := p.getRequest(ctx, url)
	if err != nil {
		return FinalityCheckpointsResponse{}, fmt.Errorf("error getting finality checkpoints: %w", err)
	}
	if status != http.StatusOK {
		return FinalityCheckpointsResponse{}, fmt.Errorf("error getting finality checkpoints: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var finalityCheckpoints FinalityCheckpointsResponse
	if err := json.Unmarshal(responseBody, &finalityCheckpoints); err != nil {
		return FinalityCheckpointsResponse{}, fmt.Errorf("error decoding finality checkpoints: %w", err)
	}
	return finalityCheckpoints, nil
}

func (p *BeaconHttpProvider) Beacon_Genesis(ctx context.Context) (GenesisResponse, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Run the request
	url := p.baseUrl.JoinPath(RequestGenesisPath)
	responseBody, status, err := p.getRequest(ctx, url)
	if err != nil {
		return GenesisResponse{}, fmt.Errorf("error getting genesis data: %w", err)
	}
	if status != http.StatusOK {
		return GenesisResponse{}, fmt.Errorf("error getting genesis data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var genesis GenesisResponse
	if err := json.Unmarshal(responseBody, &genesis); err != nil {
		return GenesisResponse{}, fmt.Errorf("error decoding genesis: %w", err)
	}
	return genesis, nil
}

func (p *BeaconHttpProvider) Beacon_Header(ctx context.Context, blockId string) (BeaconBlockHeaderResponse, bool, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Run the request
	url := p.baseUrl.JoinPath(fmt.Sprintf(RequestBeaconBlockHeaderPath, blockId))
	responseBody, status, err := p.getRequest(ctx, url)
	if err != nil {
		return BeaconBlockHeaderResponse{}, false, fmt.Errorf("error getting beacon block header data: %w", err)
	}
	if status == http.StatusNotFound {
		return BeaconBlockHeaderResponse{}, false, nil
	}
	if status != http.StatusOK {
		return BeaconBlockHeaderResponse{}, false, fmt.Errorf("error getting beacon block header data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var beaconBlock BeaconBlockHeaderResponse
	if err := json.Unmarshal(responseBody, &beaconBlock); err != nil {
		return BeaconBlockHeaderResponse{}, false, fmt.Errorf("error getting beacon block header data: %w", err)
	}
	return beaconBlock, true, nil
}

func (p *BeaconHttpProvider) Beacon_Validators(ctx context.Context, stateId string, ids []string) (ValidatorsResponse, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultSlowTimeout)
	defer cancel()

	// Create the URL
	url := p.baseUrl.JoinPath(fmt.Sprintf(RequestValidatorsPath, stateId))
	if len(ids) > 0 {
		query := url.Query()
		query.Add("id", strings.Join(ids, ","))
		url.RawQuery = query.Encode()
	}

	// Run the request
	responseBody, status, err := p.getRequest(ctx, url)
	if err != nil {
		return ValidatorsResponse{}, fmt.Errorf("error getting validators: %w", err)
	}
	if status != http.StatusOK {
		return ValidatorsResponse{}, fmt.Errorf("error getting validators: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var validators ValidatorsResponse
	if err := json.Unmarshal(responseBody, &validators); err != nil {
		return ValidatorsResponse{}, fmt.Errorf("error decoding validators: %w", err)
	}
	return validators, nil
}

func (p *BeaconHttpProvider) Beacon_VoluntaryExits_Post(ctx context.Context, request VoluntaryExitRequest) error {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Perform the post request
	url := p.baseUrl.JoinPath(RequestVoluntaryExitPath)
	responseBody, status, err := p.postRequest(ctx, url, request)
	if err != nil {
		return fmt.Errorf("error broadcasting exit for validator at index %s: %w", request.Message.ValidatorIndex, err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("error broadcasting exit for validator at index %s: HTTP status %d; response body: '%s'", request.Message.ValidatorIndex, status, string(responseBody))
	}
	return nil
}

func (p *BeaconHttpProvider) Config_DepositContract(ctx context.Context) (Eth2DepositContractResponse, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Run the request
	url := p.baseUrl.JoinPath(RequestEth2DepositContractMethod)
	responseBody, status, err := p.getRequest(ctx, url)
	if err != nil {
		return Eth2DepositContractResponse{}, fmt.Errorf("error getting eth2 deposit contract: %w", err)
	}
	if status != http.StatusOK {
		return Eth2DepositContractResponse{}, fmt.Errorf("error gettingeth2 deposit contract: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var eth2DepositContract Eth2DepositContractResponse
	if err := json.Unmarshal(responseBody, &eth2DepositContract); err != nil {
		return Eth2DepositContractResponse{}, fmt.Errorf("error decoding eth2 deposit contract: %w", err)
	}
	return eth2DepositContract, nil
}

func (p *BeaconHttpProvider) Config_Spec(ctx context.Context) (Eth2ConfigResponse, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Run the request
	url := p.baseUrl.JoinPath(RequestEth2ConfigPath)
	responseBody, status, err := p.getRequest(ctx, url)
	if err != nil {
		return Eth2ConfigResponse{}, fmt.Errorf("error getting eth2 config: %w", err)
	}
	if status != http.StatusOK {
		return Eth2ConfigResponse{}, fmt.Errorf("error getting eth2 config: HTTP status %d; response body: '%s'", status, string(responseBody))
	}

	// Unmarshal the response
	var eth2Config Eth2ConfigResponse
	if err := json.Unmarshal(responseBody, &eth2Config); err != nil {
		return Eth2ConfigResponse{}, fmt.Errorf("error decoding eth2 config: %w", err)
	}
	return eth2Config, nil
}

func (p *BeaconHttpProvider) Node_Syncing(ctx context.Context) (SyncStatusResponse, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Run the request
	url := p.baseUrl.JoinPath(RequestSyncStatusPath)
	responseBody, status, err := p.getRequest(ctx, url)
	if err != nil {
		return SyncStatusResponse{}, fmt.Errorf("error getting node sync status: %w", err)
	}
	if status != http.StatusOK {
		return SyncStatusResponse{}, fmt.Errorf("error getting node sync status: HTTP status %d; response body: '%s'", status, string(responseBody))
	}

	// Unmarshal the response
	var syncStatus SyncStatusResponse
	if err := json.Unmarshal(responseBody, &syncStatus); err != nil {
		return SyncStatusResponse{}, fmt.Errorf("error decoding node sync status: %w", err)
	}
	return syncStatus, nil
}

func (p *BeaconHttpProvider) Validator_DutiesProposer(ctx context.Context, indices []string, epoch uint64) (ProposerDutiesResponse, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Run the request
	url := p.baseUrl.JoinPath(fmt.Sprintf(RequestValidatorProposerDuties, strconv.FormatUint(epoch, 10)))
	responseBody, status, err := p.getRequest(ctx, url)
	if err != nil {
		return ProposerDutiesResponse{}, fmt.Errorf("error getting validator proposer duties: %w", err)
	}
	if status != http.StatusOK {
		return ProposerDutiesResponse{}, fmt.Errorf("error getting validator proposer duties: HTTP status %d; response body: '%s'", status, string(responseBody))
	}

	// Unmarshal the response
	var syncDuties ProposerDutiesResponse
	if err := json.Unmarshal(responseBody, &syncDuties); err != nil {
		return ProposerDutiesResponse{}, fmt.Errorf("error decoding validator proposer duties data: %w", err)
	}
	return syncDuties, nil
}

func (p *BeaconHttpProvider) Validator_DutiesSync_Post(ctx context.Context, indices []string, epoch uint64) (SyncDutiesResponse, error) {
	// Prep the context
	ctx, cancel := p.prepareContext(ctx, p.defaultFastTimeout)
	defer cancel()

	// Perform the post request
	url := p.baseUrl.JoinPath(fmt.Sprintf(RequestValidatorSyncDuties, strconv.FormatUint(epoch, 10)))
	responseBody, status, err := p.postRequest(ctx, url, indices)
	if err != nil {
		return SyncDutiesResponse{}, fmt.Errorf("error getting validator sync duties: %w", err)
	}
	if status != http.StatusOK {
		return SyncDutiesResponse{}, fmt.Errorf("error getting validator sync duties: HTTP status %d; response body: '%s'", status, string(responseBody))
	}

	// Unmarshal the response
	var syncDuties SyncDutiesResponse
	if err := json.Unmarshal(responseBody, &syncDuties); err != nil {
		return SyncDutiesResponse{}, fmt.Errorf("error decoding validator sync duties data: %w", err)
	}
	return syncDuties, nil
}

// ==========================
// === Internal Functions ===
// ==========================

// Make a GET request to the beacon node and read the body of the response
func (p *BeaconHttpProvider) getRequest(ctx context.Context, url *url.URL) ([]byte, int, error) {
	// Send request
	reader, status, err := p.getRequestReader(ctx, url)
	if err != nil {
		return []byte{}, 0, err
	}
	defer func() {
		_ = reader.Close()
	}()

	// Get response
	body, err := io.ReadAll(reader)
	if err != nil {
		return []byte{}, 0, err
	}

	// Return
	return body, status, nil
}

// Make a POST request to the beacon node
func (p *BeaconHttpProvider) postRequest(ctx context.Context, url *url.URL, requestBody any) ([]byte, int, error) {
	// Log the request and add tracing if enabled
	ctx = p.logRequest(ctx, http.MethodPost, url)

	// Get request body
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return []byte{}, 0, err
	}
	requestBodyReader := bytes.NewReader(requestBodyBytes)

	// Create the request
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), requestBodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating POST request to [%s]: %w", url, err)
	}
	request.Header.Set("Content-Type", RequestContentType)

	// Submit the request
	response, err := p.httpClient.Do(request)
	if err != nil {
		// Remove the query for readability
		trimmedPath := url.JoinPath(url.Host, url.Path)
		return []byte{}, 0, fmt.Errorf("error running POST request to [%s]: %w", trimmedPath, err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	// Get response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, 0, err
	}

	// Return
	return body, response.StatusCode, nil
}

// Make a GET request but do not read its body yet (allows buffered decoding)
func (p *BeaconHttpProvider) getRequestReader(ctx context.Context, url *url.URL) (io.ReadCloser, int, error) {
	// Log the request and add tracing if enabled
	ctx = p.logRequest(ctx, http.MethodGet, url)

	// Make the request
	path := url.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating GET request to [%s]: %w", path, err)
	}
	req.Header.Set("Content-Type", RequestContentType)

	// Submit the request
	response, err := p.httpClient.Do(req)
	if err != nil {
		// Remove the query for readability
		trimmedPath := url.JoinPath(url.Host, url.Path)
		return nil, 0, fmt.Errorf("error running GET request to [%s]: %w", trimmedPath, err)
	}
	return response.Body, response.StatusCode, nil
}

// Adds a timeout to the context if one didn't already exist
func (p *BeaconHttpProvider) prepareContext(ctx context.Context, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	// Make a new context if it wasn't provided
	if ctx == nil {
		ctx = context.Background()
	}

	// Return if there was already a deadline
	_, hasDeadline := ctx.Deadline()
	if hasDeadline {
		return ctx, func() {}
	}

	// Add a default timeout if there isn't one
	return context.WithTimeout(ctx, defaultTimeout)
}

// Log a request and prepare the context by adding HTTP tracing if the logger has it enabled
func (p *BeaconHttpProvider) logRequest(ctx context.Context, methodName string, url *url.URL) context.Context {
	logger, _ := log.FromContext(ctx)
	if logger == nil {
		return ctx
	}

	// Log the request
	logger.Debug("Running BN request",
		slog.String(log.MethodKey, methodName),
		slog.String("host", url.Host),
		slog.String("path", url.Path),
	)
	tracer := logger.GetHttpTracer()
	if tracer != nil {
		// Enable HTTP tracing if requested
		ctx = httptrace.WithClientTrace(ctx, tracer)
	}
	return ctx
}

// ==========================
// === Committees Decoder ===
// ==========================

type committeesDecoder struct {
	decoder       *json.Decoder
	currentReader *io.ReadCloser
}

// Read will be called by the json decoder to request more bytes of data from
// the beacon node's committees response. Since the decoder is reused, we
// need to avoid sending it io.EOF, or it will enter an unusable state and can
// not be reused later.
//
// On subsequent calls to Decode, the decoder resets its internal buffer, which
// means any data it reads between the last json token and EOF is correctly
// discarded.
func (c *committeesDecoder) Read(p []byte) (int, error) {
	n, err := (*c.currentReader).Read(p)
	if err == io.EOF {
		return n, nil
	}

	return n, err
}

var committeesDecoderPool sync.Pool = sync.Pool{
	New: func() any {
		var out committeesDecoder

		out.decoder = json.NewDecoder(&out)
		return &out
	},
}
