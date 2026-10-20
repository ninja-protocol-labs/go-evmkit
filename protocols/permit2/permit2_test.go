package permit2

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

const erc20ByteCodeHex = "0x60a060405234801561000f575f5ffd5b506040516114e73803806114e7833981810160405281019061003191906101ee565b828281600390816100429190610497565b5080600490816100529190610497565b5050508060ff1660808160ff1681525050505050610566565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6100ca82610084565b810181811067ffffffffffffffff821117156100e9576100e8610094565b5b80604052505050565b5f6100fb61006b565b905061010782826100c1565b919050565b5f67ffffffffffffffff82111561012657610125610094565b5b61012f82610084565b9050602081019050919050565b8281835e5f83830152505050565b5f61015c6101578461010c565b6100f2565b90508281526020810184848401111561017857610177610080565b5b61018384828561013c565b509392505050565b5f82601f83011261019f5761019e61007c565b5b81516101af84826020860161014a565b91505092915050565b5f60ff82169050919050565b6101cd816101b8565b81146101d7575f5ffd5b50565b5f815190506101e8816101c4565b92915050565b5f5f5f6060848603121561020557610204610074565b5b5f84015167ffffffffffffffff81111561022257610221610078565b5b61022e8682870161018b565b935050602084015167ffffffffffffffff81111561024f5761024e610078565b5b61025b8682870161018b565b925050604061026c868287016101da565b9150509250925092565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f60028204905060018216806102c457607f821691505b6020821081036102d7576102d6610280565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026103397fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826102fe565b61034386836102fe565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61038761038261037d8461035b565b610364565b61035b565b9050919050565b5f819050919050565b6103a08361036d565b6103b46103ac8261038e565b84845461030a565b825550505050565b5f5f905090565b6103cb6103bc565b6103d6818484610397565b505050565b5f5b828110156103fc576103f15f8284016103c3565b6001810190506103dd565b505050565b601f82111561044f578282111561044e5761041b816102dd565b610424836102ef565b61042d856102ef565b602086101561043a575f90505b808301610449828403826103db565b505050505b5b505050565b5f82821c905092915050565b5f61046f5f1984600802610454565b1980831691505092915050565b5f6104878383610460565b9150826002028217905092915050565b6104a082610276565b67ffffffffffffffff8111156104b9576104b8610094565b5b6104c382546102ad565b6104ce828285610401565b5f60209050601f8311600181146104ff575f84156104ed578287015190505b6104f7858261047c565b86555061055e565b601f19841661050d866102dd565b5f5b828110156105345784890151825560018201915060208501945060208101905061050f565b86831015610551578489015161054d601f891682610460565b8355505b6001600288020188555050505b505050505050565b608051610f6961057e5f395f6103370152610f695ff3fe608060405234801561000f575f5ffd5b50600436106100a7575f3560e01c806340c10f191161006f57806340c10f191461016557806370a082311461018157806395d89b41146101b15780639dc29fac146101cf578063a9059cbb146101eb578063dd62ed3e1461021b576100a7565b806306fdde03146100ab578063095ea7b3146100c957806318160ddd146100f957806323b872dd14610117578063313ce56714610147575b5f5ffd5b6100b361024b565b6040516100c09190610be2565b60405180910390f35b6100e360048036038101906100de9190610c93565b6102db565b6040516100f09190610ceb565b60405180910390f35b6101016102fd565b60405161010e9190610d13565b60405180910390f35b610131600480360381019061012c9190610d2c565b610306565b60405161013e9190610ceb565b60405180910390f35b61014f610334565b60405161015c9190610d97565b60405180910390f35b61017f600480360381019061017a9190610c93565b61035b565b005b61019b60048036038101906101969190610db0565b610369565b6040516101a89190610d13565b60405180910390f35b6101b96103ae565b6040516101c69190610be2565b60405180910390f35b6101e960048036038101906101e49190610c93565b61043e565b005b61020560048036038101906102009190610c93565b61044c565b6040516102129190610ceb565b60405180910390f35b61023560048036038101906102309190610ddb565b61046e565b6040516102429190610d13565b60405180910390f35b60606003805461025a90610e46565b80601f016020809104026020016040519081016040528092919081815260200182805461028690610e46565b80156102d15780601f106102a8576101008083540402835291602001916102d1565b820191905f5260205f20905b8154815290600101906020018083116102b457829003601f168201915b5050505050905090565b5f5f6102e56104f0565b90506102f28185856104f7565b600191505092915050565b5f600254905090565b5f5f6103106104f0565b905061031d858285610509565b61032885858561059c565b60019150509392505050565b5f7f0000000000000000000000000000000000000000000000000000000000000000905090565b610365828261068c565b5050565b5f5f5f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b6060600480546103bd90610e46565b80601f01602080910402602001604051908101604052809291908181526020018280546103e990610e46565b80156104345780601f1061040b57610100808354040283529160200191610434565b820191905f5260205f20905b81548152906001019060200180831161041757829003601f168201915b5050505050905090565b610448828261070b565b5050565b5f5f6104566104f0565b905061046381858561059c565b600191505092915050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f33905090565b610504838383600161078a565b505050565b5f610514848461046e565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8110156105965781811015610587578281836040517ffb8f41b200000000000000000000000000000000000000000000000000000000815260040161057e93929190610e85565b60405180910390fd5b61059584848484035f61078a565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361060c575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016106039190610eba565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361067c575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016106739190610eba565b60405180910390fd5b610687838383610959565b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16036106fc575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016106f39190610eba565b60405180910390fd5b6107075f8383610959565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361077b575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016107729190610eba565b60405180910390fd5b610786825f83610959565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff16036107fa575f6040517fe602df050000000000000000000000000000000000000000000000000000000081526004016107f19190610eba565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361086a575f6040517f94280d620000000000000000000000000000000000000000000000000000000081526004016108619190610eba565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508015610953578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9258460405161094a9190610d13565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036109a9578060025f82825461099d9190610f00565b92505081905550610a77565b5f5f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905081811015610a32578381836040517fe450d38c000000000000000000000000000000000000000000000000000000008152600401610a2993929190610e85565b60405180910390fd5b8181035f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610abe578060025f8282540392505081905550610b08565b805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051610b659190610d13565b60405180910390a3505050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f610bb482610b72565b610bbe8185610b7c565b9350610bce818560208601610b8c565b610bd781610b9a565b840191505092915050565b5f6020820190508181035f830152610bfa8184610baa565b905092915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610c2f82610c06565b9050919050565b610c3f81610c25565b8114610c49575f5ffd5b50565b5f81359050610c5a81610c36565b92915050565b5f819050919050565b610c7281610c60565b8114610c7c575f5ffd5b50565b5f81359050610c8d81610c69565b92915050565b5f5f60408385031215610ca957610ca8610c02565b5b5f610cb685828601610c4c565b9250506020610cc785828601610c7f565b9150509250929050565b5f8115159050919050565b610ce581610cd1565b82525050565b5f602082019050610cfe5f830184610cdc565b92915050565b610d0d81610c60565b82525050565b5f602082019050610d265f830184610d04565b92915050565b5f5f5f60608486031215610d4357610d42610c02565b5b5f610d5086828701610c4c565b9350506020610d6186828701610c4c565b9250506040610d7286828701610c7f565b9150509250925092565b5f60ff82169050919050565b610d9181610d7c565b82525050565b5f602082019050610daa5f830184610d88565b92915050565b5f60208284031215610dc557610dc4610c02565b5b5f610dd284828501610c4c565b91505092915050565b5f5f60408385031215610df157610df0610c02565b5b5f610dfe85828601610c4c565b9250506020610e0f85828601610c4c565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680610e5d57607f821691505b602082108103610e7057610e6f610e19565b5b50919050565b610e7f81610c25565b82525050565b5f606082019050610e985f830186610e76565b610ea56020830185610d04565b610eb26040830184610d04565b949350505050565b5f602082019050610ecd5f830184610e76565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f610f0a82610c60565b9150610f1583610c60565b9250828201905080821115610f2d57610f2c610ed3565b5b9291505056fea264697066735822122049c38efc930a01957bf084f9c67bdda53452b282675854f0c07be3a52f60fb5e64736f6c63430008230033"

type Permit2Suite struct {
	suite.Suite

	Enable          bool
	RPCURL          string
	PrivateKeyHex   string
	SpenderKeyHex   string
	Permit2Hex      string
	MulticallHex    string
	ByteCode        string
	Constructor     *abi.Function
	Args            []any
	Salt            *types.Hash
	MintValue       string
	ApproveAmount   string
	NativeSymbol    string
	Interval        int
	Timeout         int
	DeadlineSeconds int
	PriceBuffer     uint64
	LimitBuffer     uint64

	ctx    context.Context
	cancel context.CancelFunc

	client        rpc.Client
	deployer      *deploy.NicksDeployer
	signer        core.Signer
	spenderSigner core.Signer
	from          *types.Address
	spender       *types.Address
	contract      *types.Address
	multicall     *types.Address
	token         *types.Address
	chainID       *big.Int
	mintValue     *big.Int
	approveAmount *big.Int

	p     *Permit2
	erc20 *erc20.ERC20

	totalFee *big.Int
}

func TestPermit2Suite(t *testing.T) {
	suite.Run(t, &Permit2Suite{
		Enable:          false,
		RPCURL:          "",
		PrivateKeyHex:   "",
		SpenderKeyHex:   "",
		Permit2Hex:      DefaultPermit2Address.String(),
		MulticallHex:    multicall3.DefaultMulticall3Address.String(),
		ByteCode:        erc20ByteCodeHex,
		Constructor:     abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String, abi.Uint8), nil),
		Args:            []any{"evmkit permit2 test ERC20", "EVKP2", uint8(18)},
		Salt:            core.Keccak256([]byte("go-evmkit-permit2")),
		MintValue:       "1000000000000000000000000",
		ApproveAmount:   "1000000000000000000000",
		NativeSymbol:    "",
		Interval:        200,
		Timeout:         60,
		DeadlineSeconds: 600,
		PriceBuffer:     150,
		LimitBuffer:     150,
	})
}

func (s *Permit2Suite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.p = New(s.client)
	s.erc20 = erc20.New(s.client)

	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	spenderKey, err := types.NewPrivateKeyFromHex(s.SpenderKeyHex)
	s.Require().NoError(err)
	s.spenderSigner = spenderKey
	s.spender = core.PubkeyToAddress(spenderKey.PublicKey())

	s.contract, err = types.NewAddressFromHex(s.Permit2Hex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	mintValue, ok := new(big.Int).SetString(s.MintValue, 10)
	s.Require().True(ok)
	s.mintValue = mintValue

	approveAmount, ok := new(big.Int).SetString(s.ApproveAmount, 10)
	s.Require().True(ok)
	s.approveAmount = approveAmount

	var chainIDHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHChainID("chainId", &chainIDHex)))
	chainID, ok := new(big.Int).SetString(strings.TrimPrefix(chainIDHex, "0x"), 16)
	s.Require().True(ok)
	s.chainID = chainID

	s.T().Logf("SetupSuite: chainID=%s", s.chainID.String())
	s.T().Logf("\tpermit2 = %s", s.contract.String())
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tmulticall = %s", s.multicall.String())

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	var salt [32]byte
	copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())).Bytes())

	deployment := deploy.NewDeterministicContractDeployment(
		deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, s.Args),
		salt,
	)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.deployer.CreateDeployment(s.ctx, builder, s.from, deploy.DefaultNicksMethodFactory, big.NewInt(0), deployment)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.deployer.Deploy(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "deploy transaction reverted")

	token, err := s.deployer.DeployAddress(deploy.DefaultNicksMethodFactory, deployment)
	s.Require().NoError(err)
	s.token = token

	s.Require().Eventually(func() bool {
		var codeHex string
		if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
			return false
		}
		code, err := encoding.Hex.Decode(codeHex)
		return err == nil && len(code) > 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "no code visible at deployed token %s", token)

	s.T().Logf("SetupSuite: deployed test ERC20")
	s.T().Logf("\ttoken = %s", token.String())
	s.T().Logf("\ttx = %s", hash.String())

	mintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
	calldata, err := mintFn.EncodeCall(s.from, s.mintValue)
	s.Require().NoError(err)

	mintBuilder := tx.NewDynamicFeeBuilder()
	mintBuilder.PriceBuffer(s.PriceBuffer)
	mintBuilder.LimitBuffer(s.LimitBuffer)
	unsignedMint, err := mintBuilder.Build(s.from, s.token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsignedMint.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())

	raw, err := unsignedMint.EncodeRLP()
	s.Require().NoError(err)
	mintHash, err := tx.Broadcast(s.ctx, s.client, raw)
	s.Require().NoError(err)

	mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(mintReceipt.Status, "mint transaction reverted")

	s.T().Logf("SetupSuite: minted test ERC20")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tto = %s", s.from.String())
	s.T().Logf("\tamount = %s", s.mintValue.String())
	s.T().Logf("\ttx = %s", mintHash.String())

	approveBuilder := tx.NewDynamicFeeBuilder()
	approveBuilder.PriceBuffer(s.PriceBuffer)
	approveBuilder.LimitBuffer(s.LimitBuffer)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, approveBuilder, s.token, s.from, s.contract, s.mintValue)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")

	s.Require().Eventually(func() bool {
		a, err := s.erc20.Allowance(s.ctx, s.token, s.from, s.contract, rpc.BlockTagLatest)
		return err == nil && a.Cmp(s.mintValue) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "token allowance to permit2 not visible after approve")

	s.T().Logf("SetupSuite: approved permit2 to spend test ERC20")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\tspender = %s", s.contract.String())
	s.T().Logf("\tamount = %s", s.mintValue.String())
	s.T().Logf("\ttx = %s", approveHash.String())
}

func (s *Permit2Suite) TearDownSuite() {
	if s.token != nil {
		balance, err := s.erc20.BalanceOf(s.ctx, s.token, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.T().Logf("TearDownSuite: remaining balance")
		s.T().Logf("\ttoken = %s", s.token.String())
		s.T().Logf("\towner = %s", s.from.String())
		s.T().Logf("\tbalance = %s", balance.String())
		if balance.Sign() > 0 {
			burnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)
			calldata, err := burnFn.EncodeCall(s.from, balance)
			s.Require().NoError(err)

			builder := tx.NewDynamicFeeBuilder()
			builder.PriceBuffer(s.PriceBuffer)
			builder.LimitBuffer(s.LimitBuffer)
			unsignedBurn, err := builder.Build(s.from, s.token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")
			s.T().Logf("TearDownSuite: burned remaining balance")
			s.T().Logf("\ttoken = %s", s.token.String())
			s.T().Logf("\towner = %s", s.from.String())
			s.T().Logf("\tamount = %s", balance.String())
			s.T().Logf("\ttx = %s", burnHash.String())
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

		s.T().Logf("TearDownSuite: total fee spent")
		s.T().Logf("\tamount = %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Permit2Suite) Test01ReadMethods() {
	s.T().Logf("Test01ReadMethods: fixture")
	s.T().Logf("\tpermit2 = %s", s.contract.String())
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tchainID = %s", s.chainID.String())

	domainSep, err := s.p.DomainSeparator(s.ctx, s.contract, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(domainSep.Bytes(), 32)
	s.T().Logf("Test01ReadMethods: DomainSeparator")
	s.T().Logf("\tvalue = 0x%x", domainSep.Bytes())

	allowance, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test01ReadMethods: Allowance")
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())
	s.Require().Zero(allowance.Amount.Sign())
	s.Require().Zero(allowance.Expiration.Sign())
	s.Require().Zero(allowance.Nonce.Sign())

	wordPos := big.NewInt(0)
	bitmap, err := s.p.NonceBitmap(s.ctx, s.contract, s.from, wordPos, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test01ReadMethods: NonceBitmap")
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\twordPos = %s", wordPos.String())
	s.T().Logf("\tbitmap = %s", bitmap.String())
	s.Require().Zero(bitmap.Sign())
}

func (s *Permit2Suite) Test02Approve() {
	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreateApprove(s.ctx, builder, s.contract, s.from, s.token, s.spender, s.approveAmount, expiration)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.Approve(s.ctx, unsigned)
	s.Require().NoError(err)

	s.T().Logf("Test02Approve: broadcast approve")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", s.approveAmount.String())
	s.T().Logf("\texpiration = %s", expiration.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "approve transaction reverted")

	approvals, err := ExtractApprovals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().True(approvals[0].Contract.Equal(s.contract))
	s.Require().True(approvals[0].Owner.Equal(s.from))
	s.Require().True(approvals[0].Token.Equal(s.token))
	s.Require().True(approvals[0].Spender.Equal(s.spender))
	s.Require().Zero(s.approveAmount.Cmp(approvals[0].Amount))
	s.Require().Zero(expiration.Cmp(approvals[0].Expiration))

	s.T().Logf("Test02Approve: Approval event")
	s.T().Logf("\towner = %s", approvals[0].Owner.String())
	s.T().Logf("\ttoken = %s", approvals[0].Token.String())
	s.T().Logf("\tspender = %s", approvals[0].Spender.String())
	s.T().Logf("\tamount = %s", approvals[0].Amount.String())
	s.T().Logf("\texpiration = %s", approvals[0].Expiration.String())

	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
		if err != nil || a.Amount.Cmp(s.approveAmount) != 0 || a.Expiration.Cmp(expiration) != 0 {
			return false
		}
		allowance = a
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	s.T().Logf("Test02Approve: Allowance after approve")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())
}

func (s *Permit2Suite) Test03PermitSingle() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	details := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	permitSingle := NewPermitSingle(details, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permitSingle.Sign(domain, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test03PermitSingle: signed PermitSingle")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", s.approveAmount.String())
	s.T().Logf("\texpiration = %s", expiration.String())
	s.T().Logf("\tnonce = %s", before.Nonce.String())
	s.T().Logf("\tsigDeadline = %s", sigDeadline.String())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreatePermitSingle(s.ctx, builder, s.contract, s.spender, s.from, permitSingle, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.PermitSingle(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test03PermitSingle: broadcast permit")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permit transaction reverted")

	permits, err := ExtractPermits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(permits, 1)
	s.Require().True(permits[0].Contract.Equal(s.contract))
	s.Require().True(permits[0].Owner.Equal(s.from))
	s.Require().True(permits[0].Token.Equal(s.token))
	s.Require().True(permits[0].Spender.Equal(s.spender))
	s.Require().Zero(s.approveAmount.Cmp(permits[0].Amount))
	s.Require().Zero(expiration.Cmp(permits[0].Expiration))
	s.Require().Zero(before.Nonce.Cmp(permits[0].Nonce))

	s.T().Logf("Test03PermitSingle: Permit event")
	s.T().Logf("\towner = %s", permits[0].Owner.String())
	s.T().Logf("\ttoken = %s", permits[0].Token.String())
	s.T().Logf("\tspender = %s", permits[0].Spender.String())
	s.T().Logf("\tamount = %s", permits[0].Amount.String())
	s.T().Logf("\texpiration = %s", permits[0].Expiration.String())
	s.T().Logf("\tnonce = %s", permits[0].Nonce.String())

	wantNonce := new(big.Int).Add(before.Nonce, big.NewInt(1))
	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
		if err != nil || a.Nonce.Cmp(wantNonce) != 0 {
			return false
		}
		allowance = a
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance nonce not advanced after permit")

	s.T().Logf("Test03PermitSingle: Allowance after permit")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())
}

func (s *Permit2Suite) Test04TransferFrom() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)

	amount := new(big.Int).Div(before.Amount, big.NewInt(2))

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreateTransferFrom(s.ctx, builder, s.contract, s.spender, s.from, s.spender, amount, s.token)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.TransferFrom(s.ctx, unsigned)
	s.Require().NoError(err)

	s.T().Logf("Test04TransferFrom: broadcast transferFrom")
	s.T().Logf("\tfrom = %s", s.from.String())
	s.T().Logf("\tto = %s", s.spender.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", amount.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transferFrom transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(amount.Cmp(transfers[0].Value))

	s.T().Logf("Test04TransferFrom: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantAllowance := new(big.Int).Sub(before.Amount, amount)
	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
		if err != nil || a.Amount.Cmp(wantAllowance) != 0 {
			return false
		}
		allowance = a
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not decremented after transferFrom")

	s.T().Logf("Test04TransferFrom: Allowance after transferFrom")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())

	balancesAfter, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	fromBalanceAfter := balancesAfter.Addresses[0].Amount
	toBalanceAfter := balancesAfter.Addresses[1].Amount

	s.T().Logf("Test04TransferFrom: balances after transferFrom")
	s.T().Logf("\tfrom = %s", fromBalanceAfter.String())
	s.T().Logf("\tto = %s", toBalanceAfter.String())

	wantFromBalance := new(big.Int).Sub(fromBalanceBefore, amount)
	wantToBalance := new(big.Int).Add(toBalanceBefore, amount)
	s.Require().Zero(wantFromBalance.Cmp(fromBalanceAfter))
	s.Require().Zero(wantToBalance.Cmp(toBalanceAfter))
}

func (s *Permit2Suite) Test05PermitBatch() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	details := []PermitDetails{*NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)}
	permitBatch := NewPermitBatch(details, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permitBatch.Sign(domain, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test05PermitBatch: signed PermitBatch")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", s.approveAmount.String())
	s.T().Logf("\texpiration = %s", expiration.String())
	s.T().Logf("\tnonce = %s", before.Nonce.String())
	s.T().Logf("\tsigDeadline = %s", sigDeadline.String())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreatePermitBatch(s.ctx, builder, s.contract, s.spender, s.from, permitBatch, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.PermitBatch(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test05PermitBatch: broadcast permit")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permit transaction reverted")

	permits, err := ExtractPermits(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(permits, 1)
	s.Require().True(permits[0].Contract.Equal(s.contract))
	s.Require().True(permits[0].Owner.Equal(s.from))
	s.Require().True(permits[0].Token.Equal(s.token))
	s.Require().True(permits[0].Spender.Equal(s.spender))
	s.Require().Zero(s.approveAmount.Cmp(permits[0].Amount))
	s.Require().Zero(expiration.Cmp(permits[0].Expiration))
	s.Require().Zero(before.Nonce.Cmp(permits[0].Nonce))

	s.T().Logf("Test05PermitBatch: Permit event")
	s.T().Logf("\towner = %s", permits[0].Owner.String())
	s.T().Logf("\ttoken = %s", permits[0].Token.String())
	s.T().Logf("\tspender = %s", permits[0].Spender.String())
	s.T().Logf("\tamount = %s", permits[0].Amount.String())
	s.T().Logf("\texpiration = %s", permits[0].Expiration.String())
	s.T().Logf("\tnonce = %s", permits[0].Nonce.String())

	wantNonce := new(big.Int).Add(before.Nonce, big.NewInt(1))
	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
		if err != nil || a.Nonce.Cmp(wantNonce) != 0 {
			return false
		}
		allowance = a
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance nonce not advanced after permitBatch")

	s.T().Logf("Test05PermitBatch: Allowance after permitBatch")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())
}

func (s *Permit2Suite) Test06TransferFromBatch() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)

	amount := new(big.Int).Div(before.Amount, big.NewInt(2))
	transferDetails := []AllowanceTransferDetails{*NewAllowanceTransferDetails(s.from, s.spender, amount, s.token)}

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreateTransferFromBatch(s.ctx, builder, s.contract, s.spender, transferDetails)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.TransferFromBatch(s.ctx, unsigned)
	s.Require().NoError(err)

	s.T().Logf("Test06TransferFromBatch: broadcast transferFrom")
	s.T().Logf("\tfrom = %s", s.from.String())
	s.T().Logf("\tto = %s", s.spender.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", amount.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transferFrom(batch) transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(amount.Cmp(transfers[0].Value))

	s.T().Logf("Test06TransferFromBatch: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantAllowance := new(big.Int).Sub(before.Amount, amount)
	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
		if err != nil || a.Amount.Cmp(wantAllowance) != 0 {
			return false
		}
		allowance = a
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not decremented after transferFrom(batch)")

	s.T().Logf("Test06TransferFromBatch: Allowance after transferFrom")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())

	balancesAfter, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	fromBalanceAfter := balancesAfter.Addresses[0].Amount
	toBalanceAfter := balancesAfter.Addresses[1].Amount

	s.T().Logf("Test06TransferFromBatch: balances after transferFrom")
	s.T().Logf("\tfrom = %s", fromBalanceAfter.String())
	s.T().Logf("\tto = %s", toBalanceAfter.String())

	wantFromBalance := new(big.Int).Sub(fromBalanceBefore, amount)
	wantToBalance := new(big.Int).Add(toBalanceBefore, amount)
	s.Require().Zero(wantFromBalance.Cmp(fromBalanceAfter))
	s.Require().Zero(wantToBalance.Cmp(toBalanceAfter))
}

func (s *Permit2Suite) Test07Lockdown() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Positive(before.Amount.Sign(), "expected a nonzero allowance to lock down")

	approvals := []TokenSpenderPair{*NewTokenSpenderPair(s.token, s.spender)}

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreateLockdown(s.ctx, builder, s.contract, s.from, approvals)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.Lockdown(s.ctx, unsigned)
	s.Require().NoError(err)

	s.T().Logf("Test07Lockdown: broadcast lockdown")
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "lockdown transaction reverted")

	lockdowns, err := ExtractLockdowns(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(lockdowns, 1)
	s.Require().True(lockdowns[0].Contract.Equal(s.contract))
	s.Require().True(lockdowns[0].Owner.Equal(s.from))
	s.Require().True(lockdowns[0].Token.Equal(s.token))
	s.Require().True(lockdowns[0].Spender.Equal(s.spender))

	s.T().Logf("Test07Lockdown: Lockdown event")
	s.T().Logf("\towner = %s", lockdowns[0].Owner.String())
	s.T().Logf("\ttoken = %s", lockdowns[0].Token.String())
	s.T().Logf("\tspender = %s", lockdowns[0].Spender.String())

	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
		if err != nil || a.Amount.Sign() != 0 {
			return false
		}
		allowance = a
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not zeroed after lockdown")

	s.T().Logf("Test07Lockdown: Allowance after lockdown")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())
}

func (s *Permit2Suite) Test08InvalidateNonces() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)

	newNonce := new(big.Int).Add(before.Nonce, big.NewInt(5))

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreateInvalidateNonces(s.ctx, builder, s.contract, s.from, s.token, s.spender, newNonce)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.InvalidateNonces(s.ctx, unsigned)
	s.Require().NoError(err)

	s.T().Logf("Test08InvalidateNonces: broadcast invalidateNonces")
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\toldNonce = %s", before.Nonce.String())
	s.T().Logf("\tnewNonce = %s", newNonce.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "invalidateNonces transaction reverted")

	invalidations, err := ExtractNonceInvalidations(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(invalidations, 1)
	s.Require().True(invalidations[0].Contract.Equal(s.contract))
	s.Require().True(invalidations[0].Owner.Equal(s.from))
	s.Require().True(invalidations[0].Token.Equal(s.token))
	s.Require().True(invalidations[0].Spender.Equal(s.spender))
	s.Require().Zero(newNonce.Cmp(invalidations[0].NewNonce))
	s.Require().Zero(before.Nonce.Cmp(invalidations[0].OldNonce))

	s.T().Logf("Test08InvalidateNonces: NonceInvalidation event")
	s.T().Logf("\towner = %s", invalidations[0].Owner.String())
	s.T().Logf("\ttoken = %s", invalidations[0].Token.String())
	s.T().Logf("\tspender = %s", invalidations[0].Spender.String())
	s.T().Logf("\tnewNonce = %s", invalidations[0].NewNonce.String())
	s.T().Logf("\toldNonce = %s", invalidations[0].OldNonce.String())

	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
		if err != nil || a.Nonce.Cmp(newNonce) != 0 {
			return false
		}
		allowance = a
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance nonce not advanced after invalidateNonces")

	s.T().Logf("Test08InvalidateNonces: Allowance after invalidateNonces")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	staleDetails := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	stalePermitSingle := NewPermitSingle(staleDetails, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	staleSig, err := stalePermitSingle.Sign(domain, s.signer)
	s.Require().NoError(err)

	staleBuilder := tx.NewDynamicFeeBuilder()
	staleBuilder.PriceBuffer(s.PriceBuffer)
	staleBuilder.LimitBuffer(s.LimitBuffer)
	_, err = s.p.CreatePermitSingle(s.ctx, staleBuilder, s.contract, s.spender, s.from, stalePermitSingle, staleSig)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidNonce)
	s.T().Logf("Test08InvalidateNonces: stale-nonce PermitSingle correctly rejected: %v", err)
}

func (s *Permit2Suite) Test09PermitTransferFrom() {
	requestedAmount := new(big.Int).Div(s.approveAmount, big.NewInt(10))
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	permitted := NewTokenPermissions(s.token, requestedAmount)
	permit := NewPermitTransferFrom(permitted, nonce, deadline)
	transferDetails := NewSignatureTransferDetails(s.spender, requestedAmount)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.Sign(domain, s.spender, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test09PermitTransferFrom: signed PermitTransferFrom")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", requestedAmount.String())
	s.T().Logf("\tnonce = %s", nonce.String())
	s.T().Logf("\tdeadline = %s", deadline.String())
	s.T().Logf("\tspender = %s", s.spender.String())

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreatePermitTransferFrom(s.ctx, builder, s.contract, s.spender, permit, transferDetails, s.from, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.PermitTransferFrom(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test09PermitTransferFrom: broadcast permitTransferFrom")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permitTransferFrom transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(requestedAmount.Cmp(transfers[0].Value))

	s.T().Logf("Test09PermitTransferFrom: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantFrom := new(big.Int).Sub(fromBalanceBefore, requestedAmount)
	wantTo := new(big.Int).Add(toBalanceBefore, requestedAmount)
	var balancesAfter *erc20.TokenBalances
	s.Require().Eventually(func() bool {
		b, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
		if err != nil || b.Addresses[0].Amount.Cmp(wantFrom) != 0 || b.Addresses[1].Amount.Cmp(wantTo) != 0 {
			return false
		}
		balancesAfter = b
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not updated after permitTransferFrom")

	s.T().Logf("Test09PermitTransferFrom: balances after permitTransferFrom")
	s.T().Logf("\tfrom = %s", balancesAfter.Addresses[0].Amount.String())
	s.T().Logf("\tto = %s", balancesAfter.Addresses[1].Amount.String())
}

func (s *Permit2Suite) Test10PermitWitnessTransferFrom() {
	requestedAmount := new(big.Int).Div(s.approveAmount, big.NewInt(10))
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	witness := types.NewHashFromBytes([]byte("test witness payload, 32 bytes!"))
	witnessTypeString := "Witness witness)Witness(uint256 value)"

	permitted := NewTokenPermissions(s.token, requestedAmount)
	permit := NewPermitTransferFrom(permitted, nonce, deadline)
	transferDetails := NewSignatureTransferDetails(s.spender, requestedAmount)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.SignWithWitness(domain, s.spender, witness, witnessTypeString, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test10PermitWitnessTransferFrom: signed PermitWitnessTransferFrom")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", requestedAmount.String())
	s.T().Logf("\tnonce = %s", nonce.String())
	s.T().Logf("\tdeadline = %s", deadline.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\twitness = 0x%x", witness.Bytes())
	s.T().Logf("\twitnessTypeString = %s", witnessTypeString)

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreatePermitWitnessTransferFrom(s.ctx, builder, s.contract, s.spender, permit, transferDetails, s.from, witness, witnessTypeString, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.PermitWitnessTransferFrom(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test10PermitWitnessTransferFrom: broadcast permitWitnessTransferFrom")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permitWitnessTransferFrom transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(requestedAmount.Cmp(transfers[0].Value))

	s.T().Logf("Test10PermitWitnessTransferFrom: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantFrom := new(big.Int).Sub(fromBalanceBefore, requestedAmount)
	wantTo := new(big.Int).Add(toBalanceBefore, requestedAmount)
	var balancesAfter *erc20.TokenBalances
	s.Require().Eventually(func() bool {
		b, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
		if err != nil || b.Addresses[0].Amount.Cmp(wantFrom) != 0 || b.Addresses[1].Amount.Cmp(wantTo) != 0 {
			return false
		}
		balancesAfter = b
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not updated after permitWitnessTransferFrom")

	s.T().Logf("Test10PermitWitnessTransferFrom: balances after permitWitnessTransferFrom")
	s.T().Logf("\tfrom = %s", balancesAfter.Addresses[0].Amount.String())
	s.T().Logf("\tto = %s", balancesAfter.Addresses[1].Amount.String())
}

func (s *Permit2Suite) Test11PermitTransferFromBatch() {
	requestedAmount := new(big.Int).Div(s.approveAmount, big.NewInt(10))
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	permitted := []TokenPermissions{*NewTokenPermissions(s.token, requestedAmount)}
	permit := NewPermitBatchTransferFrom(permitted, nonce, deadline)
	transferDetails := []SignatureTransferDetails{*NewSignatureTransferDetails(s.spender, requestedAmount)}

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.Sign(domain, s.spender, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test11PermitTransferFromBatch: signed PermitBatchTransferFrom")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", requestedAmount.String())
	s.T().Logf("\tnonce = %s", nonce.String())
	s.T().Logf("\tdeadline = %s", deadline.String())
	s.T().Logf("\tspender = %s", s.spender.String())

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreatePermitTransferFromBatch(s.ctx, builder, s.contract, s.spender, permit, transferDetails, s.from, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.PermitTransferFromBatch(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test11PermitTransferFromBatch: broadcast permitTransferFrom(batch)")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permitTransferFrom(batch) transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(requestedAmount.Cmp(transfers[0].Value))

	s.T().Logf("Test11PermitTransferFromBatch: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantFrom := new(big.Int).Sub(fromBalanceBefore, requestedAmount)
	wantTo := new(big.Int).Add(toBalanceBefore, requestedAmount)
	var balancesAfter *erc20.TokenBalances
	s.Require().Eventually(func() bool {
		b, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
		if err != nil || b.Addresses[0].Amount.Cmp(wantFrom) != 0 || b.Addresses[1].Amount.Cmp(wantTo) != 0 {
			return false
		}
		balancesAfter = b
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not updated after permitTransferFrom(batch)")

	s.T().Logf("Test11PermitTransferFromBatch: balances after permitTransferFrom(batch)")
	s.T().Logf("\tfrom = %s", balancesAfter.Addresses[0].Amount.String())
	s.T().Logf("\tto = %s", balancesAfter.Addresses[1].Amount.String())
}

func (s *Permit2Suite) Test12PermitWitnessTransferFromBatch() {
	requestedAmount := new(big.Int).Div(s.approveAmount, big.NewInt(10))
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	witness := types.NewHashFromBytes([]byte("test witness payload, 32 bytes!"))
	witnessTypeString := "Witness witness)Witness(uint256 value)"

	permitted := []TokenPermissions{*NewTokenPermissions(s.token, requestedAmount)}
	permit := NewPermitBatchTransferFrom(permitted, nonce, deadline)
	transferDetails := []SignatureTransferDetails{*NewSignatureTransferDetails(s.spender, requestedAmount)}

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.SignWithWitness(domain, s.spender, witness, witnessTypeString, s.signer)
	s.Require().NoError(err)

	s.T().Logf("Test12PermitWitnessTransferFromBatch: signed PermitBatchTransferFrom with witness")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tamount = %s", requestedAmount.String())
	s.T().Logf("\tnonce = %s", nonce.String())
	s.T().Logf("\tdeadline = %s", deadline.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\twitness = 0x%x", witness.Bytes())
	s.T().Logf("\twitnessTypeString = %s", witnessTypeString)

	balancesBefore, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	fromBalanceBefore := balancesBefore.Addresses[0].Amount
	toBalanceBefore := balancesBefore.Addresses[1].Amount

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreatePermitWitnessTransferFromBatch(s.ctx, builder, s.contract, s.spender, permit, transferDetails, s.from, witness, witnessTypeString, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.PermitWitnessTransferFromBatch(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test12PermitWitnessTransferFromBatch: broadcast permitWitnessTransferFrom(batch)")
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "permitWitnessTransferFrom(batch) transaction reverted")

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.token))
	s.Require().True(transfers[0].From.Equal(s.from))
	s.Require().True(transfers[0].To.Equal(s.spender))
	s.Require().Zero(requestedAmount.Cmp(transfers[0].Value))

	s.T().Logf("Test12PermitWitnessTransferFromBatch: Transfer event")
	s.T().Logf("\tfrom = %s", transfers[0].From.String())
	s.T().Logf("\tto = %s", transfers[0].To.String())
	s.T().Logf("\tvalue = %s", transfers[0].Value.String())

	wantFrom := new(big.Int).Sub(fromBalanceBefore, requestedAmount)
	wantTo := new(big.Int).Add(toBalanceBefore, requestedAmount)
	var balancesAfter *erc20.TokenBalances
	s.Require().Eventually(func() bool {
		b, err := s.erc20.TokenBalances(s.ctx, s.multicall, s.token, []*types.Address{s.from, s.spender}, false, rpc.BlockTagLatest)
		if err != nil || b.Addresses[0].Amount.Cmp(wantFrom) != 0 || b.Addresses[1].Amount.Cmp(wantTo) != 0 {
			return false
		}
		balancesAfter = b
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not updated after permitWitnessTransferFrom(batch)")

	s.T().Logf("Test12PermitWitnessTransferFromBatch: balances after permitWitnessTransferFrom(batch)")
	s.T().Logf("\tfrom = %s", balancesAfter.Addresses[0].Amount.String())
	s.T().Logf("\tto = %s", balancesAfter.Addresses[1].Amount.String())
}

func (s *Permit2Suite) Test13InvalidateUnorderedNonces() {
	wordPos := big.NewInt(time.Now().UnixNano())
	mask := big.NewInt(0b1011)

	before, err := s.p.NonceBitmap(s.ctx, s.contract, s.from, wordPos, rpc.BlockTagLatest)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreateInvalidateUnorderedNonces(s.ctx, builder, s.contract, s.from, wordPos, mask)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.InvalidateUnorderedNonces(s.ctx, unsigned)
	s.Require().NoError(err)

	s.T().Logf("Test13InvalidateUnorderedNonces: broadcast invalidateUnorderedNonces")
	s.T().Logf("\towner = %s", s.from.String())
	s.T().Logf("\twordPos = %s", wordPos.String())
	s.T().Logf("\tmask = %s", mask.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "invalidateUnorderedNonces transaction reverted")

	invalidations, err := ExtractUnorderedNonceInvalidations(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(invalidations, 1)
	s.Require().True(invalidations[0].Contract.Equal(s.contract))
	s.Require().True(invalidations[0].Owner.Equal(s.from))
	s.Require().Zero(wordPos.Cmp(invalidations[0].Word))
	s.Require().Zero(mask.Cmp(invalidations[0].Mask))

	s.T().Logf("Test13InvalidateUnorderedNonces: UnorderedNonceInvalidation event")
	s.T().Logf("\towner = %s", invalidations[0].Owner.String())
	s.T().Logf("\tword = %s", invalidations[0].Word.String())
	s.T().Logf("\tmask = %s", invalidations[0].Mask.String())

	wantBitmap := new(big.Int).Or(before, mask)
	var bitmap *big.Int
	s.Require().Eventually(func() bool {
		b, err := s.p.NonceBitmap(s.ctx, s.contract, s.from, wordPos, rpc.BlockTagLatest)
		if err != nil || b.Cmp(wantBitmap) != 0 {
			return false
		}
		bitmap = b
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "bitmap not updated after invalidateUnorderedNonces")

	s.T().Logf("Test13InvalidateUnorderedNonces: NonceBitmap after invalidateUnorderedNonces")
	s.T().Logf("\tbitmap = %s", bitmap.String())

	usedNonce := new(big.Int).Lsh(wordPos, 8)

	permitted := NewTokenPermissions(s.token, big.NewInt(1))
	permit := NewPermitTransferFrom(permitted, usedNonce, big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds)*time.Second).Unix()))
	transferDetails := NewSignatureTransferDetails(s.spender, big.NewInt(1))

	domain := NewDomain(s.chainID, s.contract)
	staleSig, err := permit.Sign(domain, s.spender, s.signer)
	s.Require().NoError(err)

	staleBuilder := tx.NewDynamicFeeBuilder()
	staleBuilder.PriceBuffer(s.PriceBuffer)
	staleBuilder.LimitBuffer(s.LimitBuffer)
	_, err = s.p.CreatePermitTransferFrom(s.ctx, staleBuilder, s.contract, s.spender, permit, transferDetails, s.from, staleSig)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidNonce)
	s.T().Logf("Test13InvalidateUnorderedNonces: reused nonce correctly rejected: %v", err)
}

func (s *Permit2Suite) Test14AllowanceExpired() {
	amount := big.NewInt(1)
	expiration := big.NewInt(time.Now().Add(5 * time.Second).Unix())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreateApprove(s.ctx, builder, s.contract, s.from, s.token, s.spender, amount, expiration)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.Approve(s.ctx, unsigned)
	s.Require().NoError(err)

	s.T().Logf("Test14AllowanceExpired: broadcast approve with short expiration")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", amount.String())
	s.T().Logf("\texpiration = %s", expiration.String())
	s.T().Logf("\ttx = %s", hash.String())

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "approve transaction reverted")

	var allowance *PackedAllowance
	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
		if err != nil || a.Amount.Cmp(amount) != 0 {
			return false
		}
		allowance = a
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	s.T().Logf("Test14AllowanceExpired: Allowance before expiry")
	s.T().Logf("\tamount = %s", allowance.Amount.String())
	s.T().Logf("\texpiration = %s", allowance.Expiration.String())
	s.T().Logf("\tnonce = %s", allowance.Nonce.String())

	s.T().Logf("Test14AllowanceExpired: waiting for allowance to expire")
	time.Sleep(10 * time.Second)

	expiredBuilder := tx.NewDynamicFeeBuilder()
	expiredBuilder.PriceBuffer(s.PriceBuffer)
	expiredBuilder.LimitBuffer(s.LimitBuffer)
	_, err = s.p.CreateTransferFrom(s.ctx, expiredBuilder, s.contract, s.spender, s.from, s.spender, amount, s.token)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrAllowanceExpired)
	s.T().Logf("Test14AllowanceExpired: expired-allowance transferFrom correctly rejected: %v", err)
}

func (s *Permit2Suite) Test15SignatureExpired() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(-1 * time.Hour).Unix())
	details := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	permitSingle := NewPermitSingle(details, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permitSingle.Sign(domain, s.signer)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err = s.p.CreatePermitSingle(s.ctx, builder, s.contract, s.spender, s.from, permitSingle, sig)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrSignatureExpired)
	s.T().Logf("Test15SignatureExpired: expired-deadline PermitSingle correctly rejected: %v", err)
}

func (s *Permit2Suite) Test16ExcessiveInvalidation() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)

	newNonce := new(big.Int).Add(before.Nonce, big.NewInt(70000))

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err = s.p.CreateInvalidateNonces(s.ctx, builder, s.contract, s.from, s.token, s.spender, newNonce)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrExcessiveInvalidation)
	s.T().Logf("Test16ExcessiveInvalidation: oversized nonce jump correctly rejected: %v", err)
}

func (s *Permit2Suite) Test17InvalidSignatureLength() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	details := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	permitSingle := NewPermitSingle(details, s.spender, sigDeadline)

	shortSig := make([]byte, 63)
	calldata, err := permitSingleFn.EncodeCall(s.from, permitSingle.tuple(), shortSig)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err = builder.Build(s.spender, s.contract, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().Error(err)

	wrapped := s.p.wrapRevertError(err)
	s.Require().ErrorIs(wrapped, ErrInvalidSignatureLength)
	s.T().Logf("Test17InvalidSignatureLength: short signature correctly rejected: %v", wrapped)
}

func (s *Permit2Suite) Test18InvalidSignature() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	details := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	permitSingle := NewPermitSingle(details, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permitSingle.Sign(domain, s.signer)
	s.Require().NoError(err)

	badSig := sig.LegacyBytes()
	badSig[64] = 17

	calldata, err := permitSingleFn.EncodeCall(s.from, permitSingle.tuple(), badSig)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err = builder.Build(s.spender, s.contract, big.NewInt(0), calldata).Pack(s.ctx, s.client)
	s.Require().Error(err)

	wrapped := s.p.wrapRevertError(err)
	s.Require().ErrorIs(wrapped, ErrInvalidSignature)
	s.T().Logf("Test18InvalidSignature: malformed-v signature correctly rejected: %v", wrapped)
}

func (s *Permit2Suite) Test19InvalidSigner() {
	before, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
	s.Require().NoError(err)

	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
	sigDeadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())
	details := NewPermitDetails(s.token, s.approveAmount, expiration, before.Nonce)
	permitSingle := NewPermitSingle(details, s.spender, sigDeadline)

	domain := NewDomain(s.chainID, s.contract)
	wrongSig, err := permitSingle.Sign(domain, s.spenderSigner)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err = s.p.CreatePermitSingle(s.ctx, builder, s.contract, s.spender, s.from, permitSingle, wrongSig)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidSigner)
	s.T().Logf("Test19InvalidSigner: wrong-signer PermitSingle correctly rejected: %v", err)
}

func (s *Permit2Suite) Test20InvalidAmount() {
	permittedAmount := big.NewInt(100)
	requestedAmount := big.NewInt(101)
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	permitted := NewTokenPermissions(s.token, permittedAmount)
	permit := NewPermitTransferFrom(permitted, nonce, deadline)
	transferDetails := NewSignatureTransferDetails(s.spender, requestedAmount)

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.Sign(domain, s.spender, s.signer)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err = s.p.CreatePermitTransferFrom(s.ctx, builder, s.contract, s.spender, permit, transferDetails, s.from, sig)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidAmount)
	s.T().Logf("Test20InvalidAmount: over-requested PermitTransferFrom correctly rejected: %v", err)
}

func (s *Permit2Suite) Test21LengthMismatch() {
	nonce := big.NewInt(time.Now().UnixNano())
	deadline := big.NewInt(time.Now().Add(time.Duration(s.DeadlineSeconds) * time.Second).Unix())

	permitted := []TokenPermissions{*NewTokenPermissions(s.token, big.NewInt(1))}
	permit := NewPermitBatchTransferFrom(permitted, nonce, deadline)
	transferDetails := []SignatureTransferDetails{
		*NewSignatureTransferDetails(s.spender, big.NewInt(1)),
		*NewSignatureTransferDetails(s.spender, big.NewInt(1)),
	}

	domain := NewDomain(s.chainID, s.contract)
	sig, err := permit.Sign(domain, s.spender, s.signer)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	_, err = s.p.CreatePermitTransferFromBatch(s.ctx, builder, s.contract, s.spender, permit, transferDetails, s.from, sig)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrLengthMismatch)
	s.T().Logf("Test21LengthMismatch: mismatched transferDetails length correctly rejected: %v", err)
}

func (s *Permit2Suite) Test22InsufficientAllowance() {
	amount := big.NewInt(1)
	expiration := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(s.PriceBuffer)
	builder.LimitBuffer(s.LimitBuffer)
	unsigned, err := s.p.CreateApprove(s.ctx, builder, s.contract, s.from, s.token, s.spender, amount, expiration)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())

	hash, err := s.p.Approve(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "approve transaction reverted")

	s.Require().Eventually(func() bool {
		a, err := s.p.Allowance(s.ctx, s.contract, s.from, s.token, s.spender, rpc.BlockTagLatest)
		return err == nil && a.Amount.Cmp(amount) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")

	s.T().Logf("Test22InsufficientAllowance: approved small allowance")
	s.T().Logf("\ttoken = %s", s.token.String())
	s.T().Logf("\tspender = %s", s.spender.String())
	s.T().Logf("\tamount = %s", amount.String())
	s.T().Logf("\texpiration = %s", expiration.String())

	overAmount := new(big.Int).Add(amount, big.NewInt(1))
	overBuilder := tx.NewDynamicFeeBuilder()
	overBuilder.PriceBuffer(s.PriceBuffer)
	overBuilder.LimitBuffer(s.LimitBuffer)
	_, err = s.p.CreateTransferFrom(s.ctx, overBuilder, s.contract, s.spender, s.from, s.spender, overAmount, s.token)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientAllowance)
	s.T().Logf("Test22InsufficientAllowance: over-amount transferFrom correctly rejected: %v", err)
}
