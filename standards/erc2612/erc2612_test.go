package erc2612

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
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// erc2612ByteCodeHex is OpenZeppelin's ERC20Permit plus mint/burn, constructor(name, symbol).
const erc2612ByteCodeHex = "0x610160604052348015610010575f5ffd5b506040516125d83803806125d8833981810160405281019061003291906103a2565b81806040518060400160405280600181526020017f31000000000000000000000000000000000000000000000000000000000000008152508484816003908161007b9190610639565b50806004908161008b9190610639565b5050506100a260058361014060201b90919060201c565b61012081815250506100be60068261014060201b90919060201c565b6101408181525050818051906020012060e08181525050808051906020012061010081815250504660a081815250506100fb61018d60201b60201c565b608081815250503073ffffffffffffffffffffffffffffffffffffffff1660c08173ffffffffffffffffffffffffffffffffffffffff168152505050505050506108ba565b5f6020835110156101615761015a836101e760201b60201c565b9050610187565b826101718361024c60201b60201c565b5f01908161017f9190610639565b5060ff5f1b90505b92915050565b5f7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60e0516101005146306040516020016101cc95949392919061076e565b60405160208183030381529060405280519060200120905090565b5f5f829050601f8151111561023357826040517f305a27a900000000000000000000000000000000000000000000000000000000815260040161022a9190610807565b60405180910390fd5b80518161023f90610854565b5f1c175f1b915050919050565b5f819050919050565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6102b48261026e565b810181811067ffffffffffffffff821117156102d3576102d261027e565b5b80604052505050565b5f6102e5610255565b90506102f182826102ab565b919050565b5f67ffffffffffffffff8211156103105761030f61027e565b5b6103198261026e565b9050602081019050919050565b8281835e5f83830152505050565b5f610346610341846102f6565b6102dc565b9050828152602081018484840111156103625761036161026a565b5b61036d848285610326565b509392505050565b5f82601f83011261038957610388610266565b5b8151610399848260208601610334565b91505092915050565b5f5f604083850312156103b8576103b761025e565b5b5f83015167ffffffffffffffff8111156103d5576103d4610262565b5b6103e185828601610375565b925050602083015167ffffffffffffffff81111561040257610401610262565b5b61040e85828601610375565b9150509250929050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061046657607f821691505b60208210810361047957610478610422565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026104db7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826104a0565b6104e586836104a0565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61052961052461051f846104fd565b610506565b6104fd565b9050919050565b5f819050919050565b6105428361050f565b61055661054e82610530565b8484546104ac565b825550505050565b5f5f905090565b61056d61055e565b610578818484610539565b505050565b5f5b8281101561059e576105935f828401610565565b60018101905061057f565b505050565b601f8211156105f157828211156105f0576105bd8161047f565b6105c683610491565b6105cf85610491565b60208610156105dc575f90505b8083016105eb8284038261057d565b505050505b5b505050565b5f82821c905092915050565b5f6106115f19846008026105f6565b1980831691505092915050565b5f6106298383610602565b9150826002028217905092915050565b61064282610418565b67ffffffffffffffff81111561065b5761065a61027e565b5b610665825461044f565b6106708282856105a3565b5f60209050601f8311600181146106a1575f841561068f578287015190505b610699858261061e565b865550610700565b601f1984166106af8661047f565b5f5b828110156106d6578489015182556001820191506020850194506020810190506106b1565b868310156106f357848901516106ef601f891682610602565b8355505b6001600288020188555050505b505050505050565b5f819050919050565b61071a81610708565b82525050565b610729816104fd565b82525050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6107588261072f565b9050919050565b6107688161074e565b82525050565b5f60a0820190506107815f830188610711565b61078e6020830187610711565b61079b6040830186610711565b6107a86060830185610720565b6107b5608083018461075f565b9695505050505050565b5f82825260208201905092915050565b5f6107d982610418565b6107e381856107bf565b93506107f3818560208601610326565b6107fc8161026e565b840191505092915050565b5f6020820190508181035f83015261081f81846107cf565b905092915050565b5f81519050919050565b5f819050602082019050919050565b5f61084b8251610708565b80915050919050565b5f61085e82610827565b8261086884610831565b905061087381610840565b925060208210156108b3576108ae7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff836020036008026104a0565b831692505b5050919050565b60805160a05160c05160e051610100516101205161014051611ccd61090b5f395f610b0f01525f610ad401525f61108701525f61106601525f61095301525f6109a901525f6109d20152611ccd5ff3fe608060405234801561000f575f5ffd5b50600436106100f3575f3560e01c806370a08231116100955780639dc29fac116100645780639dc29fac1461028d578063a9059cbb146102a9578063d505accf146102d9578063dd62ed3e146102f5576100f3565b806370a08231146101eb5780637ecebe001461021b57806384b0196e1461024b57806395d89b411461026f576100f3565b806323b872dd116100d157806323b872dd14610163578063313ce567146101935780633644e515146101b157806340c10f19146101cf576100f3565b806306fdde03146100f7578063095ea7b31461011557806318160ddd14610145575b5f5ffd5b6100ff610325565b60405161010c919061153d565b60405180910390f35b61012f600480360381019061012a91906115ee565b6103b5565b60405161013c9190611646565b60405180910390f35b61014d6103d7565b60405161015a919061166e565b60405180910390f35b61017d60048036038101906101789190611687565b6103e0565b60405161018a9190611646565b60405180910390f35b61019b61040e565b6040516101a891906116f2565b60405180910390f35b6101b9610416565b6040516101c69190611723565b60405180910390f35b6101e960048036038101906101e491906115ee565b610424565b005b6102056004803603810190610200919061173c565b610432565b604051610212919061166e565b60405180910390f35b6102356004803603810190610230919061173c565b610477565b604051610242919061166e565b60405180910390f35b610253610488565b6040516102669796959493929190611867565b60405180910390f35b61027761052d565b604051610284919061153d565b60405180910390f35b6102a760048036038101906102a291906115ee565b6105bd565b005b6102c360048036038101906102be91906115ee565b6105cb565b6040516102d09190611646565b60405180910390f35b6102f360048036038101906102ee919061193d565b6105ed565b005b61030f600480360381019061030a91906119da565b610732565b60405161031c919061166e565b60405180910390f35b60606003805461033490611a45565b80601f016020809104026020016040519081016040528092919081815260200182805461036090611a45565b80156103ab5780601f10610382576101008083540402835291602001916103ab565b820191905f5260205f20905b81548152906001019060200180831161038e57829003601f168201915b5050505050905090565b5f5f6103bf6107b4565b90506103cc8185856107bb565b600191505092915050565b5f600254905090565b5f5f6103ea6107b4565b90506103f78582856107cd565b610402858585610860565b60019150509392505050565b5f6012905090565b5f61041f610950565b905090565b61042e8282610a06565b5050565b5f5f5f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b5f61048182610a85565b9050919050565b5f6060805f5f5f6060610499610acb565b6104a1610b06565b46305f5f1b5f67ffffffffffffffff8111156104c0576104bf611a75565b5b6040519080825280602002602001820160405280156104ee5781602001602082028036833780820191505090505b507f0f00000000000000000000000000000000000000000000000000000000000000959493929190965096509650965096509650965090919293949596565b60606004805461053c90611a45565b80601f016020809104026020016040519081016040528092919081815260200182805461056890611a45565b80156105b35780601f1061058a576101008083540402835291602001916105b3565b820191905f5260205f20905b81548152906001019060200180831161059657829003601f168201915b5050505050905090565b6105c78282610b41565b5050565b5f5f6105d56107b4565b90506105e2818585610860565b600191505092915050565b8342111561063257836040517f62791302000000000000000000000000000000000000000000000000000000008152600401610629919061166e565b60405180910390fd5b5f7f6e71edae12b1b97f4d1f60370fef10105fa2faae0126114a169c64845d6126c98888886106608c610bc0565b8960405160200161067696959493929190611aa2565b6040516020818303038152906040528051906020012090505f61069882610c13565b90505f6106a782878787610c2c565b90508973ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161461071b57808a6040517f4b800e46000000000000000000000000000000000000000000000000000000008152600401610712929190611b01565b60405180910390fd5b6107268a8a8a6107bb565b50505050505050505050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f33905090565b6107c88383836001610c5a565b505050565b5f6107d88484610732565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81101561085a578181101561084b578281836040517ffb8f41b200000000000000000000000000000000000000000000000000000000815260040161084293929190611b28565b60405180910390fd5b61085984848484035f610c5a565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16036108d0575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016108c79190611b5d565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610940575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016109379190611b5d565b60405180910390fd5b61094b838383610e29565b505050565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff161480156109cb57507f000000000000000000000000000000000000000000000000000000000000000046145b156109f8577f00000000000000000000000000000000000000000000000000000000000000009050610a03565b610a00611042565b90505b90565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610a76575f6040517fec442f05000000000000000000000000000000000000000000000000000000008152600401610a6d9190611b5d565b60405180910390fd5b610a815f8383610e29565b5050565b5f60075f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b6060610b0160057f00000000000000000000000000000000000000000000000000000000000000006110d790919063ffffffff16565b905090565b6060610b3c60067f00000000000000000000000000000000000000000000000000000000000000006110d790919063ffffffff16565b905090565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610bb1575f6040517f96c6fd1e000000000000000000000000000000000000000000000000000000008152600401610ba89190611b5d565b60405180910390fd5b610bbc825f83610e29565b5050565b5f60075f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f815480929190600101919050559050919050565b5f610c25610c1f610950565b83611184565b9050919050565b5f5f5f5f610c3c888888886111c4565b925092509250610c4c82826112ab565b829350505050949350505050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610cca575f6040517fe602df05000000000000000000000000000000000000000000000000000000008152600401610cc19190611b5d565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610d3a575f6040517f94280d62000000000000000000000000000000000000000000000000000000008152600401610d319190611b5d565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508015610e23578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92584604051610e1a919061166e565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610e79578060025f828254610e6d9190611ba3565b92505081905550610f47565b5f5f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905081811015610f02578381836040517fe450d38c000000000000000000000000000000000000000000000000000000008152600401610ef993929190611b28565b60405180910390fd5b8181035f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610f8e578060025f8282540392505081905550610fd8565b805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051611035919061166e565b60405180910390a3505050565b5f7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f7f00000000000000000000000000000000000000000000000000000000000000007f000000000000000000000000000000000000000000000000000000000000000046306040516020016110bc959493929190611bd6565b60405160208183030381529060405280519060200120905090565b606060ff5f1b83146110f3576110ec8361140d565b905061117e565b8180546110ff90611a45565b80601f016020809104026020016040519081016040528092919081815260200182805461112b90611a45565b80156111765780601f1061114d57610100808354040283529160200191611176565b820191905f5260205f20905b81548152906001019060200180831161115957829003601f168201915b505050505090505b92915050565b5f6040517f190100000000000000000000000000000000000000000000000000000000000081528360028201528260228201526042812091505092915050565b5f5f5f7f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0845f1c1115611200575f6003859250925092506112a1565b5f6001888888886040515f81526020016040526040516112239493929190611c27565b6020604051602081039080840390855afa158015611243573d5f5f3e3d5ffd5b5050506020604051035190505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603611294575f60015f5f1b935093509350506112a1565b805f5f5f1b935093509350505b9450945094915050565b5f60038111156112be576112bd611c6a565b5b8260038111156112d1576112d0611c6a565b5b031561140957600160038111156112eb576112ea611c6a565b5b8260038111156112fe576112fd611c6a565b5b03611335576040517ff645eedf00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6002600381111561134957611348611c6a565b5b82600381111561135c5761135b611c6a565b5b036113a057805f1c6040517ffce698f7000000000000000000000000000000000000000000000000000000008152600401611397919061166e565b60405180910390fd5b6003808111156113b3576113b2611c6a565b5b8260038111156113c6576113c5611c6a565b5b0361140857806040517fd78bce0c0000000000000000000000000000000000000000000000000000000081526004016113ff9190611723565b60405180910390fd5b5b5050565b60605f6114198361147f565b90505f602067ffffffffffffffff81111561143757611436611a75565b5b6040519080825280601f01601f1916602001820160405280156114695781602001600182028036833780820191505090505b5090508181528360208201528092505050919050565b5f5f60ff835f1c169050601f8111156114c4576040517fb3512b0c00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b80915050919050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f61150f826114cd565b61151981856114d7565b93506115298185602086016114e7565b611532816114f5565b840191505092915050565b5f6020820190508181035f8301526115558184611505565b905092915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61158a82611561565b9050919050565b61159a81611580565b81146115a4575f5ffd5b50565b5f813590506115b581611591565b92915050565b5f819050919050565b6115cd816115bb565b81146115d7575f5ffd5b50565b5f813590506115e8816115c4565b92915050565b5f5f604083850312156116045761160361155d565b5b5f611611858286016115a7565b9250506020611622858286016115da565b9150509250929050565b5f8115159050919050565b6116408161162c565b82525050565b5f6020820190506116595f830184611637565b92915050565b611668816115bb565b82525050565b5f6020820190506116815f83018461165f565b92915050565b5f5f5f6060848603121561169e5761169d61155d565b5b5f6116ab868287016115a7565b93505060206116bc868287016115a7565b92505060406116cd868287016115da565b9150509250925092565b5f60ff82169050919050565b6116ec816116d7565b82525050565b5f6020820190506117055f8301846116e3565b92915050565b5f819050919050565b61171d8161170b565b82525050565b5f6020820190506117365f830184611714565b92915050565b5f602082840312156117515761175061155d565b5b5f61175e848285016115a7565b91505092915050565b5f7fff0000000000000000000000000000000000000000000000000000000000000082169050919050565b61179b81611767565b82525050565b6117aa81611580565b82525050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b6117e2816115bb565b82525050565b5f6117f383836117d9565b60208301905092915050565b5f602082019050919050565b5f611815826117b0565b61181f81856117ba565b935061182a836117ca565b805f5b8381101561185a57815161184188826117e8565b975061184c836117ff565b92505060018101905061182d565b5085935050505092915050565b5f60e08201905061187a5f83018a611792565b818103602083015261188c8189611505565b905081810360408301526118a08188611505565b90506118af606083018761165f565b6118bc60808301866117a1565b6118c960a0830185611714565b81810360c08301526118db818461180b565b905098975050505050505050565b6118f2816116d7565b81146118fc575f5ffd5b50565b5f8135905061190d816118e9565b92915050565b61191c8161170b565b8114611926575f5ffd5b50565b5f8135905061193781611913565b92915050565b5f5f5f5f5f5f5f60e0888a0312156119585761195761155d565b5b5f6119658a828b016115a7565b97505060206119768a828b016115a7565b96505060406119878a828b016115da565b95505060606119988a828b016115da565b94505060806119a98a828b016118ff565b93505060a06119ba8a828b01611929565b92505060c06119cb8a828b01611929565b91505092959891949750929550565b5f5f604083850312156119f0576119ef61155d565b5b5f6119fd858286016115a7565b9250506020611a0e858286016115a7565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680611a5c57607f821691505b602082108103611a6f57611a6e611a18565b5b50919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b5f60c082019050611ab55f830189611714565b611ac260208301886117a1565b611acf60408301876117a1565b611adc606083018661165f565b611ae9608083018561165f565b611af660a083018461165f565b979650505050505050565b5f604082019050611b145f8301856117a1565b611b2160208301846117a1565b9392505050565b5f606082019050611b3b5f8301866117a1565b611b48602083018561165f565b611b55604083018461165f565b949350505050565b5f602082019050611b705f8301846117a1565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611bad826115bb565b9150611bb8836115bb565b9250828201905080821115611bd057611bcf611b76565b5b92915050565b5f60a082019050611be95f830188611714565b611bf66020830187611714565b611c036040830186611714565b611c10606083018561165f565b611c1d60808301846117a1565b9695505050505050565b5f608082019050611c3a5f830187611714565b611c4760208301866116e3565b611c546040830185611714565b611c616060830184611714565b95945050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffdfea264697066735822122044347b76c25dd816dbe46eb154b880013571f2602582d57a3ff7e772d124eb4e64736f6c63430008230033"

type ERC2612Suite struct {
	suite.Suite

	Enable               bool
	RPCURL               string
	Interval             int
	PrivateKeyHex        string
	SpenderPrivateKeyHex string
	ToHex                string
	MulticallHex         string
	FactoryHex           string
	Value                string
	MintValue            string
	NativeSymbol         string
	TokenCount           int
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
	spender       *types.Address
	to            *types.Address
	multicall     *types.Address
	factory       *types.Address
	tokens        []*types.Address
	value         *big.Int
	mintValue     *big.Int
	chainID       *big.Int

	erc20   *erc20.ERC20
	erc2612 *ERC2612

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestERC2612Suite(t *testing.T) {
	suite.Run(t, &ERC2612Suite{
		Enable:               false,
		RPCURL:               "",
		Interval:             200,
		PrivateKeyHex:        "",
		SpenderPrivateKeyHex: "",
		ToHex:                "",
		MulticallHex:         multicall3.DefaultMulticall3Address.String(),
		FactoryHex:           deploy.DefaultNicksMethodFactory.String(),
		Value:                "100000",
		MintValue:            "1000000000000000000000000",
		NativeSymbol:         "",
		TokenCount:           5,
		ByteCode:             erc2612ByteCodeHex,
		Constructor:          abi.NewFunction("constructor", abi.Types{abi.String, abi.String}, nil),
		Args:                 []any{"evmkit test ERC2612", "EVK2612"},
		Salt:                 core.Keccak256([]byte("go-evmkit")),
	})
}

func (s *ERC2612Suite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.erc20 = erc20.New(s.client)
	s.erc2612 = New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

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

	s.factory, err = types.NewAddressFromHex(s.FactoryHex)
	s.Require().NoError(err)

	value, ok := new(big.Int).SetString(s.Value, 10)
	s.Require().True(ok)
	s.value = value

	mintValue, ok := new(big.Int).SetString(s.MintValue, 10)
	s.Require().True(ok)
	s.mintValue = mintValue

	var chainIDHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHChainID("chainId", &chainIDHex)))
	chainID, ok := new(big.Int).SetString(strings.TrimPrefix(chainIDHex, "0x"), 16)
	s.Require().True(ok)
	s.chainID = chainID

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

		unsigned, err := s.deployer.CreateDeployment(s.ctx, tx.NewDynamicFeeBuilder(), s.from, s.factory, big.NewInt(0), deployment)
		s.Require().NoError(err)

		s.Require().NoError(unsigned.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsigned.MaxFee())
		s.mu.Unlock()

		hash, err := s.deployer.Deploy(s.ctx, unsigned)
		s.Require().NoError(err)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, 200, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

		token, err := s.deployer.DeployAddress(s.factory, deployment)
		s.Require().NoError(err)
		s.tokens[i] = token

		s.Require().Eventually(func() bool {
			var codeHex string
			if err := s.client.Call(s.ctx, rpc.ETHGetCode("code", token.String(), rpc.BlockTagLatest, &codeHex)); err != nil {
				return false
			}
			code, err := encoding.Hex.Decode(codeHex)
			return err == nil && len(code) > 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "no code visible at deployed token %s", token)

		s.T().Logf("SetupSuite: deployed token %d/%d %s (nonce %d, tx %s)", i+1, s.TokenCount, token, unsigned.AddressNonce(), hash)

		calldata, err := mintFn.EncodeCall(s.from, s.mintValue)
		s.Require().NoError(err)

		var unsignedMint core.Transaction
		s.Require().Eventually(func() bool {
			u, err := tx.NewDynamicFeeBuilder().Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			if err != nil || u.Gas() < 40000 {
				return false
			}
			unsignedMint = u
			return true
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "mint gas estimate looked like a no-code call")
		s.Require().NoError(unsignedMint.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())
		s.mu.Unlock()

		raw, err := unsignedMint.EncodeRLP()
		s.Require().NoError(err)
		mintHash, err := tx.Broadcast(s.ctx, s.client, raw)
		s.Require().NoError(err)

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")

		s.T().Logf("SetupSuite: minted %s of %s to owner %s (tx %s)", s.mintValue, token, s.from, mintHash)
	}
}

func (s *ERC2612Suite) TearDownSuite() {
	if s.tokens != nil {
		burnFn := abi.NewFunction("burn", abi.Types{abi.Address, abi.Uint256}, nil)

		for _, token := range s.tokens {
			for _, holder := range []*types.Address{s.from, s.spender, s.to} {
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

				burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
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

		s.T().Logf("ERC2612Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC2612Suite) TestReadMethods() {
	for _, token := range s.tokens {
		nonce, err := s.erc2612.Nonces(s.ctx, token, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		domainSeparator, err := s.erc2612.DomainSeparator(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)

		balances, err := s.erc20.TokenBalances(s.ctx, s.multicall, token, []*types.Address{s.from, s.spender, s.to}, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.T().Logf("ERC2612 %s TestReadMethods Result", token)
		s.T().Logf("\tNonces(%s): %s", s.from, nonce)
		s.T().Logf("\tDomainSeparator: %s", encoding.Hex.EncodePrefixed(domainSeparator))
		s.T().Logf("\tBalances")
		for _, ab := range balances.Addresses {
			s.T().Logf("\t\t%s: %s", ab.Address, ab.Amount)
		}
	}
}

func (s *ERC2612Suite) TestDomainSeparatorMatchesOnChain() {
	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)

		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)
		want, err := eip712.DomainSeparator(domain)
		s.Require().NoError(err)

		got, err := s.erc2612.DomainSeparator(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().Equal(want.Bytes(), got)

		s.T().Logf("ERC2612 %s TestDomainSeparatorMatchesOnChain Result", token)
		s.T().Logf("\tName: %s", name)
		s.T().Logf("\tChainID: %s", s.chainID)
		s.T().Logf("\tDomainSeparator: %s", encoding.Hex.EncodePrefixed(got))
	}
}

func (s *ERC2612Suite) TestPermit() {
	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)

		nonceBefore, err := s.erc2612.Nonces(s.ctx, token, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
		permit := NewPermit(s.from, s.spender, s.value, nonceBefore, deadline)
		sig, err := permit.Sign(domain, s.signer)
		s.Require().NoError(err)

		permitTx, err := s.erc2612.CreatePermit(s.ctx, tx.NewDynamicFeeBuilder(), token, s.spender, s.from, s.spender, s.value, deadline, sig)
		s.Require().NoError(err)
		s.Require().NoError(permitTx.Sign(s.spenderSigner))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, permitTx.MaxFee())
		s.mu.Unlock()

		permitHash, err := s.erc2612.Permit(s.ctx, permitTx)
		s.Require().NoError(err)

		permitReceipt, err := tx.WaitForReceipt(s.ctx, s.client, permitHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(permitReceipt.Status, "permit transaction reverted")

		approvals, err := erc20.ExtractApprovals(permitReceipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(approvals, 1)
		s.Require().True(approvals[0].Contract.Equal(token))
		s.Require().True(approvals[0].Owner.Equal(s.from))
		s.Require().True(approvals[0].Spender.Equal(s.spender))
		s.Require().Zero(s.value.Cmp(approvals[0].Value))

		wantNonceAfter := new(big.Int).Add(nonceBefore, big.NewInt(1))
		var nonceAfter *big.Int
		s.Require().Eventually(func() bool {
			nonceAfter, err = s.erc2612.Nonces(s.ctx, token, s.from, rpc.BlockTagLatest)
			return err == nil && wantNonceAfter.Cmp(nonceAfter) == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "nonce not visible after permit")

		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.spender, rpc.BlockTagLatest)
			return err == nil && s.value.Cmp(allowance) == 0
		}, 10*time.Second, 500*time.Millisecond, "allowance not visible after permit")

		s.T().Logf("ERC2612 %s TestPermit Permit Result", token)
		s.T().Logf("\tOwner: %s", s.from)
		s.T().Logf("\tSpender: %s", s.spender)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tNonce: %s -> %s", nonceBefore, nonceAfter)
		s.T().Logf("\tTx: %s", permitHash)

		transferFromTx, err := s.erc20.CreateTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), token, s.spender, s.from, s.to, s.value)
		s.Require().NoError(err)
		s.Require().NoError(transferFromTx.Sign(s.spenderSigner))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, transferFromTx.MaxFee())
		s.mu.Unlock()

		transferHash, err := s.erc20.TransferFrom(s.ctx, transferFromTx)
		s.Require().NoError(err)

		transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(transferReceipt.Status, "transferFrom transaction reverted")

		transfers, err := erc20.ExtractTransfers(transferReceipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].Contract.Equal(token))
		s.Require().True(transfers[0].From.Equal(s.from))
		s.Require().True(transfers[0].To.Equal(s.to))
		s.Require().Zero(s.value.Cmp(transfers[0].Value))

		var allowanceAfter *big.Int
		s.Require().Eventually(func() bool {
			allowanceAfter, err = s.erc20.Allowance(s.ctx, token, s.from, s.spender, rpc.BlockTagLatest)
			return err == nil && allowanceAfter.Sign() == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible as consumed after transferFrom")

		var toBalance *big.Int
		s.Require().Eventually(func() bool {
			toBalance, err = s.erc20.BalanceOf(s.ctx, token, s.to, rpc.BlockTagLatest)
			return err == nil && s.value.Cmp(toBalance) == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "balance not visible after transferFrom")

		s.T().Logf("ERC2612 %s TestPermit TransferFrom Result", token)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tValue: %s", s.value)
		s.T().Logf("\tTx: %s", transferHash)
		s.T().Logf("\tAllowance remaining: %s", allowanceAfter)
		s.T().Logf("\tTo balance: %s", toBalance)
	}
}

func (s *ERC2612Suite) TestPermitExpiredDeadline() {
	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)

		nonce, err := s.erc2612.Nonces(s.ctx, token, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		deadline := big.NewInt(time.Now().Add(-1 * time.Hour).Unix())
		permit := NewPermit(s.from, s.spender, s.value, nonce, deadline)
		sig, err := permit.Sign(domain, s.signer)
		s.Require().NoError(err)

		_, err = s.erc2612.CreatePermit(s.ctx, tx.NewDynamicFeeBuilder(), token, s.spender, s.from, s.spender, s.value, deadline, sig)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrExpiredSignature)

		s.T().Logf("ERC2612 %s TestPermitExpiredDeadline Result", token)
		s.T().Logf("\tDeadline: %s", deadline)
		s.T().Logf("\tError: %v", err)
	}
}

func (s *ERC2612Suite) TestPermitWrongSigner() {
	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)

		nonce, err := s.erc2612.Nonces(s.ctx, token, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
		permit := NewPermit(s.from, s.spender, s.value, nonce, deadline)
		sig, err := permit.Sign(domain, s.spenderSigner)
		s.Require().NoError(err)

		_, err = s.erc2612.CreatePermit(s.ctx, tx.NewDynamicFeeBuilder(), token, s.spender, s.from, s.spender, s.value, deadline, sig)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidSigner)

		s.T().Logf("ERC2612 %s TestPermitWrongSigner Result", token)
		s.T().Logf("\tError: %v", err)
	}
}

func (s *ERC2612Suite) TestPermitReplayRejected() {
	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)

		nonce, err := s.erc2612.Nonces(s.ctx, token, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())
		permit := NewPermit(s.from, s.spender, s.value, nonce, deadline)
		sig, err := permit.Sign(domain, s.signer)
		s.Require().NoError(err)

		permitTx, err := s.erc2612.CreatePermit(s.ctx, tx.NewDynamicFeeBuilder(), token, s.spender, s.from, s.spender, s.value, deadline, sig)
		s.Require().NoError(err)
		s.Require().NoError(permitTx.Sign(s.spenderSigner))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, permitTx.MaxFee())
		s.mu.Unlock()

		permitHash, err := s.erc2612.Permit(s.ctx, permitTx)
		s.Require().NoError(err)

		permitReceipt, err := tx.WaitForReceipt(s.ctx, s.client, permitHash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
		s.Require().NoError(err)
		s.Require().True(permitReceipt.Status, "permit transaction reverted")

		var nonceAfter *big.Int
		wantNonceAfter := new(big.Int).Add(nonce, big.NewInt(1))
		s.Require().Eventually(func() bool {
			nonceAfter, err = s.erc2612.Nonces(s.ctx, token, s.from, rpc.BlockTagLatest)
			return err == nil && wantNonceAfter.Cmp(nonceAfter) == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "nonce not visible after permit")

		_, err = s.erc2612.CreatePermit(s.ctx, tx.NewDynamicFeeBuilder(), token, s.spender, s.from, s.spender, s.value, deadline, sig)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidSigner)

		s.T().Logf("ERC2612 %s TestPermitReplayRejected Result", token)
		s.T().Logf("\tFirst use Tx: %s", permitHash)
		s.T().Logf("\tReplay error: %v", err)
	}
}

func (s *ERC2612Suite) TestPermitOverwritesAllowance() {
	secondValue := new(big.Int).Quo(s.value, big.NewInt(2))

	for _, token := range s.tokens {
		name, err := s.erc20.Name(s.ctx, token, rpc.BlockTagLatest)
		s.Require().NoError(err)
		domain := eip712.NewDomain(name, DefaultVersion, s.chainID, token, nil)

		submit := func(value, deadline *big.Int) *types.Hash {
			nonce, err := s.erc2612.Nonces(s.ctx, token, s.from, rpc.BlockTagLatest)
			s.Require().NoError(err)

			permit := NewPermit(s.from, s.spender, value, nonce, deadline)
			sig, err := permit.Sign(domain, s.signer)
			s.Require().NoError(err)

			permitTx, err := s.erc2612.CreatePermit(s.ctx, tx.NewDynamicFeeBuilder(), token, s.spender, s.from, s.spender, value, deadline, sig)
			s.Require().NoError(err)
			s.Require().NoError(permitTx.Sign(s.spenderSigner))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, permitTx.MaxFee())
			s.mu.Unlock()

			hash, err := s.erc2612.Permit(s.ctx, permitTx)
			s.Require().NoError(err)

			receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, 15*time.Second)
			s.Require().NoError(err)
			s.Require().True(receipt.Status, "permit transaction reverted")

			return hash
		}

		deadline := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

		firstHash := submit(s.value, deadline)
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.spender, rpc.BlockTagLatest)
			return err == nil && s.value.Cmp(allowance) == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after first permit")

		secondHash := submit(secondValue, deadline)
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.spender, rpc.BlockTagLatest)
			return err == nil && secondValue.Cmp(allowance) == 0
		}, 10*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after second permit")

		s.T().Logf("ERC2612 %s TestPermitOverwritesAllowance Result", token)
		s.T().Logf("\tFirst permit (value %s): %s", s.value, firstHash)
		s.T().Logf("\tSecond permit (value %s): %s", secondValue, secondHash)
		s.T().Logf("\tFinal allowance: %s", allowance)
	}
}
