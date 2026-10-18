package v2

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
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// erc20ByteCodeHex is OpenZeppelin's ERC20 plus mint/burn, constructor(name, symbol, decimals);
// the same test token every other standards suite self-deploys, used here as tokenA/tokenB so
// createPair has two real, disposable ERC20s to pair -- never an existing real-world token.
const erc20ByteCodeHex = "0x60a060405234801561000f575f5ffd5b506040516114e73803806114e7833981810160405281019061003191906101ee565b828281600390816100429190610497565b5080600490816100529190610497565b5050508060ff1660808160ff1681525050505050610566565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6100ca82610084565b810181811067ffffffffffffffff821117156100e9576100e8610094565b5b80604052505050565b5f6100fb61006b565b905061010782826100c1565b919050565b5f67ffffffffffffffff82111561012657610125610094565b5b61012f82610084565b9050602081019050919050565b8281835e5f83830152505050565b5f61015c6101578461010c565b6100f2565b90508281526020810184848401111561017857610177610080565b5b61018384828561013c565b509392505050565b5f82601f83011261019f5761019e61007c565b5b81516101af84826020860161014a565b91505092915050565b5f60ff82169050919050565b6101cd816101b8565b81146101d7575f5ffd5b50565b5f815190506101e8816101c4565b92915050565b5f5f5f6060848603121561020557610204610074565b5b5f84015167ffffffffffffffff81111561022257610221610078565b5b61022e8682870161018b565b935050602084015167ffffffffffffffff81111561024f5761024e610078565b5b61025b8682870161018b565b925050604061026c868287016101da565b9150509250925092565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f60028204905060018216806102c457607f821691505b6020821081036102d7576102d6610280565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026103397fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826102fe565b61034386836102fe565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61038761038261037d8461035b565b610364565b61035b565b9050919050565b5f819050919050565b6103a08361036d565b6103b46103ac8261038e565b84845461030a565b825550505050565b5f5f905090565b6103cb6103bc565b6103d6818484610397565b505050565b5f5b828110156103fc576103f15f8284016103c3565b6001810190506103dd565b505050565b601f82111561044f578282111561044e5761041b816102dd565b610424836102ef565b61042d856102ef565b602086101561043a575f90505b808301610449828403826103db565b505050505b5b505050565b5f82821c905092915050565b5f61046f5f1984600802610454565b1980831691505092915050565b5f6104878383610460565b9150826002028217905092915050565b6104a082610276565b67ffffffffffffffff8111156104b9576104b8610094565b5b6104c382546102ad565b6104ce828285610401565b5f60209050601f8311600181146104ff575f84156104ed578287015190505b6104f7858261047c565b86555061055e565b601f19841661050d866102dd565b5f5b828110156105345784890151825560018201915060208501945060208101905061050f565b86831015610551578489015161054d601f891682610460565b8355505b6001600288020188555050505b505050505050565b608051610f6961057e5f395f6103370152610f695ff3fe608060405234801561000f575f5ffd5b50600436106100a7575f3560e01c806340c10f191161006f57806340c10f191461016557806370a082311461018157806395d89b41146101b15780639dc29fac146101cf578063a9059cbb146101eb578063dd62ed3e1461021b576100a7565b806306fdde03146100ab578063095ea7b3146100c957806318160ddd146100f957806323b872dd14610117578063313ce56714610147575b5f5ffd5b6100b361024b565b6040516100c09190610be2565b60405180910390f35b6100e360048036038101906100de9190610c93565b6102db565b6040516100f09190610ceb565b60405180910390f35b6101016102fd565b60405161010e9190610d13565b60405180910390f35b610131600480360381019061012c9190610d2c565b610306565b60405161013e9190610ceb565b60405180910390f35b61014f610334565b60405161015c9190610d97565b60405180910390f35b61017f600480360381019061017a9190610c93565b61035b565b005b61019b60048036038101906101969190610db0565b610369565b6040516101a89190610d13565b60405180910390f35b6101b96103ae565b6040516101c69190610be2565b60405180910390f35b6101e960048036038101906101e49190610c93565b61043e565b005b61020560048036038101906102009190610c93565b61044c565b6040516102129190610ceb565b60405180910390f35b61023560048036038101906102309190610ddb565b61046e565b6040516102429190610d13565b60405180910390f35b60606003805461025a90610e46565b80601f016020809104026020016040519081016040528092919081815260200182805461028690610e46565b80156102d15780601f106102a8576101008083540402835291602001916102d1565b820191905f5260205f20905b8154815290600101906020018083116102b457829003601f168201915b5050505050905090565b5f5f6102e56104f0565b90506102f28185856104f7565b600191505092915050565b5f600254905090565b5f5f6103106104f0565b905061031d858285610509565b61032885858561059c565b60019150509392505050565b5f7f0000000000000000000000000000000000000000000000000000000000000000905090565b610365828261068c565b5050565b5f5f5f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b6060600480546103bd90610e46565b80601f01602080910402602001604051908101604052809291908181526020018280546103e990610e46565b80156104345780601f1061040b57610100808354040283529160200191610434565b820191905f5260205f20905b81548152906001019060200180831161041757829003601f168201915b5050505050905090565b610448828261070b565b5050565b5f5f6104566104f0565b905061046381858561059c565b600191505092915050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f33905090565b610504838383600161078a565b505050565b5f610514848461046e565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8110156105965781811015610587578281836040517ffb8f41b200000000000000000000000000000000000000000000000000000000815260040161057e93929190610e85565b60405180910390fd5b61059584848484035f61078a565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361060c575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016106039190610eba565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361067c575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016106739190610eba565b60405180910390fd5b610687838383610959565b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16036106fc575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016106f39190610eba565b60405180910390fd5b6107075f8383610959565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361077b575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016107729190610eba565b60405180910390fd5b610786825f83610959565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff16036107fa575f6040517fe602df050000000000000000000000000000000000000000000000000000000081526004016107f19190610eba565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361086a575f6040517f94280d620000000000000000000000000000000000000000000000000000000081526004016108619190610eba565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508015610953578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9258460405161094a9190610d13565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036109a9578060025f82825461099d9190610f00565b92505081905550610a77565b5f5f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905081811015610a32578381836040517fe450d38c000000000000000000000000000000000000000000000000000000008152600401610a2993929190610e85565b60405180910390fd5b8181035f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610abe578060025f8282540392505081905550610b08565b805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051610b659190610d13565b60405180910390a3505050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f610bb482610b72565b610bbe8185610b7c565b9350610bce818560208601610b8c565b610bd781610b9a565b840191505092915050565b5f6020820190508181035f830152610bfa8184610baa565b905092915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610c2f82610c06565b9050919050565b610c3f81610c25565b8114610c49575f5ffd5b50565b5f81359050610c5a81610c36565b92915050565b5f819050919050565b610c7281610c60565b8114610c7c575f5ffd5b50565b5f81359050610c8d81610c69565b92915050565b5f5f60408385031215610ca957610ca8610c02565b5b5f610cb685828601610c4c565b9250506020610cc785828601610c7f565b9150509250929050565b5f8115159050919050565b610ce581610cd1565b82525050565b5f602082019050610cfe5f830184610cdc565b92915050565b610d0d81610c60565b82525050565b5f602082019050610d265f830184610d04565b92915050565b5f5f5f60608486031215610d4357610d42610c02565b5b5f610d5086828701610c4c565b9350506020610d6186828701610c4c565b9250506040610d7286828701610c7f565b9150509250925092565b5f60ff82169050919050565b610d9181610d7c565b82525050565b5f602082019050610daa5f830184610d88565b92915050565b5f60208284031215610dc557610dc4610c02565b5b5f610dd284828501610c4c565b91505092915050565b5f5f60408385031215610df157610df0610c02565b5b5f610dfe85828601610c4c565b9250506020610e0f85828601610c4c565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680610e5d57607f821691505b602082108103610e7057610e6f610e19565b5b50919050565b610e7f81610c25565b82525050565b5f606082019050610e985f830186610e76565b610ea56020830185610d04565b610eb26040830184610d04565b949350505050565b5f602082019050610ecd5f830184610e76565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f610f0a82610c60565b9150610f1583610c60565b9250828201905080821115610f2d57610f2c610ed3565b5b9291505056fea264697066735822122049c38efc930a01957bf084f9c67bdda53452b282675854f0c07be3a52f60fb5e64736f6c63430008230033"

type FactorySuite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	FactoryHex    string
	Interval      int
	Timeout       int
	ByteCode      string
	Constructor   *abi.Function
	Salt          *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client   rpc.Client
	signer   core.Signer
	deployer *deploy.NicksDeployer
	from     *types.Address
	factory  *types.Address
	tokenA   *types.Address
	tokenB   *types.Address

	v2Factory *Factory

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestFactorySuite(t *testing.T) {
	suite.Run(t, &FactorySuite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		FactoryHex:    "",
		Interval:      1500,
		Timeout:       60,
		ByteCode:      erc20ByteCodeHex,
		Constructor:   abi.NewFunction("constructor", abi.Types{abi.String, abi.String, abi.Uint8}, nil),
		Salt:          core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *FactorySuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v2Factory = NewFactory(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.factory, err = types.NewAddressFromHex(s.FactoryHex)
	s.Require().NoError(err)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [2]string{"evmkit test token A", "evmkit test token B"}
	symbols := [2]string{"EVKA", "EVKB"}
	tokens := [2]*types.Address{}

	for i := range 2 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i], uint8(18)}),
			salt,
		)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)

		unsigned, err := s.deployer.CreateDeployment(
			s.ctx,
			builder,
			s.from,
			deploy.DefaultNicksMethodFactory,
			big.NewInt(0),
			deployment,
		)
		s.Require().NoError(err)

		s.Require().NoError(unsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsigned.MaxFee())
		s.mu.Unlock()

		hash, err := s.deployer.Deploy(s.ctx, unsigned)
		s.Require().NoError(err)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

		token, err := s.deployer.DeployAddress(deploy.DefaultNicksMethodFactory, deployment)
		s.Require().NoError(err)
		tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 20*time.Second, time.Duration(s.Interval)*time.Millisecond, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %s %s (nonce %d, tx %s)", symbols[i], token, unsigned.AddressNonce(), hash)
	}

	s.tokenA, s.tokenB = tokens[0], tokens[1]
	if s.tokenA.Big().Cmp(s.tokenB.Big()) > 0 {
		s.tokenA, s.tokenB = s.tokenB, s.tokenA
	}
}

func (s *FactorySuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("FactorySuite total fee spent: %s", decimal)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *FactorySuite) TestReadMethods() {
	feeTo, err := s.v2Factory.FeeTo(s.ctx, s.factory, rpc.BlockTagLatest)
	s.Require().NoError(err)

	feeToSetter, err := s.v2Factory.FeeToSetter(s.ctx, s.factory, rpc.BlockTagLatest)
	s.Require().NoError(err)

	allPairsLength, err := s.v2Factory.AllPairsLength(s.ctx, s.factory, rpc.BlockTagLatest)
	s.Require().NoError(err)

	s.T().Logf("UniswapV2Factory %s TestReadMethods Result", s.factory)
	s.T().Logf("\tFeeTo: %s", feeTo)
	s.T().Logf("\tFeeToSetter: %s", feeToSetter)
	s.T().Logf("\tAllPairsLength: %s", allPairsLength)
}

func (s *FactorySuite) TestCreatePair() {
	before, err := s.v2Factory.GetPair(s.ctx, s.factory, s.tokenA, s.tokenB, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(before.IsZero(), "tokenA/tokenB shouldn't have a pair yet")

	allPairsLengthBefore, err := s.v2Factory.AllPairsLength(s.ctx, s.factory, rpc.BlockTagLatest)
	s.Require().NoError(err)

	predicted, err := s.v2Factory.CallCreatePair(s.ctx, s.factory, s.from, s.tokenA, s.tokenB, rpc.BlockTagLatest)
	s.Require().NoError(err)

	unsigned, err := s.v2Factory.CreateCreatePair(s.ctx, tx.NewDynamicFeeBuilder(), s.factory, s.from, s.tokenA, s.tokenB)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	hash, err := s.v2Factory.CreatePair(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createPair transaction reverted")

	created, err := ExtractPairCreated(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(created, 1)
	s.Require().True(created[0].Contract.Equal(s.factory))
	s.Require().True(created[0].Token0.Equal(s.tokenA))
	s.Require().True(created[0].Token1.Equal(s.tokenB))
	s.Require().True(created[0].Pair.Equal(predicted))
	s.Require().Zero(new(big.Int).Add(allPairsLengthBefore, big.NewInt(1)).Cmp(created[0].AllPairsLength))

	var after *types.Address
	s.Require().Eventually(func() bool {
		after, err = s.v2Factory.GetPair(s.ctx, s.factory, s.tokenA, s.tokenB, rpc.BlockTagLatest)
		return err == nil && after.Equal(predicted)
	}, 20*time.Second, time.Duration(s.Interval)*time.Millisecond, "pair not visible after createPair")

	allPairsLengthAfter, err := s.v2Factory.AllPairsLength(s.ctx, s.factory, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(new(big.Int).Add(allPairsLengthBefore, big.NewInt(1)).Cmp(allPairsLengthAfter))

	s.T().Logf("UniswapV2Factory %s TestCreatePair Result", s.factory)
	s.T().Logf("\tToken0: %s", s.tokenA)
	s.T().Logf("\tToken1: %s", s.tokenB)
	s.T().Logf("\tPredicted: %s", predicted)
	s.T().Logf("\tPair: %s", created[0].Pair)
	s.T().Logf("\tGetPair: %s", after)
	s.T().Logf("\tAllPairsLength: %s -> %s", allPairsLengthBefore, allPairsLengthAfter)
	s.T().Logf("\tTx: %s", hash)
}

func (s *FactorySuite) TestCreatePairExists() {
	_, err := s.v2Factory.CreateCreatePair(s.ctx, tx.NewDynamicFeeBuilder(), s.factory, s.from, s.tokenA, s.tokenB)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrPairExists)
}

func (s *FactorySuite) TestCreatePairIdenticalAddresses() {
	_, err := s.v2Factory.CreateCreatePair(s.ctx, tx.NewDynamicFeeBuilder(), s.factory, s.from, s.tokenA, s.tokenA)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrIdenticalAddresses)
}

func (s *FactorySuite) TestCreatePairZeroAddress() {
	zero := types.MustNewAddressFromHex("0x0000000000000000000000000000000000000000")
	_, err := s.v2Factory.CreateCreatePair(s.ctx, tx.NewDynamicFeeBuilder(), s.factory, s.from, s.tokenA, zero)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrZeroAddress)
}

func (s *FactorySuite) TestSetFeeToForbidden() {
	_, err := s.v2Factory.CreateSetFeeTo(s.ctx, tx.NewDynamicFeeBuilder(), s.factory, s.from, s.from)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrForbidden)
}

func (s *FactorySuite) TestSetFeeToSetterForbidden() {
	_, err := s.v2Factory.CreateSetFeeToSetter(s.ctx, tx.NewDynamicFeeBuilder(), s.factory, s.from, s.from)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrForbidden)
}
