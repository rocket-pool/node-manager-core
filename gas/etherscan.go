package gas

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/goccy/go-json"
)

const gasOracleUrl string = "https://api.etherscan.io/api?module=gastracker&action=gasoracle"

// Standard response
type gasOracleResponse struct {
	Status  uinteger `json:"status"`
	Message string   `json:"message"`
	Result  struct {
		SafeGasPrice    uinteger `json:"SafeGasPrice"`
		ProposeGasPrice uinteger `json:"ProposeGasPrice"`
		FastGasPrice    uinteger `json:"FastGasPrice"`
	} `json:"result"`
}

type EtherscanGasFeeSuggestion struct {
	SlowGwei     float64
	StandardGwei float64
	FastGwei     float64
}

// Get gas prices
func GetEtherscanGasPrices() (EtherscanGasFeeSuggestion, error) {
	// Send request
	response, err := http.Get(gasOracleUrl)
	if err != nil {
		return EtherscanGasFeeSuggestion{}, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	// Check the response code
	if response.StatusCode != http.StatusOK {
		return EtherscanGasFeeSuggestion{}, fmt.Errorf("request failed with code %d", response.StatusCode)
	}

	// Get response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return EtherscanGasFeeSuggestion{}, err
	}

	// Deserialize response
	var oracleResponse gasOracleResponse
	if err := json.Unmarshal(body, &oracleResponse); err != nil {
		return EtherscanGasFeeSuggestion{}, fmt.Errorf("error deserializing Etherscan gas oracle response: %w", err)
	}
	if oracleResponse.Status != 1 {
		return EtherscanGasFeeSuggestion{}, fmt.Errorf("error retrieving Etherscan gas oracle response: %s", oracleResponse.Message)
	}

	suggestion := EtherscanGasFeeSuggestion{
		SlowGwei:     float64(oracleResponse.Result.SafeGasPrice),
		StandardGwei: float64(oracleResponse.Result.ProposeGasPrice),
		FastGwei:     float64(oracleResponse.Result.FastGasPrice),
	}

	// Return
	return suggestion, nil
}

// Unsigned integer type
type uinteger uint64

func (i uinteger) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Itoa(int(i)))
}

func (i *uinteger) UnmarshalJSON(data []byte) error {

	// Unmarshal string
	var dataStr string
	if err := json.Unmarshal(data, &dataStr); err != nil {
		return err
	}

	// Parse integer value
	value, err := strconv.ParseUint(dataStr, 10, 64)
	if err != nil {
		return err
	}

	// Set value and return
	*i = uinteger(value)
	return nil
}
