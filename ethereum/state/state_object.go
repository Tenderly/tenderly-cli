package state

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tenderly/tenderly-cli/ethereum"
	"github.com/tenderly/tenderly-cli/ethereum/types"
)

var emptyCodeHash = crypto.Keccak256(nil)

type Code []byte

func (c Code) String() string {
	return string(c) //strings.Join(Disassemble(c), " ")
}

type Storage map[common.Hash]common.Hash

func (s Storage) String() (str string) {
	for key, value := range s {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

func (s Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range s {
		cpy[key] = value
	}

	return cpy
}

// stateObject represents an Ethereum account which is being modified.
//
// The usage pattern is as follows:
// First you need to obtain a state object.
// Account values can be accessed and modified through the object.
// Finally, call CommitTrie to write the modified storage trie into a database.
type stateObject struct {
	address  common.Address
	addrHash common.Hash // hash of ethereum address of the account
	data     Account
	db       *StateDB

	used bool
	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Write caches.
	code Code // contract bytecode, which gets set when code is loaded

	originStorage Storage // Storage cache of original entries to dedup rewrites
	dirtyStorage  Storage // Storage entries that need to be flushed to disk

	// Cache flags.
	// When an object is marked suicided it will be delete from the trie
	// during the "update" phase of the state transition.
	dirtyCode bool // true if the code was updated
	suicided  bool
	deleted   bool
}

// empty returns whether the account is considered empty.
func (s *stateObject) empty() bool {
	return s.data.DirtyNonce == 0 && s.data.DirtyBalance.Sign() == 0 && bytes.Equal(s.data.DirtyCodeHash, emptyCodeHash)
}

func (s *stateObject) Used() bool {
	return s.used
}

// Account is the Ethereum consensus representation of accounts.
// These objects are stored in the main account trie.
type Account struct {
	OriginalNonce    uint64
	DirtyNonce       uint64
	OriginalBalance  *big.Int
	DirtyBalance     *big.Int
	Root             common.Hash // merkle root of the storage trie
	OriginalCodeHash []byte
	DirtyCodeHash    []byte
}

// newObject creates a state object.
func newObject(db *StateDB, address common.Address, data Account, code Code) *stateObject {
	if data.OriginalBalance == nil {
		data.OriginalBalance = new(big.Int)
	}
	if data.OriginalCodeHash == nil {
		data.OriginalCodeHash = emptyCodeHash
	}

	data.DirtyNonce = data.OriginalNonce
	data.DirtyBalance = data.OriginalBalance
	data.DirtyCodeHash = crypto.Keccak256Hash(code).Bytes()

	return &stateObject{
		used:          true,
		db:            db,
		address:       address,
		addrHash:      crypto.Keccak256Hash(address[:]),
		data:          data,
		code:          code,
		originStorage: make(Storage),
		dirtyStorage:  make(Storage),
	}
}

// EncodeRLP implements rlp.Encoder.
func (s *stateObject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, s.data)
}

// setError remembers the first non-nil error it is called with.
func (s *stateObject) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *stateObject) markSuicided() {
	s.suicided = true
}

func (s *stateObject) touch() {
	s.db.journal.append(touchChange{
		account: &s.address,
	})
	if s.address == ripemd {
		// Explicitly put it in the dirty-cache, which is otherwise generated from
		// flattened journals.
		s.db.journal.dirty(s.address)
	}
}

// GetState retrieves a value from the account storage trie.
func (s *stateObject) GetState(client *ethereum.Client, blockNumber int64, key common.Hash) common.Hash {
	// If we have a dirty value for this state entry, return it
	value, dirty := s.dirtyStorage[key]
	if dirty {
		return value
	}
	// Otherwise return the entry's original value
	return s.GetCommittedState(client, blockNumber, key)
}

// GetCommittedState retrieves a value from the committed account storage trie.
func (s *stateObject) GetCommittedState(client *ethereum.Client, blockNumber int64, key common.Hash) common.Hash {
	// If we have the original value cached, return that
	value, cached := s.originStorage[key]
	if cached {
		return value
	}
	// Otherwise load the value from the database
	number := types.Number(blockNumber)
	state, err := client.GetStorageAt(s.address.String(), key, &number)
	if err != nil {
		s.db.setDbErr(err)
		return common.Hash{}
	}
	s.originStorage[key] = *state
	return *state
}

// SetState updates a value in account storage.
func (s *stateObject) SetState(client *ethereum.Client, blockNumber int64, key, value common.Hash) {
	// If the new value is the same as old, don't set
	prev := s.GetState(client, blockNumber, key)
	if prev == value {
		return
	}
	// New value is different, update and journal the change
	s.db.journal.append(storageChange{
		account:  &s.address,
		key:      key,
		prevalue: prev,
	})
	s.setState(key, value)
}

func (s *stateObject) setState(key, value common.Hash) {
	s.dirtyStorage[key] = value
}

// AddBalance removes amount from c's balance.
// It is used to add funds to the destination account of a transfer.
func (s *stateObject) AddBalance(amount *big.Int) {
	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount.Sign() == 0 {
		if s.empty() {
			s.touch()
		}

		return
	}
	s.SetBalance(new(big.Int).Add(s.Balance(), amount))
}

// SubBalance removes amount from c's balance.
// It is used to remove funds from the origin account of a transfer.
func (s *stateObject) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	s.SetBalance(new(big.Int).Sub(s.Balance(), amount))
}

func (s *stateObject) SetBalance(amount *big.Int) {
	s.db.journal.append(balanceChange{
		account: &s.address,
		prev:    new(big.Int).Set(s.data.DirtyBalance),
	})
	s.setBalance(amount)
}

func (s *stateObject) setBalance(amount *big.Int) {
	s.data.DirtyBalance = amount
}

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (s *stateObject) ReturnGas(gas *big.Int) {}

//
// Attribute accessors
//

// Returns the address of the contract/account
func (s *stateObject) Address() common.Address {
	return s.address
}

// Code returns the contract code associated with this object, if any.
func (s *stateObject) Code(client *ethereum.Client, blockNumber int64) []byte {
	if s.code != nil {
		return s.code
	}
	if bytes.Equal(s.CodeHash(), emptyCodeHash) {
		return nil
	}

	number := types.Number(blockNumber)
	code, err := client.GetCode(s.address.String(), &number)
	if err != nil {
		s.db.setDbErr(err)
	}

	raw := code
	if strings.HasPrefix(raw, "0x") {
		raw = raw[2:]
	}
	bin, err := hex.DecodeString(raw)
	if err != nil {
		return []byte{}
	}

	s.code = Code(bin)
	return bin
}

func (s *stateObject) SetCode(client *ethereum.Client, blockNumber int64, codeHash common.Hash, code []byte) {
	prevcode := s.Code(client, blockNumber)
	s.db.journal.append(codeChange{
		account:  &s.address,
		prevhash: s.CodeHash(),
		prevcode: prevcode,
	})
	s.setCode(codeHash, code)
}

func (s *stateObject) setCode(codeHash common.Hash, code []byte) {
	s.code = code
	s.data.DirtyCodeHash = codeHash[:]
	s.dirtyCode = true
}

func (s *stateObject) SetNonce(nonce uint64) {
	s.db.journal.append(nonceChange{
		account: &s.address,
		prev:    s.data.DirtyNonce,
	})
	s.setNonce(nonce)
}

func (s *stateObject) setNonce(nonce uint64) {
	s.data.DirtyNonce = nonce
}

func (s *stateObject) OriginalCodeHash() []byte {
	return s.data.OriginalCodeHash
}

func (s *stateObject) CodeHash() []byte {
	return s.data.DirtyCodeHash
}

func (s *stateObject) Root() common.Hash {
	return s.data.Root
}

func (s *stateObject) OriginalBalance() *big.Int {
	return s.data.OriginalBalance
}

func (s *stateObject) Balance() *big.Int {
	return s.data.DirtyBalance
}

func (s *stateObject) OriginalNonce() uint64 {
	return s.data.OriginalNonce
}

func (s *stateObject) Nonce() uint64 {
	return s.data.DirtyNonce
}

func (s *stateObject) GetStorage() map[string][]byte {
	storage := make(map[string][]byte)
	for k, v := range s.originStorage {
		storage[k.String()] = v.Bytes()
	}

	return storage
}

func (s *stateObject) GetCode() []byte {
	if bytes.Compare(s.OriginalCodeHash(), emptyCodeHash) == 0 {
		return []byte{}
	}

	return s.code
}

func (s *stateObject) finalise() {
	if !s.used {
		return
	}

	s.used = false

	s.data.OriginalNonce = s.data.DirtyNonce
	s.data.OriginalBalance = s.data.DirtyBalance
	s.data.OriginalCodeHash = s.data.DirtyCodeHash

	if len(s.dirtyStorage) > 0 {
		for k, v := range s.dirtyStorage {
			s.originStorage[k] = v
		}
	}
}

// Never called, but must be present to allow stateObject to be used
// as a vm.Account interface that also satisfies the vm.ContractRef
// interface. Interfaces are awesome.
func (s *stateObject) Value() *big.Int {
	panic("Value on stateObject should never be called")
}
