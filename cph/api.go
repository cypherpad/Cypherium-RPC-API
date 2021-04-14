// Copyright 2015 The cypherBFT Authors
// This file is part of the cypherBFT library.
//
// The cypherBFT library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The cypherBFT library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the cypherBFT library. If not, see <http://www.gnu.org/licenses/>.

package cph

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cypherium/cypherBFT/common"
	"github.com/cypherium/cypherBFT/common/hexutil"
	"github.com/cypherium/cypherBFT/core"
	"github.com/cypherium/cypherBFT/core/rawdb"
	"github.com/cypherium/cypherBFT/core/state"
	"github.com/cypherium/cypherBFT/core/types"
	"github.com/cypherium/cypherBFT/internal/cphapi"
	"github.com/cypherium/cypherBFT/log"
	"github.com/cypherium/cypherBFT/params"
	"github.com/cypherium/cypherBFT/reconfig/bftview"
	"github.com/cypherium/cypherBFT/rlp"
	"github.com/cypherium/cypherBFT/rpc"
	"github.com/cypherium/cypherBFT/trie"

	"golang.org/x/crypto/ed25519"
)

// PublicCphereumAPI provides an API to access Cypherium full node-related
// information.
type PublicCphereumAPI struct {
	e *Cypherium
}

// NewPublicCphereumAPI creates a new Cypherium protocol API for full nodes.
func NewPublicCphereumAPI(e *Cypherium) *PublicCphereumAPI {
	return &PublicCphereumAPI{e}
}

// Cpherbase is the address that mining rewards will be send to
func (api *PublicCphereumAPI) Cpherbase() (common.Address, error) {
	return api.e.Cpherbase()
}

// Coinbase is the address that mining rewards will be send to (alias for Cpherbase)
func (api *PublicCphereumAPI) Coinbase() (common.Address, error) {
	return api.Cpherbase()
}

// Hashrate returns the POW hashrate
func (api *PublicCphereumAPI) Hashrate() hexutil.Uint64 {
	return hexutil.Uint64(api.e.Miner().HashRate())
}

// PublicMinerAPI provides an API to control the miner.
// It offers only methods that operate on data that pose no security risk when it is publicly accessible.
//type PublicMinerAPI struct {
//	e     *Cypherium
//	agent *miner.RemoteAgent
//}

// NewPublicMinerAPI create a new PublicMinerAPI instance.
//func NewPublicMinerAPI(e *Cypherium) *PublicMinerAPI {
//	agent := miner.NewRemoteAgent(e.KeyBlockChain(), e.Engine())
//	e.Miner().Register(agent)
//
//	return &PublicMinerAPI{e, agent}
//}

// Mining returns an indication if this node is currently mining.
//func (api *PublicMinerAPI) Mining() bool {
//	return api.e.IsMining()
//}

// SubmitWork can be used by external miner to submit their POW solution. It returns an indication if the work was
// accepted. Note, this is not an indication if the provided work was valid!
//func (api *PublicMinerAPI) SubmitWork(nonce types.BlockNonce, solution, digest common.Hash) bool {
//	return api.agent.SubmitWork(nonce, digest, solution)
//}

// GetWork returns a work package for external miner. The work package consists of 3 strings
// result[0], 32 bytes hex encoded current block header pow-hash
// result[1], 32 bytes hex encoded seed hash used for DAG
// result[2], 32 bytes hex encoded boundary condition ("target"), 2^256/difficulty
//func (api *PublicMinerAPI) GetWork() ([3]string, error) {
//	if !api.e.IsMining() {
//		return [3]string{}, errors.New("miner is not running")
//	}
//	work, err := api.agent.GetWork()
//	if err != nil {
//		return work, fmt.Errorf("mining not ready: %v", err)
//	}
//	return work, nil
//}

// SubmitHashrate can be used for remote miners to submit their hash rate. This enables the node to report the combined
// hash rate of all miners which submit work through this node. It accepts the miner hash rate and an identifier which
// must be unique between nodes.
//func (api *PublicMinerAPI) SubmitHashrate(hashrate hexutil.Uint64, id common.Hash) bool {
//	api.agent.SubmitHashrate(id, uint64(hashrate))
//	return true
//}

// PrivateMinerAPI provides private RPC methods to control the miner.
// These methods can be abused by external users and must be considered insecure for use by untrusted users.
type PrivateMinerAPI struct {
	e *Cypherium
}

// NewPrivateMinerAPI create a new RPC service which controls the miner of this node.
func NewPrivateMinerAPI(e *Cypherium) *PrivateMinerAPI {
	return &PrivateMinerAPI{e: e}
}

// Start the miner with the given number of threads. If threads is nil the number
// of workers started is equal to the number of logical CPUs that are usable by
// this process. If mining is already running, this method adjust the number of
// threads allowed to use.
func (api *PrivateMinerAPI) Start(threads *int, addr common.Address, password string) error {
	var (
		err    error
		eb     common.Address
		prvKey ed25519.PrivateKey
		pubKey ed25519.PublicKey
	)
	server := &common.NodeConfig{}

	if addr != (common.Address{}) {
		eb = addr
	}

	for _, wallet := range api.e.AccountManager().Wallets() {
		for _, account := range wallet.Accounts() {
			if account.Address == eb {
				//wallet.GetPubKey(account, passwd)
				pubKey, prvKey, err = wallet.GetKeyPair(account, password)
				if err != nil {
					log.Error("Cannot start reconfig without public key of cpherbase", "err", err)
					return fmt.Errorf("cpherbase missing public key: %v", err)
				}
				server.Public = common.HexString(pubKey)
				server.Private = common.HexString(prvKey)
			}
		}
	}

	if pubKey == nil || prvKey == nil {
		log.Error("Cannot start reconfig without correct public key")
		return errors.New("missing public key")
	}
	log.Warn("pubKey", "pubKey", server.Public, "prvKey", server.Private)
	log.Warn("exip", "ip", api.e.ExtIP(), "port", api.e.config.OnetPort)
	server.Ip = api.e.ExtIP().String()
	server.Port = api.e.config.OnetPort
	api.e.reconfig.Start(server)
	// Start the miner and return
	// Set the number of threads if the seal engine supports it
	//if threads == nil {
	//	threads = new(int)
	//} else if *threads == 0 {
	//	*threads = -1 // Disable the miner from within
	//}
	//type threaded interface {
	//	SetThreads(threads int)
	//}
	//if th, ok := api.e.engine.(threaded); ok {
	//	log.Info("Updated mining threads", "threads", *threads)
	//	th.SetThreads(*threads)
	//}
	if !api.e.IsMining() {
		return api.e.StartMining(true, eb, pubKey)
	}
	return nil
}

// Stop the miner
func (api *PrivateMinerAPI) Stop() bool {
	type threaded interface {
		SetThreads(threads int)
	}
	if th, ok := api.e.engine.(threaded); ok {
		th.SetThreads(-1)
	}
	api.e.StopMining()
	api.e.reconfig.Stop()
	return true
}

func (api *PrivateMinerAPI) Status() string {
	var s string
	i := bftview.IamMember()
	if i >= 0 {
		if i == 0 {
			s = "I'm leader."
		} else {
			s = "I'm committee member."
		}
	} else {
		s += "I'm common node."
	}
	running := api.e.IsMining() || api.e.reconfigIsRunning()
	if running {
		s += "is Running."
	} else {
		s += "Stopped."
	}
	return s
}

// PrivateReconfigAPI
type PrivateReconfigAPI struct {
	e *Cypherium
}

// NewPrivateMinerAPI
func NewPrivateReconfigAPI(e *Cypherium) *PrivateReconfigAPI {
	return &PrivateReconfigAPI{e: e}
}

// Start .
func (api *PrivateReconfigAPI) Start(threads *int, addr common.Address) error {
	log.Info("PrivateReconfigAPI start")
	return nil
}

func (api *PublicCphereumAPI) Status() string {
	var s string
	i := bftview.IamMember()

	if i >= 0 {
		if i == 0 {
			s = "I'm leader."
		} else {
			s = "I'm committee member."
		}
	} else {
		s += "I'm common node."

	}
	if api.e.IsMining() {
		s += "is Running."
	} else {
		s += "Stopped."
	}
	if api.e.reconfigIsRunning() {
		s += "&& in service."
	} else {
		s += "&& not in service."
	}
	return s
}

// PrivateAdminAPI is the collection of Cypherium full node-related APIs
// exposed over the private admin endpoint.
type PrivateAdminAPI struct {
	cph *Cypherium
}

// NewPrivateAdminAPI creates a new API definition for the full node private
// admin methods of the Cypherium service.
func NewPrivateAdminAPI(cph *Cypherium) *PrivateAdminAPI {
	return &PrivateAdminAPI{cph: cph}
}

// ExportChain exports the current blockchain into a local file.
func (api *PrivateAdminAPI) ExportChain(file string) (bool, error) {
	// Make sure we can create the file to export into
	out, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return false, err
	}
	defer out.Close()

	var writer io.Writer = out
	if strings.HasSuffix(file, ".gz") {
		writer = gzip.NewWriter(writer)
		defer writer.(*gzip.Writer).Close()
	}

	// Export the blockchain
	if err := api.cph.BlockChain().Export(writer); err != nil {
		return false, err
	}
	return true, nil
}

func hasAllBlocks(chain *core.BlockChain, bs []*types.Block) bool {
	for _, b := range bs {
		if !chain.HasBlock(b.Hash(), b.NumberU64()) {
			return false
		}
	}

	return true
}

// ImportChain imports a blockchain from a local file.
func (api *PrivateAdminAPI) ImportChain(file string) (bool, error) {
	// Make sure the can access the file to import
	in, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer in.Close()

	var reader io.Reader = in
	if strings.HasSuffix(file, ".gz") {
		if reader, err = gzip.NewReader(reader); err != nil {
			return false, err
		}
	}

	// Run actual the import in pre-configured batches
	stream := rlp.NewStream(reader, 0)

	blocks, index := make([]*types.Block, 0, 2500), 0
	for batch := 0; ; batch++ {
		// Load a batch of blocks from the input file
		for len(blocks) < cap(blocks) {
			block := new(types.Block)
			if err := stream.Decode(block); err == io.EOF {
				break
			} else if err != nil {
				return false, fmt.Errorf("block %d: failed to parse: %v", index, err)
			}
			blocks = append(blocks, block)
			index++
		}
		if len(blocks) == 0 {
			break
		}

		if hasAllBlocks(api.cph.BlockChain(), blocks) {
			blocks = blocks[:0]
			continue
		}
		// Import the batch and reset the buffer
		if _, err := api.cph.BlockChain().InsertChain(blocks); err != nil {
			return false, fmt.Errorf("batch %d: failed to insert: %v", batch, err)
		}
		blocks = blocks[:0]
	}
	return true, nil
}

// PublicDebugAPI is the collection of Cypherium full node APIs exposed
// over the public debugging endpoint.
type PublicDebugAPI struct {
	cph *Cypherium
}

// NewPublicDebugAPI creates a new API definition for the full node-
// related public debug methods of the Cypherium service.
func NewPublicDebugAPI(cph *Cypherium) *PublicDebugAPI {
	return &PublicDebugAPI{cph: cph}
}

// DumpBlock retrieves the entire state of the database at a given block.
func (api *PublicDebugAPI) DumpBlock(blockNr rpc.BlockNumber) (state.Dump, error) {
	if blockNr == rpc.PendingBlockNumber {
		// If we're dumping the pending state, we need to request
		// both the pending block as well as the pending state from
		// the miner and operate on those
		//_, stateDb := api.cph.miner.Pending()
		//return stateDb.RawDump(), nil

		return state.Dump{}, errors.New("No pending block for Cypherium")
	}
	var block *types.Block
	if blockNr == rpc.LatestBlockNumber {
		block = api.cph.blockchain.CurrentBlock()
	} else {
		block = api.cph.blockchain.GetBlockByNumber(uint64(blockNr))
	}
	if block == nil {
		return state.Dump{}, fmt.Errorf("block #%d not found", blockNr)
	}
	stateDb, err := api.cph.BlockChain().StateAt(block.Root())
	if err != nil {
		return state.Dump{}, err
	}
	return stateDb.RawDump(), nil
}

// PrivateDebugAPI is the collection of Cypherium full node APIs exposed over
// the private debugging endpoint.
type PrivateDebugAPI struct {
	config *params.ChainConfig
	cph    *Cypherium
}

// NewPrivateDebugAPI creates a new API definition for the full node-related
// private debug methods of the Cypherium service.
func NewPrivateDebugAPI(config *params.ChainConfig, cph *Cypherium) *PrivateDebugAPI {
	return &PrivateDebugAPI{config: config, cph: cph}
}

// Preimage is a debug API function that returns the preimage for a sha3 hash, if known.
func (api *PrivateDebugAPI) Preimage(ctx context.Context, hash common.Hash) (hexutil.Bytes, error) {
	if preimage := rawdb.ReadPreimage(api.cph.ChainDb(), hash); preimage != nil {
		return preimage, nil
	}
	return nil, errors.New("unknown preimage")
}

// BadBlockArgs represents the entries in the list returned when bad blocks are queried.
type BadBlockArgs struct {
	Hash  common.Hash            `json:"hash"`
	Block map[string]interface{} `json:"block"`
	RLP   string                 `json:"rlp"`
}

// GetBadBlocks returns a list of the last 'bad blocks' that the client has seen on the network
// and returns them as a JSON list of block-hashes
func (api *PrivateDebugAPI) GetBadBlocks(ctx context.Context) ([]*BadBlockArgs, error) {
	blocks := api.cph.BlockChain().BadBlocks()
	results := make([]*BadBlockArgs, len(blocks))

	var err error
	for i, block := range blocks {
		results[i] = &BadBlockArgs{
			Hash: block.Hash(),
		}
		if rlpBytes, err := rlp.EncodeToBytes(block); err != nil {
			results[i].RLP = err.Error() // Hacky, but hey, it works
		} else {
			results[i].RLP = fmt.Sprintf("0x%x", rlpBytes)
		}
		if results[i].Block, err = cphapi.RPCMarshalBlock(block, true, true); err != nil {
			results[i].Block = map[string]interface{}{"error": err.Error()}
		}
	}
	return results, nil
}

// StorageRangeResult is the result of a debug_storageRangeAt API call.
type StorageRangeResult struct {
	Storage storageMap   `json:"storage"`
	NextKey *common.Hash `json:"nextKey"` // nil if Storage includes the last key in the trie.
}

type storageMap map[common.Hash]storageEntry

type storageEntry struct {
	Key   *common.Hash `json:"key"`
	Value common.Hash  `json:"value"`
}

// StorageRangeAt returns the storage at the given block height and transaction index.
func (api *PrivateDebugAPI) StorageRangeAt(ctx context.Context, blockHash common.Hash, txIndex int, contractAddress common.Address, keyStart hexutil.Bytes, maxResult int) (StorageRangeResult, error) {
	_, _, statedb, err := api.computeTxEnv(blockHash, txIndex, 0)
	if err != nil {
		return StorageRangeResult{}, err
	}
	st := statedb.StorageTrie(contractAddress)
	if st == nil {
		return StorageRangeResult{}, fmt.Errorf("account %x doesn't exist", contractAddress)
	}
	return storageRangeAt(st, keyStart, maxResult)
}

func storageRangeAt(st state.Trie, start []byte, maxResult int) (StorageRangeResult, error) {
	it := trie.NewIterator(st.NodeIterator(start))
	result := StorageRangeResult{Storage: storageMap{}}
	for i := 0; i < maxResult && it.Next(); i++ {
		_, content, _, err := rlp.Split(it.Value)
		if err != nil {
			return StorageRangeResult{}, err
		}
		e := storageEntry{Value: common.BytesToHash(content)}
		if preimage := st.GetKey(it.Key); preimage != nil {
			preimage := common.BytesToHash(preimage)
			e.Key = &preimage
		}
		result.Storage[common.BytesToHash(it.Key)] = e
	}
	// Add the 'next key' so clients can continue downloading.
	if it.Next() {
		next := common.BytesToHash(it.Key)
		result.NextKey = &next
	}
	return result, nil
}

// GetModifiedAccountsByNumber returns all accounts that have changed between the
// two blocks specified. A change is defined as a difference in nonce, balance,
// code hash, or storage hash.
//
// With one parameter, returns the list of accounts modified in the specified block.
func (api *PrivateDebugAPI) GetModifiedAccountsByNumber(startNum uint64, endNum *uint64) ([]common.Address, error) {
	var startBlock, endBlock *types.Block

	startBlock = api.cph.blockchain.GetBlockByNumber(startNum)
	if startBlock == nil {
		return nil, fmt.Errorf("start block %x not found", startNum)
	}

	if endNum == nil {
		endBlock = startBlock
		startBlock = api.cph.blockchain.GetBlockByHash(startBlock.ParentHash())
		if startBlock == nil {
			return nil, fmt.Errorf("block %x has no parent", endBlock.Number())
		}
	} else {
		endBlock = api.cph.blockchain.GetBlockByNumber(*endNum)
		if endBlock == nil {
			return nil, fmt.Errorf("end block %d not found", *endNum)
		}
	}
	return api.getModifiedAccounts(startBlock, endBlock)
}

// GetModifiedAccountsByHash returns all accounts that have changed between the
// two blocks specified. A change is defined as a difference in nonce, balance,
// code hash, or storage hash.
//
// With one parameter, returns the list of accounts modified in the specified block.
func (api *PrivateDebugAPI) GetModifiedAccountsByHash(startHash common.Hash, endHash *common.Hash) ([]common.Address, error) {
	var startBlock, endBlock *types.Block
	startBlock = api.cph.blockchain.GetBlockByHash(startHash)
	if startBlock == nil {
		return nil, fmt.Errorf("start block %x not found", startHash)
	}

	if endHash == nil {
		endBlock = startBlock
		startBlock = api.cph.blockchain.GetBlockByHash(startBlock.ParentHash())
		if startBlock == nil {
			return nil, fmt.Errorf("block %x has no parent", endBlock.Number())
		}
	} else {
		endBlock = api.cph.blockchain.GetBlockByHash(*endHash)
		if endBlock == nil {
			return nil, fmt.Errorf("end block %x not found", *endHash)
		}
	}
	return api.getModifiedAccounts(startBlock, endBlock)
}

func (api *PrivateDebugAPI) getModifiedAccounts(startBlock, endBlock *types.Block) ([]common.Address, error) {
	if startBlock.Number().Uint64() >= endBlock.Number().Uint64() {
		return nil, fmt.Errorf("start block height (%d) must be less than end block height (%d)", startBlock.Number().Uint64(), endBlock.Number().Uint64())
	}

	oldTrie, err := trie.NewSecure(startBlock.Root(), trie.NewDatabase(api.cph.chainDb), 0)
	if err != nil {
		return nil, err
	}
	newTrie, err := trie.NewSecure(endBlock.Root(), trie.NewDatabase(api.cph.chainDb), 0)
	if err != nil {
		return nil, err
	}

	diff, _ := trie.NewDifferenceIterator(oldTrie.NodeIterator([]byte{}), newTrie.NodeIterator([]byte{}))
	iter := trie.NewIterator(diff)

	var dirty []common.Address
	for iter.Next() {
		key := newTrie.GetKey(iter.Key)
		if key == nil {
			return nil, fmt.Errorf("no preimage found for hash %x", iter.Key)
		}
		dirty = append(dirty, common.BytesToAddress(key))
	}
	return dirty, nil
}
