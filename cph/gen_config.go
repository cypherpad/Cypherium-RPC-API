// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package cph

import (
	"math/big"
	"time"

	"github.com/cypherium/cypherBFT/common/hexutil"
	"github.com/cypherium/cypherBFT/core"
	"github.com/cypherium/cypherBFT/cph/downloader"
	"github.com/cypherium/cypherBFT/cph/gasprice"
	"github.com/cypherium/cypherBFT/pow/cphash"
)

var _ = (*configMarshaling)(nil)

// MarshalTOML marshals as TOML.
func (c Config) MarshalTOML() (interface{}, error) {
	type Config struct {
		GenesisKey              *core.GenesisKey `toml:",omitempty"`
		Genesis                 *core.Genesis    `toml:",omitempty"`
		NetworkId               uint64
		SyncMode                downloader.SyncMode
		NoPruning               bool
		LightServ               int  `toml:",omitempty"`
		LightPeers              int  `toml:",omitempty"`
		SkipBcVersionCheck      bool `toml:"-"`
		DatabaseHandles         int  `toml:"-"`
		DatabaseCache           int
		TrieCache               int
		TrieTimeout             time.Duration
		MinerThreads            int           `toml:",omitempty"`
		ExtraData               hexutil.Bytes `toml:",omitempty"`
		GasPrice                *big.Int
		Cphash                  cphash.Config
		TxPool                  core.TxPoolConfig
		GPO                     gasprice.Config
		EnablePreimageRecording bool
		DocRoot                 string `toml:"-"`
		RnetPort                string
	}
	var enc Config
	enc.GenesisKey = c.GenesisKey
	enc.Genesis = c.Genesis
	enc.NetworkId = c.NetworkId
	enc.SyncMode = c.SyncMode
	enc.NoPruning = c.NoPruning
	enc.LightServ = c.LightServ
	enc.LightPeers = c.LightPeers
	enc.SkipBcVersionCheck = c.SkipBcVersionCheck
	enc.DatabaseHandles = c.DatabaseHandles
	enc.DatabaseCache = c.DatabaseCache
	enc.TrieCache = c.TrieCache
	enc.TrieTimeout = c.TrieTimeout
	enc.MinerThreads = c.MinerThreads
	enc.ExtraData = c.ExtraData
	enc.GasPrice = c.GasPrice
	enc.Cphash = c.Cphash
	enc.TxPool = c.TxPool
	enc.GPO = c.GPO
	enc.EnablePreimageRecording = c.EnablePreimageRecording
	enc.DocRoot = c.DocRoot
	enc.RnetPort = c.RnetPort
	return &enc, nil
}

// UnmarshalTOML unmarshals from TOML.
func (c *Config) UnmarshalTOML(unmarshal func(interface{}) error) error {
	type Config struct {
		GenesisKey              *core.GenesisKey `toml:",omitempty"`
		Genesis                 *core.Genesis    `toml:",omitempty"`
		NetworkId               *uint64
		SyncMode                *downloader.SyncMode
		NoPruning               *bool
		LightServ               *int  `toml:",omitempty"`
		LightPeers              *int  `toml:",omitempty"`
		SkipBcVersionCheck      *bool `toml:"-"`
		DatabaseHandles         *int  `toml:"-"`
		DatabaseCache           *int
		TrieCache               *int
		TrieTimeout             *time.Duration
		MinerThreads            *int           `toml:",omitempty"`
		ExtraData               *hexutil.Bytes `toml:",omitempty"`
		GasPrice                *big.Int
		Cphash                  *cphash.Config
		TxPool                  *core.TxPoolConfig
		GPO                     *gasprice.Config
		EnablePreimageRecording *bool
		DocRoot                 *string `toml:"-"`
		RnetPort                *string
	}
	var dec Config
	if err := unmarshal(&dec); err != nil {
		return err
	}
	if dec.GenesisKey != nil {
		c.GenesisKey = dec.GenesisKey
	}
	if dec.Genesis != nil {
		c.Genesis = dec.Genesis
	}
	if dec.NetworkId != nil {
		c.NetworkId = *dec.NetworkId
	}
	if dec.SyncMode != nil {
		c.SyncMode = *dec.SyncMode
	}
	if dec.NoPruning != nil {
		c.NoPruning = *dec.NoPruning
	}
	if dec.LightServ != nil {
		c.LightServ = *dec.LightServ
	}
	if dec.LightPeers != nil {
		c.LightPeers = *dec.LightPeers
	}
	if dec.SkipBcVersionCheck != nil {
		c.SkipBcVersionCheck = *dec.SkipBcVersionCheck
	}
	if dec.DatabaseHandles != nil {
		c.DatabaseHandles = *dec.DatabaseHandles
	}
	if dec.DatabaseCache != nil {
		c.DatabaseCache = *dec.DatabaseCache
	}
	if dec.TrieCache != nil {
		c.TrieCache = *dec.TrieCache
	}
	if dec.TrieTimeout != nil {
		c.TrieTimeout = *dec.TrieTimeout
	}
	if dec.MinerThreads != nil {
		c.MinerThreads = *dec.MinerThreads
	}
	if dec.ExtraData != nil {
		c.ExtraData = *dec.ExtraData
	}
	if dec.GasPrice != nil {
		c.GasPrice = dec.GasPrice
	}
	if dec.Cphash != nil {
		c.Cphash = *dec.Cphash
	}
	if dec.TxPool != nil {
		c.TxPool = *dec.TxPool
	}
	if dec.GPO != nil {
		c.GPO = *dec.GPO
	}
	if dec.EnablePreimageRecording != nil {
		c.EnablePreimageRecording = *dec.EnablePreimageRecording
	}
	if dec.DocRoot != nil {
		c.DocRoot = *dec.DocRoot
	}
	if dec.RnetPort != nil {
		c.RnetPort = *dec.RnetPort
	}

	return nil
}
