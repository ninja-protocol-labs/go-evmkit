package erc6909

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

var interfaceIDERC6909 = [4]byte{0x0f, 0x63, 0x2f, 0xb3}

const erc6909ByteCodeHex = "0x608060405234801561000f575f5ffd5b50604051611df1380380611df183398181016040528101906100319190610343565b610041848461006a60201b60201c565b61005184836100c760201b60201c565b610061848261012560201b60201c565b50505050610756565b8060035f8481526020019081526020015f205f01908161008a91906105f7565b50817f6846707b16f55a3623d38bf118700af63d0784698270a53ce28b116dbaa8b794826040516100bb919061070e565b60405180910390a25050565b8060035f8481526020019081526020015f2060010190816100e891906105f7565b50817f8fabcb7b4fbdde605303a7339bc80fb6956cd01a42e9cdf2d15d0f2fdd3b06fa82604051610119919061070e565b60405180910390a25050565b8060035f8481526020019081526020015f206002015f6101000a81548160ff021916908360ff160217905550817fc7435874568017e6134c851002a405d733521c0f4c63f9f25f3a32a4d595b72982604051610181919061073d565b60405180910390a25050565b5f604051905090565b5f5ffd5b5f5ffd5b5f819050919050565b6101b08161019e565b81146101ba575f5ffd5b50565b5f815190506101cb816101a7565b92915050565b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b61021f826101d9565b810181811067ffffffffffffffff8211171561023e5761023d6101e9565b5b80604052505050565b5f61025061018d565b905061025c8282610216565b919050565b5f67ffffffffffffffff82111561027b5761027a6101e9565b5b610284826101d9565b9050602081019050919050565b8281835e5f83830152505050565b5f6102b16102ac84610261565b610247565b9050828152602081018484840111156102cd576102cc6101d5565b5b6102d8848285610291565b509392505050565b5f82601f8301126102f4576102f36101d1565b5b815161030484826020860161029f565b91505092915050565b5f60ff82169050919050565b6103228161030d565b811461032c575f5ffd5b50565b5f8151905061033d81610319565b92915050565b5f5f5f5f6080858703121561035b5761035a610196565b5b5f610368878288016101bd565b945050602085015167ffffffffffffffff8111156103895761038861019a565b5b610395878288016102e0565b935050604085015167ffffffffffffffff8111156103b6576103b561019a565b5b6103c2878288016102e0565b92505060606103d38782880161032f565b91505092959194509250565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061042d57607f821691505b6020821081036104405761043f6103e9565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026104a27fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610467565b6104ac8683610467565b95508019841693508086168417925050509392505050565b5f819050919050565b5f6104e76104e26104dd8461019e565b6104c4565b61019e565b9050919050565b5f819050919050565b610500836104cd565b61051461050c826104ee565b848454610473565b825550505050565b5f5f905090565b61052b61051c565b6105368184846104f7565b505050565b5f5b8281101561055c576105515f828401610523565b60018101905061053d565b505050565b601f8211156105af57828211156105ae5761057b81610446565b61058483610458565b61058d85610458565b602086101561059a575f90505b8083016105a98284038261053b565b505050505b5b505050565b5f82821c905092915050565b5f6105cf5f19846008026105b4565b1980831691505092915050565b5f6105e783836105c0565b9150826002028217905092915050565b610600826103df565b67ffffffffffffffff811115610619576106186101e9565b5b6106238254610416565b61062e828285610561565b5f60209050601f83116001811461065f575f841561064d578287015190505b61065785826105dc565b8655506106be565b601f19841661066d86610446565b5f5b828110156106945784890151825560018201915060208501945060208101905061066f565b868310156106b157848901516106ad601f8916826105c0565b8355505b6001600288020188555050505b505050505050565b5f82825260208201905092915050565b5f6106e0826103df565b6106ea81856106c6565b93506106fa818560208601610291565b610703816101d9565b840191505092915050565b5f6020820190508181035f83015261072681846106d6565b905092915050565b6107378161030d565b82525050565b5f6020820190506107505f83018461072e565b92915050565b61168e806107635f395ff3fe608060405234801561000f575f5ffd5b50600436106100cb575f3560e01c8063426a84931161008a578063598af9e711610064578063598af9e71461026b578063b6363cf21461029b578063f5298aca146102cb578063fe99049a146102e7576100cb565b8063426a8493146101db5780634e41a1fb1461020b578063558a72971461023b576100cb565b8062ad800c146100cf578062fdd58e146100ff57806301ffc9a71461012f578063095bcdb61461015f578063156e29f61461018f5780633f47e662146101ab575b5f5ffd5b6100e960048036038101906100e491906110fd565b610317565b6040516100f69190611198565b60405180910390f35b61011960048036038101906101149190611212565b6103ba565b604051610126919061125f565b60405180910390f35b610149600480360381019061014491906112cd565b61040f565b6040516101569190611312565b60405180910390f35b6101796004803603810190610174919061132b565b610488565b6040516101869190611312565b60405180910390f35b6101a960048036038101906101a4919061132b565b6104a7565b005b6101c560048036038101906101c091906110fd565b6104b7565b6040516101d29190611396565b60405180910390f35b6101f560048036038101906101f0919061132b565b6104e0565b6040516102029190611312565b60405180910390f35b610225600480360381019061022091906110fd565b6104ff565b6040516102329190611198565b60405180910390f35b610255600480360381019061025091906113d9565b6105a3565b6040516102629190611312565b60405180910390f35b61028560048036038101906102809190611417565b6105c0565b604051610292919061125f565b60405180910390f35b6102b560048036038101906102b09190611467565b610652565b6040516102c29190611312565b60405180910390f35b6102e560048036038101906102e0919061132b565b6106e0565b005b61030160048036038101906102fc91906114a5565b6106f0565b60405161030e9190611312565b60405180910390f35b606060035f8381526020019081526020015f205f01805461033790611536565b80601f016020809104026020016040519081016040528092919081815260200182805461036390611536565b80156103ae5780601f10610385576101008083540402835291602001916103ae565b820191905f5260205f20905b81548152906001019060200180831161039157829003601f168201915b50505050509050919050565b5f5f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8381526020019081526020015f2054905092915050565b5f7f0f632fb3000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916148061048157506104808261076a565b5b9050919050565b5f61049c6104946107d3565b8585856107da565b600190509392505050565b6104b28383836108cc565b505050565b5f60035f8381526020019081526020015f206002015f9054906101000a900460ff169050919050565b5f6104f46104ec6107d3565b85858561094d565b600190509392505050565b606060035f8381526020019081526020015f20600101805461052090611536565b80601f016020809104026020016040519081016040528092919081815260200182805461054c90611536565b80156105975780601f1061056e57610100808354040283529160200191610597565b820191905f5260205f20905b81548152906001019060200180831161057a57829003601f168201915b50505050509050919050565b5f6105b66105af6107d3565b8484610b25565b6001905092915050565b5f60025f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8381526020019081526020015f205490509392505050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff16905092915050565b6106eb838383610cfe565b505050565b5f5f6106fa6107d3565b90508073ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff161415801561073f575061073d8682610652565b155b156107515761075086828686610d7f565b5b61075d868686866107da565b6001915050949350505050565b5f7f01ffc9a7000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916149050919050565b5f33905090565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff160361084a575f6040517fa43520800000000000000000000000000000000000000000000000000000000081526004016108419190611575565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036108ba575f6040517fb8bbd6100000000000000000000000000000000000000000000000000000000081526004016108b19190611575565b60405180910390fd5b6108c684848484610e96565b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361093c575f6040517fb8bbd6100000000000000000000000000000000000000000000000000000000081526004016109339190611575565b60405180910390fd5b6109485f848484610e96565b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff16036109bd575f6040517fcc766a980000000000000000000000000000000000000000000000000000000081526004016109b49190611575565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610a2d575f6040517f6f65f465000000000000000000000000000000000000000000000000000000008152600401610a249190611575565b60405180910390fd5b8060025f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8481526020019081526020015f2081905550818373ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff167fb3fd5071835887567a0671151121894ddccc2842f1d10bedad13e0d17cace9a784604051610b17919061125f565b60405180910390a450505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610b95575f6040517fcc766a98000000000000000000000000000000000000000000000000000000008152600401610b8c9190611575565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610c05575f6040517f6f65f465000000000000000000000000000000000000000000000000000000008152600401610bfc9190611575565b60405180910390fd5b8060015f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff0219169083151502179055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fceb576d9f15e4e200fdb5096d64d5dfd667e16def20c1eefd14256d8e3faa26783604051610cf19190611312565b60405180910390a3505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610d6e575f6040517fa4352080000000000000000000000000000000000000000000000000000000008152600401610d659190611575565b60405180910390fd5b610d7a835f8484610e96565b505050565b5f610d8b8585856105c0565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff811015610e8f5781811015610e0057838183856040517f58a3fd5a000000000000000000000000000000000000000000000000000000008152600401610df7949392919061158e565b60405180910390fd5b81810360025f8773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8581526020019081526020015f20819055505b5050505050565b5f610e9f6107d3565b90505f73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff1614610fc2575f5f5f8773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8581526020019081526020015f2054905082811015610f6e57858184866040517fb1b4fec0000000000000000000000000000000000000000000000000000000008152600401610f65949392919061158e565b60405180910390fd5b8281035f5f8873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff161461105757815f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8581526020019081526020015f205f82825461104f91906115fe565b925050819055505b828473ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff167f1b3d7edb2e9c0b0e7c525b20aaaef0f5940d2ed71663c7d39266ecafac72885984866040516110b7929190611631565b60405180910390a45050505050565b5f5ffd5b5f819050919050565b6110dc816110ca565b81146110e6575f5ffd5b50565b5f813590506110f7816110d3565b92915050565b5f60208284031215611112576111116110c6565b5b5f61111f848285016110e9565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f61116a82611128565b6111748185611132565b9350611184818560208601611142565b61118d81611150565b840191505092915050565b5f6020820190508181035f8301526111b08184611160565b905092915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6111e1826111b8565b9050919050565b6111f1816111d7565b81146111fb575f5ffd5b50565b5f8135905061120c816111e8565b92915050565b5f5f60408385031215611228576112276110c6565b5b5f611235858286016111fe565b9250506020611246858286016110e9565b9150509250929050565b611259816110ca565b82525050565b5f6020820190506112725f830184611250565b92915050565b5f7fffffffff0000000000000000000000000000000000000000000000000000000082169050919050565b6112ac81611278565b81146112b6575f5ffd5b50565b5f813590506112c7816112a3565b92915050565b5f602082840312156112e2576112e16110c6565b5b5f6112ef848285016112b9565b91505092915050565b5f8115159050919050565b61130c816112f8565b82525050565b5f6020820190506113255f830184611303565b92915050565b5f5f5f60608486031215611342576113416110c6565b5b5f61134f868287016111fe565b9350506020611360868287016110e9565b9250506040611371868287016110e9565b9150509250925092565b5f60ff82169050919050565b6113908161137b565b82525050565b5f6020820190506113a95f830184611387565b92915050565b6113b8816112f8565b81146113c2575f5ffd5b50565b5f813590506113d3816113af565b92915050565b5f5f604083850312156113ef576113ee6110c6565b5b5f6113fc858286016111fe565b925050602061140d858286016113c5565b9150509250929050565b5f5f5f6060848603121561142e5761142d6110c6565b5b5f61143b868287016111fe565b935050602061144c868287016111fe565b925050604061145d868287016110e9565b9150509250925092565b5f5f6040838503121561147d5761147c6110c6565b5b5f61148a858286016111fe565b925050602061149b858286016111fe565b9150509250929050565b5f5f5f5f608085870312156114bd576114bc6110c6565b5b5f6114ca878288016111fe565b94505060206114db878288016111fe565b93505060406114ec878288016110e9565b92505060606114fd878288016110e9565b91505092959194509250565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061154d57607f821691505b6020821081036115605761155f611509565b5b50919050565b61156f816111d7565b82525050565b5f6020820190506115885f830184611566565b92915050565b5f6080820190506115a15f830187611566565b6115ae6020830186611250565b6115bb6040830185611250565b6115c86060830184611250565b95945050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611608826110ca565b9150611613836110ca565b925082820190508082111561162b5761162a6115d1565b5b92915050565b5f6040820190506116445f830185611566565b6116516020830184611250565b939250505056fea26469706673582212209537e0a5290687f730731a9834842827affca3bf1cd80ae1bb37f2189f03e8ee64736f6c63430008230033"

type ERC6909Suite struct {
	suite.Suite

	Enable               bool
	RPCURL               string
	PrivateKeyHex        string
	SpenderPrivateKeyHex string
	ToHex                string
	MulticallHex         string
	NativeSymbol         string
	ByteCode             string
	Constructor          *abi.Function
	Args                 []any
	Salt                 *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client        rpc.Client
	signer        core.Signer
	spenderSigner core.Signer
	deployer      *deploy.NicksDeployer
	from          *types.Address
	contract      *types.Address
	spender       *types.Address
	to            *types.Address
	multicall     *types.Address

	erc6909 *ERC6909

	id     *big.Int
	amount *big.Int

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestERC6909Suite(t *testing.T) {
	suite.Run(t, &ERC6909Suite{
		Enable:               false,
		RPCURL:               "",
		PrivateKeyHex:        "",
		SpenderPrivateKeyHex: "",
		ToHex:                "",
		MulticallHex:         multicall3.DefaultMulticall3Address.String(),
		NativeSymbol:         "",
		ByteCode:             erc6909ByteCodeHex,
		Constructor:          abi.NewFunction("constructor", abi.Types{abi.Uint256, abi.String, abi.String, abi.Uint8}, nil),
		Args:                 []any{big.NewInt(1), "evmkit test token", "EVK6909", uint8(18)},
		Salt:                 core.Keccak256([]byte("hello")),
	})
}

func (s *ERC6909Suite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.erc6909 = New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	s.id = big.NewInt(1)
	s.amount = big.NewInt(100)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	spenderKey, err := types.NewPrivateKeyFromHex(s.SpenderPrivateKeyHex)
	s.Require().NoError(err)
	s.spenderSigner = spenderKey
	s.spender = core.PubkeyToAddress(spenderKey.PublicKey())

	s.to, err = types.NewAddressFromHex(s.ToHex)
	s.Require().NoError(err)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	var salt [32]byte
	copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())).Bytes())

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

	contract, err := s.deployer.DeployAddress(deploy.DefaultNicksMethodFactory, deployment)
	s.Require().NoError(err)
	s.contract = contract

	s.Require().Eventually(func() bool {
		var codeHex string
		if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", contract.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
			return false
		}
		code, err := encoding.Hex.Decode(codeHex)
		return err == nil && len(code) > 0
	}, 10*time.Second, 500*time.Millisecond, "no code visible at deployed contract %s", contract)

	s.T().Logf("SetupSuite: deployed contract %s (nonce %d, tx %s)", contract, unsigned.AddressNonce(), hash)

	mintFn := abi.NewFunction("mint", abi.Types{abi.Address, abi.Uint256, abi.Uint256}, nil)
	calldata, err := mintFn.EncodeCall(s.from, s.id, s.amount)
	s.Require().NoError(err)

	var unsignedMint core.Transaction
	s.Require().Eventually(func() bool {
		u, err := tx.NewDynamicFeeBuilder().Build(s.from, contract, big.NewInt(0), calldata).Pack(s.ctx, s.client)
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

	s.T().Logf("SetupSuite: minted id %s (amount %s) of %s to %s (tx %s)", s.id, s.amount, contract, s.from, mintHash)
}

func (s *ERC6909Suite) TearDownSuite() {
	if s.contract != nil {
		burnFn := abi.NewFunction("burn", abi.Types{abi.Address, abi.Uint256, abi.Uint256}, nil)

		for _, holder := range []*types.Address{s.from, s.to} {
			balance, err := s.erc6909.BalanceOf(s.ctx, s.contract, holder, s.id, rpc.BlockTagLatest)
			s.Require().NoError(err)
			if balance.Sign() == 0 {
				continue
			}

			calldata, err := burnFn.EncodeCall(holder, s.id, balance)
			s.Require().NoError(err)

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, s.contract, big.NewInt(0), calldata).Pack(s.ctx, s.client)
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

			s.T().Logf("TearDownSuite: burned id %s (amount %s) of %s from %s (tx %s)", s.id, balance, s.contract, holder, burnHash)
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

		s.T().Logf("ERC6909Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC6909Suite) Test01ReadMethods() {
	supports, err := s.erc6909.SupportsInterface(s.ctx, s.contract, interfaceIDERC6909, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(supports)

	fromBalance, err := s.erc6909.BalanceOf(s.ctx, s.contract, s.from, s.id, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.amount.Cmp(fromBalance))

	toBalance, err := s.erc6909.BalanceOf(s.ctx, s.contract, s.to, s.id, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(0).Cmp(toBalance))

	allowance, err := s.erc6909.Allowance(s.ctx, s.contract, s.from, s.to, s.id, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(0).Cmp(allowance))

	isOperator, err := s.erc6909.IsOperator(s.ctx, s.contract, s.from, s.to, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(isOperator)

	name, err := s.erc6909.Name(s.ctx, s.contract, s.id, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Equal("evmkit test token", name)

	symbol, err := s.erc6909.Symbol(s.ctx, s.contract, s.id, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Equal("EVK6909", symbol)

	decimals, err := s.erc6909.Decimals(s.ctx, s.contract, s.id, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Equal(uint8(18), decimals)

	s.T().Logf("ERC6909 %s Test01ReadMethods Result", s.contract)
	s.T().Logf("\tSupports: %v", supports)
	s.T().Logf("\tBalanceOf(id %s)", s.id)
	s.T().Logf("\t\tFrom(%s): %s", s.from, fromBalance)
	s.T().Logf("\t\tTo(%s): %s", s.to, toBalance)
	s.T().Logf("\tAllowance(from, to, id %s): %s", s.id, allowance)
	s.T().Logf("\tIsOperator(from, to): %v", isOperator)
	s.T().Logf("\tName(id %s): %s", s.id, name)
	s.T().Logf("\tSymbol(id %s): %s", s.id, symbol)
	s.T().Logf("\tDecimals(id %s): %d", s.id, decimals)
}

func (s *ERC6909Suite) Test02SetOperator() {
	unsigned, err := s.erc6909.CreateSetOperator(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.to, true)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	hash, err := s.erc6909.SetOperator(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	operatorSets, err := ExtractOperatorSets(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(operatorSets, 1)
	s.Require().True(operatorSets[0].Contract.Equal(s.contract))
	s.Require().True(operatorSets[0].Owner.Equal(s.from))
	s.Require().True(operatorSets[0].Operator.Equal(s.to))
	s.Require().True(operatorSets[0].Approved)

	var after bool
	s.Require().Eventually(func() bool {
		after, err = s.erc6909.IsOperator(s.ctx, s.contract, s.from, s.to, rpc.BlockTagLatest)
		return err == nil && after
	}, 10*time.Second, 500*time.Millisecond, "operator not visible after setOperator")

	s.T().Logf("ERC6909 %s Test02SetOperator Result", s.contract)
	s.T().Logf("\tOwner: %s", s.from)
	s.T().Logf("\tOperator: %s", s.to)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", unsigned.Gas())
	s.T().Logf("\tNonce: %d", unsigned.AddressNonce())
	s.T().Logf("\tMaxFee: %s", unsigned.MaxFee())
	s.T().Logf("\tIsOperator")
	s.T().Logf("\t\tAfter: %v", after)

	revertUnsigned, err := s.erc6909.CreateSetOperator(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.to, false)
	s.Require().NoError(err)
	s.Require().NoError(revertUnsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, revertUnsigned.MaxFee())
	s.mu.Unlock()

	revertHash, err := s.erc6909.SetOperator(s.ctx, revertUnsigned)
	s.Require().NoError(err)

	revertReceipt, err := tx.WaitForReceipt(s.ctx, s.client, revertHash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(revertReceipt.Status, "revert transaction reverted")

	revertOperatorSets, err := ExtractOperatorSets(revertReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(revertOperatorSets, 1)
	s.Require().False(revertOperatorSets[0].Approved)

	var reverted bool
	s.Require().Eventually(func() bool {
		reverted, err = s.erc6909.IsOperator(s.ctx, s.contract, s.from, s.to, rpc.BlockTagLatest)
		return err == nil && !reverted
	}, 10*time.Second, 500*time.Millisecond, "operator not reverted to false")

	s.T().Logf("ERC6909 %s Test02SetOperator Revert Result", s.contract)
	s.T().Logf("\tRevertTx: %s", revertHash)
	s.T().Logf("\t\tMaxFee: %s", revertUnsigned.MaxFee())
	s.T().Logf("\t\tReverted: %v", reverted)
}

func (s *ERC6909Suite) Test03Transfer() {
	half := new(big.Int).Div(s.amount, big.NewInt(2))

	unsigned, err := s.erc6909.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.to, s.id, half)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	hash, err := s.erc6909.Transfer(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	transfers, err := ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.contract))
	s.Require().True(transfers[0].Caller.Equal(s.from))
	s.Require().True(transfers[0].Sender.Equal(s.from))
	s.Require().True(transfers[0].Receiver.Equal(s.to))
	s.Require().Zero(s.id.Cmp(transfers[0].Id))
	s.Require().Zero(half.Cmp(transfers[0].Amount))

	var toBalance *big.Int
	s.Require().Eventually(func() bool {
		toBalance, err = s.erc6909.BalanceOf(s.ctx, s.contract, s.to, s.id, rpc.BlockTagLatest)
		return err == nil && half.Cmp(toBalance) == 0
	}, 10*time.Second, 500*time.Millisecond, "balance not visible after transfer")

	fromBalance, err := s.erc6909.BalanceOf(s.ctx, s.contract, s.from, s.id, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(half.Cmp(fromBalance))

	s.T().Logf("ERC6909 %s Test03Transfer Result", s.contract)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.to)
	s.T().Logf("\tId: %s", s.id)
	s.T().Logf("\tAmount: %s", half)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", unsigned.Gas())
	s.T().Logf("\tMaxFee: %s", unsigned.MaxFee())
	s.T().Logf("\tBalanceOf")
	s.T().Logf("\t\tFrom: %s", fromBalance)
	s.T().Logf("\t\tTo: %s", toBalance)
}

func (s *ERC6909Suite) Test04Approve() {
	half := new(big.Int).Div(s.amount, big.NewInt(2))

	unsigned, err := s.erc6909.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.spender, s.id, half)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	hash, err := s.erc6909.Approve(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	approvals, err := ExtractApprovals(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(approvals, 1)
	s.Require().True(approvals[0].Contract.Equal(s.contract))
	s.Require().True(approvals[0].Owner.Equal(s.from))
	s.Require().True(approvals[0].Spender.Equal(s.spender))
	s.Require().Zero(s.id.Cmp(approvals[0].Id))
	s.Require().Zero(half.Cmp(approvals[0].Amount))

	var after *big.Int
	s.Require().Eventually(func() bool {
		after, err = s.erc6909.Allowance(s.ctx, s.contract, s.from, s.spender, s.id, rpc.BlockTagLatest)
		return err == nil && half.Cmp(after) == 0
	}, 10*time.Second, 500*time.Millisecond, "allowance not visible after approve")

	s.T().Logf("ERC6909 %s Test04Approve Result", s.contract)
	s.T().Logf("\tOwner: %s", s.from)
	s.T().Logf("\tSpender: %s", s.spender)
	s.T().Logf("\tId: %s", s.id)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", unsigned.Gas())
	s.T().Logf("\tNonce: %d", unsigned.AddressNonce())
	s.T().Logf("\tMaxFee: %s", unsigned.MaxFee())
	s.T().Logf("\tAllowance")
	s.T().Logf("\t\tAfter: %s", after)
}

func (s *ERC6909Suite) Test05TransferFrom() {
	half := new(big.Int).Div(s.amount, big.NewInt(2))

	unsigned, err := s.erc6909.CreateTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.spender, s.from, s.to, s.id, half)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.spenderSigner))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()

	hash, err := s.erc6909.TransferFrom(s.ctx, unsigned)
	s.Require().NoError(err)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 0, 0)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "transaction reverted")

	transfers, err := ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(transfers, 1)
	s.Require().True(transfers[0].Contract.Equal(s.contract))
	s.Require().True(transfers[0].Caller.Equal(s.spender))
	s.Require().True(transfers[0].Sender.Equal(s.from))
	s.Require().True(transfers[0].Receiver.Equal(s.to))
	s.Require().Zero(s.id.Cmp(transfers[0].Id))
	s.Require().Zero(half.Cmp(transfers[0].Amount))

	var allowanceAfter *big.Int
	s.Require().Eventually(func() bool {
		allowanceAfter, err = s.erc6909.Allowance(s.ctx, s.contract, s.from, s.spender, s.id, rpc.BlockTagLatest)
		return err == nil && big.NewInt(0).Cmp(allowanceAfter) == 0
	}, 10*time.Second, 500*time.Millisecond, "allowance not spent after transferFrom")

	fromBalance, err := s.erc6909.BalanceOf(s.ctx, s.contract, s.from, s.id, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(big.NewInt(0).Cmp(fromBalance))

	toBalance, err := s.erc6909.BalanceOf(s.ctx, s.contract, s.to, s.id, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.amount.Cmp(toBalance))

	s.T().Logf("ERC6909 %s Test05TransferFrom Result", s.contract)
	s.T().Logf("\tSpender: %s", s.spender)
	s.T().Logf("\tFrom: %s", s.from)
	s.T().Logf("\tTo: %s", s.to)
	s.T().Logf("\tId: %s", s.id)
	s.T().Logf("\tAmount: %s", half)
	s.T().Logf("\tTx: %s", hash)
	s.T().Logf("\tGas: %d", unsigned.Gas())
	s.T().Logf("\tMaxFee: %s", unsigned.MaxFee())
	s.T().Logf("\tAllowance")
	s.T().Logf("\t\tAfter: %s", allowanceAfter)
	s.T().Logf("\tBalanceOf")
	s.T().Logf("\t\tFrom: %s", fromBalance)
	s.T().Logf("\t\tTo: %s", toBalance)
}

func (s *ERC6909Suite) Test06TransferInvalidReceiver() {
	_, err := s.erc6909.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, types.ZeroAddress(), s.id, big.NewInt(1))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidReceiver)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC6909 %s Test06TransferInvalidReceiver Result", s.contract)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC6909Suite) Test07ApproveInvalidSpender() {
	_, err := s.erc6909.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, types.ZeroAddress(), s.id, s.amount)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInvalidSpender)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC6909 %s Test07ApproveInvalidSpender Result", s.contract)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC6909Suite) Test08TransferInsufficientBalance() {
	_, err := s.erc6909.CreateTransfer(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.from, s.to, s.id, big.NewInt(1))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientBalance)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC6909 %s Test08TransferInsufficientBalance Result", s.contract)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}

func (s *ERC6909Suite) Test09TransferFromInsufficientAllowance() {
	_, err := s.erc6909.CreateTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.contract, s.spender, s.from, s.to, s.id, big.NewInt(1))
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrInsufficientAllowance)

	var rpcErr *rpc.ResponseError
	s.Require().ErrorAs(err, &rpcErr)
	s.Require().NotEmpty(rpcErr.Data)

	s.T().Logf("ERC6909 %s Test09TransferFromInsufficientAllowance Result", s.contract)
	s.T().Logf("\tError: %v", err)
	s.T().Logf("\tRevertData: %s", rpcErr.Data)
}
