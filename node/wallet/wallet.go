package wallet

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math/big"
	"os"
	"sync"

	"github.com/goccy/go-json"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/wallet"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
)

// Config
const (
	EntropyBits = 256
	FileMode    = 0600
)

// Errors
var (
	// Attempted to do an operation requiring a wallet, but it's not loaded
	ErrWalletNotLoaded = errors.New("wallet is not loaded")

	// Attempted to load a wallet keystore, but it's already loaded
	ErrWalletAlreadyLoaded = errors.New("wallet is already loaded, nothing to do")

	// Attempted to create a new wallet, but one is already present
	ErrKeystoreAlreadyPresent = errors.New("wallet keystore is already present - please delete it before creating a new wallet")

	// Attempted to load a keystore, but it's not on disk
	ErrKeystoreNotPresent = errors.New("keystore not present, wallet must be initialized or recovered first")

	// Attempted to do an operation that is not supported by the loaded wallet type
	ErrNotSupported = errors.New("loaded wallet type does not support this operation")

	// Provided password is not correct to unlock the wallet keystore
	ErrInvalidPassword = errors.New("provided password is not correct for the loaded wallet")
)

// Wallet
type Wallet struct {
	// Managers
	walletManager   IWalletManager
	addressManager  *addressManager
	passwordManager *passwordManager

	// Misc cache
	chainID        uint
	walletDataPath string

	// Sync
	lock *sync.Mutex
}

// Create new wallet
func NewWallet(logger *slog.Logger, walletDataPath string, walletAddressPath string, passwordFilePath string, chainID uint) (*Wallet, error) {
	// Create the wallet
	w := &Wallet{
		// Create managers
		addressManager:  newAddressManager(walletAddressPath),
		passwordManager: newPasswordManager(passwordFilePath),

		// Initialize other fields
		chainID:        chainID,
		walletDataPath: walletDataPath,
		lock:           &sync.Mutex{},
	}

	// Load the wallet
	return w, w.Reload(logger)
}

// Gets the status of the wallet and its artifacts
func (w *Wallet) GetStatus() (wallet.WalletStatus, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	// Make a status wrapper
	status := wallet.WalletStatus{}

	// Get the password details
	var err error
	_, status.Password.IsPasswordSaved, err = w.passwordManager.GetPasswordFromDisk()
	if err != nil {
		return status, fmt.Errorf("error checking password manager status: %w", err)
	}

	// Get the wallet details
	if w.walletManager != nil {
		status.Wallet.IsLoaded = true
		status.Wallet.Type = w.walletManager.GetType()
		status.Wallet.IsOnDisk = true
		status.Wallet.WalletAddress, err = w.walletManager.GetAddress()
		if err != nil {
			return status, fmt.Errorf("error getting wallet address: %w", err)
		}
	} else {
		status.Wallet.IsOnDisk, err = w.isWalletDataOnDisk()
		if err != nil {
			return status, fmt.Errorf("error checking if wallet data is on disk: %w", err)
		}
	}

	// Get the address details
	status.Address.NodeAddress, status.Address.HasAddress = w.addressManager.GetAddress()
	return status, nil
}

// Reloads the wallet artifacts from disk
func (w *Wallet) Reload(logger *slog.Logger) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	// Load the password
	password, isPasswordSaved, err := w.passwordManager.GetPasswordFromDisk()
	if err != nil {
		return fmt.Errorf("error loading password: %w", err)
	}

	// Load the wallet
	if isPasswordSaved {
		walletMgr, err := w.loadWalletData(password)
		if err != nil && logger != nil {
			logger.Warn("Loading wallet with stored node password failed", slog.String(log.PathKey, w.walletDataPath), log.Err(err))
		} else if walletMgr != nil {
			w.walletManager = walletMgr
		}
	} else {
		w.walletManager = nil
	}

	// Load the node address
	_, _, err = w.addressManager.LoadAddress()
	if err != nil {
		return fmt.Errorf("error loading node address: %w", err)
	}
	return nil
}

// Get the node address, if one is loaded
func (w *Wallet) GetAddress() (common.Address, bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	return w.addressManager.GetAddress()
}

// Get the transactor that can sign transactions
func (w *Wallet) GetTransactor() (*bind.TransactOpts, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.walletManager == nil {
		return nil, ErrWalletNotLoaded
	}

	opts, err := w.walletManager.GetTransactor()
	if err != nil {
		return nil, err
	}

	// Create a copy of the transactor so mods to it don't propagate to the underlying struct
	clone := &bind.TransactOpts{
		From:   opts.From,
		Signer: opts.Signer,

		GasLimit: opts.GasLimit,
		Context:  opts.Context,
		NoSend:   opts.NoSend,
	}
	if opts.Nonce != nil {
		clone.Nonce = big.NewInt(0).Set(opts.Nonce)
	}
	if opts.Value != nil {
		clone.Value = big.NewInt(0).Set(opts.Value)
	}
	if opts.GasPrice != nil {
		clone.GasPrice = big.NewInt(0).Set(opts.GasPrice)
	}
	if opts.GasFeeCap != nil {
		clone.GasFeeCap = big.NewInt(0).Set(opts.GasFeeCap)
	}
	if opts.GasTipCap != nil {
		clone.GasFeeCap = big.NewInt(0).Set(opts.GasTipCap)
	}
	return clone, nil
}

// Sign a message with the wallet's private key
func (w *Wallet) SignMessage(message []byte) ([]byte, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.walletManager == nil {
		return nil, ErrWalletNotLoaded
	}
	return w.walletManager.SignMessage(message)
}

// Sign a transaction with the wallet's private key
func (w *Wallet) SignTransaction(serializedTx []byte) ([]byte, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.walletManager == nil {
		return nil, ErrWalletNotLoaded
	}
	return w.walletManager.SignTransaction(serializedTx)
}

// Masquerade as another node address, running all node functions as that address (in read only mode)
func (w *Wallet) MasqueradeAsAddress(newAddress common.Address) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	return w.masqueradeImpl(newAddress)
}

// End masquerading as another node address, and use the wallet's address (returning to read/write mode)
func (w *Wallet) RestoreAddressToWallet() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.addressManager == nil {
		return ErrWalletNotLoaded
	}

	walletAddress, err := w.walletManager.GetAddress()
	if err != nil {
		return fmt.Errorf("error getting wallet address: %w", err)
	}

	return w.masqueradeImpl(walletAddress)
}

// Initialize the wallet from a random seed
func (w *Wallet) CreateNewLocalWallet(derivationPath string, walletIndex uint, password string, savePassword bool) (string, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.walletManager != nil {
		return "", ErrKeystoreAlreadyPresent
	}

	// Make a mnemonic
	mnemonic, err := GenerateNewMnemonic()
	if err != nil {
		return "", err
	}

	// Initialize the wallet with it
	err = w.buildLocalWallet(derivationPath, walletIndex, mnemonic, password, savePassword, false)
	if err != nil {
		return "", fmt.Errorf("error initializing new wallet keystore: %w", err)
	}
	return mnemonic, nil
}

// Recover a local wallet from a mnemonic
func (w *Wallet) Recover(derivationPath string, walletIndex uint, mnemonic string, password string, savePassword bool, testMode bool) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.walletManager != nil {
		return ErrKeystoreAlreadyPresent
	}

	// Check the mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("invalid mnemonic '%s'", mnemonic)
	}

	return w.buildLocalWallet(derivationPath, walletIndex, mnemonic, password, savePassword, testMode)
}

// Attempts to load the wallet keystore with the provided password if not set
func (w *Wallet) SetPassword(password string, save bool) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.walletManager != nil {
		if !save {
			return ErrWalletAlreadyLoaded
		}

		switch w.walletManager.GetType() {
		case wallet.WalletType_Local:
			// Make sure the password is correct
			localMgr := w.walletManager.(*localWalletManager)
			isValid, err := localMgr.VerifyPassword(password)
			if err != nil {
				return fmt.Errorf("error setting password: %w", err)
			}
			if !isValid {
				return ErrInvalidPassword
			}

			// Save and exit
			return w.passwordManager.SavePassword(password)
		default:
			return ErrNotSupported
		}
	}

	// Try to load the wallet with the new password
	isWalletOnDisk, err := w.isWalletDataOnDisk()
	if err != nil {
		return fmt.Errorf("error checking if wallet data is on disk: %w", err)
	}
	if !isWalletOnDisk {
		return ErrKeystoreNotPresent
	}
	mgr, err := w.loadWalletData(password)
	if err != nil {
		return fmt.Errorf("error loading wallet with provided password: %w", err)
	}

	// Save if requested
	if save {
		err := w.passwordManager.SavePassword(password)
		if err != nil {
			return err
		}
	}

	// Set the wallet manager
	w.walletManager = mgr
	return nil
}

// Retrieves the wallet's password
func (w *Wallet) GetPassword() (string, bool, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	return w.passwordManager.GetPasswordFromDisk()
}

// Delete the wallet password from disk, but retain it in memory if a local keystore is already loaded
func (w *Wallet) DeletePassword() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	err := w.passwordManager.DeletePassword()
	if err != nil {
		return fmt.Errorf("error deleting wallet password: %w", err)
	}
	return nil
}

// Get the node account private key bytes
func (w *Wallet) GetNodePrivateKeyBytes() ([]byte, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.walletManager == nil {
		return nil, ErrWalletNotLoaded
	}

	switch w.walletManager.GetType() {
	case wallet.WalletType_Local:
		localMgr := w.walletManager.(*localWalletManager)
		return crypto.FromECDSA(localMgr.GetPrivateKey()), nil
	default:
		return nil, ErrNotSupported
	}
}

// Get the node account private key bytes
func (w *Wallet) GetEthKeystore(password string) ([]byte, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.walletManager == nil {
		return nil, ErrWalletNotLoaded
	}

	switch w.walletManager.GetType() {
	case wallet.WalletType_Local:
		localMgr := w.walletManager.(*localWalletManager)
		return localMgr.GetEthKeystore(password)
	default:
		return nil, ErrNotSupported
	}
}

// Serialize the wallet data as JSON
func (w *Wallet) SerializeData() (string, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.walletManager == nil {
		return "", ErrWalletNotLoaded
	}
	return w.walletManager.SerializeData()
}

// Generate a BLS validator key from the provided path, using the node wallet's seed as a basis
func (w *Wallet) GenerateValidatorKey(path string) ([]byte, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.walletManager == nil {
		return nil, ErrWalletNotLoaded
	}

	switch w.walletManager.GetType() {
	case wallet.WalletType_Local:
		localMgr := w.walletManager.(*localWalletManager)
		return localMgr.GenerateValidatorKey(path)
	default:
		return nil, ErrNotSupported
	}
}

// Builds a local wallet keystore and saves its artifacts to disk
func (w *Wallet) buildLocalWallet(derivationPath string, walletIndex uint, mnemonic string, password string, savePassword bool, testMode bool) error {
	// Initialize the wallet with it
	localMgr := newLocalWalletManager(w.chainID)
	localData, err := localMgr.InitializeKeystore(derivationPath, walletIndex, mnemonic, password)
	if err != nil {
		return fmt.Errorf("error initializing wallet keystore with recovered data: %w", err)
	}

	// Get the wallet address
	walletAddress, _ := localMgr.GetAddress()

	if !testMode {
		// Create data
		data := &wallet.WalletData{
			Type:      wallet.WalletType_Local,
			LocalData: *localData,
		}

		// Save the wallet data
		err = w.saveWalletData(data)
		if err != nil {
			return fmt.Errorf("error saving wallet data: %w", err)
		}
		// Update the address file
		err = w.addressManager.SetAndSaveAddress(walletAddress)
		if err != nil {
			return fmt.Errorf("error saving wallet address to node address file: %w", err)
		}

		if savePassword {
			err := w.passwordManager.SavePassword(password)
			if err != nil {
				return fmt.Errorf("error saving password: %w", err)
			}
		}
	} else {
		w.addressManager.SetAddress(walletAddress)
	}

	w.walletManager = localMgr
	return nil
}

// Check if the wallet file is saved to disk
func (w *Wallet) isWalletDataOnDisk() (bool, error) {
	// Read the file
	_, err := os.Stat(w.walletDataPath)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking if wallet file [%s] exists: %w", w.walletDataPath, err)
	}
	return true, nil
}

// Load the wallet data from disk
func (w *Wallet) loadWalletData(password string) (IWalletManager, error) {
	// Read the file
	bytes, err := os.ReadFile(w.walletDataPath)
	if err != nil {
		return nil, fmt.Errorf("error reading wallet data at [%s]: %w", w.walletDataPath, err)
	}

	// Deserialize it
	data := new(wallet.WalletData)
	err = json.Unmarshal(bytes, data)
	if err != nil {
		return nil, fmt.Errorf("error deserializing wallet data at [%s]: %w", w.walletDataPath, err)
	}

	// Load the proper type
	var manager IWalletManager
	switch data.Type {
	case wallet.WalletType_Local:
		localMgr := newLocalWalletManager(w.chainID)
		err = localMgr.LoadWallet(&data.LocalData, password)
		if err != nil {
			return nil, fmt.Errorf("error loading local wallet data at %s: %w", w.walletDataPath, err)
		}
		manager = localMgr
	default:
		return nil, fmt.Errorf("unsupported wallet type: %s", data.Type)
	}

	// Data loaded!
	return manager, nil
}

// Save the wallet data to disk
func (w *Wallet) saveWalletData(data *wallet.WalletData) error {
	// Serialize it
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error serializing wallet data: %w", err)
	}

	// Write the file
	err = os.WriteFile(w.walletDataPath, bytes, FileMode)
	if err != nil {
		return fmt.Errorf("error writing wallet data to [%s]: %w", w.walletDataPath, err)
	}
	return nil
}

// Masquerade as another node address, running all node functions as that address (in read only mode)
func (w *Wallet) masqueradeImpl(newAddress common.Address) error {
	return w.addressManager.SetAndSaveAddress(newAddress)
}

// =============
// === Utils ===
// =============

// Generate a new random mnemonic and seed
func GenerateNewMnemonic() (string, error) {
	// Generate random entropy for the mnemonic
	entropy, err := bip39.NewEntropy(EntropyBits)
	if err != nil {
		return "", fmt.Errorf("error generating wallet mnemonic entropy bytes: %w", err)
	}

	// Generate a new mnemonic
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("error generating wallet mnemonic: %w", err)
	}
	return mnemonic, nil
}
