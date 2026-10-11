package erc721

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/contracts/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/contracts/multicall3"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var interfaceIDERC721 = [4]byte{0x80, 0xac, 0x58, 0xcd}

// erc721ByteCodeHex is OpenZeppelin's ERC721 plus mint/burn, constructor(name, symbol).
const erc721ByteCodeHex = "0x608060405234801561000f575f5ffd5b50604051612452380380612452833981810160405281019061003191906101a8565b8181815f9081610041919061043f565b508060019081610051919061043f565b505050505061050e565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6100ba82610074565b810181811067ffffffffffffffff821117156100d9576100d8610084565b5b80604052505050565b5f6100eb61005b565b90506100f782826100b1565b919050565b5f67ffffffffffffffff82111561011657610115610084565b5b61011f82610074565b9050602081019050919050565b8281835e5f83830152505050565b5f61014c610147846100fc565b6100e2565b90508281526020810184848401111561016857610167610070565b5b61017384828561012c565b509392505050565b5f82601f83011261018f5761018e61006c565b5b815161019f84826020860161013a565b91505092915050565b5f5f604083850312156101be576101bd610064565b5b5f83015167ffffffffffffffff8111156101db576101da610068565b5b6101e78582860161017b565b925050602083015167ffffffffffffffff81111561020857610207610068565b5b6102148582860161017b565b9150509250929050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061026c57607f821691505b60208210810361027f5761027e610228565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026102e17fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826102a6565b6102eb86836102a6565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61032f61032a61032584610303565b61030c565b610303565b9050919050565b5f819050919050565b61034883610315565b61035c61035482610336565b8484546102b2565b825550505050565b5f5f905090565b610373610364565b61037e81848461033f565b505050565b5f5b828110156103a4576103995f82840161036b565b600181019050610385565b505050565b601f8211156103f757828211156103f6576103c381610285565b6103cc83610297565b6103d585610297565b60208610156103e2575f90505b8083016103f182840382610383565b505050505b5b505050565b5f82821c905092915050565b5f6104175f19846008026103fc565b1980831691505092915050565b5f61042f8383610408565b9150826002028217905092915050565b6104488261021e565b67ffffffffffffffff81111561046157610460610084565b5b61046b8254610255565b6104768282856103a9565b5f60209050601f8311600181146104a7575f8415610495578287015190505b61049f8582610424565b865550610506565b601f1984166104b586610285565b5f5b828110156104dc578489015182556001820191506020850194506020810190506104b7565b868310156104f957848901516104f5601f891682610408565b8355505b6001600288020188555050505b505050505050565b611f378061051b5f395ff3fe608060405234801561000f575f5ffd5b5060043610610109575f3560e01c80636352211e116100a057806395d89b411161006f57806395d89b41146102c5578063a22cb465146102e3578063b88d4fde146102ff578063c87b56dd1461031b578063e985e9c51461034b57610109565b80636352211e146102175780636a6278421461024757806370a082311461027757806375794a3c146102a757610109565b806323b872dd116100dc57806323b872dd146101a757806340c10f19146101c357806342842e0e146101df57806342966c68146101fb57610109565b806301ffc9a71461010d57806306fdde031461013d578063081812fc1461015b578063095ea7b31461018b575b5f5ffd5b61012760048036038101906101229190611717565b61037b565b604051610134919061175c565b60405180910390f35b61014561045c565b60405161015291906117e5565b60405180910390f35b61017560048036038101906101709190611838565b6104eb565b60405161018291906118a2565b60405180910390f35b6101a560048036038101906101a091906118e5565b610506565b005b6101c160048036038101906101bc9190611923565b61051c565b005b6101dd60048036038101906101d891906118e5565b61061b565b005b6101f960048036038101906101f49190611923565b610646565b005b61021560048036038101906102109190611838565b610665565b005b610231600480360381019061022c9190611838565b610671565b60405161023e91906118a2565b60405180910390f35b610261600480360381019061025c9190611973565b610682565b60405161026e91906119ad565b60405180910390f35b610291600480360381019061028c9190611973565b6106aa565b60405161029e91906119ad565b60405180910390f35b6102af610760565b6040516102bc91906119ad565b60405180910390f35b6102cd610766565b6040516102da91906117e5565b60405180910390f35b6102fd60048036038101906102f891906119f0565b6107f6565b005b61031960048036038101906103149190611b5a565b61080c565b005b61033560048036038101906103309190611838565b610831565b60405161034291906117e5565b60405180910390f35b61036560048036038101906103609190611bda565b610897565b604051610372919061175c565b60405180910390f35b5f7f80ac58cd000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916148061044557507f5b5e139f000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916145b80610455575061045482610925565b5b9050919050565b60605f805461046a90611c45565b80601f016020809104026020016040519081016040528092919081815260200182805461049690611c45565b80156104e15780601f106104b8576101008083540402835291602001916104e1565b820191905f5260205f20905b8154815290600101906020018083116104c457829003601f168201915b5050505050905090565b5f6104f58261098e565b506104ff82610a14565b9050919050565b6105188282610513610a4d565b610a54565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361058c575f6040517f64a0ae9200000000000000000000000000000000000000000000000000000000815260040161058391906118a2565b60405180910390fd5b5f61059f838361059a610a4d565b610a66565b90508373ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614610615578382826040517f64283d7b00000000000000000000000000000000000000000000000000000000815260040161060c93929190611c75565b60405180910390fd5b50505050565b6106258282610c71565b60065481106106425760018161063b9190611cd7565b6006819055505b5050565b61066083838360405180602001604052805f81525061080c565b505050565b61066e81610d64565b50565b5f61067b8261098e565b9050919050565b5f60065f81548092919061069590611d0a565b9190505590506106a58282610c71565b919050565b5f5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361071b575f6040517f89c62b6400000000000000000000000000000000000000000000000000000000815260040161071291906118a2565b60405180910390fd5b60035f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b60065481565b60606001805461077590611c45565b80601f01602080910402602001604051908101604052809291908181526020018280546107a190611c45565b80156107ec5780601f106107c3576101008083540402835291602001916107ec565b820191905f5260205f20905b8154815290600101906020018083116107cf57829003601f168201915b5050505050905090565b610808610801610a4d565b8383610de6565b5050565b61081784848461051c565b61082b610822610a4d565b85858585610f4f565b50505050565b606061083c8261098e565b505f6108466110fb565b90505f8151116108645760405180602001604052805f81525061088f565b8061086e8461111b565b60405160200161087f929190611d8b565b6040516020818303038152906040525b915050919050565b5f60055f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff16905092915050565b5f7f01ffc9a7000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916149050919050565b5f5f610999836111e5565b90505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603610a0b57826040517f7e273289000000000000000000000000000000000000000000000000000000008152600401610a0291906119ad565b60405180910390fd5b80915050919050565b5f60045f8381526020019081526020015f205f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050919050565b5f33905090565b610a61838383600161121e565b505050565b5f5f610a71846111e5565b90505f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1614610ab257610ab18184866113dd565b5b5f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614610b3d57610af15f855f5f61121e565b600160035f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825403925050819055505b5f73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff1614610bbc57600160035f8773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8460025f8681526020019081526020015f205f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550838573ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef60405160405180910390a4809150509392505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610ce1575f6040517f64a0ae92000000000000000000000000000000000000000000000000000000008152600401610cd891906118a2565b60405180910390fd5b5f610ced83835f610a66565b90505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614610d5f575f6040517f73c6ac6e000000000000000000000000000000000000000000000000000000008152600401610d5691906118a2565b60405180910390fd5b505050565b5f610d705f835f610a66565b90505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603610de257816040517f7e273289000000000000000000000000000000000000000000000000000000008152600401610dd991906119ad565b60405180910390fd5b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610e5657816040517f5b08ba18000000000000000000000000000000000000000000000000000000008152600401610e4d91906118a2565b60405180910390fd5b8060055f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff0219169083151502179055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c3183604051610f42919061175c565b60405180910390a3505050565b5f8373ffffffffffffffffffffffffffffffffffffffff163b11156110f4578273ffffffffffffffffffffffffffffffffffffffff1663150b7a02868685856040518563ffffffff1660e01b8152600401610fad9493929190611e00565b6020604051808303815f875af1925050508015610fe857506040513d601f19601f82011682018060405250810190610fe59190611e5e565b60015b611069573d805f8114611016576040519150601f19603f3d011682016040523d82523d5f602084013e61101b565b606091505b505f81510361106157836040517f64a0ae9200000000000000000000000000000000000000000000000000000000815260040161105891906118a2565b60405180910390fd5b805160208201fd5b63150b7a0260e01b7bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916817bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916146110f257836040517f64a0ae920000000000000000000000000000000000000000000000000000000081526004016110e991906118a2565b60405180910390fd5b505b5050505050565b6060604051806060016040528060248152602001611ede60249139905090565b60605f6001611129846114a0565b0190505f8167ffffffffffffffff81111561114757611146611a36565b5b6040519080825280601f01601f1916602001820160405280156111795781602001600182028036833780820191505090505b5090505f82602083010190505b6001156111da578080600190039150507f3031323334353637383961626364656600000000000000000000000000000000600a86061a8153600a85816111cf576111ce611e89565b5b0494505f8503611186575b819350505050919050565b5f60025f8381526020019081526020015f205f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050919050565b808061125657505f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1614155b15611388575f6112658461098e565b90505f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16141580156112cf57508273ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614155b80156112e257506112e08184610897565b155b1561132457826040517fa9fbf51f00000000000000000000000000000000000000000000000000000000815260040161131b91906118a2565b60405180910390fd5b811561138657838573ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92560405160405180910390a45b505b8360045f8581526020019081526020015f205f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050505050565b6113e88383836115f1565b61149b575f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361145c57806040517f7e27328900000000000000000000000000000000000000000000000000000000815260040161145391906119ad565b60405180910390fd5b81816040517f177e802f000000000000000000000000000000000000000000000000000000008152600401611492929190611eb6565b60405180910390fd5b505050565b5f5f5f90507a184f03e93ff9f4daa797ed6e38ed64bf6a1f01000000000000000083106114fc577a184f03e93ff9f4daa797ed6e38ed64bf6a1f01000000000000000083816114f2576114f1611e89565b5b0492506040810190505b6d04ee2d6d415b85acef81000000008310611539576d04ee2d6d415b85acef8100000000838161152f5761152e611e89565b5b0492506020810190505b662386f26fc10000831061156857662386f26fc10000838161155e5761155d611e89565b5b0492506010810190505b6305f5e1008310611591576305f5e100838161158757611586611e89565b5b0492506008810190505b61271083106115b65761271083816115ac576115ab611e89565b5b0492506004810190505b606483106115d957606483816115cf576115ce611e89565b5b0492506002810190505b600a83106115e8576001810190505b80915050919050565b5f5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16141580156116a857508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff16148061166957506116688484610897565b5b806116a757508273ffffffffffffffffffffffffffffffffffffffff1661168f83610a14565b73ffffffffffffffffffffffffffffffffffffffff16145b5b90509392505050565b5f604051905090565b5f5ffd5b5f5ffd5b5f7fffffffff0000000000000000000000000000000000000000000000000000000082169050919050565b6116f6816116c2565b8114611700575f5ffd5b50565b5f81359050611711816116ed565b92915050565b5f6020828403121561172c5761172b6116ba565b5b5f61173984828501611703565b91505092915050565b5f8115159050919050565b61175681611742565b82525050565b5f60208201905061176f5f83018461174d565b92915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f6117b782611775565b6117c1818561177f565b93506117d181856020860161178f565b6117da8161179d565b840191505092915050565b5f6020820190508181035f8301526117fd81846117ad565b905092915050565b5f819050919050565b61181781611805565b8114611821575f5ffd5b50565b5f813590506118328161180e565b92915050565b5f6020828403121561184d5761184c6116ba565b5b5f61185a84828501611824565b91505092915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61188c82611863565b9050919050565b61189c81611882565b82525050565b5f6020820190506118b55f830184611893565b92915050565b6118c481611882565b81146118ce575f5ffd5b50565b5f813590506118df816118bb565b92915050565b5f5f604083850312156118fb576118fa6116ba565b5b5f611908858286016118d1565b925050602061191985828601611824565b9150509250929050565b5f5f5f6060848603121561193a576119396116ba565b5b5f611947868287016118d1565b9350506020611958868287016118d1565b925050604061196986828701611824565b9150509250925092565b5f60208284031215611988576119876116ba565b5b5f611995848285016118d1565b91505092915050565b6119a781611805565b82525050565b5f6020820190506119c05f83018461199e565b92915050565b6119cf81611742565b81146119d9575f5ffd5b50565b5f813590506119ea816119c6565b92915050565b5f5f60408385031215611a0657611a056116ba565b5b5f611a13858286016118d1565b9250506020611a24858286016119dc565b9150509250929050565b5f5ffd5b5f5ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b611a6c8261179d565b810181811067ffffffffffffffff82111715611a8b57611a8a611a36565b5b80604052505050565b5f611a9d6116b1565b9050611aa98282611a63565b919050565b5f67ffffffffffffffff821115611ac857611ac7611a36565b5b611ad18261179d565b9050602081019050919050565b828183375f83830152505050565b5f611afe611af984611aae565b611a94565b905082815260208101848484011115611b1a57611b19611a32565b5b611b25848285611ade565b509392505050565b5f82601f830112611b4157611b40611a2e565b5b8135611b51848260208601611aec565b91505092915050565b5f5f5f5f60808587031215611b7257611b716116ba565b5b5f611b7f878288016118d1565b9450506020611b90878288016118d1565b9350506040611ba187828801611824565b925050606085013567ffffffffffffffff811115611bc257611bc16116be565b5b611bce87828801611b2d565b91505092959194509250565b5f5f60408385031215611bf057611bef6116ba565b5b5f611bfd858286016118d1565b9250506020611c0e858286016118d1565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680611c5c57607f821691505b602082108103611c6f57611c6e611c18565b5b50919050565b5f606082019050611c885f830186611893565b611c95602083018561199e565b611ca26040830184611893565b949350505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611ce182611805565b9150611cec83611805565b9250828201905080821115611d0457611d03611caa565b5b92915050565b5f611d1482611805565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203611d4657611d45611caa565b5b600182019050919050565b5f81905092915050565b5f611d6582611775565b611d6f8185611d51565b9350611d7f81856020860161178f565b80840191505092915050565b5f611d968285611d5b565b9150611da28284611d5b565b91508190509392505050565b5f81519050919050565b5f82825260208201905092915050565b5f611dd282611dae565b611ddc8185611db8565b9350611dec81856020860161178f565b611df58161179d565b840191505092915050565b5f608082019050611e135f830187611893565b611e206020830186611893565b611e2d604083018561199e565b8181036060830152611e3f8184611dc8565b905095945050505050565b5f81519050611e58816116ed565b92915050565b5f60208284031215611e7357611e726116ba565b5b5f611e8084828501611e4a565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601260045260245ffd5b5f604082019050611ec95f830185611893565b611ed6602083018461199e565b939250505056fe68747470733a2f2f65766d6b69742d66697874757265732e746573742f6572633732312fa2646970667358221220e20a1ef0e04f61cf6dffab4e6246cc65315b37433e63f53cecb31a0e6209b8ee64736f6c63430008230033"

type ERC721Suite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	ToHex         string
	MulticallHex  string
	NativeSymbol  string
	NFTCount      int
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
	contracts []*types.Address
	to        *types.Address
	multicall *types.Address

	erc721 *ERC721

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestERC721Suite(t *testing.T) {
	suite.Run(t, &ERC721Suite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		ToHex:         "",
		MulticallHex:  multicall3.DefaultMulticall3Address.String(),
		NativeSymbol:  "",
		NFTCount:      5,
		ByteCode:      erc721ByteCodeHex,
		Constructor:   abi.NewFunction("constructor", abi.Types{abi.String, abi.String}, nil),
		Args:          []any{"evmkit test NFT", "EVKNFT"},
		Salt:          core.Keccak256([]byte("hello")),
	})
}

func (s *ERC721Suite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.erc721 = New(s.client)

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

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	mintFn := abi.NewFunction("mint", abi.Types{abi.Address, abi.Uint256}, nil)
	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	s.contracts = make([]*types.Address, s.NFTCount)
	for i := 0; i < s.NFTCount; i++ {
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

		contract, err := s.deployer.DeployAddress(deploy.DefaultNicksMethodFactory, deployment)
		s.Require().NoError(err)
		s.contracts[i] = contract

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", contract.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 10*time.Second, 500*time.Millisecond, "no code visible at deployed contract %s", contract)

		s.T().Logf("SetupSuite: deployed contract %d/%d %s (nonce %d, tx %s)", i+1, s.NFTCount, contract, unsigned.AddressNonce(), hash)

		for _, tokenId := range []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)} {
			calldata, err := mintFn.EncodeCall(s.from, tokenId)
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

			s.T().Logf("SetupSuite: minted tokenId %s of %s to %s (tx %s)", tokenId, contract, s.from, mintHash)
		}
	}
}

func (s *ERC721Suite) TearDownSuite() {
	if s.contracts != nil {
		burnFn := abi.NewFunction("burn", abi.Types{abi.Uint256}, nil)

		for _, contract := range s.contracts {
			for _, tokenId := range []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)} {
				calldata, err := burnFn.EncodeCall(tokenId)
				s.Require().NoError(err)

				unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, contract, big.NewInt(0), calldata).Pack(s.ctx, s.client)
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

				s.T().Logf("TearDownSuite: burned tokenId %s of %s (tx %s)", tokenId, contract, burnHash)
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

		s.T().Logf("ERC721Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC721Suite) TestReadMethods() {
	tokenId := big.NewInt(0)

	for _, contract := range s.contracts {
		supports, err := s.erc721.SupportsInterface(s.ctx, contract, interfaceIDERC721, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(supports)

		name, err := s.erc721.Name(s.ctx, contract, rpc.BlockTagLatest)
		s.Require().NoError(err)

		symbol, err := s.erc721.Symbol(s.ctx, contract, rpc.BlockTagLatest)
		s.Require().NoError(err)

		tokenURI, err := s.erc721.TokenURI(s.ctx, contract, tokenId, rpc.BlockTagLatest)
		s.Require().NoError(err)

		fromBalance, err := s.erc721.BalanceOf(s.ctx, contract, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		toBalance, err := s.erc721.BalanceOf(s.ctx, contract, s.to, rpc.BlockTagLatest)
		s.Require().NoError(err)

		owner, err := s.erc721.OwnerOf(s.ctx, contract, tokenId, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(owner.Equal(s.from))

		approved, err := s.erc721.GetApproved(s.ctx, contract, tokenId, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(approved.Equal(s.to))

		isApprovedForAll, err := s.erc721.IsApprovedForAll(s.ctx, contract, s.from, s.to, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.T().Logf("ERC721 %s TestReadMethods Result", contract)
		s.T().Logf("\tSupports: %v", supports)
		s.T().Logf("\tName: %s", name)
		s.T().Logf("\tSymbol: %s", symbol)
		s.T().Logf("\tTokenURI(%s): %s", tokenId, tokenURI)
		s.T().Logf("\tBalanceOf")
		s.T().Logf("\t\tFrom(%s): %s", s.from, fromBalance)
		s.T().Logf("\t\tTo(%s): %s", s.to, toBalance)
		s.T().Logf("\tOwnerOf(%s): %s", tokenId, owner)
		s.T().Logf("\tGetApproved(%s): %s", tokenId, approved)
		s.T().Logf("\tIsApprovedForAll(from, to): %v", isApprovedForAll)
	}
}

func (s *ERC721Suite) TestApprove() {
	tokenId := big.NewInt(0)

	first, err := s.erc721.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), s.contracts[0], s.from, s.to, tokenId)
	s.Require().NoError(err)
	baseNonce := first.AddressNonce()

	unsigned := make([]core.Transaction, len(s.contracts))
	unsigned[0] = first
	for i := 1; i < len(s.contracts); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(baseNonce + uint64(i))

		u, err := s.erc721.CreateApprove(s.ctx, builder, s.contracts[i], s.from, s.to, tokenId)
		s.Require().NoError(err)
		unsigned[i] = u
	}

	hashes := make([]*types.Hash, len(s.contracts))
	for i, u := range unsigned {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc721.Approve(s.ctx, u)
		s.Require().NoError(err)
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		approvals, err := ExtractApprovals(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(approvals, 1)
		s.Require().True(approvals[0].Contract.Equal(contract))
		s.Require().True(approvals[0].Owner.Equal(s.from))
		s.Require().True(approvals[0].Approved.Equal(s.to))
		s.Require().Zero(tokenId.Cmp(approvals[0].TokenId))

		var after *types.Address
		s.Require().Eventually(func() bool {
			after, err = s.erc721.GetApproved(s.ctx, contract, tokenId, rpc.BlockTagLatest)
			return err == nil && after.Equal(s.to)
		}, 10*time.Second, 500*time.Millisecond, "approval not visible after approve")

		s.T().Logf("ERC721 %s TestApprove Result", contract)
		s.T().Logf("\tOwner: %s", s.from)
		s.T().Logf("\tApproved: %s", s.to)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", unsigned[i].Gas())
		s.T().Logf("\tNonce: %d", unsigned[i].AddressNonce())
		s.T().Logf("\tMaxFee: %s", unsigned[i].MaxFee())
		s.T().Logf("\tGetApproved")
		s.T().Logf("\t\tAfter: %s", after)
	}
}

func (s *ERC721Suite) TestApproveInvalidApprover() {
	tokenId := big.NewInt(0)

	for _, contract := range s.contracts {
		_, err := s.erc721.CreateApprove(s.ctx, tx.NewDynamicFeeBuilder(), contract, s.to, s.from, tokenId)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidApprover)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC721 %s TestApproveInvalidApprover Result", contract)
		s.T().Logf("\tApprover (unauthorized): %s", s.to)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC721Suite) TestSetApprovalForAll() {
	first, err := s.erc721.CreateSetApprovalForAll(s.ctx, tx.NewDynamicFeeBuilder(), s.contracts[0], s.from, s.to, true)
	s.Require().NoError(err)
	baseNonce := first.AddressNonce()

	unsigned := make([]core.Transaction, len(s.contracts))
	unsigned[0] = first
	for i := 1; i < len(s.contracts); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(baseNonce + uint64(i))

		u, err := s.erc721.CreateSetApprovalForAll(s.ctx, builder, s.contracts[i], s.from, s.to, true)
		s.Require().NoError(err)
		unsigned[i] = u
	}

	hashes := make([]*types.Hash, len(s.contracts))
	for i, u := range unsigned {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc721.SetApprovalForAll(s.ctx, u)
		s.Require().NoError(err)
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		approvals, err := ExtractApprovalForAll(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(approvals, 1)
		s.Require().True(approvals[0].Contract.Equal(contract))
		s.Require().True(approvals[0].Owner.Equal(s.from))
		s.Require().True(approvals[0].Operator.Equal(s.to))
		s.Require().True(approvals[0].Approved)
	}

	firstRevert, err := s.erc721.CreateSetApprovalForAll(s.ctx, tx.NewDynamicFeeBuilder(), s.contracts[0], s.from, s.to, false)
	s.Require().NoError(err)
	revertBaseNonce := firstRevert.AddressNonce()

	unsignedRevert := make([]core.Transaction, len(s.contracts))
	unsignedRevert[0] = firstRevert
	for i := 1; i < len(s.contracts); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(revertBaseNonce + uint64(i))

		u, err := s.erc721.CreateSetApprovalForAll(s.ctx, builder, s.contracts[i], s.from, s.to, false)
		s.Require().NoError(err)
		unsignedRevert[i] = u
	}

	revertHashes := make([]*types.Hash, len(s.contracts))
	for i, u := range unsignedRevert {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc721.SetApprovalForAll(s.ctx, u)
		s.Require().NoError(err)
		revertHashes[i] = hash
	}

	for i, contract := range s.contracts {
		revertReceipt, err := tx.WaitForReceipt(s.ctx, s.client, revertHashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(revertReceipt.Status, "revert transaction reverted")

		revertApprovals, err := ExtractApprovalForAll(revertReceipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(revertApprovals, 1)
		s.Require().False(revertApprovals[0].Approved)

		var reverted bool
		s.Require().Eventually(func() bool {
			reverted, err = s.erc721.IsApprovedForAll(s.ctx, contract, s.from, s.to, rpc.BlockTagLatest)
			return err == nil && !reverted
		}, 10*time.Second, 500*time.Millisecond, "approval-for-all not reverted to false")

		s.T().Logf("ERC721 %s TestSetApprovalForAll Result", contract)
		s.T().Logf("\tOwner: %s", s.from)
		s.T().Logf("\tOperator: %s", s.to)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", unsigned[i].Gas())
		s.T().Logf("\tMaxFee: %s", unsigned[i].MaxFee())
		s.T().Logf("\tRevertTx: %s", revertHashes[i])
		s.T().Logf("\t\tMaxFee: %s", unsignedRevert[i].MaxFee())
		s.T().Logf("\t\tReverted: %v", reverted)
	}
}

func (s *ERC721Suite) TestSetApprovalForAllInvalidOperator() {
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		_, err := s.erc721.CreateSetApprovalForAll(s.ctx, tx.NewDynamicFeeBuilder(), contract, s.from, zero, true)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidOperator)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC721 %s TestSetApprovalForAllInvalidOperator Result", contract)
		s.T().Logf("\tOperator (zero): %s", zero)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC721Suite) TestTransferFrom() {
	tokenId := big.NewInt(2)

	first, err := s.erc721.CreateTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.contracts[0], s.from, s.from, s.to, tokenId)
	s.Require().NoError(err)
	baseNonce := first.AddressNonce()

	unsigned := make([]core.Transaction, len(s.contracts))
	unsigned[0] = first
	for i := 1; i < len(s.contracts); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(baseNonce + uint64(i))

		u, err := s.erc721.CreateTransferFrom(s.ctx, builder, s.contracts[i], s.from, s.from, s.to, tokenId)
		s.Require().NoError(err)
		unsigned[i] = u
	}

	hashes := make([]*types.Hash, len(s.contracts))
	for i, u := range unsigned {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc721.TransferFrom(s.ctx, u)
		s.Require().NoError(err)
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		transfers, err := ExtractTransfers(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].Contract.Equal(contract))
		s.Require().True(transfers[0].From.Equal(s.from))
		s.Require().True(transfers[0].To.Equal(s.to))
		s.Require().Zero(tokenId.Cmp(transfers[0].TokenId))

		var owner *types.Address
		s.Require().Eventually(func() bool {
			owner, err = s.erc721.OwnerOf(s.ctx, contract, tokenId, rpc.BlockTagLatest)
			return err == nil && owner.Equal(s.to)
		}, 10*time.Second, 500*time.Millisecond, "owner not visible after transferFrom")

		s.T().Logf("ERC721 %s TestTransferFrom Result", contract)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", unsigned[i].Gas())
		s.T().Logf("\tMaxFee: %s", unsigned[i].MaxFee())
		s.T().Logf("\tOwnerOf: %s", owner)
	}
}

func (s *ERC721Suite) TestTransferFromInvalidReceiver() {
	tokenId := big.NewInt(2)
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		_, err := s.erc721.CreateTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), contract, s.from, s.from, zero, tokenId)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidReceiver)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC721 %s TestTransferFromInvalidReceiver Result", contract)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC721Suite) TestSafeTransferFrom() {
	tokenId := big.NewInt(0)

	first, err := s.erc721.CreateSafeTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.contracts[0], s.from, s.from, s.to, tokenId)
	s.Require().NoError(err)
	baseNonce := first.AddressNonce()

	unsigned := make([]core.Transaction, len(s.contracts))
	unsigned[0] = first
	for i := 1; i < len(s.contracts); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(baseNonce + uint64(i))

		u, err := s.erc721.CreateSafeTransferFrom(s.ctx, builder, s.contracts[i], s.from, s.from, s.to, tokenId)
		s.Require().NoError(err)
		unsigned[i] = u
	}

	hashes := make([]*types.Hash, len(s.contracts))
	for i, u := range unsigned {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc721.SafeTransferFrom(s.ctx, u)
		s.Require().NoError(err)
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		transfers, err := ExtractTransfers(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].Contract.Equal(contract))
		s.Require().True(transfers[0].From.Equal(s.from))
		s.Require().True(transfers[0].To.Equal(s.to))
		s.Require().Zero(tokenId.Cmp(transfers[0].TokenId))

		var owner *types.Address
		s.Require().Eventually(func() bool {
			owner, err = s.erc721.OwnerOf(s.ctx, contract, tokenId, rpc.BlockTagLatest)
			return err == nil && owner.Equal(s.to)
		}, 10*time.Second, 500*time.Millisecond, "owner not visible after safeTransferFrom")

		s.T().Logf("ERC721 %s TestSafeTransferFrom Result", contract)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", unsigned[i].Gas())
		s.T().Logf("\tMaxFee: %s", unsigned[i].MaxFee())
		s.T().Logf("\tOwnerOf: %s", owner)
	}
}

func (s *ERC721Suite) TestSafeTransferFromWithData() {
	tokenId := big.NewInt(1)
	var callData []byte

	first, err := s.erc721.CreateSafeTransferFromWithData(s.ctx, tx.NewDynamicFeeBuilder(), s.contracts[0], s.from, s.from, s.to, tokenId, callData)
	s.Require().NoError(err)
	baseNonce := first.AddressNonce()

	unsigned := make([]core.Transaction, len(s.contracts))
	unsigned[0] = first
	for i := 1; i < len(s.contracts); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(baseNonce + uint64(i))

		u, err := s.erc721.CreateSafeTransferFromWithData(s.ctx, builder, s.contracts[i], s.from, s.from, s.to, tokenId, callData)
		s.Require().NoError(err)
		unsigned[i] = u
	}

	hashes := make([]*types.Hash, len(s.contracts))
	for i, u := range unsigned {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc721.SafeTransferFrom(s.ctx, u)
		s.Require().NoError(err)
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		transfers, err := ExtractTransfers(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].Contract.Equal(contract))
		s.Require().Zero(tokenId.Cmp(transfers[0].TokenId))

		s.T().Logf("ERC721 %s TestSafeTransferFromWithData Result", contract)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", unsigned[i].Gas())
		s.T().Logf("\tMaxFee: %s", unsigned[i].MaxFee())
	}
}

func (s *ERC721Suite) TestSafeTransferFromInvalidReceiver() {
	tokenId := big.NewInt(1)
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		_, err := s.erc721.CreateSafeTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), contract, s.from, s.from, zero, tokenId)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidReceiver)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC721 %s TestSafeTransferFromInvalidReceiver Result", contract)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC721Suite) TestMetadata() {
	for _, contract := range s.contracts {
		m, err := s.erc721.Metadata(s.ctx, s.multicall, contract, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		name, err := s.erc721.Name(s.ctx, contract, rpc.BlockTagLatest)
		s.Require().NoError(err)
		symbol, err := s.erc721.Symbol(s.ctx, contract, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().Equal(name, m.Name)
		s.Require().Equal(symbol, m.Symbol)

		s.T().Logf("ERC721 %s TestMetadata Result", contract)
		s.T().Logf("\tName: %s", m.Name)
		s.T().Logf("\tSymbol: %s", m.Symbol)
	}
}

func (s *ERC721Suite) TestMetadataAllowFailure() {
	m, err := s.erc721.Metadata(s.ctx, s.multicall, s.multicall, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Empty(m.Name)
	s.Require().Empty(m.Symbol)

	s.T().Logf("ERC721 %s TestMetadataAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", m.Name)
	s.T().Logf("\tSymbol: %q", m.Symbol)
}

func (s *ERC721Suite) TestMetadataWithTokenID() {
	tokenId := big.NewInt(0)

	for _, contract := range s.contracts {
		mt, err := s.erc721.MetadataWithTokenID(s.ctx, s.multicall, contract, tokenId, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		name, err := s.erc721.Name(s.ctx, contract, rpc.BlockTagLatest)
		s.Require().NoError(err)
		symbol, err := s.erc721.Symbol(s.ctx, contract, rpc.BlockTagLatest)
		s.Require().NoError(err)
		tokenURI, err := s.erc721.TokenURI(s.ctx, contract, tokenId, rpc.BlockTagLatest)
		s.Require().NoError(err)
		owner, err := s.erc721.OwnerOf(s.ctx, contract, tokenId, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().Equal(name, mt.Name)
		s.Require().Equal(symbol, mt.Symbol)
		s.Require().Zero(tokenId.Cmp(mt.TokenId))
		s.Require().Equal(tokenURI, mt.TokenURI)
		s.Require().True(owner.Equal(mt.Owner))

		s.T().Logf("ERC721 %s TestMetadataWithTokenID Result", contract)
		s.T().Logf("\tName: %s", mt.Name)
		s.T().Logf("\tSymbol: %s", mt.Symbol)
		s.T().Logf("\tTokenId: %s", mt.TokenId)
		s.T().Logf("\tTokenURI: %s", mt.TokenURI)
		s.T().Logf("\tOwner: %s", mt.Owner)
	}
}

func (s *ERC721Suite) TestMetadataWithTokenIDAllowFailure() {
	tokenId := big.NewInt(0)

	mt, err := s.erc721.MetadataWithTokenID(s.ctx, s.multicall, s.multicall, tokenId, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Empty(mt.Name)
	s.Require().Empty(mt.Symbol)
	s.Require().Empty(mt.TokenURI)
	s.Require().Nil(mt.Owner)

	s.T().Logf("ERC721 %s TestMetadataWithTokenIDAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", mt.Name)
	s.T().Logf("\tSymbol: %q", mt.Symbol)
	s.T().Logf("\tTokenURI: %q", mt.TokenURI)
	s.T().Logf("\tOwner: %v", mt.Owner)
}

func (s *ERC721Suite) TestAllowanceWithTokenID() {
	tokenId := big.NewInt(0)

	for _, contract := range s.contracts {
		awt, err := s.erc721.AllowanceWithTokenID(s.ctx, s.multicall, contract, tokenId, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		owner, err := s.erc721.OwnerOf(s.ctx, contract, tokenId, rpc.BlockTagLatest)
		s.Require().NoError(err)
		operator, err := s.erc721.GetApproved(s.ctx, contract, tokenId, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().Zero(tokenId.Cmp(awt.TokenId))
		s.Require().True(owner.Equal(awt.Owner))
		s.Require().True(operator.Equal(awt.Operator))

		s.T().Logf("ERC721 %s TestAllowanceWithTokenID Result", contract)
		s.T().Logf("\tTokenId: %s", awt.TokenId)
		s.T().Logf("\tOwner: %s", awt.Owner)
		s.T().Logf("\tOperator: %s", awt.Operator)
	}
}

func (s *ERC721Suite) TestAllowanceWithTokenIDAllowFailure() {
	tokenId := big.NewInt(0)

	awt, err := s.erc721.AllowanceWithTokenID(s.ctx, s.multicall, s.multicall, tokenId, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Nil(awt.Owner)
	s.Require().Nil(awt.Operator)

	s.T().Logf("ERC721 %s TestAllowanceWithTokenIDAllowFailure Result", s.multicall)
	s.T().Logf("\tTokenId: %s", awt.TokenId)
	s.T().Logf("\tOwner: %v", awt.Owner)
	s.T().Logf("\tOperator: %v", awt.Operator)
}

func (s *ERC721Suite) TestApprovalForAllWithBalance() {
	for _, contract := range s.contracts {
		afawb, err := s.erc721.ApprovalForAllWithBalance(s.ctx, s.multicall, contract, s.from, s.to, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		isApprovedForAll, err := s.erc721.IsApprovedForAll(s.ctx, contract, s.from, s.to, rpc.BlockTagLatest)
		s.Require().NoError(err)
		ownerBalance, err := s.erc721.BalanceOf(s.ctx, contract, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)
		operatorBalance, err := s.erc721.BalanceOf(s.ctx, contract, s.to, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().True(afawb.Owner.Equal(s.from))
		s.Require().True(afawb.Operator.Equal(s.to))
		s.Require().Equal(isApprovedForAll, afawb.Approved)
		s.Require().Zero(ownerBalance.Cmp(afawb.OwnerBalance))
		s.Require().Zero(operatorBalance.Cmp(afawb.OperatorBalance))

		s.T().Logf("ERC721 %s TestApprovalForAllWithBalance Result", contract)
		s.T().Logf("\tOwner: %s", afawb.Owner)
		s.T().Logf("\tOperator: %s", afawb.Operator)
		s.T().Logf("\tApproved: %v", afawb.Approved)
		s.T().Logf("\tOwnerBalance: %s", afawb.OwnerBalance)
		s.T().Logf("\tOperatorBalance: %s", afawb.OperatorBalance)
	}
}

func (s *ERC721Suite) TestApprovalForAllWithBalanceAllowFailure() {
	afawb, err := s.erc721.ApprovalForAllWithBalance(s.ctx, s.multicall, s.multicall, s.from, s.to, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(afawb.Approved)
	s.Require().Nil(afawb.OwnerBalance)
	s.Require().Nil(afawb.OperatorBalance)

	s.T().Logf("ERC721 %s TestApprovalForAllWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tApproved: %v", afawb.Approved)
	s.T().Logf("\tOwnerBalance: %v", afawb.OwnerBalance)
	s.T().Logf("\tOperatorBalance: %v", afawb.OperatorBalance)
}

func (s *ERC721Suite) TestTokenOwners() {
	tokenIds := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)}

	for _, contract := range s.contracts {
		tos, err := s.erc721.TokenOwners(s.ctx, s.multicall, contract, tokenIds, false, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(tos.Contract.Equal(contract))
		s.Require().Len(tos.Tokens, len(tokenIds))

		s.T().Logf("ERC721 %s TestTokenOwners Result", contract)
		for i, tokenId := range tokenIds {
			owner, err := s.erc721.OwnerOf(s.ctx, contract, tokenId, rpc.BlockTagLatest)
			s.Require().NoError(err)

			s.Require().Zero(tokenId.Cmp(tos.Tokens[i].TokenId))
			s.Require().True(tos.Tokens[i].Owner.Equal(owner))
			s.T().Logf("\t%s: %s", tokenId, tos.Tokens[i].Owner)
		}
	}
}

func (s *ERC721Suite) TestTokenOperators() {
	tokenIds := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)}

	for _, contract := range s.contracts {
		tops, err := s.erc721.TokenOperators(s.ctx, s.multicall, contract, tokenIds, false, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(tops.Contract.Equal(contract))
		s.Require().Len(tops.Tokens, len(tokenIds))

		s.T().Logf("ERC721 %s TestTokenOperators Result", contract)
		for i, tokenId := range tokenIds {
			operator, err := s.erc721.GetApproved(s.ctx, contract, tokenId, rpc.BlockTagLatest)
			s.Require().NoError(err)

			s.Require().Zero(tokenId.Cmp(tops.Tokens[i].TokenId))
			s.Require().True(tops.Tokens[i].Operator.Equal(operator))
			s.T().Logf("\t%s: %s", tokenId, tops.Tokens[i].Operator)
		}
	}
}

func (s *ERC721Suite) TestAddressBalances() {
	ab, err := s.erc721.AddressBalances(s.ctx, s.multicall, s.from, s.contracts, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(ab.Address.Equal(s.from))
	s.Require().Len(ab.Tokens, len(s.contracts))

	s.T().Logf("ERC721 %s TestAddressBalances Result", s.from)
	for i, contract := range s.contracts {
		balance, err := s.erc721.BalanceOf(s.ctx, contract, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().True(ab.Tokens[i].Token.Equal(contract))
		s.Require().Zero(balance.Cmp(ab.Tokens[i].Amount))
		s.T().Logf("\t%s: %s", contract, ab.Tokens[i].Amount)
	}
}

func (s *ERC721Suite) TestTokenBalances() {
	addresses := []*types.Address{s.from, s.to}

	for _, contract := range s.contracts {
		tb, err := s.erc721.TokenBalances(s.ctx, s.multicall, contract, addresses, false, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(tb.Contract.Equal(contract))
		s.Require().Len(tb.Addresses, len(addresses))

		s.T().Logf("ERC721 %s TestTokenBalances Result", contract)
		for i, addr := range addresses {
			balance, err := s.erc721.BalanceOf(s.ctx, contract, addr, rpc.BlockTagLatest)
			s.Require().NoError(err)

			s.Require().True(tb.Addresses[i].Address.Equal(addr))
			s.Require().Zero(balance.Cmp(tb.Addresses[i].Amount))
			s.T().Logf("\t%s: %s", addr, tb.Addresses[i].Amount)
		}
	}
}

func (s *ERC721Suite) TestOwnerPairs() {
	tokenIds := make([]*big.Int, len(s.contracts))
	for i := range tokenIds {
		tokenIds[i] = big.NewInt(0)
	}

	pairs, err := s.erc721.OwnerPairs(s.ctx, s.multicall, s.contracts, tokenIds, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(pairs, len(s.contracts))

	s.T().Logf("ERC721 TestOwnerPairs Result")
	for i, contract := range s.contracts {
		owner, err := s.erc721.OwnerOf(s.ctx, contract, tokenIds[i], rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().True(pairs[i].Contract.Equal(contract))
		s.Require().Zero(tokenIds[i].Cmp(pairs[i].TokenId))
		s.Require().True(pairs[i].Owner.Equal(owner))
		s.T().Logf("\t%s / %s: %s", pairs[i].Contract, pairs[i].TokenId, pairs[i].Owner)
	}
}
