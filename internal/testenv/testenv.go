// Package testenv holds per-network settings for the live test suites.
package testenv

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type NetworkType int

const (
	UnknownNetworkType NetworkType = iota
	Mainnet
	Testnet
)

func (nt NetworkType) String() string {
	switch nt {
	case Mainnet:
		return "mainnet"
	case Testnet:
		return "testnet"
	default:
		return "unknown"
	}
}

type Network int

const (
	UnknownNetwork Network = iota
	Arbitrum
	Arc
	Avalanche
	Base
	BSC
	Celo
	Ethereum
	HyperEVM
	Monad
	Optimism
	Polygon
	Robinhood
	Unichain
)

func (n Network) String() string {
	switch n {
	case Arbitrum:
		return "arbitrum"
	case Arc:
		return "arc"
	case Avalanche:
		return "avalanche"
	case Base:
		return "base"
	case BSC:
		return "bsc"
	case Celo:
		return "celo"
	case Ethereum:
		return "ethereum"
	case HyperEVM:
		return "hyperevm"
	case Monad:
		return "monad"
	case Optimism:
		return "optimism"
	case Polygon:
		return "polygon"
	case Robinhood:
		return "robinhood"
	case Unichain:
		return "unichain"
	default:
		return "unknown"
	}
}

type Config struct {
	RPCURL             string
	ActorPrivateKeyHex string

	Constant  ChainConstants
	Addresses AddressBook
}

type ChainConstants struct {
	ChainID             int
	NativeSymbol        string
	NativeDecimals      int
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
}

type AddressBook struct {
	Standard StandardAddressBook
	Protocol ProtocolAddressBook
}

type StandardAddressBook struct {
	Multicall    string
	NicksFactory string
	WETH         string
	Permit2      string
}

type ProtocolAddressBook struct {
	Uniswap UniswapAddressBook
}

type UniswapAddressBook struct {
	V2              UniswapV2AddressBook
	V3              UniswapV3AddressBook
	V4              UniswapV4AddressBook
	UniversalRouter string
}

type UniswapV2AddressBook struct {
	Factory string
	Router  string
}

type UniswapV3AddressBook struct {
	Factory string
	Router  string
	Quoter  string
	NFPM    string
}

type UniswapV4AddressBook struct {
	PoolManager             string
	PositionManager         string
	StateView               string
	Quoter                  string
	PoolSwapTest            string
	PoolModifyLiquidityTest string
}

// MustLoad returns n's config on nt with the API key and actor key filled in.
func MustLoad(t *testing.T, nt NetworkType, n Network) *Config {
	t.Helper()

	if nt == UnknownNetworkType || n == UnknownNetwork {
		t.Skip("testenv: no network selected")
	}

	ntm, ok := configs[nt]
	if !ok {
		t.Fatalf("testenv: network type %s not supported", nt)
	}

	cfg, ok := ntm[n]
	if !ok {
		t.Fatalf("testenv: network %s %s not supported", n, nt)
	}

	if strings.Contains(cfg.RPCURL, "%s") {
		cfg.RPCURL = fmt.Sprintf(cfg.RPCURL, apiKey)
	}

	cfg.ActorPrivateKeyHex = actorPrivateKeyHex

	return &cfg
}
