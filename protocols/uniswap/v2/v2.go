// Package v2 encodes calls to and decodes results from Uniswap V2 (Factory,
// Pair, Router02): a constant-product (x*y=k) AMM for swapping and providing
// liquidity between two ERC-20 tokens.
//
// Official Factory/Router02 deployment addresses per network, from
// https://docs.uniswap.org/contracts/v2/reference/smart-contracts/v2-deployments:
//
//	Network          Factory                                     V2Router02
//	Mainnet          0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f   0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D
//	Sepolia          0xF62c03E08ada871A0bEb309762E260a7a6a880E6   0xeE567Fe1712Faf6149d80dA1E6934E354124CfE3
//	Unichain         0x1F98400000000000000000000000000000000002   0x284f11109359a7E1306C3e447ef14D38400063FF
//	Arbitrum         0xf1D7CC64Fb4452F05c498126312eBE29f30Fbcf9   0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24
//	Avalanche        0x9e5A52f57b3038F1B8EeE45F28b3C1967e22799C   0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24
//	BNB Chain        0x8909Dc15e40173Ff4699343b6eB8132c65e18eC6   0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24
//	Base             0x8909Dc15e40173Ff4699343b6eB8132c65e18eC6   0x4752ba5dbc23f44d87826276bf6fd6b1c372ad24
//	Optimism         0x0c3c1c532F1e39EdF36BE9Fe0bE1410313E074Bf   0x4A7b5Da61326A6379179b40d00F57E5bbDC962c2
//	Polygon          0x9e5A52f57b3038F1B8EeE45F28b3C1967e22799C   0xedf6066a2b290C185783862C7F4776A2C8077AD1
//	WorldChain       0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f   0x541aB7c31A119441eF3575F6973277DE0eF460bd
//	Monad            0x182a927119d56008d921126764bf884221b10f59   0x4b2ab38dbf28d31d467aa8993f6c2585981d6804
//	Robinhood Chain  0x8bceaa40b9acdfaedf85adf4ff01f5ad6517937f   0x89e5db8b5aa49aa85ac63f691524311aeb649eba
//	HyperEVM         0x89e5DB8B5aA49aA85AC63f691524311AEB649eba   0x1f7d7550B1b028f7571E69A784071F0205FD2EfA
package v2
