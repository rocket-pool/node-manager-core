package validator

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/beacon/ssz_types"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Get deposit data & root for a given validator key and withdrawal credentials
func GetDepositData(validatorKey *eth2types.BLSPrivateKey, withdrawalCredentials common.Hash, genesisForkVersion []byte, depositAmount uint64, networkName string) (beacon.ExtendedDepositData, error) {
	// Build deposit data
	dd := ssz_types.DepositDataNoSignature{
		PublicKey:             validatorKey.PublicKey().Marshal(),
		WithdrawalCredentials: withdrawalCredentials[:],
		Amount:                depositAmount,
	}
	domain, err := eth2types.ComputeDomain(eth2types.DomainDeposit, genesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		return beacon.ExtendedDepositData{}, fmt.Errorf("error computing domain: %w", err)
	}

	// Get signing root
	messageRoot, err := dd.HashTreeRoot()
	if err != nil {
		return beacon.ExtendedDepositData{}, fmt.Errorf("error getting message root: %w", err)
	}
	dataRoot := ssz_types.SigningRoot{
		ObjectRoot: messageRoot[:],
		Domain:     domain,
	}

	// Get signing root with domain
	dataRootHash, err := dataRoot.HashTreeRoot()
	if err != nil {
		return beacon.ExtendedDepositData{}, err
	}

	// Build deposit data struct (with signature)
	var depositData = ssz_types.DepositData{
		PublicKey:             dd.PublicKey,
		WithdrawalCredentials: dd.WithdrawalCredentials,
		Amount:                dd.Amount,
		Signature:             validatorKey.Sign(dataRootHash[:]).Marshal(),
	}

	// Get deposit data root
	depositDataRoot, err := depositData.HashTreeRoot()
	if err != nil {
		return beacon.ExtendedDepositData{}, err
	}

	// Make sure everything is correct
	err = ValidateDepositInfo(genesisForkVersion, depositAmount, dd.PublicKey, dd.WithdrawalCredentials, depositData.Signature)
	if err != nil {
		return beacon.ExtendedDepositData{}, fmt.Errorf("deposit data failed signature validation: %w", err)
	}

	// Create the extended data
	return beacon.ExtendedDepositData{
		PublicKey:             depositData.PublicKey,
		WithdrawalCredentials: depositData.WithdrawalCredentials,
		Amount:                depositData.Amount,
		Signature:             depositData.Signature,
		DepositMessageRoot:    messageRoot[:],
		DepositDataRoot:       depositDataRoot[:],
		ForkVersion:           genesisForkVersion,
		NetworkName:           networkName,
	}, nil
}

func ValidateDepositInfo(genesisForkVersion []byte, depositAmount uint64, pubkey []byte, withdrawalCredentials []byte, signature []byte) error {
	// Get the deposit domain based on the eth2 config
	depositDomain, err := signing.ComputeDomain(eth2types.DomainDeposit, genesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		return err
	}

	// Create the deposit struct
	depositData := new(ethpb.Deposit_Data)
	depositData.Amount = depositAmount
	depositData.PublicKey = pubkey
	depositData.WithdrawalCredentials = withdrawalCredentials
	depositData.Signature = signature

	// Validate the signature
	err = prdeposit.VerifyDepositSignature(depositData, depositDomain)
	return err
}
