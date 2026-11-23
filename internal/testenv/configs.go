package testenv

import "time"

const (
	apiKey             string = ""
	actorPrivateKeyHex string = ""
)

// configs holds public per-network values; a %s in RPCURL is replaced with apiKey.
var configs = map[NetworkType]map[Network]Config{
	Mainnet: {
		Arbitrum: {
			RPCURL: "https://arb-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             42161,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x82aF49447D8a07e3bd95BD0d56f35241523fBab1",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0xf1D7CC64Fb4452F05c498126312eBE29f30Fbcf9",
							Router:  "0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x1F98431c8aD98523631AE4a59f267346ea31F984",
							Router:  "0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45",
							Quoter:  "0x61fFE014bA17989E743c5F6cB21bF9697530B21e",
							NFPM:    "0xC36442b4a4522E871399CD717aBDD847Ab11FE88",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x360E68faCcca8cA495c1B759Fd9EEe466db9FB32",
							PositionManager:         "0xd88F38F930b7952f2DB2432Cb002E7abbF3dD869",
							StateView:               "0x76Fd297e2D437cd7f76d50F01AfE6160f86e9990",
							Quoter:                  "0x8928074CA1b241D8Ec02815881c1Af11E8bC5219",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0x2d01411773c8C24805306E89A41F7855C3c4Fe65",
					},
				},
			},
		},
		Arc: {
			RPCURL: "https://arc-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             5042,
				NativeSymbol:        "USDC",
				NativeDecimals:      6,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "", // Not supported
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x89e5DB8B5aA49aA85AC63f691524311AEB649eba",
							Router:  "0x1f7d7550B1b028f7571E69A784071F0205FD2EfA",
						},
						V3: UniswapV3AddressBook{
							Factory: "0xf0db7b58379503491d857dB50AC9ece64c653918",
							Router:  "0x53BF6B0684Ec7eF91e1387Da3D1a1769bC5A6F77",
							Quoter:  "0x7DfD4F31be6814D2906BDE155c3e1B146EAc1468",
							NFPM:    "0x39654A85A4C05127f5Fd6ED22CAeC077A0fB1377",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x8366a39CC670B4001A1121B8F6A443A643e40951",
							PositionManager:         "0x6049c9a0e26405C0985f9E3685C87d0aE917f82B",
							StateView:               "0xF3334192D15450CdD385c8B70e03f9A6bD9E673b",
							Quoter:                  "0x8Dc178eFB8111BB0973Dd9d722ebeFF267c98F94",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0x8702463e73f74d0b6765aBceb314Ef07aCb92650",
					},
				},
			},
		},
		Avalanche: {
			RPCURL: "https://avax-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             43114,
				NativeSymbol:        "AVAX",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x9e5A52f57b3038F1B8EeE45F28b3C1967e22799C",
							Router:  "0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x740b1c1de25031C31FF4fC9A62f554A55cdC1baD",
							Router:  "0xbb00FF08d01D300023C629E8fFfFcb65A5a578cE",
							Quoter:  "0xbe0F5544EC67e9B3b2D979aaA43f18Fd87E6257F",
							NFPM:    "0x655C406EBFa14EE2006250925e54ec43AD184f8B",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x06380C0e0912312B5150364B9DC4542BA0DbBc85",
							PositionManager:         "0xB74b1F14d2754AcfcbBe1a221023a5cf50Ab8ACD",
							StateView:               "0xc3c9e198C735a4b97e3e683f391cCBDD60B69286",
							Quoter:                  "0xf27e532878B622d6b916F524E8711880107D4eF0",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0x661E93cca42AfacB172121EF892830cA3b70F08d",
					},
				},
			},
		},
		Base: {
			RPCURL: "https://base-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             8453,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x4200000000000000000000000000000000000006",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x8909Dc15e40173Ff4699343b6eB8132c65e18eC6",
							Router:  "0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x33128a8fC17869897dcE68Ed026d694621f6FDfD",
							Router:  "0x2626664c2603336E57B271c5C0b26F421741e481",
							Quoter:  "0x3d4e44Eb1374240CE5F1B871ab261CD16335B76a",
							NFPM:    "0x03a520b32C04BF3bEEf7BEb72E919cf822Ed34f1",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x498581fF718922c3f8e6A244956aF099B2652b2b",
							PositionManager:         "0x7C5f5A4bBd8fD63184577525326123B519429bDc",
							StateView:               "0xA3c0c9b65baD0b08107Aa264b0f3dB444b867A71",
							Quoter:                  "0xE1A6C5803e3c44368b171BD680c7D32dfc92CA47",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0xd6145b2D3F379919E8CdEda7B97e37c4b2Ca9c40",
					},
				},
			},
		},
		BSC: {
			RPCURL: "https://bnb-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             56,
				NativeSymbol:        "BNB",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x8909Dc15e40173Ff4699343b6eB8132c65e18eC6",
							Router:  "0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24",
						},
						V3: UniswapV3AddressBook{
							Factory: "0xdB1d10011AD0Ff90774D0C6Bb92e5C5c8b4461F7",
							Router:  "0xB971eF87ede563556b2ED4b1C0b0019111Dd85d2",
							Quoter:  "0x78D78E420Da98ad378D7799bE8f4AF69033EB077",
							NFPM:    "0x7b8A01B39D58278b5DE7e48c8449c9f4F5170613",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x28e2Ea090877bF75740558f6BFB36A5ffeE9e9dF",
							PositionManager:         "0x7A4a5c919aE2541AeD11041A1AEeE68f1287f95b",
							StateView:               "0xd13Dd3D6E93f276FAfc9Db9E6BB47C1180aeE0c4",
							Quoter:                  "0x95C086d7fE28AdD842ee220fef225a777474d781",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0xDc264714F68d84CF29BC605589405E78bDBE7C9f",
					},
				},
			},
		},
		Celo: {
			RPCURL: "https://celo-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             42220,
				NativeSymbol:        "CELO",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "", // Not supported
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x79a530c8e2fA8748B7B40dd3629C0520c2cCf03f",
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "0xAfE208a311B21f13EF87E33A90049fC17A7acDEc",
							Router:  "0x5615CDAb10dc425a742d643d949a7F474C01abc4",
							Quoter:  "0x82825d0554fA07f7FC52Ab63c961F330fdEFa8E8",
							NFPM:    "0x3d79EdAaBC0EaB6F08ED885C05Fc0B014290D95A",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x288dc841A52FCA2707c6947B3A777c5E56cd87BC",
							PositionManager:         "0xf7965f3981e4D5BC383BfBCb61501763e9068CA9",
							StateView:               "0xbc21f8720BABf4b20d195eE5C6e99c52b76F2bfb",
							Quoter:                  "0x28566da1093609182dFf2cB2A91CFD72e61d66cd",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0xe2023F3FA515cF070e07fD9d51c1d236e07843f4",
					},
				},
			},
		},
		Ethereum: {
			RPCURL: "https://eth-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             1,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: time.Second,
				ReceiptTimeout:      30 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f",
							Router:  "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x1F98431c8aD98523631AE4a59f267346ea31F984",
							Router:  "0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45",
							Quoter:  "0x61fFE014bA17989E743c5F6cB21bF9697530B21e",
							NFPM:    "0xC36442b4a4522E871399CD717aBDD847Ab11FE88",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x000000000004444c5dc75cB358380D2e3dE08A90",
							PositionManager:         "0xbD216513d74C8cf14cf4747E6AaA6420FF64ee9e",
							StateView:               "0x7fFE42C4a5DEeA5b0feC41C94C136Cf115597227",
							Quoter:                  "0x99bd42B2B22e531f283b7F8a86dFf631EB16A9D1",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0x23617e59A5925b2A4Bf75d73ff6711cD0b29De85",
					},
				},
			},
		},
		HyperEVM: {
			RPCURL: "https://hyperliquid-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             999,
				NativeSymbol:        "HYPE",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x5555555555555555555555555555555555555555",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x89e5DB8B5aA49aA85AC63f691524311AEB649eba",
							Router:  "0x1f7d7550B1b028f7571E69A784071F0205FD2EfA",
						},
						V3: UniswapV3AddressBook{
							Factory: "0xf0db7b58379503491d857dB50AC9ece64c653918",
							Router:  "0x7AdF4701AbCDBc5Dcf5Cb58B526f897e048F0D11",
							Quoter:  "0x7DfD4F31be6814D2906BDE155c3e1B146EAc1468",
							NFPM:    "0x39654A85A4C05127f5Fd6ED22CAeC077A0fB1377",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x12D4Fd9C5DeDd00ab8a0bCe2CF0167bbf94b6B1F",
							PositionManager:         "0x0d7Ab5B3db668128Aff6F70C4eBC71D7d4DA9bf9",
							StateView:               "0x1656326235cb9E34Cb58CaDe53AE30789AB32A1A",
							Quoter:                  "0x108Bfe38532C98f8Cd8D7Ab49c9dDAa3675a1c5e",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0xeD270A1Bcdc63cf1356D695e7f40961d4819d6bF",
					},
				},
			},
		},
		Ink: {
			RPCURL: "https://ink-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             57073,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x4200000000000000000000000000000000000006",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0xfe57A6BA1951F69aE2Ed4abe23e0f095DF500C04",
							Router:  "0xB3FB126ACDd5AdCA2f50Ac644a7a2303745f18b4",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x640887A9ba3A9C53Ed27D0F7e8246A4F933f3424",
							Router:  "0x177778F19E89dD1012BdBe603F144088A95C4B53",
							Quoter:  "0x96b572D2d880cf2Fa2563651BD23ADE6f5516652",
							NFPM:    "0xC0836E5B058BBE22ae2266e1AC488A1A0fD8DCE8",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x360E68faCcca8cA495c1B759Fd9EEe466db9FB32",
							PositionManager:         "0x1b35d13a2E2528f192637F14B05f0Dc0e7dEB566",
							StateView:               "0x76Fd297e2D437cd7f76d50F01AfE6160f86e9990",
							Quoter:                  "0x8928074CA1b241D8Ec02815881c1Af11E8bC5219",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0x661E93cca42AfacB172121EF892830cA3b70F08d",
					},
				},
			},
		},
		Linea: {
			RPCURL: "https://linea-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             59144,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0xe5D7C2a44FfDDf6b295A15c148167daaAf5Cf34f",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x114A43DF6C5f54EBB8A9d70Cd1951D3dD68004c7",
							Router:  "0x8702463e73f74d0b6765aBceb314Ef07aCb92650",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x31FAfd4889FA1269F7a13A66eE0fB458f27D72A9",
							Router:  "0x3d4e44Eb1374240CE5F1B871ab261CD16335B76a",
							Quoter:  "0x42bE4D6527829FeFA1493e1fb9F3676d2425C3C1",
							NFPM:    "0x4615C383F85D0a2BbED973d83ccecf5CB7121463",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x248083Fb965359d82b06C1F5322480Dcfc1AD857",
							PositionManager:         "0xdDCAD5775B2816a87495f207731b3571D7EE3c76",
							StateView:               "0xE861de206E460A8b936b05ad3816520B58ccDf9b",
							Quoter:                  "0x2C125569C0BeE20A66E33E5491C552B37EBD9934",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0xDc264714F68d84CF29BC605589405E78bDBE7C9f",
					},
				},
			},
		},
		Monad: {
			RPCURL: "https://monad-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             143,
				NativeSymbol:        "MON",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x3bd359C1119dA7Da1D913D1C4D2B7c461115433A",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x182a927119D56008d921126764bF884221b10f59",
							Router:  "0x4B2ab38DBF28D31D467aA8993f6c2585981D6804",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x204FAca1764B154221e35c0d20aBb3c525710498",
							Router:  "0xfE31F71C1b106EAc32F1A19239c9a9A72ddfb900",
							Quoter:  "0x661E93cca42AfacB172121EF892830cA3b70F08d",
							NFPM:    "0x7197E214c0b767cFB76Fb734ab638E2c192F4E53",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x188d586Ddcf52439676Ca21A244753fA19F9Ea8e",
							PositionManager:         "0x5b7eC4a94fF9beDb700fb82aB09d5846972F4016",
							StateView:               "0x77395F3b2E73aE90843717371294fa97cC419D64",
							Quoter:                  "0xa222Dd357A9076d1091Ed6Aa2e16C9742dD26891",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0xa6CE4F10d83dBdDAc17E68e1837ca9cE6a1b596e",
					},
				},
			},
		},
		Optimism: {
			RPCURL: "https://opt-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             10,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x4200000000000000000000000000000000000006",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x0c3c1c532F1e39EdF36BE9Fe0bE1410313E074Bf",
							Router:  "0x4A7b5Da61326A6379179b40d00F57E5bbDC962c2",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x1F98431c8aD98523631AE4a59f267346ea31F984",
							Router:  "0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45",
							Quoter:  "0x61fFE014bA17989E743c5F6cB21bF9697530B21e",
							NFPM:    "0xC36442b4a4522E871399CD717aBDD847Ab11FE88",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x9a13F98Cb987694C9F086b1F5eB990EeA8264Ec3",
							PositionManager:         "0x3C3Ea4B57a46241e54610e5f022E5c45859A1017",
							StateView:               "0xc18a3169788F4F75A170290584ECA6395C75Ecdb",
							Quoter:                  "0xd673F731EDFe5AaF1bDa7FD43D13Bbb5150B87Ca",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0xC09255D86DB563cBc11C2fCf4a0C512e160111B4",
					},
				},
			},
		},
		Polygon: {
			RPCURL: "https://polygon-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             137,
				NativeSymbol:        "POL",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x9e5A52f57b3038F1B8EeE45F28b3C1967e22799C",
							Router:  "0xedf6066a2b290C185783862C7F4776A2C8077AD1",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x1F98431c8aD98523631AE4a59f267346ea31F984",
							Router:  "0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45",
							Quoter:  "0x61fFE014bA17989E743c5F6cB21bF9697530B21e",
							NFPM:    "0xC36442b4a4522E871399CD717aBDD847Ab11FE88",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x67366782805870060151383F4BbFF9daB53e5cD6",
							PositionManager:         "0x1Ec2eBf4F37E7363FDfe3551602425af0B3ceef9",
							StateView:               "0x5eA1bD7974c8A611cBAB0bDCAFcB1D9CC9b3BA5a",
							Quoter:                  "0x31AB0Dea31FC0ad5A54F03cD59C896eCD067447B",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0xDc264714F68d84CF29BC605589405E78bDBE7C9f",
					},
				},
			},
		},
		Robinhood: {
			RPCURL: "https://robinhood-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             4663,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x0Bd7D308f8E1639FAb988df18A8011f41EAcAD73",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x8bcEaA40B9AcdfAedF85AdF4FF01F5Ad6517937f",
							Router:  "0x89e5DB8B5aA49aA85AC63f691524311AEB649eba",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x1f7d7550B1b028f7571E69A784071F0205FD2EfA",
							Router:  "0xCaf681a66D020601342297493863E78C959E5cb2",
							Quoter:  "0x33e885eD0Ec9bF04EcfB19341582aADCb4c8A9E7",
							NFPM:    "0x73991a25C818Bf1f1128dEAaB1492D45638DE0D3",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x8366a39CC670B4001A1121B8F6A443A643e40951",
							PositionManager:         "0x58daec3116aae6D93017bAAea7749052E8a04fA7",
							StateView:               "0xF3334192D15450CdD385c8B70e03f9A6bD9E673b",
							Quoter:                  "0x8Dc178eFB8111BB0973Dd9d722ebeFF267c98F94",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0x204FAca1764B154221e35c0d20aBb3c525710498",
					},
				},
			},
		},
		Unichain: {
			RPCURL: "https://unichain-mainnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             130,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x4200000000000000000000000000000000000006",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x1F98400000000000000000000000000000000002",
							Router:  "0x284F11109359a7e1306C3e447ef14D38400063FF",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x1F98400000000000000000000000000000000003",
							Router:  "0x73855d06DE49d0fe4A9c42636Ba96c62da12FF9C",
							Quoter:  "0x385A5cf5F83e99f7BB2852b6A19C3538b9FA7658",
							NFPM:    "0x943e6e07a7E8E791dAFC44083e54041D743C46E9",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x1F98400000000000000000000000000000000004",
							PositionManager:         "0x4529A01c7A0410167c5740C487A8DE60232617bf",
							StateView:               "0x86e8631A016F9068C3f085fAF484Ee3F5fDee8f2",
							Quoter:                  "0xF885781cB68A0BE66AA2AC7eDa5Fe72Bde7dA907",
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "0xD1b797D92d87B688193A2B976eFc8D577D204343",
					},
				},
			},
		},
	},
	Testnet: {
		Arbitrum: {
			RPCURL: "https://arb-sepolia.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             421614,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x980B62Da83eFf3D4576C647993b0c1D7faf17c73",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "0x248AB79Bbb9bC29bB72f7Cd42F17e054Fc40188e",
							Router:  "0x101F443B4d1b059569D643917553c771E1b9663E",
							Quoter:  "0x2779a0CC1c3e0E44D2542EC3e79e3864Ae93Ef0B",
							NFPM:    "0x6b2937Bde17889EDCf8fbD8dE31C3C2a70Bc4d65",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0xFB3e0C6F74eB1a21CC1Da29aeC80D2Dfe6C9a317",
							PositionManager:         "0xAc631556d3d4019C95769033B5E719dD77124BAc",
							StateView:               "0x9D467FA9062b6e9B1a46E26007aD82db116c67cB",
							Quoter:                  "0x421A00F1d80D41a8904Ba6079A92A848696f331E",
							PoolSwapTest:            "0xf3A39C86dbd13C45365E57FB90fe413371F65AF8",
							PoolModifyLiquidityTest: "0x9A8ca723F5dcCb7926D00B71deC55c2fEa1F50f7",
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Arc: {
			RPCURL: "https://arc-testnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             5042002,
				NativeSymbol:        "USDC",
				NativeDecimals:      6,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "", // Not supported
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
							Quoter:  "", // Not supported
							NFPM:    "", // Not supported
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Avalanche: {
			RPCURL: "https://avax-fuji.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             43113,
				NativeSymbol:        "AVAX",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0xd00ae08403B9bbb9124bB305C09058E32C39A48c",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
							Quoter:  "", // Not supported
							NFPM:    "", // Not supported
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Base: {
			RPCURL: "https://base-sepolia.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             84532,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x4200000000000000000000000000000000000006",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24",
							Router:  "0x94cC0AaC535CCDB3C01d6787D6413C739ae12bc4",
							Quoter:  "0xC5290058841028F1614F3A6F0F5816cAd0df5E27",
							NFPM:    "0x27F971cb582BF9E50F397e4d29a5C7A34f11faA2",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x05E73354cFDd6745C338b50BcFDfA3Aa6fA03408",
							PositionManager:         "0x4B2C77d209D3405F41a037Ec6c77F7F5b8e2ca80",
							StateView:               "0x571291b572ed32ce6751a2Cb2486EbEe8DEfB9B4",
							Quoter:                  "0x5e6B8B8AFA7e5CEF6C09dd2A67aBA9FB12C75623",
							PoolSwapTest:            "0x8B5bcC363ddE2614281aD875bad385E0A785D3B9",
							PoolModifyLiquidityTest: "0x37429cD17Cb1454C34E7F50b09725202Fd533039",
						},
						UniversalRouter: "0x8702463e73f74d0b6765aBceb314Ef07aCb92650",
					},
				},
			},
		},
		BSC: {
			RPCURL: "https://bnb-testnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             97,
				NativeSymbol:        "BNB",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0xae13d989daC2f0dEbFf460aC112a837C89BAa7cd",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
							Quoter:  "", // Not supported
							NFPM:    "", // Not supported
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Celo: {
			RPCURL: "https://celo-sepolia.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             11142220,
				NativeSymbol:        "CELO",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "", // Not supported
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
							Quoter:  "", // Not supported
							NFPM:    "", // Not supported
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Ethereum: {
			RPCURL: "https://eth-sepolia.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             11155111,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: time.Second,
				ReceiptTimeout:      30 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0xfff9976782d46cc05630d1f6ebab18b2324d6b14",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0xF62c03E08ada871A0bEb309762E260a7a6a880E6",
							Router:  "0xeE567Fe1712Faf6149d80dA1E6934E354124CfE3",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x0227628f3F023bb0B980b67D528571c95c6DaC1c",
							Router:  "0x3bFA4769FB09eefC5a80d6E87c3B9C650f7Ae48E",
							Quoter:  "0xEd1f6473345F45b75F8179591dd5bA1888cf2FB3",
							NFPM:    "0x1238536071E1c677A632429e3655c799b22cDA52",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0xE03A1074c86CFeDd5C142C4F04F1a1536e203543",
							PositionManager:         "0x429ba70129df741B2Ca2a85BC3A2a3328e5c09b4",
							StateView:               "0xE1Dd9c3fA50EDB962E442f60DfBc432e24537E4C",
							Quoter:                  "0x68aFa8b853cAa7a2D40B0F6AEE5C4DbC147C8244",
							PoolSwapTest:            "0x9B6b46e2c869aa39918Db7f52f5557FE577B6eEe",
							PoolModifyLiquidityTest: "0x0C478023803a644c94c4CE1C1e7b9A087e411B0A",
						},
						UniversalRouter: "0x7E4f6c5e954Da5c61B3423D81E2277431Ac043f3",
					},
				},
			},
		},
		HyperEVM: {
			RPCURL: "https://hyperliquid-testnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             998,
				NativeSymbol:        "HYPE",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x5555555555555555555555555555555555555555",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
							Quoter:  "", // Not supported
							NFPM:    "", // Not supported
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Ink: {
			RPCURL: "https://ink-sepolia.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             763373,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x4200000000000000000000000000000000000006",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
							Quoter:  "", // Not supported
							NFPM:    "", // Not supported
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Linea: {
			RPCURL: "https://linea-sepolia.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             59141,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x06565ed324Ee9fb4DB0FF80B7eDbE4Cb007555a3",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
							Quoter:  "", // Not supported
							NFPM:    "", // Not supported
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Monad: {
			RPCURL: "https://monad-testnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             10143,
				NativeSymbol:        "MON",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0xFb8bf4c1CC7a94c73D209a149eA2AbEa852BC541",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
							Quoter:  "", // Not supported
							NFPM:    "", // Not supported
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Optimism: {
			RPCURL: "https://opt-sepolia.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             11155420,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x4200000000000000000000000000000000000006",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "0x8CE191193D15ea94e11d327b4c7ad8bbE520f6aF",
							Router:  "", // Not supported
							Quoter:  "0x0FBEa6cf957d95ee9313490050F6A0DA68039404",
							NFPM:    "0xdA75cEf1C93078e8b736FCA5D5a30adb97C8957d",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Polygon: {
			RPCURL: "https://polygon-amoy.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             80002,
				NativeSymbol:        "POL",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x360ad4f9a9A8EFe9A8DCB5f461c4Cc1047E1Dcf9",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
							Quoter:  "", // Not supported
							NFPM:    "", // Not supported
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Robinhood: {
			RPCURL: "https://robinhood-testnet.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             46630,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x7943e237c7F95DA44E0301572D358911207852Fa",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
						},
						V3: UniswapV3AddressBook{
							Factory: "", // Not supported
							Router:  "", // Not supported
							Quoter:  "", // Not supported
							NFPM:    "", // Not supported
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "", // Not supported
							PositionManager:         "", // Not supported
							StateView:               "", // Not supported
							Quoter:                  "", // Not supported
							PoolSwapTest:            "", // Not supported
							PoolModifyLiquidityTest: "", // Not supported
						},
						UniversalRouter: "", // Not supported
					},
				},
			},
		},
		Unichain: {
			RPCURL: "https://unichain-sepolia.g.alchemy.com/v2/%s",
			Constant: ChainConstants{
				ChainID:             1301,
				NativeSymbol:        "ETH",
				NativeDecimals:      18,
				ReceiptPollInterval: 500 * time.Millisecond,
				ReceiptTimeout:      10 * time.Second,
			},
			Addresses: AddressBook{
				Standard: StandardAddressBook{
					Multicall:    "0xcA11bde05977b3631167028862bE2a173976CA11",
					NicksFactory: "0x4e59b44847b379578588920cA78FbF26c0B4956C",
					WETH:         "0x4200000000000000000000000000000000000006",
					Permit2:      "0x000000000022D473030F116dDEE9F6B43aC78BA3",
				},
				Protocol: ProtocolAddressBook{
					Uniswap: UniswapAddressBook{
						V2: UniswapV2AddressBook{
							Factory: "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f",
							Router:  "0x6e8C9D62A16419357eA998229126D2fD6B1CCfbF",
						},
						V3: UniswapV3AddressBook{
							Factory: "0x1F98431c8aD98523631AE4a59f267346ea31F984",
							Router:  "0xd1AAE39293221B77B0C71fBD6dCb7Ea29Bb5B166",
							Quoter:  "0x6Dd37329A1A225a6Fca658265D460423DCafBF89",
							NFPM:    "0xB7F724d6dDDFd008eFf5cc2834edDE5F9eF0d075",
						},
						V4: UniswapV4AddressBook{
							PoolManager:             "0x00B036B58a818B1BC34d502D3fE730Db729e62AC",
							PositionManager:         "0xf969Aee60879C54bAAed9F3eD26147Db216Fd664",
							StateView:               "0xc199F1072a74D4e905ABa1A84d9a45E2546B6222",
							Quoter:                  "0x572B9cDBcd8c2b8F3Fe849B9d250F35284455d7f",
							PoolSwapTest:            "0x9140a78c1A137c7fF1c151EC8231272aF78a99A4",
							PoolModifyLiquidityTest: "0x5fa728C0A5cfd51BEe4B060773f50554c0C8A7AB",
						},
						UniversalRouter: "0xDf38F24fE153761634Be942F9d859f3DBA857E95",
					},
				},
			},
		},
	},
}

const (
	// ERC20ByteCodeHex is OpenZeppelin's ERC20 plus mint/burn, constructor(name, symbol, decimals).
	ERC20ByteCodeHex = "0x60a060405234801561000f575f5ffd5b506040516114e73803806114e7833981810160405281019061003191906101ee565b828281600390816100429190610497565b5080600490816100529190610497565b5050508060ff1660808160ff1681525050505050610566565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6100ca82610084565b810181811067ffffffffffffffff821117156100e9576100e8610094565b5b80604052505050565b5f6100fb61006b565b905061010782826100c1565b919050565b5f67ffffffffffffffff82111561012657610125610094565b5b61012f82610084565b9050602081019050919050565b8281835e5f83830152505050565b5f61015c6101578461010c565b6100f2565b90508281526020810184848401111561017857610177610080565b5b61018384828561013c565b509392505050565b5f82601f83011261019f5761019e61007c565b5b81516101af84826020860161014a565b91505092915050565b5f60ff82169050919050565b6101cd816101b8565b81146101d7575f5ffd5b50565b5f815190506101e8816101c4565b92915050565b5f5f5f6060848603121561020557610204610074565b5b5f84015167ffffffffffffffff81111561022257610221610078565b5b61022e8682870161018b565b935050602084015167ffffffffffffffff81111561024f5761024e610078565b5b61025b8682870161018b565b925050604061026c868287016101da565b9150509250925092565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f60028204905060018216806102c457607f821691505b6020821081036102d7576102d6610280565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026103397fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826102fe565b61034386836102fe565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61038761038261037d8461035b565b610364565b61035b565b9050919050565b5f819050919050565b6103a08361036d565b6103b46103ac8261038e565b84845461030a565b825550505050565b5f5f905090565b6103cb6103bc565b6103d6818484610397565b505050565b5f5b828110156103fc576103f15f8284016103c3565b6001810190506103dd565b505050565b601f82111561044f578282111561044e5761041b816102dd565b610424836102ef565b61042d856102ef565b602086101561043a575f90505b808301610449828403826103db565b505050505b5b505050565b5f82821c905092915050565b5f61046f5f1984600802610454565b1980831691505092915050565b5f6104878383610460565b9150826002028217905092915050565b6104a082610276565b67ffffffffffffffff8111156104b9576104b8610094565b5b6104c382546102ad565b6104ce828285610401565b5f60209050601f8311600181146104ff575f84156104ed578287015190505b6104f7858261047c565b86555061055e565b601f19841661050d866102dd565b5f5b828110156105345784890151825560018201915060208501945060208101905061050f565b86831015610551578489015161054d601f891682610460565b8355505b6001600288020188555050505b505050505050565b608051610f6961057e5f395f6103370152610f695ff3fe608060405234801561000f575f5ffd5b50600436106100a7575f3560e01c806340c10f191161006f57806340c10f191461016557806370a082311461018157806395d89b41146101b15780639dc29fac146101cf578063a9059cbb146101eb578063dd62ed3e1461021b576100a7565b806306fdde03146100ab578063095ea7b3146100c957806318160ddd146100f957806323b872dd14610117578063313ce56714610147575b5f5ffd5b6100b361024b565b6040516100c09190610be2565b60405180910390f35b6100e360048036038101906100de9190610c93565b6102db565b6040516100f09190610ceb565b60405180910390f35b6101016102fd565b60405161010e9190610d13565b60405180910390f35b610131600480360381019061012c9190610d2c565b610306565b60405161013e9190610ceb565b60405180910390f35b61014f610334565b60405161015c9190610d97565b60405180910390f35b61017f600480360381019061017a9190610c93565b61035b565b005b61019b60048036038101906101969190610db0565b610369565b6040516101a89190610d13565b60405180910390f35b6101b96103ae565b6040516101c69190610be2565b60405180910390f35b6101e960048036038101906101e49190610c93565b61043e565b005b61020560048036038101906102009190610c93565b61044c565b6040516102129190610ceb565b60405180910390f35b61023560048036038101906102309190610ddb565b61046e565b6040516102429190610d13565b60405180910390f35b60606003805461025a90610e46565b80601f016020809104026020016040519081016040528092919081815260200182805461028690610e46565b80156102d15780601f106102a8576101008083540402835291602001916102d1565b820191905f5260205f20905b8154815290600101906020018083116102b457829003601f168201915b5050505050905090565b5f5f6102e56104f0565b90506102f28185856104f7565b600191505092915050565b5f600254905090565b5f5f6103106104f0565b905061031d858285610509565b61032885858561059c565b60019150509392505050565b5f7f0000000000000000000000000000000000000000000000000000000000000000905090565b610365828261068c565b5050565b5f5f5f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b6060600480546103bd90610e46565b80601f01602080910402602001604051908101604052809291908181526020018280546103e990610e46565b80156104345780601f1061040b57610100808354040283529160200191610434565b820191905f5260205f20905b81548152906001019060200180831161041757829003601f168201915b5050505050905090565b610448828261070b565b5050565b5f5f6104566104f0565b905061046381858561059c565b600191505092915050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f33905090565b610504838383600161078a565b505050565b5f610514848461046e565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8110156105965781811015610587578281836040517ffb8f41b200000000000000000000000000000000000000000000000000000000815260040161057e93929190610e85565b60405180910390fd5b61059584848484035f61078a565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361060c575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016106039190610eba565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361067c575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016106739190610eba565b60405180910390fd5b610687838383610959565b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16036106fc575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016106f39190610eba565b60405180910390fd5b6107075f8383610959565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361077b575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016107729190610eba565b60405180910390fd5b610786825f83610959565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff16036107fa575f6040517fe602df050000000000000000000000000000000000000000000000000000000081526004016107f19190610eba565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361086a575f6040517f94280d620000000000000000000000000000000000000000000000000000000081526004016108619190610eba565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508015610953578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9258460405161094a9190610d13565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036109a9578060025f82825461099d9190610f00565b92505081905550610a77565b5f5f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905081811015610a32578381836040517fe450d38c000000000000000000000000000000000000000000000000000000008152600401610a2993929190610e85565b60405180910390fd5b8181035f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610abe578060025f8282540392505081905550610b08565b805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051610b659190610d13565b60405180910390a3505050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f610bb482610b72565b610bbe8185610b7c565b9350610bce818560208601610b8c565b610bd781610b9a565b840191505092915050565b5f6020820190508181035f830152610bfa8184610baa565b905092915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610c2f82610c06565b9050919050565b610c3f81610c25565b8114610c49575f5ffd5b50565b5f81359050610c5a81610c36565b92915050565b5f819050919050565b610c7281610c60565b8114610c7c575f5ffd5b50565b5f81359050610c8d81610c69565b92915050565b5f5f60408385031215610ca957610ca8610c02565b5b5f610cb685828601610c4c565b9250506020610cc785828601610c7f565b9150509250929050565b5f8115159050919050565b610ce581610cd1565b82525050565b5f602082019050610cfe5f830184610cdc565b92915050565b610d0d81610c60565b82525050565b5f602082019050610d265f830184610d04565b92915050565b5f5f5f60608486031215610d4357610d42610c02565b5b5f610d5086828701610c4c565b9350506020610d6186828701610c4c565b9250506040610d7286828701610c7f565b9150509250925092565b5f60ff82169050919050565b610d9181610d7c565b82525050565b5f602082019050610daa5f830184610d88565b92915050565b5f60208284031215610dc557610dc4610c02565b5b5f610dd284828501610c4c565b91505092915050565b5f5f60408385031215610df157610df0610c02565b5b5f610dfe85828601610c4c565b9250506020610e0f85828601610c4c565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680610e5d57607f821691505b602082108103610e7057610e6f610e19565b5b50919050565b610e7f81610c25565b82525050565b5f606082019050610e985f830186610e76565b610ea56020830185610d04565b610eb26040830184610d04565b949350505050565b5f602082019050610ecd5f830184610e76565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f610f0a82610c60565b9150610f1583610c60565b9250828201905080821115610f2d57610f2c610ed3565b5b9291505056fea264697066735822122049c38efc930a01957bf084f9c67bdda53452b282675854f0c07be3a52f60fb5e64736f6c63430008230033"

	// ERC721ByteCodeHex is OpenZeppelin's ERC721 plus mint/burn, constructor(name, symbol).
	ERC721ByteCodeHex = "0x608060405234801561000f575f5ffd5b50604051612452380380612452833981810160405281019061003191906101a8565b8181815f9081610041919061043f565b508060019081610051919061043f565b505050505061050e565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6100ba82610074565b810181811067ffffffffffffffff821117156100d9576100d8610084565b5b80604052505050565b5f6100eb61005b565b90506100f782826100b1565b919050565b5f67ffffffffffffffff82111561011657610115610084565b5b61011f82610074565b9050602081019050919050565b8281835e5f83830152505050565b5f61014c610147846100fc565b6100e2565b90508281526020810184848401111561016857610167610070565b5b61017384828561012c565b509392505050565b5f82601f83011261018f5761018e61006c565b5b815161019f84826020860161013a565b91505092915050565b5f5f604083850312156101be576101bd610064565b5b5f83015167ffffffffffffffff8111156101db576101da610068565b5b6101e78582860161017b565b925050602083015167ffffffffffffffff81111561020857610207610068565b5b6102148582860161017b565b9150509250929050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061026c57607f821691505b60208210810361027f5761027e610228565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026102e17fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826102a6565b6102eb86836102a6565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61032f61032a61032584610303565b61030c565b610303565b9050919050565b5f819050919050565b61034883610315565b61035c61035482610336565b8484546102b2565b825550505050565b5f5f905090565b610373610364565b61037e81848461033f565b505050565b5f5b828110156103a4576103995f82840161036b565b600181019050610385565b505050565b601f8211156103f757828211156103f6576103c381610285565b6103cc83610297565b6103d585610297565b60208610156103e2575f90505b8083016103f182840382610383565b505050505b5b505050565b5f82821c905092915050565b5f6104175f19846008026103fc565b1980831691505092915050565b5f61042f8383610408565b9150826002028217905092915050565b6104488261021e565b67ffffffffffffffff81111561046157610460610084565b5b61046b8254610255565b6104768282856103a9565b5f60209050601f8311600181146104a7575f8415610495578287015190505b61049f8582610424565b865550610506565b601f1984166104b586610285565b5f5b828110156104dc578489015182556001820191506020850194506020810190506104b7565b868310156104f957848901516104f5601f891682610408565b8355505b6001600288020188555050505b505050505050565b611f378061051b5f395ff3fe608060405234801561000f575f5ffd5b5060043610610109575f3560e01c80636352211e116100a057806395d89b411161006f57806395d89b41146102c5578063a22cb465146102e3578063b88d4fde146102ff578063c87b56dd1461031b578063e985e9c51461034b57610109565b80636352211e146102175780636a6278421461024757806370a082311461027757806375794a3c146102a757610109565b806323b872dd116100dc57806323b872dd146101a757806340c10f19146101c357806342842e0e146101df57806342966c68146101fb57610109565b806301ffc9a71461010d57806306fdde031461013d578063081812fc1461015b578063095ea7b31461018b575b5f5ffd5b61012760048036038101906101229190611717565b61037b565b604051610134919061175c565b60405180910390f35b61014561045c565b60405161015291906117e5565b60405180910390f35b61017560048036038101906101709190611838565b6104eb565b60405161018291906118a2565b60405180910390f35b6101a560048036038101906101a091906118e5565b610506565b005b6101c160048036038101906101bc9190611923565b61051c565b005b6101dd60048036038101906101d891906118e5565b61061b565b005b6101f960048036038101906101f49190611923565b610646565b005b61021560048036038101906102109190611838565b610665565b005b610231600480360381019061022c9190611838565b610671565b60405161023e91906118a2565b60405180910390f35b610261600480360381019061025c9190611973565b610682565b60405161026e91906119ad565b60405180910390f35b610291600480360381019061028c9190611973565b6106aa565b60405161029e91906119ad565b60405180910390f35b6102af610760565b6040516102bc91906119ad565b60405180910390f35b6102cd610766565b6040516102da91906117e5565b60405180910390f35b6102fd60048036038101906102f891906119f0565b6107f6565b005b61031960048036038101906103149190611b5a565b61080c565b005b61033560048036038101906103309190611838565b610831565b60405161034291906117e5565b60405180910390f35b61036560048036038101906103609190611bda565b610897565b604051610372919061175c565b60405180910390f35b5f7f80ac58cd000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916148061044557507f5b5e139f000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916145b80610455575061045482610925565b5b9050919050565b60605f805461046a90611c45565b80601f016020809104026020016040519081016040528092919081815260200182805461049690611c45565b80156104e15780601f106104b8576101008083540402835291602001916104e1565b820191905f5260205f20905b8154815290600101906020018083116104c457829003601f168201915b5050505050905090565b5f6104f58261098e565b506104ff82610a14565b9050919050565b6105188282610513610a4d565b610a54565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361058c575f6040517f64a0ae9200000000000000000000000000000000000000000000000000000000815260040161058391906118a2565b60405180910390fd5b5f61059f838361059a610a4d565b610a66565b90508373ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614610615578382826040517f64283d7b00000000000000000000000000000000000000000000000000000000815260040161060c93929190611c75565b60405180910390fd5b50505050565b6106258282610c71565b60065481106106425760018161063b9190611cd7565b6006819055505b5050565b61066083838360405180602001604052805f81525061080c565b505050565b61066e81610d64565b50565b5f61067b8261098e565b9050919050565b5f60065f81548092919061069590611d0a565b9190505590506106a58282610c71565b919050565b5f5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361071b575f6040517f89c62b6400000000000000000000000000000000000000000000000000000000815260040161071291906118a2565b60405180910390fd5b60035f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b60065481565b60606001805461077590611c45565b80601f01602080910402602001604051908101604052809291908181526020018280546107a190611c45565b80156107ec5780601f106107c3576101008083540402835291602001916107ec565b820191905f5260205f20905b8154815290600101906020018083116107cf57829003601f168201915b5050505050905090565b610808610801610a4d565b8383610de6565b5050565b61081784848461051c565b61082b610822610a4d565b85858585610f4f565b50505050565b606061083c8261098e565b505f6108466110fb565b90505f8151116108645760405180602001604052805f81525061088f565b8061086e8461111b565b60405160200161087f929190611d8b565b6040516020818303038152906040525b915050919050565b5f60055f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff16905092915050565b5f7f01ffc9a7000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916149050919050565b5f5f610999836111e5565b90505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603610a0b57826040517f7e273289000000000000000000000000000000000000000000000000000000008152600401610a0291906119ad565b60405180910390fd5b80915050919050565b5f60045f8381526020019081526020015f205f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050919050565b5f33905090565b610a61838383600161121e565b505050565b5f5f610a71846111e5565b90505f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1614610ab257610ab18184866113dd565b5b5f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614610b3d57610af15f855f5f61121e565b600160035f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825403925050819055505b5f73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff1614610bbc57600160035f8773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8460025f8681526020019081526020015f205f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550838573ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef60405160405180910390a4809150509392505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610ce1575f6040517f64a0ae92000000000000000000000000000000000000000000000000000000008152600401610cd891906118a2565b60405180910390fd5b5f610ced83835f610a66565b90505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614610d5f575f6040517f73c6ac6e000000000000000000000000000000000000000000000000000000008152600401610d5691906118a2565b60405180910390fd5b505050565b5f610d705f835f610a66565b90505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603610de257816040517f7e273289000000000000000000000000000000000000000000000000000000008152600401610dd991906119ad565b60405180910390fd5b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610e5657816040517f5b08ba18000000000000000000000000000000000000000000000000000000008152600401610e4d91906118a2565b60405180910390fd5b8060055f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff0219169083151502179055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c3183604051610f42919061175c565b60405180910390a3505050565b5f8373ffffffffffffffffffffffffffffffffffffffff163b11156110f4578273ffffffffffffffffffffffffffffffffffffffff1663150b7a02868685856040518563ffffffff1660e01b8152600401610fad9493929190611e00565b6020604051808303815f875af1925050508015610fe857506040513d601f19601f82011682018060405250810190610fe59190611e5e565b60015b611069573d805f8114611016576040519150601f19603f3d011682016040523d82523d5f602084013e61101b565b606091505b505f81510361106157836040517f64a0ae9200000000000000000000000000000000000000000000000000000000815260040161105891906118a2565b60405180910390fd5b805160208201fd5b63150b7a0260e01b7bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916817bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916146110f257836040517f64a0ae920000000000000000000000000000000000000000000000000000000081526004016110e991906118a2565b60405180910390fd5b505b5050505050565b6060604051806060016040528060248152602001611ede60249139905090565b60605f6001611129846114a0565b0190505f8167ffffffffffffffff81111561114757611146611a36565b5b6040519080825280601f01601f1916602001820160405280156111795781602001600182028036833780820191505090505b5090505f82602083010190505b6001156111da578080600190039150507f3031323334353637383961626364656600000000000000000000000000000000600a86061a8153600a85816111cf576111ce611e89565b5b0494505f8503611186575b819350505050919050565b5f60025f8381526020019081526020015f205f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050919050565b808061125657505f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1614155b15611388575f6112658461098e565b90505f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16141580156112cf57508273ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614155b80156112e257506112e08184610897565b155b1561132457826040517fa9fbf51f00000000000000000000000000000000000000000000000000000000815260040161131b91906118a2565b60405180910390fd5b811561138657838573ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92560405160405180910390a45b505b8360045f8581526020019081526020015f205f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050505050565b6113e88383836115f1565b61149b575f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361145c57806040517f7e27328900000000000000000000000000000000000000000000000000000000815260040161145391906119ad565b60405180910390fd5b81816040517f177e802f000000000000000000000000000000000000000000000000000000008152600401611492929190611eb6565b60405180910390fd5b505050565b5f5f5f90507a184f03e93ff9f4daa797ed6e38ed64bf6a1f01000000000000000083106114fc577a184f03e93ff9f4daa797ed6e38ed64bf6a1f01000000000000000083816114f2576114f1611e89565b5b0492506040810190505b6d04ee2d6d415b85acef81000000008310611539576d04ee2d6d415b85acef8100000000838161152f5761152e611e89565b5b0492506020810190505b662386f26fc10000831061156857662386f26fc10000838161155e5761155d611e89565b5b0492506010810190505b6305f5e1008310611591576305f5e100838161158757611586611e89565b5b0492506008810190505b61271083106115b65761271083816115ac576115ab611e89565b5b0492506004810190505b606483106115d957606483816115cf576115ce611e89565b5b0492506002810190505b600a83106115e8576001810190505b80915050919050565b5f5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16141580156116a857508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff16148061166957506116688484610897565b5b806116a757508273ffffffffffffffffffffffffffffffffffffffff1661168f83610a14565b73ffffffffffffffffffffffffffffffffffffffff16145b5b90509392505050565b5f604051905090565b5f5ffd5b5f5ffd5b5f7fffffffff0000000000000000000000000000000000000000000000000000000082169050919050565b6116f6816116c2565b8114611700575f5ffd5b50565b5f81359050611711816116ed565b92915050565b5f6020828403121561172c5761172b6116ba565b5b5f61173984828501611703565b91505092915050565b5f8115159050919050565b61175681611742565b82525050565b5f60208201905061176f5f83018461174d565b92915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f6117b782611775565b6117c1818561177f565b93506117d181856020860161178f565b6117da8161179d565b840191505092915050565b5f6020820190508181035f8301526117fd81846117ad565b905092915050565b5f819050919050565b61181781611805565b8114611821575f5ffd5b50565b5f813590506118328161180e565b92915050565b5f6020828403121561184d5761184c6116ba565b5b5f61185a84828501611824565b91505092915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61188c82611863565b9050919050565b61189c81611882565b82525050565b5f6020820190506118b55f830184611893565b92915050565b6118c481611882565b81146118ce575f5ffd5b50565b5f813590506118df816118bb565b92915050565b5f5f604083850312156118fb576118fa6116ba565b5b5f611908858286016118d1565b925050602061191985828601611824565b9150509250929050565b5f5f5f6060848603121561193a576119396116ba565b5b5f611947868287016118d1565b9350506020611958868287016118d1565b925050604061196986828701611824565b9150509250925092565b5f60208284031215611988576119876116ba565b5b5f611995848285016118d1565b91505092915050565b6119a781611805565b82525050565b5f6020820190506119c05f83018461199e565b92915050565b6119cf81611742565b81146119d9575f5ffd5b50565b5f813590506119ea816119c6565b92915050565b5f5f60408385031215611a0657611a056116ba565b5b5f611a13858286016118d1565b9250506020611a24858286016119dc565b9150509250929050565b5f5ffd5b5f5ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b611a6c8261179d565b810181811067ffffffffffffffff82111715611a8b57611a8a611a36565b5b80604052505050565b5f611a9d6116b1565b9050611aa98282611a63565b919050565b5f67ffffffffffffffff821115611ac857611ac7611a36565b5b611ad18261179d565b9050602081019050919050565b828183375f83830152505050565b5f611afe611af984611aae565b611a94565b905082815260208101848484011115611b1a57611b19611a32565b5b611b25848285611ade565b509392505050565b5f82601f830112611b4157611b40611a2e565b5b8135611b51848260208601611aec565b91505092915050565b5f5f5f5f60808587031215611b7257611b716116ba565b5b5f611b7f878288016118d1565b9450506020611b90878288016118d1565b9350506040611ba187828801611824565b925050606085013567ffffffffffffffff811115611bc257611bc16116be565b5b611bce87828801611b2d565b91505092959194509250565b5f5f60408385031215611bf057611bef6116ba565b5b5f611bfd858286016118d1565b9250506020611c0e858286016118d1565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680611c5c57607f821691505b602082108103611c6f57611c6e611c18565b5b50919050565b5f606082019050611c885f830186611893565b611c95602083018561199e565b611ca26040830184611893565b949350505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611ce182611805565b9150611cec83611805565b9250828201905080821115611d0457611d03611caa565b5b92915050565b5f611d1482611805565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203611d4657611d45611caa565b5b600182019050919050565b5f81905092915050565b5f611d6582611775565b611d6f8185611d51565b9350611d7f81856020860161178f565b80840191505092915050565b5f611d968285611d5b565b9150611da28284611d5b565b91508190509392505050565b5f81519050919050565b5f82825260208201905092915050565b5f611dd282611dae565b611ddc8185611db8565b9350611dec81856020860161178f565b611df58161179d565b840191505092915050565b5f608082019050611e135f830187611893565b611e206020830186611893565b611e2d604083018561199e565b8181036060830152611e3f8184611dc8565b905095945050505050565b5f81519050611e58816116ed565b92915050565b5f60208284031215611e7357611e726116ba565b5b5f611e8084828501611e4a565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601260045260245ffd5b5f604082019050611ec95f830185611893565b611ed6602083018461199e565b939250505056fe68747470733a2f2f65766d6b69742d66697874757265732e746573742f6572633732312fa2646970667358221220e20a1ef0e04f61cf6dffab4e6246cc65315b37433e63f53cecb31a0e6209b8ee64736f6c63430008230033"

	// ERC1155ByteCodeHex is OpenZeppelin's ERC1155 plus mint/mintBatch/burn/burnBatch, constructor(uri).
	ERC1155ByteCodeHex = "0x608060405234801561000f575f5ffd5b5060405161272f38038061272f833981810160405281019061003191906101a8565b806100418161004860201b60201c565b50506104df565b80600290816100579190610410565b5050565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6100ba82610074565b810181811067ffffffffffffffff821117156100d9576100d8610084565b5b80604052505050565b5f6100eb61005b565b90506100f782826100b1565b919050565b5f67ffffffffffffffff82111561011657610115610084565b5b61011f82610074565b9050602081019050919050565b8281835e5f83830152505050565b5f61014c610147846100fc565b6100e2565b90508281526020810184848401111561016857610167610070565b5b61017384828561012c565b509392505050565b5f82601f83011261018f5761018e61006c565b5b815161019f84826020860161013a565b91505092915050565b5f602082840312156101bd576101bc610064565b5b5f82015167ffffffffffffffff8111156101da576101d9610068565b5b6101e68482850161017b565b91505092915050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061023d57607f821691505b6020821081036102505761024f6101f9565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026102b27fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610277565b6102bc8683610277565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f6103006102fb6102f6846102d4565b6102dd565b6102d4565b9050919050565b5f819050919050565b610319836102e6565b61032d61032582610307565b848454610283565b825550505050565b5f5f905090565b610344610335565b61034f818484610310565b505050565b5f5b828110156103755761036a5f82840161033c565b600181019050610356565b505050565b601f8211156103c857828211156103c75761039481610256565b61039d83610268565b6103a685610268565b60208610156103b3575f90505b8083016103c282840382610354565b505050505b5b505050565b5f82821c905092915050565b5f6103e85f19846008026103cd565b1980831691505092915050565b5f61040083836103d9565b9150826002028217905092915050565b610419826101ef565b67ffffffffffffffff81111561043257610431610084565b5b61043c8254610226565b61044782828561037a565b5f60209050601f831160018114610478575f8415610466578287015190505b61047085826103f5565b8655506104d7565b601f19841661048686610256565b5f5b828110156104ad57848901518255600182019150602085019450602081019050610488565b868310156104ca57848901516104c6601f8916826103d9565b8355505b6001600288020188555050505b505050505050565b612243806104ec5f395ff3fe608060405234801561000f575f5ffd5b50600436106100b1575f3560e01c80636b20c4541161006f5780636b20c454146101ad578063a22cb465146101c9578063d81d0a15146101e5578063e985e9c514610201578063f242432a14610231578063f5298aca1461024d576100b1565b8062fdd58e146100b557806301ffc9a7146100e55780630e89341c14610115578063156e29f6146101455780632eb2c2d6146101615780634e1273f41461017d575b5f5ffd5b6100cf60048036038101906100ca9190611639565b610269565b6040516100dc9190611686565b60405180910390f35b6100ff60048036038101906100fa91906116f4565b6102be565b60405161010c9190611739565b60405180910390f35b61012f600480360381019061012a9190611752565b61039f565b60405161013c91906117ed565b60405180910390f35b61015f600480360381019061015a919061180d565b610431565b005b61017b60048036038101906101769190611a4d565b610450565b005b61019760048036038101906101929190611bd8565b6104f7565b6040516101a49190611d05565b60405180910390f35b6101c760048036038101906101c29190611d25565b610601565b005b6101e360048036038101906101de9190611dd7565b610611565b005b6101ff60048036038101906101fa9190611d25565b610627565b005b61021b60048036038101906102169190611e15565b610646565b6040516102289190611739565b60405180910390f35b61024b60048036038101906102469190611e53565b6106d4565b005b6102676004803603810190610262919061180d565b61077b565b005b5f5f5f8381526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f7fd9b67a26000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916148061038857507f0e89341c000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916145b8061039857506103978261078b565b5b9050919050565b6060600280546103ae90611f13565b80601f01602080910402602001604051908101604052809291908181526020018280546103da90611f13565b80156104255780601f106103fc57610100808354040283529160200191610425565b820191905f5260205f20905b81548152906001019060200180831161040857829003601f168201915b50505050509050919050565b61044b83838360405180602001604052805f8152506107f4565b505050565b5f610459610889565b90508073ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff161415801561049e575061049c8682610646565b155b156104e25780866040517fe237d9220000000000000000000000000000000000000000000000000000000081526004016104d9929190611f52565b60405180910390fd5b6104ef8686868686610890565b505050505050565b6060815183511461054357815183516040517f5b05999100000000000000000000000000000000000000000000000000000000815260040161053a929190611f79565b60405180910390fd5b5f835167ffffffffffffffff81111561055f5761055e611861565b5b60405190808252806020026020018201604052801561058d5781602001602082028036833780820191505090505b5090505f5f90505b84518110156105f6576105cc6105b4828761098490919063ffffffff16565b6105c7838761099790919063ffffffff16565b610269565b8282815181106105df576105de611fa0565b5b602002602001018181525050806001019050610595565b508091505092915050565b61060c8383836109aa565b505050565b61062361061c610889565b8383610a3a565b5050565b61064183838360405180602001604052805f815250610ba3565b505050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff16905092915050565b5f6106dd610889565b90508073ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff161415801561072257506107208682610646565b155b156107665780866040517fe237d92200000000000000000000000000000000000000000000000000000000815260040161075d929190611f52565b60405180910390fd5b6107738686868686610c26565b505050505050565b610786838383610d2c565b505050565b5f7f01ffc9a7000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916149050919050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610864575f6040517f57f447ce00000000000000000000000000000000000000000000000000000000815260040161085b9190611fcd565b60405180910390fd5b5f5f6108708585610dce565b915091506108815f87848487610dfe565b505050505050565b5f33905090565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610900575f6040517f57f447ce0000000000000000000000000000000000000000000000000000000081526004016108f79190611fcd565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff1603610970575f6040517f01a835140000000000000000000000000000000000000000000000000000000081526004016109679190611fcd565b60405180910390fd5b61097d8585858585610dfe565b5050505050565b5f60208202602084010151905092915050565b5f60208202602084010151905092915050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610a1a575f6040517f01a83514000000000000000000000000000000000000000000000000000000008152600401610a119190611fcd565b60405180910390fd5b610a35835f848460405180602001604052805f815250610dfe565b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610aaa575f6040517fced3e100000000000000000000000000000000000000000000000000000000008152600401610aa19190611fcd565b60405180910390fd5b8060015f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff0219169083151502179055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c3183604051610b969190611739565b60405180910390a3505050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610c13575f6040517f57f447ce000000000000000000000000000000000000000000000000000000008152600401610c0a9190611fcd565b60405180910390fd5b610c205f85858585610dfe565b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610c96575f6040517f57f447ce000000000000000000000000000000000000000000000000000000008152600401610c8d9190611fcd565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff1603610d06575f6040517f01a83514000000000000000000000000000000000000000000000000000000008152600401610cfd9190611fcd565b60405180910390fd5b5f5f610d128585610dce565b91509150610d238787848487610dfe565b50505050505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610d9c575f6040517f01a83514000000000000000000000000000000000000000000000000000000008152600401610d939190611fcd565b60405180910390fd5b5f5f610da88484610dce565b91509150610dc7855f848460405180602001604052805f815250610dfe565b5050505050565b60608060405191506001825283602083015260408201905060018152826020820152604081016040529250929050565b610e0a85858585610eaa565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1614610ea3575f610e46610889565b90506001845103610e92575f610e655f8661099790919063ffffffff16565b90505f610e7b5f8661099790919063ffffffff16565b9050610e8b83898985858961123d565b5050610ea1565b610ea08187878787876113ec565b5b505b5050505050565b8051825114610ef457815181516040517f5b059991000000000000000000000000000000000000000000000000000000008152600401610eeb929190611f79565b60405180910390fd5b5f610efd610889565b90505f5f90505b83518110156110fc575f610f21828661099790919063ffffffff16565b90505f610f37838661099790919063ffffffff16565b90505f73ffffffffffffffffffffffffffffffffffffffff168873ffffffffffffffffffffffffffffffffffffffff161461105a575f5f5f8481526020019081526020015f205f8a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205490508181101561100657888183856040517f03dee4c5000000000000000000000000000000000000000000000000000000008152600401610ffd9493929190611fe6565b60405180910390fd5b8181035f5f8581526020019081526020015f205f8b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff16146110ef57805f5f8481526020019081526020015f205f8973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8282546110e79190612056565b925050819055505b5050806001019050610f04565b5060018351036111b7575f61111a5f8561099790919063ffffffff16565b90505f6111305f8561099790919063ffffffff16565b90508573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f6285856040516111a8929190611f79565b60405180910390a45050611236565b8373ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff167f4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb868660405161122d929190612089565b60405180910390a45b5050505050565b5f8473ffffffffffffffffffffffffffffffffffffffff163b11156113e4578373ffffffffffffffffffffffffffffffffffffffff1663f23a6e6187878686866040518663ffffffff1660e01b815260040161129d959493929190612110565b6020604051808303815f875af19250505080156112d857506040513d601f19601f820116820180604052508101906112d5919061217c565b60015b611359573d805f8114611306576040519150601f19603f3d011682016040523d82523d5f602084013e61130b565b606091505b505f81510361135157846040517f57f447ce0000000000000000000000000000000000000000000000000000000081526004016113489190611fcd565b60405180910390fd5b805160208201fd5b63f23a6e6160e01b7bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916817bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916146113e257846040517f57f447ce0000000000000000000000000000000000000000000000000000000081526004016113d99190611fcd565b60405180910390fd5b505b505050505050565b5f8473ffffffffffffffffffffffffffffffffffffffff163b1115611593578373ffffffffffffffffffffffffffffffffffffffff1663bc197c8187878686866040518663ffffffff1660e01b815260040161144c9594939291906121a7565b6020604051808303815f875af192505050801561148757506040513d601f19601f82011682018060405250810190611484919061217c565b60015b611508573d805f81146114b5576040519150601f19603f3d011682016040523d82523d5f602084013e6114ba565b606091505b505f81510361150057846040517f57f447ce0000000000000000000000000000000000000000000000000000000081526004016114f79190611fcd565b60405180910390fd5b805160208201fd5b63bc197c8160e01b7bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916817bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19161461159157846040517f57f447ce0000000000000000000000000000000000000000000000000000000081526004016115889190611fcd565b60405180910390fd5b505b505050505050565b5f604051905090565b5f5ffd5b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6115d5826115ac565b9050919050565b6115e5816115cb565b81146115ef575f5ffd5b50565b5f81359050611600816115dc565b92915050565b5f819050919050565b61161881611606565b8114611622575f5ffd5b50565b5f813590506116338161160f565b92915050565b5f5f6040838503121561164f5761164e6115a4565b5b5f61165c858286016115f2565b925050602061166d85828601611625565b9150509250929050565b61168081611606565b82525050565b5f6020820190506116995f830184611677565b92915050565b5f7fffffffff0000000000000000000000000000000000000000000000000000000082169050919050565b6116d38161169f565b81146116dd575f5ffd5b50565b5f813590506116ee816116ca565b92915050565b5f60208284031215611709576117086115a4565b5b5f611716848285016116e0565b91505092915050565b5f8115159050919050565b6117338161171f565b82525050565b5f60208201905061174c5f83018461172a565b92915050565b5f60208284031215611767576117666115a4565b5b5f61177484828501611625565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f6117bf8261177d565b6117c98185611787565b93506117d9818560208601611797565b6117e2816117a5565b840191505092915050565b5f6020820190508181035f83015261180581846117b5565b905092915050565b5f5f5f60608486031215611824576118236115a4565b5b5f611831868287016115f2565b935050602061184286828701611625565b925050604061185386828701611625565b9150509250925092565b5f5ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b611897826117a5565b810181811067ffffffffffffffff821117156118b6576118b5611861565b5b80604052505050565b5f6118c861159b565b90506118d4828261188e565b919050565b5f67ffffffffffffffff8211156118f3576118f2611861565b5b602082029050602081019050919050565b5f5ffd5b5f61191a611915846118d9565b6118bf565b9050808382526020820190506020840283018581111561193d5761193c611904565b5b835b8181101561196657806119528882611625565b84526020840193505060208101905061193f565b5050509392505050565b5f82601f8301126119845761198361185d565b5b8135611994848260208601611908565b91505092915050565b5f5ffd5b5f67ffffffffffffffff8211156119bb576119ba611861565b5b6119c4826117a5565b9050602081019050919050565b828183375f83830152505050565b5f6119f16119ec846119a1565b6118bf565b905082815260208101848484011115611a0d57611a0c61199d565b5b611a188482856119d1565b509392505050565b5f82601f830112611a3457611a3361185d565b5b8135611a448482602086016119df565b91505092915050565b5f5f5f5f5f60a08688031215611a6657611a656115a4565b5b5f611a73888289016115f2565b9550506020611a84888289016115f2565b945050604086013567ffffffffffffffff811115611aa557611aa46115a8565b5b611ab188828901611970565b935050606086013567ffffffffffffffff811115611ad257611ad16115a8565b5b611ade88828901611970565b925050608086013567ffffffffffffffff811115611aff57611afe6115a8565b5b611b0b88828901611a20565b9150509295509295909350565b5f67ffffffffffffffff821115611b3257611b31611861565b5b602082029050602081019050919050565b5f611b55611b5084611b18565b6118bf565b90508083825260208201905060208402830185811115611b7857611b77611904565b5b835b81811015611ba15780611b8d88826115f2565b845260208401935050602081019050611b7a565b5050509392505050565b5f82601f830112611bbf57611bbe61185d565b5b8135611bcf848260208601611b43565b91505092915050565b5f5f60408385031215611bee57611bed6115a4565b5b5f83013567ffffffffffffffff811115611c0b57611c0a6115a8565b5b611c1785828601611bab565b925050602083013567ffffffffffffffff811115611c3857611c376115a8565b5b611c4485828601611970565b9150509250929050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b611c8081611606565b82525050565b5f611c918383611c77565b60208301905092915050565b5f602082019050919050565b5f611cb382611c4e565b611cbd8185611c58565b9350611cc883611c68565b805f5b83811015611cf8578151611cdf8882611c86565b9750611cea83611c9d565b925050600181019050611ccb565b5085935050505092915050565b5f6020820190508181035f830152611d1d8184611ca9565b905092915050565b5f5f5f60608486031215611d3c57611d3b6115a4565b5b5f611d49868287016115f2565b935050602084013567ffffffffffffffff811115611d6a57611d696115a8565b5b611d7686828701611970565b925050604084013567ffffffffffffffff811115611d9757611d966115a8565b5b611da386828701611970565b9150509250925092565b611db68161171f565b8114611dc0575f5ffd5b50565b5f81359050611dd181611dad565b92915050565b5f5f60408385031215611ded57611dec6115a4565b5b5f611dfa858286016115f2565b9250506020611e0b85828601611dc3565b9150509250929050565b5f5f60408385031215611e2b57611e2a6115a4565b5b5f611e38858286016115f2565b9250506020611e49858286016115f2565b9150509250929050565b5f5f5f5f5f60a08688031215611e6c57611e6b6115a4565b5b5f611e79888289016115f2565b9550506020611e8a888289016115f2565b9450506040611e9b88828901611625565b9350506060611eac88828901611625565b925050608086013567ffffffffffffffff811115611ecd57611ecc6115a8565b5b611ed988828901611a20565b9150509295509295909350565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680611f2a57607f821691505b602082108103611f3d57611f3c611ee6565b5b50919050565b611f4c816115cb565b82525050565b5f604082019050611f655f830185611f43565b611f726020830184611f43565b9392505050565b5f604082019050611f8c5f830185611677565b611f996020830184611677565b9392505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f602082019050611fe05f830184611f43565b92915050565b5f608082019050611ff95f830187611f43565b6120066020830186611677565b6120136040830185611677565b6120206060830184611677565b95945050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61206082611606565b915061206b83611606565b925082820190508082111561208357612082612029565b5b92915050565b5f6040820190508181035f8301526120a18185611ca9565b905081810360208301526120b58184611ca9565b90509392505050565b5f81519050919050565b5f82825260208201905092915050565b5f6120e2826120be565b6120ec81856120c8565b93506120fc818560208601611797565b612105816117a5565b840191505092915050565b5f60a0820190506121235f830188611f43565b6121306020830187611f43565b61213d6040830186611677565b61214a6060830185611677565b818103608083015261215c81846120d8565b90509695505050505050565b5f81519050612176816116ca565b92915050565b5f60208284031215612191576121906115a4565b5b5f61219e84828501612168565b91505092915050565b5f60a0820190506121ba5f830188611f43565b6121c76020830187611f43565b81810360408301526121d98186611ca9565b905081810360608301526121ed8185611ca9565b9050818103608083015261220181846120d8565b9050969550505050505056fea264697066735822122092d466ed5172d8b5e4e6079fa99f5289a55ec7abb2f9d6d7287bc60161a41d4564736f6c63430008230033"

	// ERC2612ByteCodeHex is OpenZeppelin's ERC20Permit plus mint/burn, constructor(name, symbol).
	ERC2612ByteCodeHex = "0x610160604052348015610010575f5ffd5b506040516125d83803806125d8833981810160405281019061003291906103a2565b81806040518060400160405280600181526020017f31000000000000000000000000000000000000000000000000000000000000008152508484816003908161007b9190610639565b50806004908161008b9190610639565b5050506100a260058361014060201b90919060201c565b61012081815250506100be60068261014060201b90919060201c565b6101408181525050818051906020012060e08181525050808051906020012061010081815250504660a081815250506100fb61018d60201b60201c565b608081815250503073ffffffffffffffffffffffffffffffffffffffff1660c08173ffffffffffffffffffffffffffffffffffffffff168152505050505050506108ba565b5f6020835110156101615761015a836101e760201b60201c565b9050610187565b826101718361024c60201b60201c565b5f01908161017f9190610639565b5060ff5f1b90505b92915050565b5f7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60e0516101005146306040516020016101cc95949392919061076e565b60405160208183030381529060405280519060200120905090565b5f5f829050601f8151111561023357826040517f305a27a900000000000000000000000000000000000000000000000000000000815260040161022a9190610807565b60405180910390fd5b80518161023f90610854565b5f1c175f1b915050919050565b5f819050919050565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6102b48261026e565b810181811067ffffffffffffffff821117156102d3576102d261027e565b5b80604052505050565b5f6102e5610255565b90506102f182826102ab565b919050565b5f67ffffffffffffffff8211156103105761030f61027e565b5b6103198261026e565b9050602081019050919050565b8281835e5f83830152505050565b5f610346610341846102f6565b6102dc565b9050828152602081018484840111156103625761036161026a565b5b61036d848285610326565b509392505050565b5f82601f83011261038957610388610266565b5b8151610399848260208601610334565b91505092915050565b5f5f604083850312156103b8576103b761025e565b5b5f83015167ffffffffffffffff8111156103d5576103d4610262565b5b6103e185828601610375565b925050602083015167ffffffffffffffff81111561040257610401610262565b5b61040e85828601610375565b9150509250929050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061046657607f821691505b60208210810361047957610478610422565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026104db7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826104a0565b6104e586836104a0565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61052961052461051f846104fd565b610506565b6104fd565b9050919050565b5f819050919050565b6105428361050f565b61055661054e82610530565b8484546104ac565b825550505050565b5f5f905090565b61056d61055e565b610578818484610539565b505050565b5f5b8281101561059e576105935f828401610565565b60018101905061057f565b505050565b601f8211156105f157828211156105f0576105bd8161047f565b6105c683610491565b6105cf85610491565b60208610156105dc575f90505b8083016105eb8284038261057d565b505050505b5b505050565b5f82821c905092915050565b5f6106115f19846008026105f6565b1980831691505092915050565b5f6106298383610602565b9150826002028217905092915050565b61064282610418565b67ffffffffffffffff81111561065b5761065a61027e565b5b610665825461044f565b6106708282856105a3565b5f60209050601f8311600181146106a1575f841561068f578287015190505b610699858261061e565b865550610700565b601f1984166106af8661047f565b5f5b828110156106d6578489015182556001820191506020850194506020810190506106b1565b868310156106f357848901516106ef601f891682610602565b8355505b6001600288020188555050505b505050505050565b5f819050919050565b61071a81610708565b82525050565b610729816104fd565b82525050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6107588261072f565b9050919050565b6107688161074e565b82525050565b5f60a0820190506107815f830188610711565b61078e6020830187610711565b61079b6040830186610711565b6107a86060830185610720565b6107b5608083018461075f565b9695505050505050565b5f82825260208201905092915050565b5f6107d982610418565b6107e381856107bf565b93506107f3818560208601610326565b6107fc8161026e565b840191505092915050565b5f6020820190508181035f83015261081f81846107cf565b905092915050565b5f81519050919050565b5f819050602082019050919050565b5f61084b8251610708565b80915050919050565b5f61085e82610827565b8261086884610831565b905061087381610840565b925060208210156108b3576108ae7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff836020036008026104a0565b831692505b5050919050565b60805160a05160c05160e051610100516101205161014051611ccd61090b5f395f610b0f01525f610ad401525f61108701525f61106601525f61095301525f6109a901525f6109d20152611ccd5ff3fe608060405234801561000f575f5ffd5b50600436106100f3575f3560e01c806370a08231116100955780639dc29fac116100645780639dc29fac1461028d578063a9059cbb146102a9578063d505accf146102d9578063dd62ed3e146102f5576100f3565b806370a08231146101eb5780637ecebe001461021b57806384b0196e1461024b57806395d89b411461026f576100f3565b806323b872dd116100d157806323b872dd14610163578063313ce567146101935780633644e515146101b157806340c10f19146101cf576100f3565b806306fdde03146100f7578063095ea7b31461011557806318160ddd14610145575b5f5ffd5b6100ff610325565b60405161010c919061153d565b60405180910390f35b61012f600480360381019061012a91906115ee565b6103b5565b60405161013c9190611646565b60405180910390f35b61014d6103d7565b60405161015a919061166e565b60405180910390f35b61017d60048036038101906101789190611687565b6103e0565b60405161018a9190611646565b60405180910390f35b61019b61040e565b6040516101a891906116f2565b60405180910390f35b6101b9610416565b6040516101c69190611723565b60405180910390f35b6101e960048036038101906101e491906115ee565b610424565b005b6102056004803603810190610200919061173c565b610432565b604051610212919061166e565b60405180910390f35b6102356004803603810190610230919061173c565b610477565b604051610242919061166e565b60405180910390f35b610253610488565b6040516102669796959493929190611867565b60405180910390f35b61027761052d565b604051610284919061153d565b60405180910390f35b6102a760048036038101906102a291906115ee565b6105bd565b005b6102c360048036038101906102be91906115ee565b6105cb565b6040516102d09190611646565b60405180910390f35b6102f360048036038101906102ee919061193d565b6105ed565b005b61030f600480360381019061030a91906119da565b610732565b60405161031c919061166e565b60405180910390f35b60606003805461033490611a45565b80601f016020809104026020016040519081016040528092919081815260200182805461036090611a45565b80156103ab5780601f10610382576101008083540402835291602001916103ab565b820191905f5260205f20905b81548152906001019060200180831161038e57829003601f168201915b5050505050905090565b5f5f6103bf6107b4565b90506103cc8185856107bb565b600191505092915050565b5f600254905090565b5f5f6103ea6107b4565b90506103f78582856107cd565b610402858585610860565b60019150509392505050565b5f6012905090565b5f61041f610950565b905090565b61042e8282610a06565b5050565b5f5f5f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b5f61048182610a85565b9050919050565b5f6060805f5f5f6060610499610acb565b6104a1610b06565b46305f5f1b5f67ffffffffffffffff8111156104c0576104bf611a75565b5b6040519080825280602002602001820160405280156104ee5781602001602082028036833780820191505090505b507f0f00000000000000000000000000000000000000000000000000000000000000959493929190965096509650965096509650965090919293949596565b60606004805461053c90611a45565b80601f016020809104026020016040519081016040528092919081815260200182805461056890611a45565b80156105b35780601f1061058a576101008083540402835291602001916105b3565b820191905f5260205f20905b81548152906001019060200180831161059657829003601f168201915b5050505050905090565b6105c78282610b41565b5050565b5f5f6105d56107b4565b90506105e2818585610860565b600191505092915050565b8342111561063257836040517f62791302000000000000000000000000000000000000000000000000000000008152600401610629919061166e565b60405180910390fd5b5f7f6e71edae12b1b97f4d1f60370fef10105fa2faae0126114a169c64845d6126c98888886106608c610bc0565b8960405160200161067696959493929190611aa2565b6040516020818303038152906040528051906020012090505f61069882610c13565b90505f6106a782878787610c2c565b90508973ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161461071b57808a6040517f4b800e46000000000000000000000000000000000000000000000000000000008152600401610712929190611b01565b60405180910390fd5b6107268a8a8a6107bb565b50505050505050505050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f33905090565b6107c88383836001610c5a565b505050565b5f6107d88484610732565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81101561085a578181101561084b578281836040517ffb8f41b200000000000000000000000000000000000000000000000000000000815260040161084293929190611b28565b60405180910390fd5b61085984848484035f610c5a565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036108d0575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016108c79190611b5d565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610940575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016109379190611b5d565b60405180910390fd5b61094b838383610e29565b505050565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff161480156109cb57507f000000000000000000000000000000000000000000000000000000000000000046145b156109f8577f00000000000000000000000000000000000000000000000000000000000000009050610a03565b610a00611042565b90505b90565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610a76575f6040517fec442f05000000000000000000000000000000000000000000000000000000008152600401610a6d9190611b5d565b60405180910390fd5b610a815f8383610e29565b5050565b5f60075f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b6060610b0160057f00000000000000000000000000000000000000000000000000000000000000006110d790919063ffffffff16565b905090565b6060610b3c60067f00000000000000000000000000000000000000000000000000000000000000006110d790919063ffffffff16565b905090565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610bb1575f6040517f96c6fd1e000000000000000000000000000000000000000000000000000000008152600401610ba89190611b5d565b60405180910390fd5b610bbc825f83610e29565b5050565b5f60075f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f815480929190600101919050559050919050565b5f610c25610c1f610950565b83611184565b9050919050565b5f5f5f5f610c3c888888886111c4565b925092509250610c4c82826112ab565b829350505050949350505050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610cca575f6040517fe602df05000000000000000000000000000000000000000000000000000000008152600401610cc19190611b5d565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610d3a575f6040517f94280d62000000000000000000000000000000000000000000000000000000008152600401610d319190611b5d565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508015610e23578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92584604051610e1a919061166e565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610e79578060025f828254610e6d9190611ba3565b92505081905550610f47565b5f5f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905081811015610f02578381836040517fe450d38c000000000000000000000000000000000000000000000000000000008152600401610ef993929190611b28565b60405180910390fd5b8181035f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610f8e578060025f8282540392505081905550610fd8565b805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051611035919061166e565b60405180910390a3505050565b5f7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f7f00000000000000000000000000000000000000000000000000000000000000007f000000000000000000000000000000000000000000000000000000000000000046306040516020016110bc959493929190611bd6565b60405160208183030381529060405280519060200120905090565b606060ff5f1b83146110f3576110ec8361140d565b905061117e565b8180546110ff90611a45565b80601f016020809104026020016040519081016040528092919081815260200182805461112b90611a45565b80156111765780601f1061114d57610100808354040283529160200191611176565b820191905f5260205f20905b81548152906001019060200180831161115957829003601f168201915b505050505090505b92915050565b5f6040517f190100000000000000000000000000000000000000000000000000000000000081528360028201528260228201526042812091505092915050565b5f5f5f7f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0845f1c1115611200575f6003859250925092506112a1565b5f6001888888886040515f81526020016040526040516112239493929190611c27565b6020604051602081039080840390855afa158015611243573d5f5f3e3d5ffd5b5050506020604051035190505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603611294575f60015f5f1b935093509350506112a1565b805f5f5f1b935093509350505b9450945094915050565b5f60038111156112be576112bd611c6a565b5b8260038111156112d1576112d0611c6a565b5b031561140957600160038111156112eb576112ea611c6a565b5b8260038111156112fe576112fd611c6a565b5b03611335576040517ff645eedf00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6002600381111561134957611348611c6a565b5b82600381111561135c5761135b611c6a565b5b036113a057805f1c6040517ffce698f7000000000000000000000000000000000000000000000000000000008152600401611397919061166e565b60405180910390fd5b6003808111156113b3576113b2611c6a565b5b8260038111156113c6576113c5611c6a565b5b0361140857806040517fd78bce0c0000000000000000000000000000000000000000000000000000000081526004016113ff9190611723565b60405180910390fd5b5b5050565b60605f6114198361147f565b90505f602067ffffffffffffffff81111561143757611436611a75565b5b6040519080825280601f01601f1916602001820160405280156114695781602001600182028036833780820191505090505b5090508181528360208201528092505050919050565b5f5f60ff835f1c169050601f8111156114c4576040517fb3512b0c00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b80915050919050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f61150f826114cd565b61151981856114d7565b93506115298185602086016114e7565b611532816114f5565b840191505092915050565b5f6020820190508181035f8301526115558184611505565b905092915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61158a82611561565b9050919050565b61159a81611580565b81146115a4575f5ffd5b50565b5f813590506115b581611591565b92915050565b5f819050919050565b6115cd816115bb565b81146115d7575f5ffd5b50565b5f813590506115e8816115c4565b92915050565b5f5f604083850312156116045761160361155d565b5b5f611611858286016115a7565b9250506020611622858286016115da565b9150509250929050565b5f8115159050919050565b6116408161162c565b82525050565b5f6020820190506116595f830184611637565b92915050565b611668816115bb565b82525050565b5f6020820190506116815f83018461165f565b92915050565b5f5f5f6060848603121561169e5761169d61155d565b5b5f6116ab868287016115a7565b93505060206116bc868287016115a7565b92505060406116cd868287016115da565b9150509250925092565b5f60ff82169050919050565b6116ec816116d7565b82525050565b5f6020820190506117055f8301846116e3565b92915050565b5f819050919050565b61171d8161170b565b82525050565b5f6020820190506117365f830184611714565b92915050565b5f602082840312156117515761175061155d565b5b5f61175e848285016115a7565b91505092915050565b5f7fff0000000000000000000000000000000000000000000000000000000000000082169050919050565b61179b81611767565b82525050565b6117aa81611580565b82525050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b6117e2816115bb565b82525050565b5f6117f383836117d9565b60208301905092915050565b5f602082019050919050565b5f611815826117b0565b61181f81856117ba565b935061182a836117ca565b805f5b8381101561185a57815161184188826117e8565b975061184c836117ff565b92505060018101905061182d565b5085935050505092915050565b5f60e08201905061187a5f83018a611792565b818103602083015261188c8189611505565b905081810360408301526118a08188611505565b90506118af606083018761165f565b6118bc60808301866117a1565b6118c960a0830185611714565b81810360c08301526118db818461180b565b905098975050505050505050565b6118f2816116d7565b81146118fc575f5ffd5b50565b5f8135905061190d816118e9565b92915050565b61191c8161170b565b8114611926575f5ffd5b50565b5f8135905061193781611913565b92915050565b5f5f5f5f5f5f5f60e0888a0312156119585761195761155d565b5b5f6119658a828b016115a7565b97505060206119768a828b016115a7565b96505060406119878a828b016115da565b95505060606119988a828b016115da565b94505060806119a98a828b016118ff565b93505060a06119ba8a828b01611929565b92505060c06119cb8a828b01611929565b91505092959891949750929550565b5f5f604083850312156119f0576119ef61155d565b5b5f6119fd858286016115a7565b9250506020611a0e858286016115a7565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680611a5c57607f821691505b602082108103611a6f57611a6e611a18565b5b50919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b5f60c082019050611ab55f830189611714565b611ac260208301886117a1565b611acf60408301876117a1565b611adc606083018661165f565b611ae9608083018561165f565b611af660a083018461165f565b979650505050505050565b5f604082019050611b145f8301856117a1565b611b2160208301846117a1565b9392505050565b5f606082019050611b3b5f8301866117a1565b611b48602083018561165f565b611b55604083018461165f565b949350505050565b5f602082019050611b705f8301846117a1565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611bad826115bb565b9150611bb8836115bb565b9250828201905080821115611bd057611bcf611b76565b5b92915050565b5f60a082019050611be95f830188611714565b611bf66020830187611714565b611c036040830186611714565b611c10606083018561165f565b611c1d60808301846117a1565b9695505050505050565b5f608082019050611c3a5f830187611714565b611c4760208301866116e3565b611c546040830185611714565b611c616060830184611714565b95945050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffdfea264697066735822122044347b76c25dd816dbe46eb154b880013571f2602582d57a3ff7e772d124eb4e64736f6c63430008230033"

	// ERC6909ByteCodeHex is OpenZeppelin's ERC6909 with metadata plus mint/burn, constructor(id, name, symbol, decimals).
	ERC6909ByteCodeHex = "0x608060405234801561000f575f5ffd5b50604051611df1380380611df183398181016040528101906100319190610343565b610041848461006a60201b60201c565b61005184836100c760201b60201c565b610061848261012560201b60201c565b50505050610756565b8060035f8481526020019081526020015f205f01908161008a91906105f7565b50817f6846707b16f55a3623d38bf118700af63d0784698270a53ce28b116dbaa8b794826040516100bb919061070e565b60405180910390a25050565b8060035f8481526020019081526020015f2060010190816100e891906105f7565b50817f8fabcb7b4fbdde605303a7339bc80fb6956cd01a42e9cdf2d15d0f2fdd3b06fa82604051610119919061070e565b60405180910390a25050565b8060035f8481526020019081526020015f206002015f6101000a81548160ff021916908360ff160217905550817fc7435874568017e6134c851002a405d733521c0f4c63f9f25f3a32a4d595b72982604051610181919061073d565b60405180910390a25050565b5f604051905090565b5f5ffd5b5f5ffd5b5f819050919050565b6101b08161019e565b81146101ba575f5ffd5b50565b5f815190506101cb816101a7565b92915050565b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b61021f826101d9565b810181811067ffffffffffffffff8211171561023e5761023d6101e9565b5b80604052505050565b5f61025061018d565b905061025c8282610216565b919050565b5f67ffffffffffffffff82111561027b5761027a6101e9565b5b610284826101d9565b9050602081019050919050565b8281835e5f83830152505050565b5f6102b16102ac84610261565b610247565b9050828152602081018484840111156102cd576102cc6101d5565b5b6102d8848285610291565b509392505050565b5f82601f8301126102f4576102f36101d1565b5b815161030484826020860161029f565b91505092915050565b5f60ff82169050919050565b6103228161030d565b811461032c575f5ffd5b50565b5f8151905061033d81610319565b92915050565b5f5f5f5f6080858703121561035b5761035a610196565b5b5f610368878288016101bd565b945050602085015167ffffffffffffffff8111156103895761038861019a565b5b610395878288016102e0565b935050604085015167ffffffffffffffff8111156103b6576103b561019a565b5b6103c2878288016102e0565b92505060606103d38782880161032f565b91505092959194509250565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061042d57607f821691505b6020821081036104405761043f6103e9565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026104a27fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610467565b6104ac8683610467565b95508019841693508086168417925050509392505050565b5f819050919050565b5f6104e76104e26104dd8461019e565b6104c4565b61019e565b9050919050565b5f819050919050565b610500836104cd565b61051461050c826104ee565b848454610473565b825550505050565b5f5f905090565b61052b61051c565b6105368184846104f7565b505050565b5f5b8281101561055c576105515f828401610523565b60018101905061053d565b505050565b601f8211156105af57828211156105ae5761057b81610446565b61058483610458565b61058d85610458565b602086101561059a575f90505b8083016105a98284038261053b565b505050505b5b505050565b5f82821c905092915050565b5f6105cf5f19846008026105b4565b1980831691505092915050565b5f6105e783836105c0565b9150826002028217905092915050565b610600826103df565b67ffffffffffffffff811115610619576106186101e9565b5b6106238254610416565b61062e828285610561565b5f60209050601f83116001811461065f575f841561064d578287015190505b61065785826105dc565b8655506106be565b601f19841661066d86610446565b5f5b828110156106945784890151825560018201915060208501945060208101905061066f565b868310156106b157848901516106ad601f8916826105c0565b8355505b6001600288020188555050505b505050505050565b5f82825260208201905092915050565b5f6106e0826103df565b6106ea81856106c6565b93506106fa818560208601610291565b610703816101d9565b840191505092915050565b5f6020820190508181035f83015261072681846106d6565b905092915050565b6107378161030d565b82525050565b5f6020820190506107505f83018461072e565b92915050565b61168e806107635f395ff3fe608060405234801561000f575f5ffd5b50600436106100cb575f3560e01c8063426a84931161008a578063598af9e711610064578063598af9e71461026b578063b6363cf21461029b578063f5298aca146102cb578063fe99049a146102e7576100cb565b8063426a8493146101db5780634e41a1fb1461020b578063558a72971461023b576100cb565b8062ad800c146100cf578062fdd58e146100ff57806301ffc9a71461012f578063095bcdb61461015f578063156e29f61461018f5780633f47e662146101ab575b5f5ffd5b6100e960048036038101906100e491906110fd565b610317565b6040516100f69190611198565b60405180910390f35b61011960048036038101906101149190611212565b6103ba565b604051610126919061125f565b60405180910390f35b610149600480360381019061014491906112cd565b61040f565b6040516101569190611312565b60405180910390f35b6101796004803603810190610174919061132b565b610488565b6040516101869190611312565b60405180910390f35b6101a960048036038101906101a4919061132b565b6104a7565b005b6101c560048036038101906101c091906110fd565b6104b7565b6040516101d29190611396565b60405180910390f35b6101f560048036038101906101f0919061132b565b6104e0565b6040516102029190611312565b60405180910390f35b610225600480360381019061022091906110fd565b6104ff565b6040516102329190611198565b60405180910390f35b610255600480360381019061025091906113d9565b6105a3565b6040516102629190611312565b60405180910390f35b61028560048036038101906102809190611417565b6105c0565b604051610292919061125f565b60405180910390f35b6102b560048036038101906102b09190611467565b610652565b6040516102c29190611312565b60405180910390f35b6102e560048036038101906102e0919061132b565b6106e0565b005b61030160048036038101906102fc91906114a5565b6106f0565b60405161030e9190611312565b60405180910390f35b606060035f8381526020019081526020015f205f01805461033790611536565b80601f016020809104026020016040519081016040528092919081815260200182805461036390611536565b80156103ae5780601f10610385576101008083540402835291602001916103ae565b820191905f5260205f20905b81548152906001019060200180831161039157829003601f168201915b50505050509050919050565b5f5f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8381526020019081526020015f2054905092915050565b5f7f0f632fb3000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916148061048157506104808261076a565b5b9050919050565b5f61049c6104946107d3565b8585856107da565b600190509392505050565b6104b28383836108cc565b505050565b5f60035f8381526020019081526020015f206002015f9054906101000a900460ff169050919050565b5f6104f46104ec6107d3565b85858561094d565b600190509392505050565b606060035f8381526020019081526020015f20600101805461052090611536565b80601f016020809104026020016040519081016040528092919081815260200182805461054c90611536565b80156105975780601f1061056e57610100808354040283529160200191610597565b820191905f5260205f20905b81548152906001019060200180831161057a57829003601f168201915b50505050509050919050565b5f6105b66105af6107d3565b8484610b25565b6001905092915050565b5f60025f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8381526020019081526020015f205490509392505050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff16905092915050565b6106eb838383610cfe565b505050565b5f5f6106fa6107d3565b90508073ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff161415801561073f575061073d8682610652565b155b156107515761075086828686610d7f565b5b61075d868686866107da565b6001915050949350505050565b5f7f01ffc9a7000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916149050919050565b5f33905090565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff160361084a575f6040517fa43520800000000000000000000000000000000000000000000000000000000081526004016108419190611575565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036108ba575f6040517fb8bbd6100000000000000000000000000000000000000000000000000000000081526004016108b19190611575565b60405180910390fd5b6108c684848484610e96565b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361093c575f6040517fb8bbd6100000000000000000000000000000000000000000000000000000000081526004016109339190611575565b60405180910390fd5b6109485f848484610e96565b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff16036109bd575f6040517fcc766a980000000000000000000000000000000000000000000000000000000081526004016109b49190611575565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610a2d575f6040517f6f65f465000000000000000000000000000000000000000000000000000000008152600401610a249190611575565b60405180910390fd5b8060025f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8481526020019081526020015f2081905550818373ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff167fb3fd5071835887567a0671151121894ddccc2842f1d10bedad13e0d17cace9a784604051610b17919061125f565b60405180910390a450505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610b95575f6040517fcc766a98000000000000000000000000000000000000000000000000000000008152600401610b8c9190611575565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610c05575f6040517f6f65f465000000000000000000000000000000000000000000000000000000008152600401610bfc9190611575565b60405180910390fd5b8060015f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff0219169083151502179055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fceb576d9f15e4e200fdb5096d64d5dfd667e16def20c1eefd14256d8e3faa26783604051610cf19190611312565b60405180910390a3505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610d6e575f6040517fa4352080000000000000000000000000000000000000000000000000000000008152600401610d659190611575565b60405180910390fd5b610d7a835f8484610e96565b505050565b5f610d8b8585856105c0565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff811015610e8f5781811015610e0057838183856040517f58a3fd5a000000000000000000000000000000000000000000000000000000008152600401610df7949392919061158e565b60405180910390fd5b81810360025f8773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8581526020019081526020015f20819055505b5050505050565b5f610e9f6107d3565b90505f73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff1614610fc2575f5f5f8773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8581526020019081526020015f2054905082811015610f6e57858184866040517fb1b4fec0000000000000000000000000000000000000000000000000000000008152600401610f65949392919061158e565b60405180910390fd5b8281035f5f8873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff161461105757815f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8581526020019081526020015f205f82825461104f91906115fe565b925050819055505b828473ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff167f1b3d7edb2e9c0b0e7c525b20aaaef0f5940d2ed71663c7d39266ecafac72885984866040516110b7929190611631565b60405180910390a45050505050565b5f5ffd5b5f819050919050565b6110dc816110ca565b81146110e6575f5ffd5b50565b5f813590506110f7816110d3565b92915050565b5f60208284031215611112576111116110c6565b5b5f61111f848285016110e9565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f61116a82611128565b6111748185611132565b9350611184818560208601611142565b61118d81611150565b840191505092915050565b5f6020820190508181035f8301526111b08184611160565b905092915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6111e1826111b8565b9050919050565b6111f1816111d7565b81146111fb575f5ffd5b50565b5f8135905061120c816111e8565b92915050565b5f5f60408385031215611228576112276110c6565b5b5f611235858286016111fe565b9250506020611246858286016110e9565b9150509250929050565b611259816110ca565b82525050565b5f6020820190506112725f830184611250565b92915050565b5f7fffffffff0000000000000000000000000000000000000000000000000000000082169050919050565b6112ac81611278565b81146112b6575f5ffd5b50565b5f813590506112c7816112a3565b92915050565b5f602082840312156112e2576112e16110c6565b5b5f6112ef848285016112b9565b91505092915050565b5f8115159050919050565b61130c816112f8565b82525050565b5f6020820190506113255f830184611303565b92915050565b5f5f5f60608486031215611342576113416110c6565b5b5f61134f868287016111fe565b9350506020611360868287016110e9565b9250506040611371868287016110e9565b9150509250925092565b5f60ff82169050919050565b6113908161137b565b82525050565b5f6020820190506113a95f830184611387565b92915050565b6113b8816112f8565b81146113c2575f5ffd5b50565b5f813590506113d3816113af565b92915050565b5f5f604083850312156113ef576113ee6110c6565b5b5f6113fc858286016111fe565b925050602061140d858286016113c5565b9150509250929050565b5f5f5f6060848603121561142e5761142d6110c6565b5b5f61143b868287016111fe565b935050602061144c868287016111fe565b925050604061145d868287016110e9565b9150509250925092565b5f5f6040838503121561147d5761147c6110c6565b5b5f61148a858286016111fe565b925050602061149b858286016111fe565b9150509250929050565b5f5f5f5f608085870312156114bd576114bc6110c6565b5b5f6114ca878288016111fe565b94505060206114db878288016111fe565b93505060406114ec878288016110e9565b92505060606114fd878288016110e9565b91505092959194509250565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061154d57607f821691505b6020821081036115605761155f611509565b5b50919050565b61156f816111d7565b82525050565b5f6020820190506115885f830184611566565b92915050565b5f6080820190506115a15f830187611566565b6115ae6020830186611250565b6115bb6040830185611250565b6115c86060830184611250565b95945050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611608826110ca565b9150611613836110ca565b925082820190508082111561162b5761162a6115d1565b5b92915050565b5f6040820190506116445f830185611566565b6116516020830184611250565b939250505056fea26469706673582212209537e0a5290687f730731a9834842827affca3bf1cd80ae1bb37f2189f03e8ee64736f6c63430008230033"

	// ERC20PermitFullByteCodeHex is OpenZeppelin's ERC20 with both EIP-2612 permit and DAI-style permit(holder, spender, nonce, expiry, allowed, v, r, s), plus mint/burn, constructor(name, symbol).
	ERC20PermitFullByteCodeHex = "0x610160604052348015610010575f5ffd5b50604051612937380380612937833981810160405281019061003291906103a2565b81806040518060400160405280600181526020017f31000000000000000000000000000000000000000000000000000000000000008152508484816003908161007b9190610639565b50806004908161008b9190610639565b5050506100a260058361014060201b90919060201c565b61012081815250506100be60068261014060201b90919060201c565b6101408181525050818051906020012060e08181525050808051906020012061010081815250504660a081815250506100fb61018d60201b60201c565b608081815250503073ffffffffffffffffffffffffffffffffffffffff1660c08173ffffffffffffffffffffffffffffffffffffffff168152505050505050506108ba565b5f6020835110156101615761015a836101e760201b60201c565b9050610187565b826101718361024c60201b60201c565b5f01908161017f9190610639565b5060ff5f1b90505b92915050565b5f7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60e0516101005146306040516020016101cc95949392919061076e565b60405160208183030381529060405280519060200120905090565b5f5f829050601f8151111561023357826040517f305a27a900000000000000000000000000000000000000000000000000000000815260040161022a9190610807565b60405180910390fd5b80518161023f90610854565b5f1c175f1b915050919050565b5f819050919050565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6102b48261026e565b810181811067ffffffffffffffff821117156102d3576102d261027e565b5b80604052505050565b5f6102e5610255565b90506102f182826102ab565b919050565b5f67ffffffffffffffff8211156103105761030f61027e565b5b6103198261026e565b9050602081019050919050565b8281835e5f83830152505050565b5f610346610341846102f6565b6102dc565b9050828152602081018484840111156103625761036161026a565b5b61036d848285610326565b509392505050565b5f82601f83011261038957610388610266565b5b8151610399848260208601610334565b91505092915050565b5f5f604083850312156103b8576103b761025e565b5b5f83015167ffffffffffffffff8111156103d5576103d4610262565b5b6103e185828601610375565b925050602083015167ffffffffffffffff81111561040257610401610262565b5b61040e85828601610375565b9150509250929050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061046657607f821691505b60208210810361047957610478610422565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026104db7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826104a0565b6104e586836104a0565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61052961052461051f846104fd565b610506565b6104fd565b9050919050565b5f819050919050565b6105428361050f565b61055661054e82610530565b8484546104ac565b825550505050565b5f5f905090565b61056d61055e565b610578818484610539565b505050565b5f5b8281101561059e576105935f828401610565565b60018101905061057f565b505050565b601f8211156105f157828211156105f0576105bd8161047f565b6105c683610491565b6105cf85610491565b60208610156105dc575f90505b8083016105eb8284038261057d565b505050505b5b505050565b5f82821c905092915050565b5f6106115f19846008026105f6565b1980831691505092915050565b5f6106298383610602565b9150826002028217905092915050565b61064282610418565b67ffffffffffffffff81111561065b5761065a61027e565b5b610665825461044f565b6106708282856105a3565b5f60209050601f8311600181146106a1575f841561068f578287015190505b610699858261061e565b865550610700565b601f1984166106af8661047f565b5f5b828110156106d6578489015182556001820191506020850194506020810190506106b1565b868310156106f357848901516106ef601f891682610602565b8355505b6001600288020188555050505b505050505050565b5f819050919050565b61071a81610708565b82525050565b610729816104fd565b82525050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6107588261072f565b9050919050565b6107688161074e565b82525050565b5f60a0820190506107815f830188610711565b61078e6020830187610711565b61079b6040830186610711565b6107a86060830185610720565b6107b5608083018461075f565b9695505050505050565b5f82825260208201905092915050565b5f6107d982610418565b6107e381856107bf565b93506107f3818560208601610326565b6107fc8161026e565b840191505092915050565b5f6020820190508181035f83015261081f81846107cf565b905092915050565b5f81519050919050565b5f819050602082019050919050565b5f61084b8251610708565b80915050919050565b5f61085e82610827565b8261086884610831565b905061087381610840565b925060208210156108b3576108ae7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff836020036008026104a0565b831692505b5050919050565b60805160a05160c05160e05161010051610120516101405161202c61090b5f395f610cb601525f610c7b01525f61128501525f61126401525f610afa01525f610b5001525f610b79015261202c5ff3fe608060405234801561000f575f5ffd5b50600436106100fe575f3560e01c80637ecebe00116100955780639dc29fac116100645780639dc29fac146102b4578063a9059cbb146102d0578063d505accf14610300578063dd62ed3e1461031c576100fe565b80637ecebe001461022657806384b0196e146102565780638fcbaf0c1461027a57806395d89b4114610296576100fe565b8063313ce567116100d1578063313ce5671461019e5780633644e515146101bc57806340c10f19146101da57806370a08231146101f6576100fe565b806306fdde0314610102578063095ea7b31461012057806318160ddd1461015057806323b872dd1461016e575b5f5ffd5b61010a61034c565b604051610117919061173b565b60405180910390f35b61013a600480360381019061013591906117ec565b6103dc565b6040516101479190611844565b60405180910390f35b6101586103fe565b604051610165919061186c565b60405180910390f35b61018860048036038101906101839190611885565b610407565b6040516101959190611844565b60405180910390f35b6101a6610435565b6040516101b391906118f0565b60405180910390f35b6101c461043d565b6040516101d19190611921565b60405180910390f35b6101f460048036038101906101ef91906117ec565b61044b565b005b610210600480360381019061020b919061193a565b610459565b60405161021d919061186c565b60405180910390f35b610240600480360381019061023b919061193a565b61049e565b60405161024d919061186c565b60405180910390f35b61025e6104af565b6040516102719796959493929190611a65565b60405180910390f35b610294600480360381019061028f9190611b65565b610554565b005b61029e6106d4565b6040516102ab919061173b565b60405180910390f35b6102ce60048036038101906102c991906117ec565b610764565b005b6102ea60048036038101906102e591906117ec565b610772565b6040516102f79190611844565b60405180910390f35b61031a60048036038101906103159190611c16565b610794565b005b61033660048036038101906103319190611cb3565b6108d9565b604051610343919061186c565b60405180910390f35b60606003805461035b90611d1e565b80601f016020809104026020016040519081016040528092919081815260200182805461038790611d1e565b80156103d25780601f106103a9576101008083540402835291602001916103d2565b820191905f5260205f20905b8154815290600101906020018083116103b557829003601f168201915b5050505050905090565b5f5f6103e661095b565b90506103f3818585610962565b600191505092915050565b5f600254905090565b5f5f61041161095b565b905061041e858285610974565b610429858585610a07565b60019150509392505050565b5f6012905090565b5f610446610af7565b905090565b6104558282610bad565b5050565b5f5f5f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b5f6104a882610c2c565b9050919050565b5f6060805f5f5f60606104c0610c72565b6104c8610cad565b46305f5f1b5f67ffffffffffffffff8111156104e7576104e6611d4e565b5b6040519080825280602002602001820160405280156105155781602001602082028036833780820191505090505b507f0f00000000000000000000000000000000000000000000000000000000000000959493929190965096509650965096509650965090919293949596565b5f851415801561056357508442115b156105a557846040517f9e96b26f00000000000000000000000000000000000000000000000000000000815260040161059c919061186c565b60405180910390fd5b5f7fea2aa0a1be11a07ed86d755c93467f4f82362b452371d1ba94d1715123511acb89898989896040516020016105e196959493929190611d7b565b6040516020818303038152906040528051906020012090505f61060382610ce8565b90505f61061282878787610d01565b90508a73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161461068657808b6040517f7ba5ffb500000000000000000000000000000000000000000000000000000000815260040161067d929190611dda565b60405180910390fd5b6106908b8a610d2f565b6106c78b8b896106a0575f6106c2565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff5b610962565b5050505050505050505050565b6060600480546106e390611d1e565b80601f016020809104026020016040519081016040528092919081815260200182805461070f90611d1e565b801561075a5780601f106107315761010080835404028352916020019161075a565b820191905f5260205f20905b81548152906001019060200180831161073d57829003601f168201915b5050505050905090565b61076e8282610d86565b5050565b5f5f61077c61095b565b9050610789818585610a07565b600191505092915050565b834211156107d957836040517f627913020000000000000000000000000000000000000000000000000000000081526004016107d0919061186c565b60405180910390fd5b5f7f6e71edae12b1b97f4d1f60370fef10105fa2faae0126114a169c64845d6126c98888886108078c610e05565b8960405160200161081d96959493929190611e01565b6040516020818303038152906040528051906020012090505f61083f82610ce8565b90505f61084e82878787610d01565b90508973ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16146108c257808a6040517f4b800e460000000000000000000000000000000000000000000000000000000081526004016108b9929190611dda565b60405180910390fd5b6108cd8a8a8a610962565b50505050505050505050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f33905090565b61096f8383836001610e58565b505050565b5f61097f84846108d9565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff811015610a0157818110156109f2578281836040517ffb8f41b20000000000000000000000000000000000000000000000000000000081526004016109e993929190611e60565b60405180910390fd5b610a0084848484035f610e58565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610a77575f6040517f96c6fd1e000000000000000000000000000000000000000000000000000000008152600401610a6e9190611e95565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610ae7575f6040517fec442f05000000000000000000000000000000000000000000000000000000008152600401610ade9190611e95565b60405180910390fd5b610af2838383611027565b505050565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff16148015610b7257507f000000000000000000000000000000000000000000000000000000000000000046145b15610b9f577f00000000000000000000000000000000000000000000000000000000000000009050610baa565b610ba7611240565b90505b90565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610c1d575f6040517fec442f05000000000000000000000000000000000000000000000000000000008152600401610c149190611e95565b60405180910390fd5b610c285f8383611027565b5050565b5f60075f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b6060610ca860057f00000000000000000000000000000000000000000000000000000000000000006112d590919063ffffffff16565b905090565b6060610ce360067f00000000000000000000000000000000000000000000000000000000000000006112d590919063ffffffff16565b905090565b5f610cfa610cf4610af7565b83611382565b9050919050565b5f5f5f5f610d11888888886113c2565b925092509250610d2182826114a9565b829350505050949350505050565b5f610d3983610e05565b9050808214610d815782816040517f752d88c0000000000000000000000000000000000000000000000000000000008152600401610d78929190611eae565b60405180910390fd5b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610df6575f6040517f96c6fd1e000000000000000000000000000000000000000000000000000000008152600401610ded9190611e95565b60405180910390fd5b610e01825f83611027565b5050565b5f60075f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f815480929190600101919050559050919050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610ec8575f6040517fe602df05000000000000000000000000000000000000000000000000000000008152600401610ebf9190611e95565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610f38575f6040517f94280d62000000000000000000000000000000000000000000000000000000008152600401610f2f9190611e95565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508015611021578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92584604051611018919061186c565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603611077578060025f82825461106b9190611f02565b92505081905550611145565b5f5f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905081811015611100578381836040517fe450d38c0000000000000000000000000000000000000000000000000000000081526004016110f793929190611e60565b60405180910390fd5b8181035f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361118c578060025f82825403925050819055506111d6565b805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051611233919061186c565b60405180910390a3505050565b5f7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f7f00000000000000000000000000000000000000000000000000000000000000007f000000000000000000000000000000000000000000000000000000000000000046306040516020016112ba959493929190611f35565b60405160208183030381529060405280519060200120905090565b606060ff5f1b83146112f1576112ea8361160b565b905061137c565b8180546112fd90611d1e565b80601f016020809104026020016040519081016040528092919081815260200182805461132990611d1e565b80156113745780601f1061134b57610100808354040283529160200191611374565b820191905f5260205f20905b81548152906001019060200180831161135757829003601f168201915b505050505090505b92915050565b5f6040517f190100000000000000000000000000000000000000000000000000000000000081528360028201528260228201526042812091505092915050565b5f5f5f7f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0845f1c11156113fe575f60038592509250925061149f565b5f6001888888886040515f81526020016040526040516114219493929190611f86565b6020604051602081039080840390855afa158015611441573d5f5f3e3d5ffd5b5050506020604051035190505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603611492575f60015f5f1b9350935093505061149f565b805f5f5f1b935093509350505b9450945094915050565b5f60038111156114bc576114bb611fc9565b5b8260038111156114cf576114ce611fc9565b5b031561160757600160038111156114e9576114e8611fc9565b5b8260038111156114fc576114fb611fc9565b5b03611533576040517ff645eedf00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6002600381111561154757611546611fc9565b5b82600381111561155a57611559611fc9565b5b0361159e57805f1c6040517ffce698f7000000000000000000000000000000000000000000000000000000008152600401611595919061186c565b60405180910390fd5b6003808111156115b1576115b0611fc9565b5b8260038111156115c4576115c3611fc9565b5b0361160657806040517fd78bce0c0000000000000000000000000000000000000000000000000000000081526004016115fd9190611921565b60405180910390fd5b5b5050565b60605f6116178361167d565b90505f602067ffffffffffffffff81111561163557611634611d4e565b5b6040519080825280601f01601f1916602001820160405280156116675781602001600182028036833780820191505090505b5090508181528360208201528092505050919050565b5f5f60ff835f1c169050601f8111156116c2576040517fb3512b0c00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b80915050919050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f61170d826116cb565b61171781856116d5565b93506117278185602086016116e5565b611730816116f3565b840191505092915050565b5f6020820190508181035f8301526117538184611703565b905092915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6117888261175f565b9050919050565b6117988161177e565b81146117a2575f5ffd5b50565b5f813590506117b38161178f565b92915050565b5f819050919050565b6117cb816117b9565b81146117d5575f5ffd5b50565b5f813590506117e6816117c2565b92915050565b5f5f604083850312156118025761180161175b565b5b5f61180f858286016117a5565b9250506020611820858286016117d8565b9150509250929050565b5f8115159050919050565b61183e8161182a565b82525050565b5f6020820190506118575f830184611835565b92915050565b611866816117b9565b82525050565b5f60208201905061187f5f83018461185d565b92915050565b5f5f5f6060848603121561189c5761189b61175b565b5b5f6118a9868287016117a5565b93505060206118ba868287016117a5565b92505060406118cb868287016117d8565b9150509250925092565b5f60ff82169050919050565b6118ea816118d5565b82525050565b5f6020820190506119035f8301846118e1565b92915050565b5f819050919050565b61191b81611909565b82525050565b5f6020820190506119345f830184611912565b92915050565b5f6020828403121561194f5761194e61175b565b5b5f61195c848285016117a5565b91505092915050565b5f7fff0000000000000000000000000000000000000000000000000000000000000082169050919050565b61199981611965565b82525050565b6119a88161177e565b82525050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b6119e0816117b9565b82525050565b5f6119f183836119d7565b60208301905092915050565b5f602082019050919050565b5f611a13826119ae565b611a1d81856119b8565b9350611a28836119c8565b805f5b83811015611a58578151611a3f88826119e6565b9750611a4a836119fd565b925050600181019050611a2b565b5085935050505092915050565b5f60e082019050611a785f83018a611990565b8181036020830152611a8a8189611703565b90508181036040830152611a9e8188611703565b9050611aad606083018761185d565b611aba608083018661199f565b611ac760a0830185611912565b81810360c0830152611ad98184611a09565b905098975050505050505050565b611af08161182a565b8114611afa575f5ffd5b50565b5f81359050611b0b81611ae7565b92915050565b611b1a816118d5565b8114611b24575f5ffd5b50565b5f81359050611b3581611b11565b92915050565b611b4481611909565b8114611b4e575f5ffd5b50565b5f81359050611b5f81611b3b565b92915050565b5f5f5f5f5f5f5f5f610100898b031215611b8257611b8161175b565b5b5f611b8f8b828c016117a5565b9850506020611ba08b828c016117a5565b9750506040611bb18b828c016117d8565b9650506060611bc28b828c016117d8565b9550506080611bd38b828c01611afd565b94505060a0611be48b828c01611b27565b93505060c0611bf58b828c01611b51565b92505060e0611c068b828c01611b51565b9150509295985092959890939650565b5f5f5f5f5f5f5f60e0888a031215611c3157611c3061175b565b5b5f611c3e8a828b016117a5565b9750506020611c4f8a828b016117a5565b9650506040611c608a828b016117d8565b9550506060611c718a828b016117d8565b9450506080611c828a828b01611b27565b93505060a0611c938a828b01611b51565b92505060c0611ca48a828b01611b51565b91505092959891949750929550565b5f5f60408385031215611cc957611cc861175b565b5b5f611cd6858286016117a5565b9250506020611ce7858286016117a5565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680611d3557607f821691505b602082108103611d4857611d47611cf1565b5b50919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b5f60c082019050611d8e5f830189611912565b611d9b602083018861199f565b611da8604083018761199f565b611db5606083018661185d565b611dc2608083018561185d565b611dcf60a0830184611835565b979650505050505050565b5f604082019050611ded5f83018561199f565b611dfa602083018461199f565b9392505050565b5f60c082019050611e145f830189611912565b611e21602083018861199f565b611e2e604083018761199f565b611e3b606083018661185d565b611e48608083018561185d565b611e5560a083018461185d565b979650505050505050565b5f606082019050611e735f83018661199f565b611e80602083018561185d565b611e8d604083018461185d565b949350505050565b5f602082019050611ea85f83018461199f565b92915050565b5f604082019050611ec15f83018561199f565b611ece602083018461185d565b9392505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611f0c826117b9565b9150611f17836117b9565b9250828201905080821115611f2f57611f2e611ed5565b5b92915050565b5f60a082019050611f485f830188611912565b611f556020830187611912565b611f626040830186611912565b611f6f606083018561185d565b611f7c608083018461199f565b9695505050505050565b5f608082019050611f995f830187611912565b611fa660208301866118e1565b611fb36040830185611912565b611fc06060830184611912565b95945050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffdfea2646970667358221220faca0f7cea2c9328299ad2f0bf911d29a083c468837794bcb5cfcdd6dd2bdc1e64736f6c63430008230033"
)
