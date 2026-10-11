package erc20

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// erc20ByteCodeHex is OpenZeppelin's ERC20 plus mint/burn, constructor(name, symbol, decimals).
const erc20ByteCodeHex = "0x60a060405234801561000f575f5ffd5b506040516114e73803806114e7833981810160405281019061003191906101ee565b828281600390816100429190610497565b5080600490816100529190610497565b5050508060ff1660808160ff1681525050505050610566565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6100ca82610084565b810181811067ffffffffffffffff821117156100e9576100e8610094565b5b80604052505050565b5f6100fb61006b565b905061010782826100c1565b919050565b5f67ffffffffffffffff82111561012657610125610094565b5b61012f82610084565b9050602081019050919050565b8281835e5f83830152505050565b5f61015c6101578461010c565b6100f2565b90508281526020810184848401111561017857610177610080565b5b61018384828561013c565b509392505050565b5f82601f83011261019f5761019e61007c565b5b81516101af84826020860161014a565b91505092915050565b5f60ff82169050919050565b6101cd816101b8565b81146101d7575f5ffd5b50565b5f815190506101e8816101c4565b92915050565b5f5f5f6060848603121561020557610204610074565b5b5f84015167ffffffffffffffff81111561022257610221610078565b5b61022e8682870161018b565b935050602084015167ffffffffffffffff81111561024f5761024e610078565b5b61025b8682870161018b565b925050604061026c868287016101da565b9150509250925092565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f60028204905060018216806102c457607f821691505b6020821081036102d7576102d6610280565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026103397fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826102fe565b61034386836102fe565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61038761038261037d8461035b565b610364565b61035b565b9050919050565b5f819050919050565b6103a08361036d565b6103b46103ac8261038e565b84845461030a565b825550505050565b5f5f905090565b6103cb6103bc565b6103d6818484610397565b505050565b5f5b828110156103fc576103f15f8284016103c3565b6001810190506103dd565b505050565b601f82111561044f578282111561044e5761041b816102dd565b610424836102ef565b61042d856102ef565b602086101561043a575f90505b808301610449828403826103db565b505050505b5b505050565b5f82821c905092915050565b5f61046f5f1984600802610454565b1980831691505092915050565b5f6104878383610460565b9150826002028217905092915050565b6104a082610276565b67ffffffffffffffff8111156104b9576104b8610094565b5b6104c382546102ad565b6104ce828285610401565b5f60209050601f8311600181146104ff575f84156104ed578287015190505b6104f7858261047c565b86555061055e565b601f19841661050d866102dd565b5f5b828110156105345784890151825560018201915060208501945060208101905061050f565b86831015610551578489015161054d601f891682610460565b8355505b6001600288020188555050505b505050505050565b608051610f6961057e5f395f6103370152610f695ff3fe608060405234801561000f575f5ffd5b50600436106100a7575f3560e01c806340c10f191161006f57806340c10f191461016557806370a082311461018157806395d89b41146101b15780639dc29fac146101cf578063a9059cbb146101eb578063dd62ed3e1461021b576100a7565b806306fdde03146100ab578063095ea7b3146100c957806318160ddd146100f957806323b872dd14610117578063313ce56714610147575b5f5ffd5b6100b361024b565b6040516100c09190610be2565b60405180910390f35b6100e360048036038101906100de9190610c93565b6102db565b6040516100f09190610ceb565b60405180910390f35b6101016102fd565b60405161010e9190610d13565b60405180910390f35b610131600480360381019061012c9190610d2c565b610306565b60405161013e9190610ceb565b60405180910390f35b61014f610334565b60405161015c9190610d97565b60405180910390f35b61017f600480360381019061017a9190610c93565b61035b565b005b61019b60048036038101906101969190610db0565b610369565b6040516101a89190610d13565b60405180910390f35b6101b96103ae565b6040516101c69190610be2565b60405180910390f35b6101e960048036038101906101e49190610c93565b61043e565b005b61020560048036038101906102009190610c93565b61044c565b6040516102129190610ceb565b60405180910390f35b61023560048036038101906102309190610ddb565b61046e565b6040516102429190610d13565b60405180910390f35b60606003805461025a90610e46565b80601f016020809104026020016040519081016040528092919081815260200182805461028690610e46565b80156102d15780601f106102a8576101008083540402835291602001916102d1565b820191905f5260205f20905b8154815290600101906020018083116102b457829003601f168201915b5050505050905090565b5f5f6102e56104f0565b90506102f28185856104f7565b600191505092915050565b5f600254905090565b5f5f6103106104f0565b905061031d858285610509565b61032885858561059c565b60019150509392505050565b5f7f0000000000000000000000000000000000000000000000000000000000000000905090565b610365828261068c565b5050565b5f5f5f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b6060600480546103bd90610e46565b80601f01602080910402602001604051908101604052809291908181526020018280546103e990610e46565b80156104345780601f1061040b57610100808354040283529160200191610434565b820191905f5260205f20905b81548152906001019060200180831161041757829003601f168201915b5050505050905090565b610448828261070b565b5050565b5f5f6104566104f0565b905061046381858561059c565b600191505092915050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f33905090565b610504838383600161078a565b505050565b5f610514848461046e565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8110156105965781811015610587578281836040517ffb8f41b200000000000000000000000000000000000000000000000000000000815260040161057e93929190610e85565b60405180910390fd5b61059584848484035f61078a565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361060c575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016106039190610eba565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361067c575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016106739190610eba565b60405180910390fd5b610687838383610959565b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16036106fc575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016106f39190610eba565b60405180910390fd5b6107075f8383610959565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361077b575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016107729190610eba565b60405180910390fd5b610786825f83610959565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff16036107fa575f6040517fe602df050000000000000000000000000000000000000000000000000000000081526004016107f19190610eba565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361086a575f6040517f94280d620000000000000000000000000000000000000000000000000000000081526004016108619190610eba565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508015610953578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9258460405161094a9190610d13565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036109a9578060025f82825461099d9190610f00565b92505081905550610a77565b5f5f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905081811015610a32578381836040517fe450d38c000000000000000000000000000000000000000000000000000000008152600401610a2993929190610e85565b60405180910390fd5b8181035f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610abe578060025f8282540392505081905550610b08565b805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051610b659190610d13565b60405180910390a3505050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f610bb482610b72565b610bbe8185610b7c565b9350610bce818560208601610b8c565b610bd781610b9a565b840191505092915050565b5f6020820190508181035f830152610bfa8184610baa565b905092915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610c2f82610c06565b9050919050565b610c3f81610c25565b8114610c49575f5ffd5b50565b5f81359050610c5a81610c36565b92915050565b5f819050919050565b610c7281610c60565b8114610c7c575f5ffd5b50565b5f81359050610c8d81610c69565b92915050565b5f5f60408385031215610ca957610ca8610c02565b5b5f610cb685828601610c4c565b9250506020610cc785828601610c7f565b9150509250929050565b5f8115159050919050565b610ce581610cd1565b82525050565b5f602082019050610cfe5f830184610cdc565b92915050565b610d0d81610c60565b82525050565b5f602082019050610d265f830184610d04565b92915050565b5f5f5f60608486031215610d4357610d42610c02565b5b5f610d5086828701610c4c565b9350506020610d6186828701610c4c565b9250506040610d7286828701610c7f565b9150509250925092565b5f60ff82169050919050565b610d9181610d7c565b82525050565b5f602082019050610daa5f830184610d88565b92915050565b5f60208284031215610dc557610dc4610c02565b5b5f610dd284828501610c4c565b91505092915050565b5f5f60408385031215610df157610df0610c02565b5b5f610dfe85828601610c4c565b9250506020610e0f85828601610c4c565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680610e5d57607f821691505b602082108103610e7057610e6f610e19565b5b50919050565b610e7f81610c25565b82525050565b5f606082019050610e985f830186610e76565b610ea56020830185610d04565b610eb26040830184610d04565b949350505050565b5f602082019050610ecd5f830184610e76565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f610f0a82610c60565b9150610f1583610c60565b9250828201905080821115610f2d57610f2c610ed3565b5b9291505056fea264697066735822122049c38efc930a01957bf084f9c67bdda53452b282675854f0c07be3a52f60fb5e64736f6c63430008230033"

type ERC20Suite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	ToHex         string
	MulticallHex  string
	Value         string
	MintValue     string
	NativeSymbol  string
	TokenCount    int
	ByteCode      string
	Constructor   *abi.Function
	Args          []any
	Salt          *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	deployer  *deploy.NicksDeployer
	from      *types.Address
	tokens    []*types.Address
	to        *types.Address
	multicall *types.Address
	value     *big.Int
	mintValue *big.Int

	erc20 *ERC20

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestERC20Suite(t *testing.T) {
	suite.Run(t, &ERC20Suite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		ToHex:         "",
		MulticallHex:  multicall3.DefaultMulticall3Address.String(),
		Value:         "100000",
		MintValue:     "1000000000000000000000000",
		NativeSymbol:  "",
		TokenCount:    5,
		ByteCode:      erc20ByteCodeHex,
		Constructor:   abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Args:          []any{"evmkit test ERC20", "EVK20", uint8(18)},
		Salt:          core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *ERC20Suite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.erc20 = New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.to, err = types.NewAddressFromHex(s.ToHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	value, ok := new(big.Int).SetString(s.Value, 10)
	s.Require().True(ok)
	s.value = value

	mintValue, ok := new(big.Int).SetString(s.MintValue, 10)
	s.Require().True(ok)
	s.mintValue = mintValue

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	mintFn := abi.NewFunction("mint", abi.Types{abi.Address, abi.Uint256}, nil)
	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	s.tokens = make([]*types.Address, s.TokenCount)
	for i := 0; i < s.TokenCount; i++ {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, s.Args),
			salt,
		)

		unsigned, err := s.deployer.CreateDeployment(s.ctx, tx.NewDynamicFeeBuilder(), s.from, deploy.DefaultNicksMethodFactory, big.NewInt(0), deployment)
		s.Require().NoError(err)

		s.Require().NoError(unsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsigned.MaxFee())
		s.mu.Unlock()

		hash, err := s.deployer.Deploy(s.ctx, unsigned)
		s.Require().NoError(err)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 500, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

		token, err := s.deployer.DeployAddress(deploy.DefaultNicksMethodFactory, deployment)
		s.Require().NoError(err)
		s.tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 10*time.Second, 500*time.Millisecond, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %d/%d %s (nonce %d, tx %s)", i+1, s.TokenCount, token, unsigned.AddressNonce(), hash)

		for _, recipient := range []*types.Address{s.from, s.to} {
			calldata, err := mintFn.EncodeCall(recipient, s.mintValue)
			s.Require().NoError(err)

			var unsignedMint core.Transaction
			s.Require().Eventually(func() bool {
				u, err := tx.NewDynamicFeeBuilder().Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
				if err != nil || u.Gas() < 40000 {
					return false
				}
				unsignedMint = u
				return true
			}, 10*time.Second, 500*time.Millisecond, "mint gas estimate looked like a no-code call")
			s.Require().NoError(unsignedMint.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())
			s.mu.Unlock()

			raw, err := unsignedMint.EncodeRLP()
			s.Require().NoError(err)
			mintHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)

			mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, 500, 15*time.Second)
			s.Require().NoError(err)
			s.Require().True(mintReceipt.Status, "mint transaction reverted")

			s.T().Logf("SetupSuite: minted %s of %s to %s (tx %s)", s.mintValue, token, recipient, mintHash)
		}
	}
}

func (s *ERC20Suite) TearDownSuite() {
	if s.tokens != nil {
		burnFn := abi.NewFunction("burn", abi.Types{abi.Address, abi.Uint256}, nil)

		for _, token := range s.tokens {
			for _, holder := range []*types.Address{s.from, s.to} {
				balance, err := s.erc20.BalanceOf(s.ctx, token, holder, rpc.BlockTagLatest)
				s.Require().NoError(err)
				if balance.Sign() == 0 {
					continue
				}

				calldata, err := burnFn.EncodeCall(holder, balance)
				s.Require().NoError(err)

				unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
				s.Require().NoError(err)
				s.Require().NoError(unsignedBurn.Sign(s.signer))
				s.mu.Lock()
				s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
				s.mu.Unlock()

				raw, err := unsignedBurn.EncodeRLP()
				s.Require().NoError(err)
				burnHash, err := tx.Broadcast(s.ctx, s.client, raw)
				s.Require().NoError(err)

				burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, 500, 15*time.Second)
				s.Require().NoError(err)
				s.Require().True(burnReceipt.Status, "burn transaction reverted")

				s.T().Logf("TearDownSuite: burned %s of %s from %s (tx %s)", balance, token, holder, burnHash)
			}
		}
	}

	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("ERC20Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC20Suite) TestReadMethods() {
	for _, token := range s.tokens {
		totalSupply, err := s.erc20.TotalSupply(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)

		balance, err := s.erc20.BalanceOf(s.ctx, token, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		allowance, err := s.erc20.Allowance(s.ctx, token, s.from, s.to, rpc.BlockTagLatest)
		s.Require().NoError(err)

		name, err := s.erc20.Name(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)

		symbol, err := s.erc20.Symbol(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)

		decimals, err := s.erc20.Decimals(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.T().Logf("ERC20 %s TestReadMethods Result", token)
		s.T().Logf("\tName: %s", name)
		s.T().Logf("\tSymbol: %s", symbol)
		s.T().Logf("\tDecimals: %d", decimals)
		s.T().Logf("\tTotalSupply: %s", totalSupply)
		s.T().Logf("\tBalance: %s", balance)
		s.T().Logf("\tAllowance")
		s.T().Logf("\t\tOwner: %s", s.from)
		s.T().Logf("\t\tSpender: %s", s.to)
		s.T().Logf("\t\tValue: %s", allowance)
	}
}

func (s *ERC20Suite) TestMetadata() {
	for _, token := range s.tokens {
		m, err := s.erc20.Metadata(s.ctx, s.multicall, token, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		name, err := s.erc20.Name(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		symbol, err := s.erc20.Symbol(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		decimals, err := s.erc20.Decimals(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		totalSupply, err := s.erc20.TotalSupply(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().Equal(name, m.Name)
		s.Require().Equal(symbol, m.Symbol)
		s.Require().Equal(decimals, m.Decimals)
		s.Require().Zero(totalSupply.Cmp(m.TotalSupply))

		s.T().Logf("ERC20 %s TestMetadata Result", token)
		s.T().Logf("\tName: %s", m.Name)
		s.T().Logf("\tSymbol: %s", m.Symbol)
		s.T().Logf("\tDecimals: %d", m.Decimals)
		s.T().Logf("\tTotalSupply: %s", m.TotalSupply)
	}
}

func (s *ERC20Suite) TestMetadataAllowFailure() {
	m, err := s.erc20.Metadata(s.ctx, s.multicall, s.multicall, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Empty(m.Name)
	s.Require().Empty(m.Symbol)
	s.Require().Zero(m.Decimals)
	s.Require().Nil(m.TotalSupply)

	s.T().Logf("ERC20 %s TestMetadataAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", m.Name)
	s.T().Logf("\tSymbol: %q", m.Symbol)
	s.T().Logf("\tDecimals: %d", m.Decimals)
	s.T().Logf("\tTotalSupply: %v", m.TotalSupply)
}

func (s *ERC20Suite) TestMetadataWithBalance() {
	for _, token := range s.tokens {
		mb, err := s.erc20.MetadataWithBalance(s.ctx, s.multicall, token, s.from, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		name, err := s.erc20.Name(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		symbol, err := s.erc20.Symbol(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		decimals, err := s.erc20.Decimals(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		totalSupply, err := s.erc20.TotalSupply(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		balance, err := s.erc20.BalanceOf(s.ctx, token, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().Equal(name, mb.Name)
		s.Require().Equal(symbol, mb.Symbol)
		s.Require().Equal(decimals, mb.Decimals)
		s.Require().Zero(totalSupply.Cmp(mb.TotalSupply))
		s.Require().Zero(balance.Cmp(mb.Balance))

		s.T().Logf("ERC20 %s TestMetadataWithBalance Result", token)
		s.T().Logf("\tName: %s", mb.Name)
		s.T().Logf("\tSymbol: %s", mb.Symbol)
		s.T().Logf("\tDecimals: %d", mb.Decimals)
		s.T().Logf("\tTotalSupply: %s", mb.TotalSupply)
		s.T().Logf("\tBalance: %s", mb.Balance)
	}
}

func (s *ERC20Suite) TestMetadataWithBalanceAllowFailure() {
	mb, err := s.erc20.MetadataWithBalance(s.ctx, s.multicall, s.multicall, s.from, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Empty(mb.Name)
	s.Require().Empty(mb.Symbol)
	s.Require().Zero(mb.Decimals)
	s.Require().Nil(mb.TotalSupply)
	s.Require().Nil(mb.Balance)

	s.T().Logf("ERC20 %s TestMetadataWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", mb.Name)
	s.T().Logf("\tSymbol: %q", mb.Symbol)
	s.T().Logf("\tDecimals: %d", mb.Decimals)
	s.T().Logf("\tTotalSupply: %v", mb.TotalSupply)
	s.T().Logf("\tBalance: %v", mb.Balance)
}

func (s *ERC20Suite) TestAllowanceWithBalance() {
	for _, token := range s.tokens {
		awb, err := s.erc20.AllowanceWithBalance(s.ctx, s.multicall, token, s.from, s.to, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		allowance, err := s.erc20.Allowance(s.ctx, token, s.from, s.to, rpc.BlockTagLatest)
		s.Require().NoError(err)
		ownerBalance, err := s.erc20.BalanceOf(s.ctx, token, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)
		spenderBalance, err := s.erc20.BalanceOf(s.ctx, token, s.to, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().Zero(allowance.Cmp(awb.Value))
		s.Require().Zero(ownerBalance.Cmp(awb.OwnerBalance))
		s.Require().Zero(spenderBalance.Cmp(awb.SpenderBalance))

		s.T().Logf("ERC20 %s TestAllowanceWithBalance Result", token)
		s.T().Logf("\tAllowance")
		s.T().Logf("\t\tOwner: %s", s.from)
		s.T().Logf("\t\tSpender: %s", s.to)
		s.T().Logf("\t\tValue: %s", awb.Value)
		s.T().Logf("\tOwnerBalance: %s", awb.OwnerBalance)
		s.T().Logf("\tSpenderBalance: %s", awb.SpenderBalance)
	}
}

func (s *ERC20Suite) TestAllowanceWithBalanceAllowFailure() {
	awb, err := s.erc20.AllowanceWithBalance(s.ctx, s.multicall, s.multicall, s.from, s.to, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Nil(awb.Value)
	s.Require().Nil(awb.OwnerBalance)
	s.Require().Nil(awb.SpenderBalance)

	s.T().Logf("ERC20 %s TestAllowanceWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tAllowance")
	s.T().Logf("\t\tOwner: %s", s.from)
	s.T().Logf("\t\tSpender: %s", s.to)
	s.T().Logf("\t\tValue: %v", awb.Value)
	s.T().Logf("\tOwnerBalance: %v", awb.OwnerBalance)
	s.T().Logf("\tSpenderBalance: %v", awb.SpenderBalance)
}

func (s *ERC20Suite) TestTokenBalances() {
	addresses := []*types.Address{s.from, s.to}

	for _, token := range s.tokens {
		tb, err := s.erc20.TokenBalances(s.ctx, s.multicall, token, addresses, false, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(tb.Token.Equal(token))
		s.Require().Len(tb.Addresses, len(addresses))

		s.T().Logf("ERC20 %s TestTokenBalances Result", token)
		for i, addr := range addresses {
			balance, err := s.erc20.BalanceOf(s.ctx, token, addr, rpc.BlockTagLatest)
			s.Require().NoError(err)

			s.Require().True(tb.Addresses[i].Address.Equal(addr))
			s.Require().Zero(balance.Cmp(tb.Addresses[i].Amount))
			s.T().Logf("\t%s: %s", addr, tb.Addresses[i].Amount)
		}
	}
}

func (s *ERC20Suite) TestAddressBalances() {
	for _, addr := range []*types.Address{s.from, s.to} {
		ab, err := s.erc20.AddressBalances(s.ctx, s.multicall, addr, s.tokens, false, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(ab.Address.Equal(addr))
		s.Require().Len(ab.Tokens, len(s.tokens))

		s.T().Logf("ERC20 %s TestAddressBalances Result", addr)
		for i, token := range s.tokens {
			balance, err := s.erc20.BalanceOf(s.ctx, token, addr, rpc.BlockTagLatest)
			s.Require().NoError(err)

			s.Require().True(ab.Tokens[i].Token.Equal(token))
			s.Require().Zero(balance.Cmp(ab.Tokens[i].Amount))
			s.T().Logf("\t%s: %s", token, ab.Tokens[i].Amount)
		}
	}
}

func (s *ERC20Suite) TestBalancePairs() {
	for _, addr := range []*types.Address{s.from, s.to} {
		addresses := make([]*types.Address, len(s.tokens))
		for i := range addresses {
			addresses[i] = addr
		}

		pairs, err := s.erc20.BalancePairs(s.ctx, s.multicall, s.tokens, addresses, false, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Len(pairs, len(s.tokens))

		s.T().Logf("ERC20 TestBalancePairs Result")
		for i, pair := range pairs {
			balance, err := s.erc20.BalanceOf(s.ctx, s.tokens[i], addresses[i], rpc.BlockTagLatest)
			s.Require().NoError(err)

			s.Require().True(pair.Token.Equal(s.tokens[i]))
			s.Require().True(pair.Address.Equal(addresses[i]))
			s.Require().Zero(balance.Cmp(pair.Amount))
			s.T().Logf("\t%s / %s: %s", pair.Token, pair.Address, pair.Amount)
		}
	}
}

func (s *ERC20Suite) TestTransfer() {
	befores := make([]*big.Int, len(s.tokens))
	for i, token := range s.tokens {
		before, err := s.erc20.BalanceOf(s.ctx, token, s.to, rpc.BlockTagLatest)
		s.Require().NoError(err)
		befores[i] = before

		ok, err := s.erc20.CallTransfer(s.ctx, token, s.from, s.to, s.value, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(ok)
	}

	first, err := s.erc20.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), s.tokens[0], s.from, s.to, s.value)
	s.Require().NoError(err)
	baseNonce := first.AddressNonce()

	unsigned := make([]core.Transaction, len(s.tokens))
	unsigned[0] = first
	for i := 1; i < len(s.tokens); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(baseNonce + uint64(i))

		u, err := s.erc20.CreateTransfer(s.ctx, builder, s.tokens[i], s.from, s.to, s.value)
		s.Require().NoError(err)
		unsigned[i] = u
	}

	hashes := make([]*types.Hash, len(s.tokens))
	for i, u := range unsigned {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc20.Transfer(s.ctx, u)
		s.Require().NoError(err)
		hashes[i] = hash
	}

	for i, token := range s.tokens {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		transfers, err := ExtractTransfers(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].Contract.Equal(token))
		s.Require().True(transfers[0].From.Equal(s.from))
		s.Require().True(transfers[0].To.Equal(s.to))
		s.Require().Zero(s.value.Cmp(transfers[0].Value))

		want := new(big.Int).Add(befores[i], s.value)
		var after *big.Int
		s.Require().Eventually(func() bool {
			after, err = s.erc20.BalanceOf(s.ctx, token, s.to, rpc.BlockTagLatest)
			return err == nil && want.Cmp(after) == 0
		}, 10*time.Second, 500*time.Millisecond, "balance not visible after transfer")

		s.T().Logf("ERC20 %s TestTransfer Result", token)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tNonce: %d", unsigned[i].AddressNonce())
		s.T().Logf("\tGas: %d", unsigned[i].Gas())
		s.T().Logf("\tMaxFee: %s", unsigned[i].MaxFee())
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tBalance")
		s.T().Logf("\t\tBefore: %s", befores[i])
		s.T().Logf("\t\tAfter: %s", after)
	}
}

func (s *ERC20Suite) TestTransferInsufficientBalance() {
	for _, token := range s.tokens {
		decimals, err := s.erc20.Decimals(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)

		scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
		huge := new(big.Int).Mul(big.NewInt(100_000_000), scale)

		ok, err := s.erc20.CallTransfer(s.ctx, token, s.from, s.to, huge, rpc.BlockTagLatest)
		s.Require().Error(err)
		s.Require().False(ok)
		s.Require().ErrorIs(err, ErrInsufficientBalance)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC20 %s TestTransferInsufficientBalance Result", token)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tValue: %s", huge)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC20Suite) TestApprove() {
	for _, token := range s.tokens {
		ok, err := s.erc20.CallApprove(s.ctx, token, s.from, s.to, s.value, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(ok)
	}

	first, err := s.erc20.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), s.tokens[0], s.from, s.to, s.value)
	s.Require().NoError(err)
	baseNonce := first.AddressNonce()

	unsigned := make([]core.Transaction, len(s.tokens))
	unsigned[0] = first
	for i := 1; i < len(s.tokens); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(baseNonce + uint64(i))

		u, err := s.erc20.CreateApprove(s.ctx, builder, s.tokens[i], s.from, s.to, s.value)
		s.Require().NoError(err)
		unsigned[i] = u
	}

	hashes := make([]*types.Hash, len(s.tokens))
	for i, u := range unsigned {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc20.Approve(s.ctx, u)
		s.Require().NoError(err)
		hashes[i] = hash
	}

	for i, token := range s.tokens {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		approvals, err := ExtractApprovals(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(approvals, 1)
		s.Require().True(approvals[0].Contract.Equal(token))
		s.Require().True(approvals[0].Owner.Equal(s.from))
		s.Require().True(approvals[0].Spender.Equal(s.to))
		s.Require().Zero(s.value.Cmp(approvals[0].Value))

		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.to, rpc.BlockTagLatest)
			return err == nil && s.value.Cmp(allowance) == 0
		}, 10*time.Second, 500*time.Millisecond, "allowance not visible after approve")

		s.T().Logf("ERC20 %s TestApprove Result", token)
		s.T().Logf("\tOwner: %s", s.from)
		s.T().Logf("\tSpender: %s", s.to)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tNonce: %d", unsigned[i].AddressNonce())
		s.T().Logf("\tGas: %d", unsigned[i].Gas())
		s.T().Logf("\tMaxFee: %s", unsigned[i].MaxFee())
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tAllowance: %s", allowance)
	}
}

func (s *ERC20Suite) TestTransferFrom() {
	firstApprove, err := s.erc20.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), s.tokens[0], s.from, s.from, s.value)
	s.Require().NoError(err)
	baseNonce := firstApprove.AddressNonce()

	unsignedApproves := make([]core.Transaction, len(s.tokens))
	unsignedApproves[0] = firstApprove
	for i := 1; i < len(s.tokens); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(baseNonce + uint64(i))

		u, err := s.erc20.CreateApprove(s.ctx, builder, s.tokens[i], s.from, s.from, s.value)
		s.Require().NoError(err)
		unsignedApproves[i] = u
	}

	approveHashes := make([]*types.Hash, len(s.tokens))
	for i, u := range unsignedApproves {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc20.Approve(s.ctx, u)
		s.Require().NoError(err)
		approveHashes[i] = hash
	}

	allowances := make([]*big.Int, len(s.tokens))
	for i, token := range s.tokens {
		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")

		approvals, err := ExtractApprovals(approveReceipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(approvals, 1)
		s.Require().Zero(s.value.Cmp(approvals[0].Value))

		s.Require().Eventually(func() bool {
			allowances[i], err = s.erc20.Allowance(s.ctx, token, s.from, s.from, rpc.BlockTagLatest)
			return err == nil && s.value.Cmp(allowances[i]) == 0
		}, 10*time.Second, 500*time.Millisecond, "allowance not visible after approve")
	}

	firstTransfer, err := s.erc20.CreateTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.tokens[0], s.from, s.from, s.to, s.value)
	s.Require().NoError(err)
	transferBaseNonce := firstTransfer.AddressNonce()

	unsignedTransfers := make([]core.Transaction, len(s.tokens))
	unsignedTransfers[0] = firstTransfer
	for i := 1; i < len(s.tokens); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(transferBaseNonce + uint64(i))

		u, err := s.erc20.CreateTransferFrom(s.ctx, builder, s.tokens[i], s.from, s.from, s.to, s.value)
		s.Require().NoError(err)
		unsignedTransfers[i] = u
	}

	transferHashes := make([]*types.Hash, len(s.tokens))
	for i, u := range unsignedTransfers {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc20.TransferFrom(s.ctx, u)
		s.Require().NoError(err)
		transferHashes[i] = hash
	}

	for i, token := range s.tokens {
		allowance := allowances[i]

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		transfers, err := ExtractTransfers(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].Contract.Equal(token))
		s.Require().True(transfers[0].From.Equal(s.from))
		s.Require().True(transfers[0].To.Equal(s.to))
		s.Require().Zero(s.value.Cmp(transfers[0].Value))

		s.T().Logf("ERC20 %s TestTransferFrom Result", token)
		s.T().Logf("\tApprove")
		s.T().Logf("\t\tOwner: %s", s.from)
		s.T().Logf("\t\tSpender: %s", s.from)
		s.T().Logf("\t\tValue: %s", s.value)
		s.T().Logf("\t\tNonce: %d", unsignedApproves[i].AddressNonce())
		s.T().Logf("\t\tGas: %d", unsignedApproves[i].Gas())
		s.T().Logf("\t\tMaxFee: %s", unsignedApproves[i].MaxFee())
		s.T().Logf("\t\tTx: %s", approveHashes[i])
		s.T().Logf("\t\tAllowance: %s", allowance)
		s.T().Logf("\tTransferFrom")
		s.T().Logf("\t\tSender: %s", s.from)
		s.T().Logf("\t\tFrom: %s", s.from)
		s.T().Logf("\t\tTo: %s", s.to)
		s.T().Logf("\t\tValue: %s", s.value)
		s.T().Logf("\t\tNonce: %d", unsignedTransfers[i].AddressNonce())
		s.T().Logf("\t\tGas: %d", unsignedTransfers[i].Gas())
		s.T().Logf("\t\tMaxFee: %s", unsignedTransfers[i].MaxFee())
		s.T().Logf("\t\tTx: %s", transferHashes[i])
	}
}

func (s *ERC20Suite) TestTransferFromInsufficientAllowance() {
	sender := types.DeadAddress()

	for _, token := range s.tokens {
		ok, err := s.erc20.CallTransferFrom(s.ctx, token, sender, s.from, s.to, s.value, rpc.BlockTagLatest)
		s.Require().Error(err)
		s.Require().False(ok)
		s.Require().ErrorIs(err, ErrInsufficientAllowance)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC20 %s TestTransferFromInsufficientAllowance Result", token)
		s.T().Logf("\tSender: %s", sender)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC20Suite) TestApproveInvalidSpender() {
	spender := types.ZeroAddress()

	for _, token := range s.tokens {
		ok, err := s.erc20.CallApprove(s.ctx, token, s.from, spender, s.value, rpc.BlockTagLatest)
		s.Require().Error(err)
		s.Require().False(ok)
		s.Require().ErrorIs(err, ErrInvalidSpender)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC20 %s TestApproveInvalidSpender Result", token)
		s.T().Logf("\tOwner: %s", s.from)
		s.T().Logf("\tSpender: %s", spender)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}
