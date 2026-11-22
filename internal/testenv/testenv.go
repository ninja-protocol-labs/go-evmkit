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
	Ink
	Linea
	Monad
	Optimism
	Polygon
	Robinhood
	Sei
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
	case Ink:
		return "ink"
	case Linea:
		return "linea"
	case Monad:
		return "monad"
	case Optimism:
		return "optimism"
	case Polygon:
		return "polygon"
	case Robinhood:
		return "robinhood"
	case Sei:
		return "sei"
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
	V2 UniswapV2Config
}

type UniswapV2Config struct {
	Factory string
	Router  string
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
