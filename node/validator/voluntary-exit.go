package validator

import (
	"fmt"
	"strconv"

	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/beacon/ssz_types"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Get a voluntary exit message signature for a given validator key and index.
// Validates the signature prior to returning.
func GetSignedExitMessage(validatorKey *eth2types.BLSPrivateKey, validatorIndex string, epoch uint64, signatureDomain []byte) (beacon.ValidatorSignature, error) {
	return getSignedExitMessageImpl(validatorKey, validatorIndex, epoch, signatureDomain, true)
}

// Get a voluntary exit message signature for a given validator key and index.
// NOTE: This ignores signature validation - use with caution!
func GetSignedExitMessageWithoutValidation(validatorKey *eth2types.BLSPrivateKey, validatorIndex string, epoch uint64, signatureDomain []byte) (beacon.ValidatorSignature, error) {
	return getSignedExitMessageImpl(validatorKey, validatorIndex, epoch, signatureDomain, false)
}

// Implementation for getting a voluntary exit message signature
func getSignedExitMessageImpl(validatorKey *eth2types.BLSPrivateKey, validatorIndex string, epoch uint64, signatureDomain []byte, validate bool) (beacon.ValidatorSignature, error) {
	// Parse the validator index
	indexNum, err := strconv.ParseUint(validatorIndex, 10, 64)
	if err != nil {
		return beacon.ValidatorSignature{}, fmt.Errorf("error parsing validator index (%s): %w", validatorIndex, err)
	}
	// Build voluntary exit message
	exitMessage := ssz_types.VoluntaryExit{
		Epoch:          epoch,
		ValidatorIndex: indexNum,
	}
	// Get object root
	or, err := exitMessage.HashTreeRoot()
	if err != nil {
		return beacon.ValidatorSignature{}, err
	}
	// Get signing root
	sr := ssz_types.SigningRoot{
		ObjectRoot: or[:],
		Domain:     signatureDomain,
	}

	srHash, err := sr.HashTreeRoot()
	if err != nil {
		return beacon.ValidatorSignature{}, err
	}
	// Sign message
	signature := validatorKey.Sign(srHash[:]).Marshal()

	if validate {
		// Validate the signature
		pubkey := beacon.ValidatorPubkey(validatorKey.PublicKey().Marshal())
		err = ValidateExitMessageSignature(pubkey, validatorIndex, signatureDomain, epoch, signature)
		if err != nil {
			return beacon.ValidatorSignature{}, err
		}
	}
	return beacon.ValidatorSignature(signature), nil
}

// Validate a voluntary exit message signature
func ValidateExitMessageSignature(pubkey beacon.ValidatorPubkey, validatorIndex string, signatureDomain []byte, epoch uint64, signature []byte) error {
	// Parse the index
	indexUint64, err := strconv.ParseUint(validatorIndex, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing validator index (%s): %w", validatorIndex, err)
	}

	// Use Prysm to validate the signature
	prysmExit := &ethpb.VoluntaryExit{
		Epoch:          primitives.Epoch(epoch),
		ValidatorIndex: primitives.ValidatorIndex(indexUint64),
	}
	err = signing.VerifySigningRoot(prysmExit, pubkey[:], signature, signatureDomain)
	if err != nil {
		return fmt.Errorf("error verifying exit message signature: %w", err)
	}
	return nil
}
