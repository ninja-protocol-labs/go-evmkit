package v3

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/protocols/weth"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc20"
	"github.com/ninja-protocol-labs/go-evmkit/standards/multicall3"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// erc20PermitFullByteCodeHex is OpenZeppelin's ERC20 plus both ERC-2612's permit (owner, spender, value, deadline, v, r, s) and a DAI/CHAI-style permit (holder, spender, nonce, expiry, allowed, v, r, s) -- distinct selectors, so both coexist -- plus mint/burn,
// constructor(name, symbol). Compiled from evmkit-fixtures' ERC20PermitFullKit.
const erc20PermitFullByteCodeHex = "0x610160604052348015610010575f5ffd5b50604051612937380380612937833981810160405281019061003291906103a2565b81806040518060400160405280600181526020017f31000000000000000000000000000000000000000000000000000000000000008152508484816003908161007b9190610639565b50806004908161008b9190610639565b5050506100a260058361014060201b90919060201c565b61012081815250506100be60068261014060201b90919060201c565b6101408181525050818051906020012060e08181525050808051906020012061010081815250504660a081815250506100fb61018d60201b60201c565b608081815250503073ffffffffffffffffffffffffffffffffffffffff1660c08173ffffffffffffffffffffffffffffffffffffffff168152505050505050506108ba565b5f6020835110156101615761015a836101e760201b60201c565b9050610187565b826101718361024c60201b60201c565b5f01908161017f9190610639565b5060ff5f1b90505b92915050565b5f7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60e0516101005146306040516020016101cc95949392919061076e565b60405160208183030381529060405280519060200120905090565b5f5f829050601f8151111561023357826040517f305a27a900000000000000000000000000000000000000000000000000000000815260040161022a9190610807565b60405180910390fd5b80518161023f90610854565b5f1c175f1b915050919050565b5f819050919050565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6102b48261026e565b810181811067ffffffffffffffff821117156102d3576102d261027e565b5b80604052505050565b5f6102e5610255565b90506102f182826102ab565b919050565b5f67ffffffffffffffff8211156103105761030f61027e565b5b6103198261026e565b9050602081019050919050565b8281835e5f83830152505050565b5f610346610341846102f6565b6102dc565b9050828152602081018484840111156103625761036161026a565b5b61036d848285610326565b509392505050565b5f82601f83011261038957610388610266565b5b8151610399848260208601610334565b91505092915050565b5f5f604083850312156103b8576103b761025e565b5b5f83015167ffffffffffffffff8111156103d5576103d4610262565b5b6103e185828601610375565b925050602083015167ffffffffffffffff81111561040257610401610262565b5b61040e85828601610375565b9150509250929050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061046657607f821691505b60208210810361047957610478610422565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026104db7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826104a0565b6104e586836104a0565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f61052961052461051f846104fd565b610506565b6104fd565b9050919050565b5f819050919050565b6105428361050f565b61055661054e82610530565b8484546104ac565b825550505050565b5f5f905090565b61056d61055e565b610578818484610539565b505050565b5f5b8281101561059e576105935f828401610565565b60018101905061057f565b505050565b601f8211156105f157828211156105f0576105bd8161047f565b6105c683610491565b6105cf85610491565b60208610156105dc575f90505b8083016105eb8284038261057d565b505050505b5b505050565b5f82821c905092915050565b5f6106115f19846008026105f6565b1980831691505092915050565b5f6106298383610602565b9150826002028217905092915050565b61064282610418565b67ffffffffffffffff81111561065b5761065a61027e565b5b610665825461044f565b6106708282856105a3565b5f60209050601f8311600181146106a1575f841561068f578287015190505b610699858261061e565b865550610700565b601f1984166106af8661047f565b5f5b828110156106d6578489015182556001820191506020850194506020810190506106b1565b868310156106f357848901516106ef601f891682610602565b8355505b6001600288020188555050505b505050505050565b5f819050919050565b61071a81610708565b82525050565b610729816104fd565b82525050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6107588261072f565b9050919050565b6107688161074e565b82525050565b5f60a0820190506107815f830188610711565b61078e6020830187610711565b61079b6040830186610711565b6107a86060830185610720565b6107b5608083018461075f565b9695505050505050565b5f82825260208201905092915050565b5f6107d982610418565b6107e381856107bf565b93506107f3818560208601610326565b6107fc8161026e565b840191505092915050565b5f6020820190508181035f83015261081f81846107cf565b905092915050565b5f81519050919050565b5f819050602082019050919050565b5f61084b8251610708565b80915050919050565b5f61085e82610827565b8261086884610831565b905061087381610840565b925060208210156108b3576108ae7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff836020036008026104a0565b831692505b5050919050565b60805160a05160c05160e05161010051610120516101405161202c61090b5f395f610cb601525f610c7b01525f61128501525f61126401525f610afa01525f610b5001525f610b79015261202c5ff3fe608060405234801561000f575f5ffd5b50600436106100fe575f3560e01c80637ecebe00116100955780639dc29fac116100645780639dc29fac146102b4578063a9059cbb146102d0578063d505accf14610300578063dd62ed3e1461031c576100fe565b80637ecebe001461022657806384b0196e146102565780638fcbaf0c1461027a57806395d89b4114610296576100fe565b8063313ce567116100d1578063313ce5671461019e5780633644e515146101bc57806340c10f19146101da57806370a08231146101f6576100fe565b806306fdde0314610102578063095ea7b31461012057806318160ddd1461015057806323b872dd1461016e575b5f5ffd5b61010a61034c565b604051610117919061173b565b60405180910390f35b61013a600480360381019061013591906117ec565b6103dc565b6040516101479190611844565b60405180910390f35b6101586103fe565b604051610165919061186c565b60405180910390f35b61018860048036038101906101839190611885565b610407565b6040516101959190611844565b60405180910390f35b6101a6610435565b6040516101b391906118f0565b60405180910390f35b6101c461043d565b6040516101d19190611921565b60405180910390f35b6101f460048036038101906101ef91906117ec565b61044b565b005b610210600480360381019061020b919061193a565b610459565b60405161021d919061186c565b60405180910390f35b610240600480360381019061023b919061193a565b61049e565b60405161024d919061186c565b60405180910390f35b61025e6104af565b6040516102719796959493929190611a65565b60405180910390f35b610294600480360381019061028f9190611b65565b610554565b005b61029e6106d4565b6040516102ab919061173b565b60405180910390f35b6102ce60048036038101906102c991906117ec565b610764565b005b6102ea60048036038101906102e591906117ec565b610772565b6040516102f79190611844565b60405180910390f35b61031a60048036038101906103159190611c16565b610794565b005b61033660048036038101906103319190611cb3565b6108d9565b604051610343919061186c565b60405180910390f35b60606003805461035b90611d1e565b80601f016020809104026020016040519081016040528092919081815260200182805461038790611d1e565b80156103d25780601f106103a9576101008083540402835291602001916103d2565b820191905f5260205f20905b8154815290600101906020018083116103b557829003601f168201915b5050505050905090565b5f5f6103e661095b565b90506103f3818585610962565b600191505092915050565b5f600254905090565b5f5f61041161095b565b905061041e858285610974565b610429858585610a07565b60019150509392505050565b5f6012905090565b5f610446610af7565b905090565b6104558282610bad565b5050565b5f5f5f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b5f6104a882610c2c565b9050919050565b5f6060805f5f5f60606104c0610c72565b6104c8610cad565b46305f5f1b5f67ffffffffffffffff8111156104e7576104e6611d4e565b5b6040519080825280602002602001820160405280156105155781602001602082028036833780820191505090505b507f0f00000000000000000000000000000000000000000000000000000000000000959493929190965096509650965096509650965090919293949596565b5f851415801561056357508442115b156105a557846040517f9e96b26f00000000000000000000000000000000000000000000000000000000815260040161059c919061186c565b60405180910390fd5b5f7fea2aa0a1be11a07ed86d755c93467f4f82362b452371d1ba94d1715123511acb89898989896040516020016105e196959493929190611d7b565b6040516020818303038152906040528051906020012090505f61060382610ce8565b90505f61061282878787610d01565b90508a73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161461068657808b6040517f7ba5ffb500000000000000000000000000000000000000000000000000000000815260040161067d929190611dda565b60405180910390fd5b6106908b8a610d2f565b6106c78b8b896106a0575f6106c2565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff5b610962565b5050505050505050505050565b6060600480546106e390611d1e565b80601f016020809104026020016040519081016040528092919081815260200182805461070f90611d1e565b801561075a5780601f106107315761010080835404028352916020019161075a565b820191905f5260205f20905b81548152906001019060200180831161073d57829003601f168201915b5050505050905090565b61076e8282610d86565b5050565b5f5f61077c61095b565b9050610789818585610a07565b600191505092915050565b834211156107d957836040517f627913020000000000000000000000000000000000000000000000000000000081526004016107d0919061186c565b60405180910390fd5b5f7f6e71edae12b1b97f4d1f60370fef10105fa2faae0126114a169c64845d6126c98888886108078c610e05565b8960405160200161081d96959493929190611e01565b6040516020818303038152906040528051906020012090505f61083f82610ce8565b90505f61084e82878787610d01565b90508973ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16146108c257808a6040517f4b800e460000000000000000000000000000000000000000000000000000000081526004016108b9929190611dda565b60405180910390fd5b6108cd8a8a8a610962565b50505050505050505050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f33905090565b61096f8383836001610e58565b505050565b5f61097f84846108d9565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff811015610a0157818110156109f2578281836040517ffb8f41b20000000000000000000000000000000000000000000000000000000081526004016109e993929190611e60565b60405180910390fd5b610a0084848484035f610e58565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610a77575f6040517f96c6fd1e000000000000000000000000000000000000000000000000000000008152600401610a6e9190611e95565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610ae7575f6040517fec442f05000000000000000000000000000000000000000000000000000000008152600401610ade9190611e95565b60405180910390fd5b610af2838383611027565b505050565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff16148015610b7257507f000000000000000000000000000000000000000000000000000000000000000046145b15610b9f577f00000000000000000000000000000000000000000000000000000000000000009050610baa565b610ba7611240565b90505b90565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610c1d575f6040517fec442f05000000000000000000000000000000000000000000000000000000008152600401610c149190611e95565b60405180910390fd5b610c285f8383611027565b5050565b5f60075f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b6060610ca860057f00000000000000000000000000000000000000000000000000000000000000006112d590919063ffffffff16565b905090565b6060610ce360067f00000000000000000000000000000000000000000000000000000000000000006112d590919063ffffffff16565b905090565b5f610cfa610cf4610af7565b83611382565b9050919050565b5f5f5f5f610d11888888886113c2565b925092509250610d2182826114a9565b829350505050949350505050565b5f610d3983610e05565b9050808214610d815782816040517f752d88c0000000000000000000000000000000000000000000000000000000008152600401610d78929190611eae565b60405180910390fd5b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610df6575f6040517f96c6fd1e000000000000000000000000000000000000000000000000000000008152600401610ded9190611e95565b60405180910390fd5b610e01825f83611027565b5050565b5f60075f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f815480929190600101919050559050919050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610ec8575f6040517fe602df05000000000000000000000000000000000000000000000000000000008152600401610ebf9190611e95565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610f38575f6040517f94280d62000000000000000000000000000000000000000000000000000000008152600401610f2f9190611e95565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20819055508015611021578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92584604051611018919061186c565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603611077578060025f82825461106b9190611f02565b92505081905550611145565b5f5f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905081811015611100578381836040517fe450d38c0000000000000000000000000000000000000000000000000000000081526004016110f793929190611e60565b60405180910390fd5b8181035f5f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361118c578060025f82825403925050819055506111d6565b805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051611233919061186c565b60405180910390a3505050565b5f7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f7f00000000000000000000000000000000000000000000000000000000000000007f000000000000000000000000000000000000000000000000000000000000000046306040516020016112ba959493929190611f35565b60405160208183030381529060405280519060200120905090565b606060ff5f1b83146112f1576112ea8361160b565b905061137c565b8180546112fd90611d1e565b80601f016020809104026020016040519081016040528092919081815260200182805461132990611d1e565b80156113745780601f1061134b57610100808354040283529160200191611374565b820191905f5260205f20905b81548152906001019060200180831161135757829003601f168201915b505050505090505b92915050565b5f6040517f190100000000000000000000000000000000000000000000000000000000000081528360028201528260228201526042812091505092915050565b5f5f5f7f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0845f1c11156113fe575f60038592509250925061149f565b5f6001888888886040515f81526020016040526040516114219493929190611f86565b6020604051602081039080840390855afa158015611441573d5f5f3e3d5ffd5b5050506020604051035190505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603611492575f60015f5f1b9350935093505061149f565b805f5f5f1b935093509350505b9450945094915050565b5f60038111156114bc576114bb611fc9565b5b8260038111156114cf576114ce611fc9565b5b031561160757600160038111156114e9576114e8611fc9565b5b8260038111156114fc576114fb611fc9565b5b03611533576040517ff645eedf00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6002600381111561154757611546611fc9565b5b82600381111561155a57611559611fc9565b5b0361159e57805f1c6040517ffce698f7000000000000000000000000000000000000000000000000000000008152600401611595919061186c565b60405180910390fd5b6003808111156115b1576115b0611fc9565b5b8260038111156115c4576115c3611fc9565b5b0361160657806040517fd78bce0c0000000000000000000000000000000000000000000000000000000081526004016115fd9190611921565b60405180910390fd5b5b5050565b60605f6116178361167d565b90505f602067ffffffffffffffff81111561163557611634611d4e565b5b6040519080825280601f01601f1916602001820160405280156116675781602001600182028036833780820191505090505b5090508181528360208201528092505050919050565b5f5f60ff835f1c169050601f8111156116c2576040517fb3512b0c00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b80915050919050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f61170d826116cb565b61171781856116d5565b93506117278185602086016116e5565b611730816116f3565b840191505092915050565b5f6020820190508181035f8301526117538184611703565b905092915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6117888261175f565b9050919050565b6117988161177e565b81146117a2575f5ffd5b50565b5f813590506117b38161178f565b92915050565b5f819050919050565b6117cb816117b9565b81146117d5575f5ffd5b50565b5f813590506117e6816117c2565b92915050565b5f5f604083850312156118025761180161175b565b5b5f61180f858286016117a5565b9250506020611820858286016117d8565b9150509250929050565b5f8115159050919050565b61183e8161182a565b82525050565b5f6020820190506118575f830184611835565b92915050565b611866816117b9565b82525050565b5f60208201905061187f5f83018461185d565b92915050565b5f5f5f6060848603121561189c5761189b61175b565b5b5f6118a9868287016117a5565b93505060206118ba868287016117a5565b92505060406118cb868287016117d8565b9150509250925092565b5f60ff82169050919050565b6118ea816118d5565b82525050565b5f6020820190506119035f8301846118e1565b92915050565b5f819050919050565b61191b81611909565b82525050565b5f6020820190506119345f830184611912565b92915050565b5f6020828403121561194f5761194e61175b565b5b5f61195c848285016117a5565b91505092915050565b5f7fff0000000000000000000000000000000000000000000000000000000000000082169050919050565b61199981611965565b82525050565b6119a88161177e565b82525050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b6119e0816117b9565b82525050565b5f6119f183836119d7565b60208301905092915050565b5f602082019050919050565b5f611a13826119ae565b611a1d81856119b8565b9350611a28836119c8565b805f5b83811015611a58578151611a3f88826119e6565b9750611a4a836119fd565b925050600181019050611a2b565b5085935050505092915050565b5f60e082019050611a785f83018a611990565b8181036020830152611a8a8189611703565b90508181036040830152611a9e8188611703565b9050611aad606083018761185d565b611aba608083018661199f565b611ac760a0830185611912565b81810360c0830152611ad98184611a09565b905098975050505050505050565b611af08161182a565b8114611afa575f5ffd5b50565b5f81359050611b0b81611ae7565b92915050565b611b1a816118d5565b8114611b24575f5ffd5b50565b5f81359050611b3581611b11565b92915050565b611b4481611909565b8114611b4e575f5ffd5b50565b5f81359050611b5f81611b3b565b92915050565b5f5f5f5f5f5f5f5f610100898b031215611b8257611b8161175b565b5b5f611b8f8b828c016117a5565b9850506020611ba08b828c016117a5565b9750506040611bb18b828c016117d8565b9650506060611bc28b828c016117d8565b9550506080611bd38b828c01611afd565b94505060a0611be48b828c01611b27565b93505060c0611bf58b828c01611b51565b92505060e0611c068b828c01611b51565b9150509295985092959890939650565b5f5f5f5f5f5f5f60e0888a031215611c3157611c3061175b565b5b5f611c3e8a828b016117a5565b9750506020611c4f8a828b016117a5565b9650506040611c608a828b016117d8565b9550506060611c718a828b016117d8565b9450506080611c828a828b01611b27565b93505060a0611c938a828b01611b51565b92505060c0611ca48a828b01611b51565b91505092959891949750929550565b5f5f60408385031215611cc957611cc861175b565b5b5f611cd6858286016117a5565b9250506020611ce7858286016117a5565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680611d3557607f821691505b602082108103611d4857611d47611cf1565b5b50919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b5f60c082019050611d8e5f830189611912565b611d9b602083018861199f565b611da8604083018761199f565b611db5606083018661185d565b611dc2608083018561185d565b611dcf60a0830184611835565b979650505050505050565b5f604082019050611ded5f83018561199f565b611dfa602083018461199f565b9392505050565b5f60c082019050611e145f830189611912565b611e21602083018861199f565b611e2e604083018761199f565b611e3b606083018661185d565b611e48608083018561185d565b611e5560a083018461185d565b979650505050505050565b5f606082019050611e735f83018661199f565b611e80602083018561185d565b611e8d604083018461185d565b949350505050565b5f602082019050611ea85f83018461199f565b92915050565b5f604082019050611ec15f83018561199f565b611ece602083018461185d565b9392505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611f0c826117b9565b9150611f17836117b9565b9250828201905080821115611f2f57611f2e611ed5565b5b92915050565b5f60a082019050611f485f830188611912565b611f556020830187611912565b611f626040830186611912565b611f6f606083018561185d565b611f7c608083018461199f565b9695505050505050565b5f608082019050611f995f830187611912565b611fa660208301866118e1565b611fb36040830185611912565b611fc06060830184611912565b95945050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffdfea2646970667358221220faca0f7cea2c9328299ad2f0bf911d29a083c468837794bcb5cfcdd6dd2bdc1e64736f6c63430008230033"

type RouterSwapSuite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	RouterHex     string
	QuoterHex     string
	TokenInHex    string
	TokenOutHex   string
	HopsHex       []string
	FeesHex       []string
	NativeIn      bool
	NativeOut     bool
	MulticallHex  string
	AmountIn      string
	NativeSymbol  string
	RoundTrip     bool
	Interval      int
	Timeout       int
	PriceBuffer   uint64
	LimitBuffer   uint64

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	from      *types.Address
	router    *types.Address
	quoter    *types.Address
	weth9Addr *types.Address
	tokenIn   *types.Address
	tokenOut  *types.Address
	hops      []*types.Address
	fees      []*big.Int
	multicall *types.Address
	amount    *big.Int

	v3Router *Router
	v3Quoter *Quoter
	erc20    *erc20.ERC20

	totalFee *big.Int
}

func TestRouterSwapSuite(t *testing.T) {
	suite.Run(t, &RouterSwapSuite{
		Enable:        false,
		RoundTrip:     false,
		RPCURL:        "",
		PrivateKeyHex: "",
		RouterHex:     "",
		QuoterHex:     "",
		TokenInHex:    "",
		TokenOutHex:   "",
		HopsHex:       []string{},
		FeesHex:       []string{strconv.Itoa(FeeLow)},
		NativeIn:      false,
		NativeOut:     false,
		MulticallHex:  multicall3.DefaultMulticall3Address.String(),
		AmountIn:      "",
		NativeSymbol:  "",
		Interval:      200,
		Timeout:       60,
		PriceBuffer:   150,
		LimitBuffer:   150,
	})
}

func (s *RouterSwapSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.v3Router = NewRouter(s.client)
	s.v3Quoter = NewQuoter(s.client)
	s.erc20 = erc20.New(s.client)

	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())

	s.router, err = types.NewAddressFromHex(s.RouterHex)
	s.Require().NoError(err)

	s.quoter, err = types.NewAddressFromHex(s.QuoterHex)
	s.Require().NoError(err)

	s.Require().False(s.NativeIn && s.NativeOut, "NativeIn and NativeOut can't both be set -- that's not a swap")

	s.weth9Addr, err = s.v3Router.WETH9(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)

	if s.NativeIn {
		s.tokenIn = s.weth9Addr
	} else {
		s.tokenIn, err = types.NewAddressFromHex(s.TokenInHex)
		s.Require().NoError(err)
	}

	if s.NativeOut {
		s.tokenOut = s.weth9Addr
	} else {
		s.tokenOut, err = types.NewAddressFromHex(s.TokenOutHex)
		s.Require().NoError(err)
	}

	s.hops = make([]*types.Address, 0, len(s.HopsHex))
	for _, h := range s.HopsHex {
		hop, err := types.NewAddressFromHex(h)
		s.Require().NoError(err)
		s.hops = append(s.hops, hop)
	}

	s.Require().Len(s.FeesHex, len(s.hops)+1, "need exactly one fee tier per pool in the path (len(hops)+1)")
	s.fees = make([]*big.Int, 0, len(s.FeesHex))
	for _, f := range s.FeesHex {
		fee, ok := new(big.Int).SetString(f, 10)
		s.Require().True(ok)
		s.fees = append(s.fees, fee)
	}

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	amount, ok := new(big.Int).SetString(s.AmountIn, 10)
	s.Require().True(ok)
	s.amount = amount

	path := append([]*types.Address{s.tokenIn}, s.hops...)
	path = append(path, s.tokenOut)
	previewPath := append([]byte{}, path[0].Bytes()...)
	for i, fee := range s.fees {
		feeBytes := make([]byte, 3)
		feeRaw := fee.Bytes()
		copy(feeBytes[3-len(feeRaw):], feeRaw)
		previewPath = append(previewPath, feeBytes...)
		previewPath = append(previewPath, path[i+1].Bytes()...)
	}
	if s.NativeIn {
		previewParams := NewExactInputParams(previewPath, s.from, s.amount, big.NewInt(0))
		_, err = s.v3Router.CallExactInput(s.ctx, s.router, s.from, s.amount, previewParams, rpc.BlockTagLatest)
		s.Require().NoError(err, "no route from tokenIn to tokenOut via this router, at this fee tier")
	}
}

func (s *RouterSwapSuite) TearDownSuite() {
	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%018s", frac.String()), "0")
		decimal := whole.String()
		if fracStr != "" {
			decimal += "." + fracStr
		}

		s.T().Logf("RouterSwapSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RouterSwapSuite) TestSwap() {
	tokenInMeta, err := s.erc20.MetadataWithBalance(s.ctx, s.multicall, s.tokenIn, s.from, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("TestSwap: TokenIn %s", s.tokenIn)
	s.T().Logf("\t %s (%s), decimals: %d", tokenInMeta.Name, tokenInMeta.Symbol, tokenInMeta.Decimals)
	s.T().Logf("\t %s amounts: %s", s.tokenIn, tokenInMeta.Balance.String())

	tokenOutMeta, err := s.erc20.MetadataWithBalance(s.ctx, s.multicall, s.tokenOut, s.from, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("TestSwap: TokenOut %s", s.tokenOut)
	s.T().Logf("\t %s (%s), decimals: %d", tokenOutMeta.Name, tokenOutMeta.Symbol, tokenOutMeta.Decimals)
	s.T().Logf("\t %s amounts: %s", s.tokenOut, tokenOutMeta.Balance.String())

	forwardPath := append([]*types.Address{s.tokenIn}, s.hops...)
	forwardPath = append(forwardPath, s.tokenOut)
	forwardFees := s.fees

	reversePath := make([]*types.Address, len(forwardPath))
	reverseFees := make([]*big.Int, len(forwardFees))
	for i, addr := range forwardPath {
		reversePath[len(forwardPath)-1-i] = addr
	}
	for i, fee := range forwardFees {
		reverseFees[len(forwardFees)-1-i] = fee
	}

	received := s.swap(forwardPath, forwardFees, s.amount, s.NativeIn, s.NativeOut)
	if !s.RoundTrip {
		return
	}

	final := s.swap(reversePath, reverseFees, received, s.NativeOut, s.NativeIn)
	s.T().Logf("TestSwap: round trip %s -> %s -> %s: sent %s, received back %s", s.tokenIn, s.tokenOut, s.tokenIn, s.amount, final)

	if s.NativeOut {
		return
	}
	tokenOutBalanceAfter, err := s.erc20.BalanceOf(s.ctx, s.tokenOut, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(tokenOutBalanceAfter.Cmp(tokenOutMeta.Balance), "wallet's TokenOut balance changed (%s -> %s) after swapping it all back", tokenOutMeta.Balance, tokenOutBalanceAfter)
}

func (s *RouterSwapSuite) swap(path []*types.Address, fees []*big.Int, amountIn *big.Int, nativeIn, nativeOut bool) *big.Int {
	tokenIn, tokenOut := path[0], path[len(path)-1]

	if nativeIn {
		var hex string
		s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &hex)))
		balance, ok := new(big.Int).SetString(strings.TrimPrefix(hex, "0x"), 16)
		s.Require().True(ok)
		s.Require().True(balance.Cmp(amountIn) >= 0, "wallet holds %s native, less than amountIn %s", balance, amountIn)
		s.T().Logf("swap %s -> %s: wallet native balance: %s", tokenIn, tokenOut, balance)
	} else {
		balance, err := s.erc20.BalanceOf(s.ctx, tokenIn, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(balance.Cmp(amountIn) >= 0, "wallet holds %s of tokenIn, less than amountIn %s", balance, amountIn)
		s.T().Logf("swap %s -> %s: wallet balance of tokenIn: %s", tokenIn, tokenOut, balance)

		approveBuilder := tx.NewDynamicFeeBuilder()
		approveBuilder.PriceBuffer(s.PriceBuffer)
		approveBuilder.LimitBuffer(s.LimitBuffer)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, approveBuilder, tokenIn, s.from, s.router, amountIn)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("swap %s -> %s: approve %s to spend %s, broadcast tx %s", tokenIn, tokenOut, s.router, amountIn, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, tokenIn, s.from, s.router, rpc.BlockTagLatest)
			return err == nil && amountIn.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("swap %s -> %s: approve confirmed in block %d", tokenIn, tokenOut, approveReceipt.BlockNumber)
	}

	encodedPath := append([]byte{}, path[0].Bytes()...)
	for i, fee := range fees {
		feeBytes := make([]byte, 3)
		feeRaw := fee.Bytes()
		copy(feeBytes[3-len(feeRaw):], feeRaw)
		encodedPath = append(encodedPath, feeBytes...)
		encodedPath = append(encodedPath, path[i+1].Bytes()...)
	}

	recipient := s.from
	if nativeOut {
		recipient = types.MustNewAddressFromHex("0x0000000000000000000000000000000000000002")
	}

	var previewValue *big.Int
	if nativeIn {
		previewValue = amountIn
	}

	var amountOutMin *big.Int
	var swapCalldata []byte
	var lastErr error
	attempt := 0
	s.Require().Eventually(func() bool {
		attempt++
		previewParams := NewExactInputParams(encodedPath, recipient, amountIn, big.NewInt(0))
		amountOutLast, err := s.v3Router.CallExactInput(s.ctx, s.router, s.from, previewValue, previewParams, rpc.BlockTagLatest)
		if err != nil {
			lastErr = err
			return false
		}
		amountOutMin = new(big.Int).Div(new(big.Int).Mul(amountOutLast, big.NewInt(995)), big.NewInt(1000))
		s.T().Logf("swap %s -> %s: attempt %d: CallExactInput=%s amountOutMin=%s", tokenIn, tokenOut, attempt, amountOutLast, amountOutMin)

		calldata, err := EncodeExactInput(encodedPath, recipient, amountIn, amountOutMin)
		if err != nil {
			lastErr = err
			return false
		}
		swapCalldata = calldata
		return true
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "swap not buildable: %v", lastErr)

	quoterQuote, err := s.v3Quoter.CallQuoteExactInput(s.ctx, s.quoter, s.from, encodedPath, amountIn, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("swap %s -> %s: Quoter's own quote taken right before broadcast: amountOut=%s", tokenIn, tokenOut, quoterQuote.AmountOut)

	var unsignedSwap core.Transaction
	if nativeOut {
		unwrapCalldata, encErr := EncodeRouterUnwrapWETH9(amountOutMin, s.from)
		s.Require().NoError(encErr)
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(s.PriceBuffer)
		builder.LimitBuffer(s.LimitBuffer)
		unsignedSwap, err = s.v3Router.CreateMulticall(s.ctx, builder, s.router, s.from, previewValue, [][]byte{swapCalldata, unwrapCalldata})
	} else {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(s.PriceBuffer)
		builder.LimitBuffer(s.LimitBuffer)
		unsignedSwap, err = builder.Build(s.from, s.router, previewValue, swapCalldata).Pack(s.ctx, s.client)
	}
	s.Require().NoError(err)
	s.Require().NoError(unsignedSwap.Sign(s.signer))
	s.totalFee.Add(s.totalFee, unsignedSwap.MaxFee())
	s.T().Logf("swap %s -> %s: swap tx built after %d attempt(s), amountOutMin=%s", tokenIn, tokenOut, attempt, amountOutMin)

	swapRaw, err := unsignedSwap.EncodeRLP()
	s.Require().NoError(err)
	swapHash, err := tx.Broadcast(s.ctx, s.client, swapRaw)
	s.Require().NoError(err)
	s.T().Logf("swap %s -> %s: broadcast swap, tx %s", tokenIn, tokenOut, swapHash)

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "swap transaction reverted")
	s.T().Logf("swap %s -> %s: swap confirmed in block %d", tokenIn, tokenOut, swapReceipt.BlockNumber)

	swaps, err := ExtractSwap(swapReceipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, len(path)-1)

	transfers, err := erc20.ExtractTransfers(swapReceipt.Logs)
	s.Require().NoError(err)

	lastSwap := swaps[len(swaps)-1]
	var actualAmountOut *big.Int
	if lastSwap.Amount0.Sign() < 0 {
		actualAmountOut = new(big.Int).Neg(lastSwap.Amount0)
	} else {
		actualAmountOut = new(big.Int).Neg(lastSwap.Amount1)
	}

	drift := new(big.Int).Sub(actualAmountOut, quoterQuote.AmountOut)
	driftPct := new(big.Float).Quo(new(big.Float).SetInt(drift), new(big.Float).SetInt(quoterQuote.AmountOut))
	driftPct.Mul(driftPct, big.NewFloat(100))

	s.T().Logf("Router swap %s -> %s Result", tokenIn, tokenOut)
	s.T().Logf("\tAmountIn: %s", amountIn)
	s.T().Logf("\tAmountOut: %s", actualAmountOut)
	s.T().Logf("\tQuoter's pre-swap quote: %s (drift %s, %s%%)", quoterQuote.AmountOut, drift, driftPct.Text('f', 6))
	s.T().Logf("\tTx: %s", swapHash)
	s.T().Logf("\tLogs:")
	for _, transfer := range transfers {
		s.T().Logf("\t\tTransfer: contract=%s from=%s to=%s value=%s", transfer.Contract, transfer.From, transfer.To, transfer.Value)
	}
	if nativeIn {
		deposits, err := weth.ExtractDeposits(swapReceipt.Logs)
		s.Require().NoError(err)
		for _, d := range deposits {
			s.T().Logf("\t\tDeposit: dst=%s wad=%s", d.Dst, d.Wad)
		}
	}
	if nativeOut {
		withdrawals, err := weth.ExtractWithdrawals(swapReceipt.Logs)
		s.Require().NoError(err)
		for _, w := range withdrawals {
			s.T().Logf("\t\tWithdrawal: src=%s wad=%s", w.Src, w.Wad)
		}
	}
	for _, sw := range swaps {
		s.T().Logf("\t\tSwap: contract=%s sender=%s recipient=%s amount0=%s amount1=%s", sw.Contract, sw.Sender, sw.Recipient, sw.Amount0, sw.Amount1)
	}

	return actualAmountOut
}

type RouterLifeCycleSuite struct {
	suite.Suite

	Enable          bool
	RPCURL          string
	PrivateKeyHex   string
	RouterHex       string
	NfpmHex         string
	MulticallHex    string
	NativeSymbol    string
	Fee             *big.Int
	Interval        int
	Timeout         int
	ByteCode        string
	Constructor     *abi.Function
	Salt            *types.Hash
	MintValue       string
	FeeRecipientHex string

	ctx    context.Context
	cancel context.CancelFunc

	client       rpc.Client
	signer       core.Signer
	deployer     *deploy.NicksDeployer
	from         *types.Address
	router       *types.Address
	nfpm         *types.Address
	factory      *types.Address
	weth9        *types.Address
	multicall    *types.Address
	tokenA       *types.Address
	tokenB       *types.Address
	tokenC       *types.Address
	pool         *types.Address
	poolBC       *types.Address
	poolWETH     *types.Address
	wethToken0   *types.Address
	wethToken1   *types.Address
	feeRecipient *types.Address
	chainID      *big.Int
	tokenId      *big.Int

	v3Router  *Router
	v3Nfpm    *NFPM
	v3Factory *Factory
	v3Pool    *Pool
	erc20     *erc20.ERC20

	mu            *sync.Mutex
	totalFee      *big.Int
	mintValue     *big.Int
	wethLiquidity *big.Int
}

func TestRouterLifeCycleSuite(t *testing.T) {
	suite.Run(t, &RouterLifeCycleSuite{
		Enable:          false,
		RPCURL:          "",
		PrivateKeyHex:   "",
		FeeRecipientHex: "",
		NfpmHex:         "",
		RouterHex:       "",
		MulticallHex:    multicall3.DefaultMulticall3Address.String(),
		NativeSymbol:    "",
		Fee:             big.NewInt(FeeMedium),
		Interval:        1500,
		Timeout:         60,
		ByteCode:        erc20PermitFullByteCodeHex,
		Constructor:     abi.NewFunction("constructor", abi.Types{abi.String, abi.String}, nil),
		Salt:            core.Keccak256([]byte("go-evmkit")),
		MintValue:       "1000000000000000000000000",
	})
}

func (s *RouterLifeCycleSuite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.v3Router = NewRouter(s.client)
	s.v3Nfpm = NewNFPM(s.client)
	s.v3Factory = NewFactory(s.client)
	s.v3Pool = NewPool(s.client)
	s.erc20 = erc20.New(s.client)
	s.T().Logf("SetupSuite: client ready for %s", s.RPCURL)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	privKey, err := types.NewPrivateKeyFromHex(s.PrivateKeyHex)
	s.Require().NoError(err)
	s.signer = privKey
	s.from = core.PubkeyToAddress(privKey.PublicKey())
	s.T().Logf("SetupSuite: signing as %s", s.from)

	s.router, err = types.NewAddressFromHex(s.RouterHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target SwapRouter %s (fee tier %s)", s.router, s.Fee)

	s.nfpm, err = types.NewAddressFromHex(s.NfpmHex)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: target NonfungiblePositionManager %s", s.nfpm)

	s.multicall, err = types.NewAddressFromHex(s.MulticallHex)
	s.Require().NoError(err)

	s.feeRecipient, err = types.NewAddressFromHex(s.FeeRecipientHex)
	s.Require().NoError(err)

	mintValue, ok := new(big.Int).SetString(s.MintValue, 10)
	s.Require().True(ok)
	s.mintValue = mintValue

	var chainIDHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHChainID("chainId", &chainIDHex)))
	chainID, ok := new(big.Int).SetString(strings.TrimPrefix(chainIDHex, "0x"), 16)
	s.Require().True(ok)
	s.chainID = chainID
	s.T().Logf("SetupSuite: chainID = %s", s.chainID)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	names := [3]string{"evkit test router token A", "evkit test router token B", "evkit test router token C"}
	symbols := [3]string{"EVKRA", "EVKRB", "EVKRC"}
	tokens := [3]*types.Address{}

	for i := range 3 {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, []any{names[i], symbols[i]}),
			salt,
		)

		predicted, err := s.deployer.DeployAddress(deploy.DefaultNicksMethodFactory, deployment)
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: deploying token %s %s, predicted address %s", names[i], symbols[i], predicted)

		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)

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
		s.T().Logf("SetupSuite: signed deployment of %s (nonce %d, maxFee %s)", symbols[i], unsigned.AddressNonce(), unsigned.MaxFee())

		hash, err := s.deployer.Deploy(s.ctx, unsigned)
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: broadcast deployment of %s, tx %s", symbols[i], hash)

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")
		s.T().Logf("SetupSuite: %s mined in block %d", symbols[i], receipt.BlockNumber)

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

		tokenMintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
		calldata, err := tokenMintFn.EncodeCall(s.from, s.mintValue)
		s.Require().NoError(err)

		mintBuilder := tx.NewDynamicFeeBuilder()
		mintBuilder.PriceBuffer(150)
		mintBuilder.LimitBuffer(300)
		unsignedMint, err := mintBuilder.Build(s.from, token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
		s.Require().NoError(err)
		s.Require().NoError(unsignedMint.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedMint.MaxFee())
		s.mu.Unlock()
		s.T().Logf("SetupSuite: signed mint of %s to %s (nonce %d, gasLimit %d, maxFee %s)", symbols[i], s.from, unsignedMint.AddressNonce(), unsignedMint.Gas(), unsignedMint.MaxFee())

		mintRaw, err := unsignedMint.EncodeRLP()
		s.Require().NoError(err)
		mintHash, err := tx.Broadcast(s.ctx, s.client, mintRaw)
		s.Require().NoError(err)
		s.T().Logf("SetupSuite: broadcast mint of %s, tx %s", symbols[i], mintHash)

		mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(mintReceipt.Status, "mint transaction reverted")
		s.T().Logf("SetupSuite: minted %s of %s to %s, confirmed in block %d (tx %s)", s.mintValue, token, s.from, mintReceipt.BlockNumber, mintHash)
	}

	sorted := tokens[:]
	slices.SortFunc(sorted, func(a, b *types.Address) int { return a.Big().Cmp(b.Big()) })
	s.tokenA, s.tokenB, s.tokenC = sorted[0], sorted[1], sorted[2]
	s.T().Logf("SetupSuite: sorted token0=%s token1=%s token2=%s", s.tokenA, s.tokenB, s.tokenC)

	weth9, err := s.v3Router.WETH9(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(weth9.IsZero())
	s.weth9 = weth9

	var nativeBalanceHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceHex)))
	nativeBalance, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceHex, "0x"), 16)
	s.Require().True(ok)
	s.wethLiquidity = new(big.Int).Div(nativeBalance, big.NewInt(10000))
	s.Require().True(s.wethLiquidity.Sign() > 0, "native balance %s is too small to spare 1/10000 for WETH liquidity", nativeBalance)
	s.T().Logf("SetupSuite: native balance of %s is %s, wrapping 1/10000 of it (%s) into WETH9 %s", s.from, nativeBalance, s.wethLiquidity, weth9)

	depositFn := abi.NewFunction("deposit", nil, nil)
	wrapBuilder := tx.NewDynamicFeeBuilder()
	wrapBuilder.PriceBuffer(150)
	wrapBuilder.LimitBuffer(300)
	unsignedWrap, err := wrapBuilder.Build(s.from, weth9, s.wethLiquidity, depositFn.SelectorBytes()).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsignedWrap.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedWrap.MaxFee())
	s.mu.Unlock()
	s.T().Logf("SetupSuite: signed WETH9.deposit() of %s (nonce %d, maxFee %s)", s.wethLiquidity, unsignedWrap.AddressNonce(), unsignedWrap.MaxFee())

	wrapRaw, err := unsignedWrap.EncodeRLP()
	s.Require().NoError(err)
	wrapHash, err := tx.Broadcast(s.ctx, s.client, wrapRaw)
	s.Require().NoError(err)
	s.T().Logf("SetupSuite: broadcast WETH9.deposit(), tx %s", wrapHash)

	wrapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, wrapHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(wrapReceipt.Status, "WETH9.deposit() transaction reverted")
	s.T().Logf("SetupSuite: wrapped %s into WETH9, confirmed in block %d (tx %s)", s.wethLiquidity, wrapReceipt.BlockNumber, wrapHash)
}

func (s *RouterLifeCycleSuite) TearDownSuite() {
	if s.tokenA != nil && s.tokenB != nil && s.tokenC != nil {
		balances, err := s.erc20.AddressBalances(s.ctx, s.multicall, s.from, []*types.Address{s.tokenA, s.tokenB, s.tokenC}, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)

		for _, tb := range balances.Tokens {
			if tb.Amount == nil || tb.Amount.Sign() == 0 {
				continue
			}

			s.T().Logf("TearDownSuite: leftover balance of %s is %s, burning it", tb.Token, tb.Amount)

			calldata, err := tokenBurnFn.EncodeCall(s.from, tb.Amount)
			s.Require().NoError(err)

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, tb.Token, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
			s.mu.Unlock()
			s.T().Logf("TearDownSuite: signed burn of %s (nonce %d, maxFee %s)", tb.Token, unsignedBurn.AddressNonce(), unsignedBurn.MaxFee())

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)
			s.T().Logf("TearDownSuite: broadcast burn of %s, tx %s", tb.Token, burnHash)

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")

			s.T().Logf("TearDownSuite: burned %s of %s from %s, confirmed in block %d (tx %s)", tb.Amount, tb.Token, s.from, burnReceipt.BlockNumber, burnHash)
		}
	}

	if s.feeRecipient != nil && s.tokenA != nil {
		feeRecipientBalance, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.feeRecipient, rpc.BlockTagLatest)
		s.Require().NoError(err)

		if feeRecipientBalance.Sign() > 0 {
			s.T().Logf("TearDownSuite: feeRecipient %s holds %s of %s (collected fees), burning it (this test token's burn() is unrestricted by design)", s.feeRecipient, feeRecipientBalance, s.tokenA)

			tokenBurnFn := abi.NewFunction("burn", abi.NewTypes(abi.Address, abi.Uint256), nil)
			calldata, err := tokenBurnFn.EncodeCall(s.feeRecipient, feeRecipientBalance)
			s.Require().NoError(err)

			unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, s.tokenA, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedBurn.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
			s.mu.Unlock()

			raw, err := unsignedBurn.EncodeRLP()
			s.Require().NoError(err)
			burnHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)

			burnReceipt, err := tx.WaitForReceipt(s.ctx, s.client, burnHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(burnReceipt.Status, "burn transaction reverted")
			s.T().Logf("TearDownSuite: burned %s of %s from feeRecipient %s, confirmed in block %d (tx %s)", feeRecipientBalance, s.tokenA, s.feeRecipient, burnReceipt.BlockNumber, burnHash)
		}

		var feeRecipientNativeHex string
		s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.feeRecipient.String(), rpc.BlockTagLatest, &feeRecipientNativeHex)))
		feeRecipientNative, ok := new(big.Int).SetString(strings.TrimPrefix(feeRecipientNativeHex, "0x"), 16)
		s.Require().True(ok)
		if feeRecipientNative.Sign() > 0 {
			s.T().Logf("TearDownSuite: feeRecipient %s holds %s native %s (collected WETH9 fee, now unwrapped) -- not reclaimable, we don't hold its private key", s.feeRecipient, feeRecipientNative, s.NativeSymbol)
		}
	}

	if s.weth9 != nil {
		wethBalance, err := s.erc20.BalanceOf(s.ctx, s.weth9, s.from, rpc.BlockTagLatest)
		s.Require().NoError(err)

		if wethBalance.Sign() > 0 {
			s.T().Logf("TearDownSuite: leftover WETH9 balance is %s, withdrawing it back to native", wethBalance)

			withdrawFn := abi.NewFunction("withdraw", abi.NewTypes(abi.Uint256), nil)
			calldata, err := withdrawFn.EncodeCall(wethBalance)
			s.Require().NoError(err)

			unsignedWithdraw, err := tx.NewDynamicFeeBuilder().Build(s.from, s.weth9, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			s.Require().NoError(err)
			s.Require().NoError(unsignedWithdraw.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, unsignedWithdraw.MaxFee())
			s.mu.Unlock()
			s.T().Logf("TearDownSuite: signed WETH9.withdraw(%s) (nonce %d, maxFee %s)", wethBalance, unsignedWithdraw.AddressNonce(), unsignedWithdraw.MaxFee())

			raw, err := unsignedWithdraw.EncodeRLP()
			s.Require().NoError(err)
			withdrawHash, err := tx.Broadcast(s.ctx, s.client, raw)
			s.Require().NoError(err)
			s.T().Logf("TearDownSuite: broadcast WETH9.withdraw(), tx %s", withdrawHash)

			withdrawReceipt, err := tx.WaitForReceipt(s.ctx, s.client, withdrawHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
			s.Require().NoError(err)
			s.Require().True(withdrawReceipt.Status, "WETH9.withdraw() transaction reverted")
			s.T().Logf("TearDownSuite: withdrew %s of WETH9 back to native, confirmed in block %d (tx %s)", wethBalance, withdrawReceipt.BlockNumber, withdrawHash)
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

		s.T().Logf("RouterLifeCycleSuite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RouterLifeCycleSuite) Test01ReadMethods() {
	factory, err := s.v3Router.Factory(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(factory.IsZero())
	s.factory = factory
	s.T().Logf("Test01ReadMethods: factory() = %s", factory)

	weth9, err := s.v3Router.WETH9(s.ctx, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(weth9.IsZero())
	s.weth9 = weth9
	s.T().Logf("Test01ReadMethods: WETH9() = %s", weth9)

	nfpmFactory, err := s.v3Nfpm.Factory(s.ctx, s.nfpm, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(nfpmFactory.Equal(factory), "Router and NFPM are different periphery contracts, but they must share the same core Factory")
	s.T().Logf("Test01ReadMethods: NFPM's factory() = %s matches Router's", nfpmFactory)

	nfpmWETH9, err := s.v3Nfpm.WETH9(s.ctx, s.nfpm, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(nfpmWETH9.Equal(weth9), "Router and NFPM are different periphery contracts, but they must share the same WETH9")
	s.T().Logf("Test01ReadMethods: NFPM's WETH9() = %s matches Router's", nfpmWETH9)

	balances, err := s.erc20.AddressBalances(s.ctx, s.multicall, s.from, []*types.Address{s.tokenA, s.tokenB, s.tokenC}, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(balances.Tokens, 3)
	for _, tb := range balances.Tokens {
		s.Require().Zero(s.mintValue.Cmp(tb.Amount), "freshly minted token %s must hold exactly mintValue before any swap spends it", tb.Token)
		s.T().Logf("Test01ReadMethods: deployed token %s balance=%s", tb.Token, tb.Amount)
	}
}

func (s *RouterLifeCycleSuite) Test02CreateAndInitializePoolIfNecessary() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: token0=%s token1=%s fee=%s sqrtPriceX96=%s", s.tokenA, s.tokenB, s.Fee, sqrtPriceX96)

	predicted, err := s.v3Nfpm.CallCreateAndInitializePoolIfNecessary(s.ctx, s.nfpm, s.from, s.tokenA, s.tokenB, s.Fee, sqrtPriceX96, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(predicted.IsZero())
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: CallCreateAndInitializePoolIfNecessary predicted pool=%s", predicted)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateCreateAndInitializePoolIfNecessary(s.ctx, builder, s.nfpm, s.from, s.tokenA, s.tokenB, s.Fee, sqrtPriceX96)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: signed tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.CreateAndInitializePoolIfNecessary(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createAndInitializePoolIfNecessary transaction reverted")
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: confirmed in block %d", receipt.BlockNumber)

	s.pool = predicted

	initialized, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(initialized, 1)
	s.Require().True(initialized[0].Contract.Equal(s.pool))
	s.Require().Zero(sqrtPriceX96.Cmp(initialized[0].SqrtPriceX96))
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: Initialize event confirms pool=%s sqrtPriceX96=%s tick=%s", initialized[0].Contract, initialized[0].SqrtPriceX96, initialized[0].Tick)

	var registeredPool *types.Address
	s.Require().Eventually(func() bool {
		registeredPool, err = s.v3Factory.GetPool(s.ctx, s.factory, s.tokenA, s.tokenB, s.Fee, rpc.BlockTagLatest)
		return err == nil && registeredPool.Equal(s.pool)
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "Factory.getPool not yet reflecting the new pool")
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: Factory.getPool(token0, token1, fee) = %s matches", registeredPool)

	expectedTickSpacing, err := s.v3Factory.FeeAmountTickSpacing(s.ctx, s.factory, s.Fee, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(expectedTickSpacing.Sign() > 0, "fee tier %s isn't enabled on this Factory", s.Fee)

	poolToken0, err := s.v3Pool.Token0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolToken0.Equal(s.tokenA))
	poolToken1, err := s.v3Pool.Token1(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolToken1.Equal(s.tokenB))
	poolFee, err := s.v3Pool.Fee(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.Fee.Cmp(poolFee))
	poolTickSpacing, err := s.v3Pool.TickSpacing(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(expectedTickSpacing.Cmp(poolTickSpacing), "pool's own tickSpacing must match what the Factory advertises for this fee tier")
	poolFactory, err := s.v3Pool.Factory(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolFactory.Equal(s.factory))
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: pool confirms token0=%s token1=%s fee=%s tickSpacing=%s factory=%s", poolToken0, poolToken1, poolFee, poolTickSpacing, poolFactory)

	slot0, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(sqrtPriceX96.Cmp(slot0.SqrtPriceX96))
	s.Require().Zero(slot0.Tick.Sign())
	s.Require().True(slot0.Unlocked)
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: pool slot0 confirms sqrtPriceX96=%s tick=%s unlocked=%t", slot0.SqrtPriceX96, slot0.Tick, slot0.Unlocked)

	liquidity, err := s.v3Pool.Liquidity(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign(), "freshly initialized pool must have 0 active liquidity")
	s.T().Logf("Test02CreateAndInitializePoolIfNecessary: pool liquidity() = %s", liquidity)
}

func (s *RouterLifeCycleSuite) Test03Mint() {
	amount := new(big.Int).Div(s.mintValue, big.NewInt(2))
	tickLower := big.NewInt(-60000)
	tickUpper := big.NewInt(60000)
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test03Mint: token0=%s token1=%s tickLower=%s tickUpper=%s amount0Desired=%s amount1Desired=%s deadline=%s", s.tokenA, s.tokenB, tickLower, tickUpper, amount, amount, deadline)

	for _, token := range []*types.Address{s.tokenA, s.tokenB} {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, token, s.from, s.nfpm, amount)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
		s.mu.Unlock()
		s.T().Logf("Test03Mint: signed approve of %s (nonce %d, maxFee %s)", token, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("Test03Mint: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.nfpm, rpc.BlockTagLatest)
			return err == nil && amount.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("Test03Mint: approve of %s confirmed in block %d, allowance now %s", token, approveReceipt.BlockNumber, allowance)
	}

	balance0Before, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	balance1Before, err := s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test03Mint: balances before: token0=%s token1=%s", balance0Before, balance1Before)

	params := NewMintParams(s.tokenA, s.tokenB, s.Fee, tickLower, tickUpper, amount, amount, big.NewInt(0), big.NewInt(0), s.from, deadline)

	predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallMint(s.ctx, s.nfpm, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedLiquidity.Sign() > 0)
	s.T().Logf("Test03Mint: CallMint predicted tokenId=%s liquidity=%s amount0=%s amount1=%s", predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateMint(s.ctx, builder, s.nfpm, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test03Mint: signed mint tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.Mint(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test03Mint: broadcast mint tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "mint transaction reverted")
	s.T().Logf("Test03Mint: mint confirmed in block %d", receipt.BlockNumber)

	increased, err := ExtractIncreaseLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	s.tokenId = increased[0].TokenId
	s.Require().Zero(predictedLiquidity.Cmp(increased[0].Liquidity))
	s.T().Logf("Test03Mint: IncreaseLiquidity tokenId=%s liquidity=%s amount0=%s amount1=%s", increased[0].TokenId, increased[0].Liquidity, increased[0].Amount0, increased[0].Amount1)

	var position *Position
	s.Require().Eventually(func() bool {
		position, err = s.v3Nfpm.Positions(s.ctx, s.nfpm, s.tokenId, rpc.BlockTagLatest)
		return err == nil && predictedLiquidity.Cmp(position.Liquidity) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "position not visible after mint")
	s.Require().True(position.Token0.Equal(s.tokenA))
	s.Require().True(position.Token1.Equal(s.tokenB))
	s.Require().Zero(s.Fee.Cmp(position.Fee))
	s.Require().Zero(tickLower.Cmp(position.TickLower))
	s.Require().Zero(tickUpper.Cmp(position.TickUpper))
	s.T().Logf("Test03Mint: positions(%s) confirms token0=%s token1=%s fee=%s tickLower=%s tickUpper=%s liquidity=%s", s.tokenId, position.Token0, position.Token1, position.Fee, position.TickLower, position.TickUpper, position.Liquidity)

	owner, err := s.v3Nfpm.OwnerOf(s.ctx, s.nfpm, s.tokenId, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(owner.Equal(s.from))
	s.T().Logf("Test03Mint: ownerOf(%s) = %s", s.tokenId, owner)

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.v3Pool.Liquidity(s.ctx, s.pool, rpc.BlockTagLatest)
		return err == nil && predictedLiquidity.Cmp(liquidity) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool's active liquidity not visible after mint")
	s.T().Logf("Test03Mint: pool liquidity() = %s, matching the position's own liquidity since it's the pool's sole LP and the current tick sits inside its range", liquidity)

	var balance0After, balance1After *big.Int
	s.Require().Eventually(func() bool {
		balance0After, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balance1After, err = s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(balance0Before, balance0After).Cmp(predictedAmount0) == 0 && new(big.Int).Sub(balance1Before, balance1After).Cmp(predictedAmount1) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after mint")
	s.T().Logf("Test03Mint: balances after: token0=%s (%s -> %s) token1=%s (%s -> %s)", balance0After, balance0Before, balance0After, balance1After, balance1Before, balance1After)
}

func (s *RouterLifeCycleSuite) Test04ExactInputSingle() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(100))
	s.T().Logf("Test04ExactInputSingle: tokenIn=%s tokenOut=%s amountIn=%s", s.tokenA, s.tokenB, amountIn)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenA, s.from, s.router, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test04ExactInputSingle: signed approve of %s (nonce %d, maxFee %s)", s.tokenA, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	s.T().Logf("Test04ExactInputSingle: approve %s to spend %s of %s, broadcast tx %s", s.router, amountIn, s.tokenA, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.tokenA, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountIn.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test04ExactInputSingle: approve confirmed in block %d, allowance now %s", approveReceipt.BlockNumber, allowance)

	balance0Before, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	balance1Before, err := s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test04ExactInputSingle: balances before: token0=%s token1=%s", balance0Before, balance1Before)

	slot0Before, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	liquidityBefore, err := s.v3Pool.Liquidity(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test04ExactInputSingle: pool before: sqrtPriceX96=%s tick=%s liquidity=%s", slot0Before.SqrtPriceX96, slot0Before.Tick, liquidityBefore)

	params := NewExactInputSingleParams(s.tokenA, s.tokenB, s.Fee, s.from, amountIn, big.NewInt(0), big.NewInt(0))

	var predictedAmountOut *big.Int
	s.Require().Eventually(func() bool {
		predictedAmountOut, err = s.v3Router.CallExactInputSingle(s.ctx, s.router, s.from, nil, params, rpc.BlockTagLatest)
		ok := err == nil && predictedAmountOut.Sign() > 0
		if !ok {
			allowance, allowanceErr := s.erc20.Allowance(s.ctx, s.tokenA, s.from, s.router, rpc.BlockTagLatest)
			balance, balanceErr := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
			slot0, slot0Err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
			liquidity, liquidityErr := s.v3Pool.Liquidity(s.ctx, s.pool, rpc.BlockTagLatest)
			s.T().Logf("Test04ExactInputSingle: retry diagnostics: callErr=%v amountOut=%v allowance=%v (err=%v) balance=%v (err=%v) slot0=%+v (err=%v) liquidity=%v (err=%v)", err, predictedAmountOut, allowance, allowanceErr, balance, balanceErr, slot0, slot0Err, liquidity, liquidityErr)
		}
		return ok
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "callExactInputSingle not yet succeeding, last error: %v", err)
	s.T().Logf("Test04ExactInputSingle: CallExactInputSingle predicted amountOut=%s", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateExactInputSingle(s.ctx, swapBuilder, s.router, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test04ExactInputSingle: signed exactInputSingle tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.ExactInputSingle(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test04ExactInputSingle: broadcast exactInputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInputSingle transaction reverted")
	s.T().Logf("Test04ExactInputSingle: exactInputSingle confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	s.Require().Zero(amountIn.Cmp(swaps[0].Amount0), "pool's own Amount0 (token0 in) must equal our exact amountIn, since tokenA is token0")
	s.Require().Zero(predictedAmountOut.Cmp(new(big.Int).Neg(swaps[0].Amount1)), "pool's own -Amount1 (token1 out) must equal the predicted amountOut")
	s.T().Logf("Test04ExactInputSingle: Swap sender=%s recipient=%s amount0=%s amount1=%s sqrtPriceX96=%s tick=%s", swaps[0].Sender, swaps[0].Recipient, swaps[0].Amount0, swaps[0].Amount1, swaps[0].SqrtPriceX96, swaps[0].Tick)

	var liquidityAfter *big.Int
	var slot0After *Slot0
	s.Require().Eventually(func() bool {
		slot0After, err = s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		liquidityAfter, err = s.v3Pool.Liquidity(s.ctx, s.pool, rpc.BlockTagLatest)
		return err == nil && slot0Before.SqrtPriceX96.Cmp(slot0After.SqrtPriceX96) != 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool state not visible after swap")
	s.Require().True(slot0After.SqrtPriceX96.Cmp(slot0Before.SqrtPriceX96) < 0, "selling token0 for token1 must push the price (token1 per token0) down")
	s.Require().Zero(liquidityBefore.Cmp(liquidityAfter), "swap stayed within our wide position's range, so active liquidity shouldn't have changed")
	s.T().Logf("Test04ExactInputSingle: pool after: sqrtPriceX96=%s (%s -> %s) tick=%s liquidity=%s", slot0After.SqrtPriceX96, slot0Before.SqrtPriceX96, slot0After.SqrtPriceX96, slot0After.Tick, liquidityAfter)

	var balance0After, balance1After *big.Int
	s.Require().Eventually(func() bool {
		balance0After, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balance1After, err = s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(balance0Before, balance0After).Cmp(amountIn) == 0 && new(big.Int).Sub(balance1After, balance1Before).Cmp(predictedAmountOut) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after swap")
	s.T().Logf("Test04ExactInputSingle: balances after: token0=%s (%s -> %s) token1=%s (%s -> %s)", balance0After, balance0Before, balance0After, balance1After, balance1Before, balance1After)
}

func (s *RouterLifeCycleSuite) Test05ExactOutputSingle() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(200))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(100))
	s.T().Logf("Test05ExactOutputSingle: tokenIn=%s tokenOut=%s amountOut=%s approveCap=%s", s.tokenB, s.tokenA, amountOut, approveCap)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenB, s.from, s.router, approveCap)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test05ExactOutputSingle: signed approve of %s (nonce %d, maxFee %s)", s.tokenB, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	s.T().Logf("Test05ExactOutputSingle: approve %s to spend %s of %s, broadcast tx %s", s.router, approveCap, s.tokenB, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.tokenB, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && approveCap.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test05ExactOutputSingle: approve confirmed in block %d, allowance now %s", approveReceipt.BlockNumber, allowance)

	previewParams := NewExactOutputSingleParams(s.tokenB, s.tokenA, s.Fee, s.from, amountOut, approveCap, big.NewInt(0))
	predictedAmountIn, err := s.v3Router.CallExactOutputSingle(s.ctx, s.router, s.from, nil, previewParams, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmountIn.Sign() > 0)
	s.Require().True(predictedAmountIn.Cmp(approveCap) <= 0, "predicted amountIn must fit within what we already approved")
	s.T().Logf("Test05ExactOutputSingle: CallExactOutputSingle predicted amountIn=%s", predictedAmountIn)

	balance0Before, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	balance1Before, err := s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test05ExactOutputSingle: balances before: token0=%s token1=%s", balance0Before, balance1Before)

	slot0Before, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	liquidityBefore, err := s.v3Pool.Liquidity(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test05ExactOutputSingle: pool before: sqrtPriceX96=%s tick=%s liquidity=%s", slot0Before.SqrtPriceX96, slot0Before.Tick, liquidityBefore)

	params := NewExactOutputSingleParams(s.tokenB, s.tokenA, s.Fee, s.from, amountOut, predictedAmountIn, big.NewInt(0))

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateExactOutputSingle(s.ctx, swapBuilder, s.router, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test05ExactOutputSingle: signed exactOutputSingle tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.ExactOutputSingle(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test05ExactOutputSingle: broadcast exactOutputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactOutputSingle transaction reverted")
	s.T().Logf("Test05ExactOutputSingle: exactOutputSingle confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	s.Require().Zero(amountOut.Cmp(new(big.Int).Neg(swaps[0].Amount0)), "pool's own -Amount0 (token0 out) must equal our exact amountOut, since tokenA is token0")
	actualAmountIn := swaps[0].Amount1
	s.Require().True(actualAmountIn.Sign() > 0)
	s.Require().True(actualAmountIn.Cmp(predictedAmountIn) <= 0, "actual amountIn must not exceed the predicted amountInMaximum we approved")
	s.T().Logf("Test05ExactOutputSingle: Swap sender=%s recipient=%s amount0=%s amount1=%s sqrtPriceX96=%s tick=%s", swaps[0].Sender, swaps[0].Recipient, swaps[0].Amount0, swaps[0].Amount1, swaps[0].SqrtPriceX96, swaps[0].Tick)

	var liquidityAfter *big.Int
	var slot0After *Slot0
	s.Require().Eventually(func() bool {
		slot0After, err = s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		liquidityAfter, err = s.v3Pool.Liquidity(s.ctx, s.pool, rpc.BlockTagLatest)
		return err == nil && slot0Before.SqrtPriceX96.Cmp(slot0After.SqrtPriceX96) != 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool state not visible after swap")
	s.Require().True(slot0After.SqrtPriceX96.Cmp(slot0Before.SqrtPriceX96) > 0, "buying token0 with token1 must push the price (token1 per token0) up")
	s.Require().Zero(liquidityBefore.Cmp(liquidityAfter), "swap stayed within our wide position's range, so active liquidity shouldn't have changed")
	s.T().Logf("Test05ExactOutputSingle: pool after: sqrtPriceX96=%s (%s -> %s) tick=%s liquidity=%s", slot0After.SqrtPriceX96, slot0Before.SqrtPriceX96, slot0After.SqrtPriceX96, slot0After.Tick, liquidityAfter)

	var balance0After, balance1After *big.Int
	s.Require().Eventually(func() bool {
		balance0After, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balance1After, err = s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(balance0After, balance0Before).Cmp(amountOut) == 0 && new(big.Int).Sub(balance1Before, balance1After).Cmp(actualAmountIn) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after swap")
	s.T().Logf("Test05ExactOutputSingle: balances after: token0=%s (%s -> %s) token1=%s (%s -> %s)", balance0After, balance0Before, balance0After, balance1After, balance1Before, balance1After)
}

func (s *RouterLifeCycleSuite) Test06CreateAndInitializePoolBC() {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	s.T().Logf("Test06CreateAndInitializePoolBC: token0=%s token1=%s fee=%s sqrtPriceX96=%s", s.tokenB, s.tokenC, s.Fee, sqrtPriceX96)

	predicted, err := s.v3Nfpm.CallCreateAndInitializePoolIfNecessary(s.ctx, s.nfpm, s.from, s.tokenB, s.tokenC, s.Fee, sqrtPriceX96, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(predicted.IsZero())
	s.T().Logf("Test06CreateAndInitializePoolBC: CallCreateAndInitializePoolIfNecessary predicted pool=%s", predicted)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateCreateAndInitializePoolIfNecessary(s.ctx, builder, s.nfpm, s.from, s.tokenB, s.tokenC, s.Fee, sqrtPriceX96)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test06CreateAndInitializePoolBC: signed tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.CreateAndInitializePoolIfNecessary(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test06CreateAndInitializePoolBC: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createAndInitializePoolIfNecessary transaction reverted")
	s.T().Logf("Test06CreateAndInitializePoolBC: confirmed in block %d", receipt.BlockNumber)

	s.poolBC = predicted

	initialized, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(initialized, 1)
	s.Require().True(initialized[0].Contract.Equal(s.poolBC))
	s.Require().Zero(sqrtPriceX96.Cmp(initialized[0].SqrtPriceX96))
	s.T().Logf("Test06CreateAndInitializePoolBC: Initialize event confirms pool=%s sqrtPriceX96=%s tick=%s", initialized[0].Contract, initialized[0].SqrtPriceX96, initialized[0].Tick)

	var registeredPool *types.Address
	s.Require().Eventually(func() bool {
		registeredPool, err = s.v3Factory.GetPool(s.ctx, s.factory, s.tokenB, s.tokenC, s.Fee, rpc.BlockTagLatest)
		return err == nil && registeredPool.Equal(s.poolBC)
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "Factory.getPool not yet reflecting the new pool")
	s.T().Logf("Test06CreateAndInitializePoolBC: Factory.getPool(token0, token1, fee) = %s matches", registeredPool)

	poolToken0, err := s.v3Pool.Token0(s.ctx, s.poolBC, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolToken0.Equal(s.tokenB))
	poolToken1, err := s.v3Pool.Token1(s.ctx, s.poolBC, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(poolToken1.Equal(s.tokenC))
	poolFee, err := s.v3Pool.Fee(s.ctx, s.poolBC, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(s.Fee.Cmp(poolFee))
	s.T().Logf("Test06CreateAndInitializePoolBC: pool confirms token0=%s token1=%s fee=%s", poolToken0, poolToken1, poolFee)

	slot0, err := s.v3Pool.Slot0(s.ctx, s.poolBC, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(sqrtPriceX96.Cmp(slot0.SqrtPriceX96))
	s.Require().Zero(slot0.Tick.Sign())
	s.Require().True(slot0.Unlocked)
	s.T().Logf("Test06CreateAndInitializePoolBC: pool slot0 confirms sqrtPriceX96=%s tick=%s unlocked=%t", slot0.SqrtPriceX96, slot0.Tick, slot0.Unlocked)

	liquidity, err := s.v3Pool.Liquidity(s.ctx, s.poolBC, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign(), "freshly initialized pool must have 0 active liquidity")
	s.T().Logf("Test06CreateAndInitializePoolBC: pool liquidity() = %s", liquidity)
}

func (s *RouterLifeCycleSuite) Test07MintBC() {
	amount := new(big.Int).Div(s.mintValue, big.NewInt(2))
	tickLower := big.NewInt(-60000)
	tickUpper := big.NewInt(60000)
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test07MintBC: token0=%s token1=%s tickLower=%s tickUpper=%s amount0Desired=%s amount1Desired=%s deadline=%s", s.tokenB, s.tokenC, tickLower, tickUpper, amount, amount, deadline)

	for _, token := range []*types.Address{s.tokenB, s.tokenC} {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, token, s.from, s.nfpm, amount)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
		s.mu.Unlock()
		s.T().Logf("Test07MintBC: signed approve of %s (nonce %d, maxFee %s)", token, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("Test07MintBC: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		var allowance *big.Int
		s.Require().Eventually(func() bool {
			allowance, err = s.erc20.Allowance(s.ctx, token, s.from, s.nfpm, rpc.BlockTagLatest)
			return err == nil && amount.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("Test07MintBC: approve of %s confirmed in block %d, allowance now %s", token, approveReceipt.BlockNumber, allowance)
	}

	balance1Before, err := s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	balance2Before, err := s.erc20.BalanceOf(s.ctx, s.tokenC, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test07MintBC: balances before: token1=%s token2=%s", balance1Before, balance2Before)

	params := NewMintParams(s.tokenB, s.tokenC, s.Fee, tickLower, tickUpper, amount, amount, big.NewInt(0), big.NewInt(0), s.from, deadline)

	predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallMint(s.ctx, s.nfpm, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedLiquidity.Sign() > 0)
	s.T().Logf("Test07MintBC: CallMint predicted tokenId=%s liquidity=%s amount0=%s amount1=%s", predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateMint(s.ctx, builder, s.nfpm, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test07MintBC: signed mint tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.Mint(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test07MintBC: broadcast mint tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "mint transaction reverted")
	s.T().Logf("Test07MintBC: mint confirmed in block %d", receipt.BlockNumber)

	increased, err := ExtractIncreaseLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	tokenIdBC := increased[0].TokenId
	s.Require().Zero(predictedLiquidity.Cmp(increased[0].Liquidity))
	s.T().Logf("Test07MintBC: IncreaseLiquidity tokenId=%s liquidity=%s amount0=%s amount1=%s", increased[0].TokenId, increased[0].Liquidity, increased[0].Amount0, increased[0].Amount1)

	var position *Position
	s.Require().Eventually(func() bool {
		position, err = s.v3Nfpm.Positions(s.ctx, s.nfpm, tokenIdBC, rpc.BlockTagLatest)
		return err == nil && predictedLiquidity.Cmp(position.Liquidity) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "position not visible after mint")
	s.Require().True(position.Token0.Equal(s.tokenB))
	s.Require().True(position.Token1.Equal(s.tokenC))
	s.Require().Zero(s.Fee.Cmp(position.Fee))
	s.Require().Zero(tickLower.Cmp(position.TickLower))
	s.Require().Zero(tickUpper.Cmp(position.TickUpper))
	s.T().Logf("Test07MintBC: positions(%s) confirms token0=%s token1=%s fee=%s tickLower=%s tickUpper=%s liquidity=%s", tokenIdBC, position.Token0, position.Token1, position.Fee, position.TickLower, position.TickUpper, position.Liquidity)

	owner, err := s.v3Nfpm.OwnerOf(s.ctx, s.nfpm, tokenIdBC, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(owner.Equal(s.from))
	s.T().Logf("Test07MintBC: ownerOf(%s) = %s", tokenIdBC, owner)

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.v3Pool.Liquidity(s.ctx, s.poolBC, rpc.BlockTagLatest)
		return err == nil && predictedLiquidity.Cmp(liquidity) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool's active liquidity not visible after mint")
	s.T().Logf("Test07MintBC: pool liquidity() = %s, matching the position's own liquidity since it's the pool's sole LP and the current tick sits inside its range", liquidity)

	var balance1After, balance2After *big.Int
	s.Require().Eventually(func() bool {
		balance1After, err = s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balance2After, err = s.erc20.BalanceOf(s.ctx, s.tokenC, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(balance1Before, balance1After).Cmp(predictedAmount0) == 0 && new(big.Int).Sub(balance2Before, balance2After).Cmp(predictedAmount1) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after mint")
	s.T().Logf("Test07MintBC: balances after: token1=%s (%s -> %s) token2=%s (%s -> %s)", balance1After, balance1Before, balance1After, balance2After, balance2Before, balance2After)
}

func (s *RouterLifeCycleSuite) Test08ExactInput() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))

	feeBytes := make([]byte, 3)
	feeRaw := s.Fee.Bytes()
	copy(feeBytes[3-len(feeRaw):], feeRaw)

	path := append([]byte{}, s.tokenA.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenB.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenC.Bytes()...)
	s.T().Logf("Test08ExactInput: path=tokenA(%s) -> tokenB(%s) -> tokenC(%s) fee=%s amountIn=%s", s.tokenA, s.tokenB, s.tokenC, s.Fee, amountIn)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenA, s.from, s.router, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test08ExactInput: signed approve of %s (nonce %d, maxFee %s)", s.tokenA, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	s.T().Logf("Test08ExactInput: approve %s to spend %s of %s, broadcast tx %s", s.router, amountIn, s.tokenA, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.tokenA, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountIn.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test08ExactInput: approve confirmed in block %d, allowance now %s", approveReceipt.BlockNumber, allowance)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	balanceBBefore, err := s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	balanceCBefore, err := s.erc20.BalanceOf(s.ctx, s.tokenC, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test08ExactInput: balances before: tokenA=%s tokenB=%s tokenC=%s", balanceABefore, balanceBBefore, balanceCBefore)

	slot0ABBefore, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	slot0BCBefore, err := s.v3Pool.Slot0(s.ctx, s.poolBC, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test08ExactInput: pools before: AB sqrtPriceX96=%s BC sqrtPriceX96=%s", slot0ABBefore.SqrtPriceX96, slot0BCBefore.SqrtPriceX96)

	params := NewExactInputParams(path, s.from, amountIn, big.NewInt(0))

	predictedAmountOut, err := s.v3Router.CallExactInput(s.ctx, s.router, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.T().Logf("Test08ExactInput: CallExactInput predicted amountOut=%s", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateExactInput(s.ctx, swapBuilder, s.router, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test08ExactInput: signed exactInput tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.ExactInput(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test08ExactInput: broadcast exactInput tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInput transaction reverted")
	s.T().Logf("Test08ExactInput: exactInput confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 2, "multihop exactInput over 2 pools must emit exactly 2 Swap events")

	var hopAB, hopBC *Swap
	for i := range swaps {
		if swaps[i].Contract.Equal(s.pool) {
			hopAB = &swaps[i]
		} else if swaps[i].Contract.Equal(s.poolBC) {
			hopBC = &swaps[i]
		}
	}
	s.Require().NotNil(hopAB, "one Swap event must come from the A-B pool")
	s.Require().NotNil(hopBC, "one Swap event must come from the B-C pool")

	s.Require().Zero(amountIn.Cmp(hopAB.Amount0), "A-B pool's own Amount0 (token0 in) must equal our exact amountIn, since tokenA is token0")
	intermediateAmount := new(big.Int).Neg(hopAB.Amount1)
	s.Require().True(intermediateAmount.Sign() > 0)
	s.Require().Zero(intermediateAmount.Cmp(hopBC.Amount0), "the tokenB amount the A-B pool paid out must equal exactly what the B-C pool received as input")
	s.Require().Zero(predictedAmountOut.Cmp(new(big.Int).Neg(hopBC.Amount1)), "B-C pool's own -Amount1 (token1 out) must equal the predicted final amountOut")
	s.T().Logf("Test08ExactInput: hop A-B amount0=%s amount1=%s; hop B-C amount0=%s amount1=%s; intermediate tokenB=%s", hopAB.Amount0, hopAB.Amount1, hopBC.Amount0, hopBC.Amount1, intermediateAmount)

	var slot0ABAfter, slot0BCAfter *Slot0
	s.Require().Eventually(func() bool {
		slot0ABAfter, err = s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		slot0BCAfter, err = s.v3Pool.Slot0(s.ctx, s.poolBC, rpc.BlockTagLatest)
		return err == nil && slot0ABBefore.SqrtPriceX96.Cmp(slot0ABAfter.SqrtPriceX96) != 0 && slot0BCBefore.SqrtPriceX96.Cmp(slot0BCAfter.SqrtPriceX96) != 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool state not visible after swap")
	s.Require().True(slot0ABAfter.SqrtPriceX96.Cmp(slot0ABBefore.SqrtPriceX96) < 0, "selling token0 for token1 in the A-B pool must push its price down")
	s.Require().True(slot0BCAfter.SqrtPriceX96.Cmp(slot0BCBefore.SqrtPriceX96) < 0, "selling token0 for token1 in the B-C pool must push its price down")
	s.T().Logf("Test08ExactInput: pools after: AB sqrtPriceX96=%s (%s -> %s) BC sqrtPriceX96=%s (%s -> %s)", slot0ABAfter.SqrtPriceX96, slot0ABBefore.SqrtPriceX96, slot0ABAfter.SqrtPriceX96, slot0BCAfter.SqrtPriceX96, slot0BCBefore.SqrtPriceX96, slot0BCAfter.SqrtPriceX96)

	var balanceAAfter, balanceBAfter, balanceCAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balanceBAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balanceCAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenC, s.from, rpc.BlockTagLatest)
		return err == nil &&
			new(big.Int).Sub(balanceABefore, balanceAAfter).Cmp(amountIn) == 0 &&
			balanceBBefore.Cmp(balanceBAfter) == 0 &&
			new(big.Int).Sub(balanceCAfter, balanceCBefore).Cmp(predictedAmountOut) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after swap")
	s.Require().Zero(balanceBBefore.Cmp(balanceBAfter), "tokenB is only an intermediate hop, so our own tokenB balance must be unchanged")
	s.T().Logf("Test08ExactInput: balances after: tokenA=%s (%s -> %s) tokenB=%s (unchanged) tokenC=%s (%s -> %s)", balanceAAfter, balanceABefore, balanceAAfter, balanceBAfter, balanceCAfter, balanceCBefore, balanceCAfter)
}

func (s *RouterLifeCycleSuite) Test09ExactInputSingleTooLittleReceived() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	unrealisticAmountOutMinimum := new(big.Int).Mul(amountIn, big.NewInt(2))
	s.T().Logf("Test09ExactInputSingleTooLittleReceived: amountIn=%s amountOutMinimum=%s", amountIn, unrealisticAmountOutMinimum)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenA, s.from, s.router, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test09ExactInputSingleTooLittleReceived: signed approve of %s (nonce %d, maxFee %s)", s.tokenA, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenA, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountIn.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test09ExactInputSingleTooLittleReceived: approve confirmed in block %d", approveReceipt.BlockNumber)

	params := NewExactInputSingleParams(s.tokenA, s.tokenB, s.Fee, s.from, amountIn, unrealisticAmountOutMinimum, big.NewInt(0))

	_, err = s.v3Router.CallExactInputSingle(s.ctx, s.router, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTooLittleReceived)
	s.T().Logf("Test09ExactInputSingleTooLittleReceived: callExactInputSingle correctly rejected with %v", err)
}

func (s *RouterLifeCycleSuite) Test10ExactOutputSingleTooMuchRequested() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	unrealisticAmountInMaximum := big.NewInt(1)
	s.T().Logf("Test10ExactOutputSingleTooMuchRequested: amountOut=%s amountInMaximum=%s", amountOut, unrealisticAmountInMaximum)

	approveCap := new(big.Int).Mul(amountOut, big.NewInt(2))
	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenB, s.from, s.router, approveCap)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test10ExactOutputSingleTooMuchRequested: signed approve of %s (nonce %d, maxFee %s)", s.tokenB, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenB, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && approveCap.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test10ExactOutputSingleTooMuchRequested: approve confirmed in block %d", approveReceipt.BlockNumber)

	params := NewExactOutputSingleParams(s.tokenB, s.tokenA, s.Fee, s.from, amountOut, unrealisticAmountInMaximum, big.NewInt(0))

	_, err = s.v3Router.CallExactOutputSingle(s.ctx, s.router, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTooMuchRequested)
	s.T().Logf("Test10ExactOutputSingleTooMuchRequested: callExactOutputSingle correctly rejected with %v", err)
}

func (s *RouterLifeCycleSuite) Test11ExactOutput() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(2000))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(1000))

	feeBytes := make([]byte, 3)
	feeRaw := s.Fee.Bytes()
	copy(feeBytes[3-len(feeRaw):], feeRaw)

	path := append([]byte{}, s.tokenA.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenB.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenC.Bytes()...)
	s.T().Logf("Test11ExactOutput: path=tokenC(%s) -> tokenB(%s) -> tokenA(%s) fee=%s amountOut=%s approveCap=%s", s.tokenC, s.tokenB, s.tokenA, s.Fee, amountOut, approveCap)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenC, s.from, s.router, approveCap)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test11ExactOutput: signed approve of %s (nonce %d, maxFee %s)", s.tokenC, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	s.T().Logf("Test11ExactOutput: approve %s to spend %s of %s, broadcast tx %s", s.router, approveCap, s.tokenC, approveHash)

	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.tokenC, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && approveCap.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test11ExactOutput: approve confirmed in block %d, allowance now %s", approveReceipt.BlockNumber, allowance)

	previewParams := NewExactOutputParams(path, s.from, amountOut, approveCap)
	predictedAmountIn, err := s.v3Router.CallExactOutput(s.ctx, s.router, s.from, nil, previewParams, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmountIn.Sign() > 0)
	s.Require().True(predictedAmountIn.Cmp(approveCap) <= 0, "predicted amountIn must fit within what we already approved")
	s.T().Logf("Test11ExactOutput: CallExactOutput predicted amountIn=%s", predictedAmountIn)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	balanceBBefore, err := s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	balanceCBefore, err := s.erc20.BalanceOf(s.ctx, s.tokenC, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.T().Logf("Test11ExactOutput: balances before: tokenA=%s tokenB=%s tokenC=%s", balanceABefore, balanceBBefore, balanceCBefore)

	params := NewExactOutputParams(path, s.from, amountOut, predictedAmountIn)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateExactOutput(s.ctx, swapBuilder, s.router, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test11ExactOutput: signed exactOutput tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.ExactOutput(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test11ExactOutput: broadcast exactOutput tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactOutput transaction reverted")
	s.T().Logf("Test11ExactOutput: exactOutput confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 2, "multihop exactOutput over 2 pools must emit exactly 2 Swap events")

	var hopAB, hopBC *Swap
	for i := range swaps {
		if swaps[i].Contract.Equal(s.pool) {
			hopAB = &swaps[i]
		} else if swaps[i].Contract.Equal(s.poolBC) {
			hopBC = &swaps[i]
		}
	}
	s.Require().NotNil(hopAB, "one Swap event must come from the A-B pool")
	s.Require().NotNil(hopBC, "one Swap event must come from the B-C pool")

	s.Require().Zero(amountOut.Cmp(new(big.Int).Neg(hopAB.Amount0)), "A-B pool's own -Amount0 (token0 out) must equal our exact amountOut, since tokenA is token0")
	intermediateAmount := hopAB.Amount1
	s.Require().True(intermediateAmount.Sign() > 0)
	s.Require().Zero(intermediateAmount.Cmp(new(big.Int).Neg(hopBC.Amount0)), "the tokenB amount the A-B pool pulled in must equal exactly what the B-C pool paid out")
	actualAmountIn := hopBC.Amount1
	s.Require().True(actualAmountIn.Sign() > 0)
	s.Require().True(actualAmountIn.Cmp(predictedAmountIn) <= 0, "actual amountIn must not exceed the predicted amountInMaximum we approved")
	s.T().Logf("Test11ExactOutput: hop A-B amount0=%s amount1=%s; hop B-C amount0=%s amount1=%s; intermediate tokenB=%s", hopAB.Amount0, hopAB.Amount1, hopBC.Amount0, hopBC.Amount1, intermediateAmount)

	var balanceAAfter, balanceBAfter, balanceCAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balanceBAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balanceCAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenC, s.from, rpc.BlockTagLatest)
		return err == nil &&
			new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(amountOut) == 0 &&
			balanceBBefore.Cmp(balanceBAfter) == 0 &&
			new(big.Int).Sub(balanceCBefore, balanceCAfter).Cmp(actualAmountIn) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after swap")
	s.Require().Zero(balanceBBefore.Cmp(balanceBAfter), "tokenB is only an intermediate hop, so our own tokenB balance must be unchanged")
	s.T().Logf("Test11ExactOutput: balances after: tokenA=%s (%s -> %s) tokenB=%s (unchanged) tokenC=%s (%s -> %s)", balanceAAfter, balanceABefore, balanceAAfter, balanceBAfter, balanceCAfter, balanceCBefore, balanceCAfter)
}

func (s *RouterLifeCycleSuite) Test12ExactInputTooLittleReceived() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	unrealisticAmountOutMinimum := new(big.Int).Mul(amountIn, big.NewInt(2))

	feeBytes := make([]byte, 3)
	feeRaw := s.Fee.Bytes()
	copy(feeBytes[3-len(feeRaw):], feeRaw)

	path := append([]byte{}, s.tokenA.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenB.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenC.Bytes()...)
	s.T().Logf("Test12ExactInputTooLittleReceived: path=tokenA(%s) -> tokenB(%s) -> tokenC(%s) amountIn=%s amountOutMinimum=%s", s.tokenA, s.tokenB, s.tokenC, amountIn, unrealisticAmountOutMinimum)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenA, s.from, s.router, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test12ExactInputTooLittleReceived: signed approve of %s (nonce %d, maxFee %s)", s.tokenA, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenA, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountIn.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test12ExactInputTooLittleReceived: approve confirmed in block %d", approveReceipt.BlockNumber)

	params := NewExactInputParams(path, s.from, amountIn, unrealisticAmountOutMinimum)

	_, err = s.v3Router.CallExactInput(s.ctx, s.router, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTooLittleReceived)
	s.T().Logf("Test12ExactInputTooLittleReceived: callExactInput correctly rejected with %v", err)
}

func (s *RouterLifeCycleSuite) Test13ExactOutputTooMuchRequested() {
	amountOut := new(big.Int).Div(s.mintValue, big.NewInt(2000))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	unrealisticAmountInMaximum := big.NewInt(1)

	feeBytes := make([]byte, 3)
	feeRaw := s.Fee.Bytes()
	copy(feeBytes[3-len(feeRaw):], feeRaw)

	path := append([]byte{}, s.tokenA.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenB.Bytes()...)
	path = append(path, feeBytes...)
	path = append(path, s.tokenC.Bytes()...)
	s.T().Logf("Test13ExactOutputTooMuchRequested: path=tokenC(%s) -> tokenB(%s) -> tokenA(%s) amountOut=%s amountInMaximum=%s", s.tokenC, s.tokenB, s.tokenA, amountOut, unrealisticAmountInMaximum)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenC, s.from, s.router, approveCap)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test13ExactOutputTooMuchRequested: signed approve of %s (nonce %d, maxFee %s)", s.tokenC, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenC, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && approveCap.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test13ExactOutputTooMuchRequested: approve confirmed in block %d", approveReceipt.BlockNumber)

	params := NewExactOutputParams(path, s.from, amountOut, unrealisticAmountInMaximum)

	_, err = s.v3Router.CallExactOutput(s.ctx, s.router, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrTooMuchRequested)
	s.T().Logf("Test13ExactOutputTooMuchRequested: callExactOutput correctly rejected with %v", err)
}

func (s *RouterLifeCycleSuite) Test14ExactInputSingleContractBalanceSentinel() {
	amountToSend := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	contractBalanceSentinel := big.NewInt(0)
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: sending %s of %s directly to router %s, then swapping with amountIn=CONTRACT_BALANCE(%s)", amountToSend, s.tokenA, s.router, contractBalanceSentinel)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedTransfer, err := s.erc20.CreateTransfer(s.ctx, builder, s.tokenA, s.from, s.router, amountToSend)
	s.Require().NoError(err)
	s.Require().NoError(unsignedTransfer.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedTransfer.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: signed transfer to router (nonce %d, maxFee %s)", unsignedTransfer.AddressNonce(), unsignedTransfer.MaxFee())

	transferHash, err := s.erc20.Transfer(s.ctx, unsignedTransfer)
	s.Require().NoError(err)
	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	var routerBalanceBefore *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceBefore, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.router, rpc.BlockTagLatest)
		return err == nil && amountToSend.Cmp(routerBalanceBefore) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "router's token balance not visible after transfer")
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: transfer confirmed in block %d, router's tokenA balance now %s", transferReceipt.BlockNumber, routerBalanceBefore)

	balanceBBefore, err := s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	params := NewExactInputSingleParams(s.tokenA, s.tokenB, s.Fee, s.from, contractBalanceSentinel, big.NewInt(0), big.NewInt(0))

	predictedAmountOut, err := s.v3Router.CallExactInputSingle(s.ctx, s.router, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: CallExactInputSingle predicted amountOut=%s (sourced from router's own balance, no approval needed from us)", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateExactInputSingle(s.ctx, swapBuilder, s.router, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: signed exactInputSingle tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.ExactInputSingle(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: broadcast exactInputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInputSingle transaction reverted")
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: exactInputSingle confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	s.Require().True(swaps[0].Sender.Equal(s.router), "the router itself initiated the pool swap, since it paid from its own balance")
	s.Require().True(swaps[0].Recipient.Equal(s.from))
	s.Require().Zero(amountToSend.Cmp(swaps[0].Amount0), "the pool must have pulled exactly the router's entire tokenA balance, not just some of it")
	s.Require().Zero(predictedAmountOut.Cmp(new(big.Int).Neg(swaps[0].Amount1)))
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: Swap sender=%s recipient=%s amount0=%s amount1=%s", swaps[0].Sender, swaps[0].Recipient, swaps[0].Amount0, swaps[0].Amount1)

	var routerBalanceAfter, balanceBAfter *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.router, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balanceBAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
		return err == nil && routerBalanceAfter.Sign() == 0 && new(big.Int).Sub(balanceBAfter, balanceBBefore).Cmp(predictedAmountOut) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after swap")
	s.T().Logf("Test14ExactInputSingleContractBalanceSentinel: router's tokenA balance now %s (fully drained), our tokenB balance %s (%s -> %s)", routerBalanceAfter, balanceBAfter, balanceBBefore, balanceBAfter)
}

func (s *RouterLifeCycleSuite) Test15ExactOutputSinglePriceLimitPartialFill() {
	requestedAmountOut := new(big.Int).Div(s.mintValue, big.NewInt(500))
	approveCap := new(big.Int).Div(s.mintValue, big.NewInt(100))

	slot0Before, err := s.v3Pool.Slot0(s.ctx, s.pool, rpc.BlockTagLatest)
	s.Require().NoError(err)
	sqrtPriceLimitX96 := new(big.Int).Add(slot0Before.SqrtPriceX96, new(big.Int).Div(slot0Before.SqrtPriceX96, big.NewInt(10000)))
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: requestedAmountOut=%s sqrtPriceX96Before=%s sqrtPriceLimitX96=%s (a tight +0.01%% cap)", requestedAmountOut, slot0Before.SqrtPriceX96, sqrtPriceLimitX96)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenB, s.from, s.router, approveCap)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: signed approve of %s (nonce %d, maxFee %s)", s.tokenB, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenB, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && approveCap.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: approve confirmed in block %d", approveReceipt.BlockNumber)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	params := NewExactOutputSingleParams(s.tokenB, s.tokenA, s.Fee, s.from, requestedAmountOut, approveCap, sqrtPriceLimitX96)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateExactOutputSingle(s.ctx, swapBuilder, s.router, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: signed exactOutputSingle tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.ExactOutputSingle(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: broadcast exactOutputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactOutputSingle transaction must not revert even though the price limit prevents it from fully filling amountOut")
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: exactOutputSingle confirmed in block %d despite the price limit", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	actualAmountOut := new(big.Int).Neg(swaps[0].Amount0)
	s.Require().True(actualAmountOut.Sign() > 0)
	s.Require().True(actualAmountOut.Cmp(requestedAmountOut) < 0, "the tight price limit must have capped the actual output below what we asked for")
	s.Require().Zero(sqrtPriceLimitX96.Cmp(swaps[0].SqrtPriceX96), "the pool must land exactly on our price limit when the limit (not the amount) is what stopped the swap")
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: actualAmountOut=%s (< requested %s), pool landed exactly at sqrtPriceX96=%s", actualAmountOut, requestedAmountOut, swaps[0].SqrtPriceX96)

	var balanceAAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(actualAmountOut) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after swap")
	s.T().Logf("Test15ExactOutputSinglePriceLimitPartialFill: tokenA balance %s (%s -> %s), exactly the partial amount actually delivered", balanceAAfter, balanceABefore, balanceAAfter)
}

func (s *RouterLifeCycleSuite) Test16ExactInputSingleMsgSenderRecipientSentinel() {
	amountIn := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	msgSenderSentinel := types.MustNewAddressFromHex("0x0000000000000000000000000000000000000001")
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: amountIn=%s recipient=%s (MSG_SENDER sentinel, should resolve to %s)", amountIn, msgSenderSentinel, s.from)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, s.tokenA, s.from, s.router, amountIn)
	s.Require().NoError(err)
	s.Require().NoError(unsignedApprove.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: signed approve of %s (nonce %d, maxFee %s)", s.tokenA, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

	approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
	s.Require().NoError(err)
	approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(approveReceipt.Status, "approve transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenA, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && amountIn.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: approve confirmed in block %d", approveReceipt.BlockNumber)

	balanceBBefore, err := s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	sentinelBalanceBefore, err := s.erc20.BalanceOf(s.ctx, s.tokenB, msgSenderSentinel, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(sentinelBalanceBefore.Sign(), "nothing should ever have been sent to the literal sentinel address")

	params := NewExactInputSingleParams(s.tokenA, s.tokenB, s.Fee, msgSenderSentinel, amountIn, big.NewInt(0), big.NewInt(0))

	predictedAmountOut, err := s.v3Router.CallExactInputSingle(s.ctx, s.router, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: CallExactInputSingle predicted amountOut=%s", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateExactInputSingle(s.ctx, swapBuilder, s.router, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: signed exactInputSingle tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.ExactInputSingle(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: broadcast exactInputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInputSingle transaction reverted")
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: exactInputSingle confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.pool))
	s.Require().True(swaps[0].Recipient.Equal(s.from), "the MSG_SENDER sentinel must resolve to the actual caller, not the literal address(1)")
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: Swap recipient=%s (resolved from the sentinel)", swaps[0].Recipient)

	var balanceBAfter, sentinelBalanceAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceBAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenB, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		sentinelBalanceAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenB, msgSenderSentinel, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(balanceBAfter, balanceBBefore).Cmp(predictedAmountOut) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after swap")
	s.Require().Zero(sentinelBalanceAfter.Sign(), "the literal sentinel address must still hold nothing")
	s.T().Logf("Test16ExactInputSingleMsgSenderRecipientSentinel: our tokenB balance %s (%s -> %s), sentinel address still holds %s", balanceBAfter, balanceBBefore, balanceBAfter, sentinelBalanceAfter)
}

func (s *RouterLifeCycleSuite) Test17CreateAndInitializePoolWETH() {
	s.wethToken0, s.wethToken1 = s.weth9, s.tokenA
	if s.wethToken0.Big().Cmp(s.wethToken1.Big()) > 0 {
		s.wethToken0, s.wethToken1 = s.wethToken1, s.wethToken0
	}
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	s.T().Logf("Test17CreateAndInitializePoolWETH: token0=%s token1=%s fee=%s sqrtPriceX96=%s", s.wethToken0, s.wethToken1, s.Fee, sqrtPriceX96)

	predicted, err := s.v3Nfpm.CallCreateAndInitializePoolIfNecessary(s.ctx, s.nfpm, s.from, s.wethToken0, s.wethToken1, s.Fee, sqrtPriceX96, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(predicted.IsZero())
	s.T().Logf("Test17CreateAndInitializePoolWETH: CallCreateAndInitializePoolIfNecessary predicted pool=%s", predicted)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateCreateAndInitializePoolIfNecessary(s.ctx, builder, s.nfpm, s.from, s.wethToken0, s.wethToken1, s.Fee, sqrtPriceX96)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test17CreateAndInitializePoolWETH: signed tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.CreateAndInitializePoolIfNecessary(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test17CreateAndInitializePoolWETH: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "createAndInitializePoolIfNecessary transaction reverted")
	s.T().Logf("Test17CreateAndInitializePoolWETH: confirmed in block %d", receipt.BlockNumber)

	s.poolWETH = predicted

	initialized, err := ExtractInitialize(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(initialized, 1)
	s.Require().True(initialized[0].Contract.Equal(s.poolWETH))
	s.Require().Zero(sqrtPriceX96.Cmp(initialized[0].SqrtPriceX96))
	s.T().Logf("Test17CreateAndInitializePoolWETH: Initialize event confirms pool=%s sqrtPriceX96=%s tick=%s", initialized[0].Contract, initialized[0].SqrtPriceX96, initialized[0].Tick)

	var registeredPool *types.Address
	s.Require().Eventually(func() bool {
		registeredPool, err = s.v3Factory.GetPool(s.ctx, s.factory, s.wethToken0, s.wethToken1, s.Fee, rpc.BlockTagLatest)
		return err == nil && registeredPool.Equal(s.poolWETH)
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "Factory.getPool not yet reflecting the new pool")
	s.T().Logf("Test17CreateAndInitializePoolWETH: Factory.getPool(token0, token1, fee) = %s matches", registeredPool)

	liquidity, err := s.v3Pool.Liquidity(s.ctx, s.poolWETH, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(liquidity.Sign(), "freshly initialized pool must have 0 active liquidity")
	s.T().Logf("Test17CreateAndInitializePoolWETH: pool liquidity() = %s", liquidity)
}

func (s *RouterLifeCycleSuite) Test18MintWETH() {
	amount := s.wethLiquidity
	tickLower := big.NewInt(-60000)
	tickUpper := big.NewInt(60000)
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test18MintWETH: token0=%s token1=%s tickLower=%s tickUpper=%s amount0Desired=%s amount1Desired=%s deadline=%s", s.wethToken0, s.wethToken1, tickLower, tickUpper, amount, amount, deadline)

	for _, token := range []*types.Address{s.wethToken0, s.wethToken1} {
		builder := tx.NewDynamicFeeBuilder()
		builder.PriceBuffer(150)
		builder.LimitBuffer(300)
		unsignedApprove, err := s.erc20.CreateApprove(s.ctx, builder, token, s.from, s.nfpm, amount)
		s.Require().NoError(err)
		s.Require().NoError(unsignedApprove.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, unsignedApprove.MaxFee())
		s.mu.Unlock()
		s.T().Logf("Test18MintWETH: signed approve of %s (nonce %d, maxFee %s)", token, unsignedApprove.AddressNonce(), unsignedApprove.MaxFee())

		approveHash, err := s.erc20.Approve(s.ctx, unsignedApprove)
		s.Require().NoError(err)
		s.T().Logf("Test18MintWETH: approve %s to spend %s of %s, broadcast tx %s", s.nfpm, amount, token, approveHash)

		approveReceipt, err := tx.WaitForReceipt(s.ctx, s.client, approveHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
		s.Require().NoError(err)
		s.Require().True(approveReceipt.Status, "approve transaction reverted")
		s.Require().Eventually(func() bool {
			allowance, err := s.erc20.Allowance(s.ctx, token, s.from, s.nfpm, rpc.BlockTagLatest)
			return err == nil && amount.Cmp(allowance) == 0
		}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after approve")
		s.T().Logf("Test18MintWETH: approve of %s confirmed in block %d", token, approveReceipt.BlockNumber)
	}

	params := NewMintParams(s.wethToken0, s.wethToken1, s.Fee, tickLower, tickUpper, amount, amount, big.NewInt(0), big.NewInt(0), s.from, deadline)

	predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1, err := s.v3Nfpm.CallMint(s.ctx, s.nfpm, s.from, nil, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedLiquidity.Sign() > 0)
	s.T().Logf("Test18MintWETH: CallMint predicted tokenId=%s liquidity=%s amount0=%s amount1=%s", predictedTokenId, predictedLiquidity, predictedAmount0, predictedAmount1)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Nfpm.CreateMint(s.ctx, builder, s.nfpm, s.from, nil, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test18MintWETH: signed mint tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Nfpm.Mint(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test18MintWETH: broadcast mint tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "mint transaction reverted")
	s.T().Logf("Test18MintWETH: mint confirmed in block %d", receipt.BlockNumber)

	increased, err := ExtractIncreaseLiquidity(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(increased, 1)
	s.Require().Zero(predictedLiquidity.Cmp(increased[0].Liquidity))
	s.T().Logf("Test18MintWETH: IncreaseLiquidity tokenId=%s liquidity=%s amount0=%s amount1=%s", increased[0].TokenId, increased[0].Liquidity, increased[0].Amount0, increased[0].Amount1)

	var liquidity *big.Int
	s.Require().Eventually(func() bool {
		liquidity, err = s.v3Pool.Liquidity(s.ctx, s.poolWETH, rpc.BlockTagLatest)
		return err == nil && predictedLiquidity.Cmp(liquidity) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "pool's active liquidity not visible after mint")
	s.T().Logf("Test18MintWETH: pool liquidity() = %s", liquidity)
}

func (s *RouterLifeCycleSuite) Test19ExactInputSingleNativeETH() {
	value := new(big.Int).Div(s.wethLiquidity, big.NewInt(10))
	s.T().Logf("Test19ExactInputSingleNativeETH: swapping %s native %s directly (tokenIn=WETH9=%s, tokenOut=tokenA=%s)", value, s.NativeSymbol, s.weth9, s.tokenA)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	params := NewExactInputSingleParams(s.weth9, s.tokenA, s.Fee, s.from, value, big.NewInt(0), big.NewInt(0))

	predictedAmountOut, err := s.v3Router.CallExactInputSingle(s.ctx, s.router, s.from, value, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.T().Logf("Test19ExactInputSingleNativeETH: CallExactInputSingle predicted amountOut=%s", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateExactInputSingle(s.ctx, swapBuilder, s.router, s.from, value, params)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test19ExactInputSingleNativeETH: signed exactInputSingle tx (nonce %d, gasLimit %d, maxFee %s, value %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee(), value)

	hash, err := s.v3Router.ExactInputSingle(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test19ExactInputSingleNativeETH: broadcast exactInputSingle tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "exactInputSingle transaction reverted")
	s.T().Logf("Test19ExactInputSingleNativeETH: exactInputSingle confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.poolWETH))
	s.Require().True(swaps[0].Recipient.Equal(s.from))
	wethIsToken0 := s.wethToken0.Equal(s.weth9)
	actualAmountIn := swaps[0].Amount0
	actualAmountOut := new(big.Int).Neg(swaps[0].Amount1)
	if !wethIsToken0 {
		actualAmountIn, actualAmountOut = swaps[0].Amount1, new(big.Int).Neg(swaps[0].Amount0)
	}
	s.Require().Zero(value.Cmp(actualAmountIn), "the pool must have received exactly the native value we attached, auto-wrapped into WETH9 by the router")
	s.Require().Zero(predictedAmountOut.Cmp(actualAmountOut))
	s.T().Logf("Test19ExactInputSingleNativeETH: Swap confirms amountIn(WETH9)=%s amountOut(tokenA)=%s", actualAmountIn, actualAmountOut)

	var balanceAAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(predictedAmountOut) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after swap")
	s.T().Logf("Test19ExactInputSingleNativeETH: tokenA balance %s (%s -> %s)", balanceAAfter, balanceABefore, balanceAAfter)
}

func (s *RouterLifeCycleSuite) Test20SweepToken() {
	amountToSweep := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	s.T().Logf("Test20SweepToken: sending %s of %s directly to router %s, then sweeping it back", amountToSweep, s.tokenA, s.router)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedTransfer, err := s.erc20.CreateTransfer(s.ctx, builder, s.tokenA, s.from, s.router, amountToSweep)
	s.Require().NoError(err)
	s.Require().NoError(unsignedTransfer.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedTransfer.MaxFee())
	s.mu.Unlock()
	transferHash, err := s.erc20.Transfer(s.ctx, unsignedTransfer)
	s.Require().NoError(err)
	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	var routerBalanceBefore *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceBefore, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.router, rpc.BlockTagLatest)
		return err == nil && amountToSweep.Cmp(routerBalanceBefore) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "router's token balance not visible after transfer")
	s.T().Logf("Test20SweepToken: transfer confirmed in block %d, router's tokenA balance now %s", transferReceipt.BlockNumber, routerBalanceBefore)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	sweepBuilder := tx.NewDynamicFeeBuilder()
	sweepBuilder.PriceBuffer(150)
	sweepBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateSweepToken(s.ctx, sweepBuilder, s.router, s.from, s.tokenA, amountToSweep, s.from)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test20SweepToken: signed sweepToken tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.SweepToken(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test20SweepToken: broadcast sweepToken tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "sweepToken transaction reverted")
	s.T().Logf("Test20SweepToken: sweepToken confirmed in block %d", receipt.BlockNumber)

	transfers, err := erc20.ExtractTransfers(receipt.Logs)
	s.Require().NoError(err)
	var swept *erc20.Transfer
	for i := range transfers {
		if transfers[i].Contract.Equal(s.tokenA) && transfers[i].From.Equal(s.router) {
			swept = &transfers[i]
		}
	}
	s.Require().NotNil(swept, "sweepToken must emit the token's own Transfer event, from the router")
	s.Require().True(swept.To.Equal(s.from))
	s.Require().Zero(amountToSweep.Cmp(swept.Value))
	s.T().Logf("Test20SweepToken: Transfer event confirms from=%s to=%s value=%s", swept.From, swept.To, swept.Value)

	var routerBalanceAfter, balanceAAfter *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.router, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		return err == nil && routerBalanceAfter.Sign() == 0 && new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(amountToSweep) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after sweep")
	s.T().Logf("Test20SweepToken: router's tokenA balance now %s (fully swept), our tokenA balance %s (%s -> %s)", routerBalanceAfter, balanceAAfter, balanceABefore, balanceAAfter)
}

func (s *RouterLifeCycleSuite) Test21SweepTokenWithFee() {
	amountToSweep := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	feeBips := big.NewInt(100)
	s.T().Logf("Test21SweepTokenWithFee: sending %s of %s directly to router %s, sweeping with feeBips=%s to feeRecipient=%s", amountToSweep, s.tokenA, s.router, feeBips, s.feeRecipient)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedTransfer, err := s.erc20.CreateTransfer(s.ctx, builder, s.tokenA, s.from, s.router, amountToSweep)
	s.Require().NoError(err)
	s.Require().NoError(unsignedTransfer.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedTransfer.MaxFee())
	s.mu.Unlock()
	transferHash, err := s.erc20.Transfer(s.ctx, unsignedTransfer)
	s.Require().NoError(err)
	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	var routerBalanceBefore *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceBefore, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.router, rpc.BlockTagLatest)
		return err == nil && amountToSweep.Cmp(routerBalanceBefore) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "router's token balance not visible after transfer")
	s.T().Logf("Test21SweepTokenWithFee: transfer confirmed in block %d, router's tokenA balance now %s", transferReceipt.BlockNumber, routerBalanceBefore)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	feeRecipientBalanceBefore, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.feeRecipient, rpc.BlockTagLatest)
	s.Require().NoError(err)

	sweepBuilder := tx.NewDynamicFeeBuilder()
	sweepBuilder.PriceBuffer(150)
	sweepBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateSweepTokenWithFee(s.ctx, sweepBuilder, s.router, s.from, s.tokenA, amountToSweep, s.from, feeBips, s.feeRecipient)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test21SweepTokenWithFee: signed sweepTokenWithFee tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.SweepTokenWithFee(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test21SweepTokenWithFee: broadcast sweepTokenWithFee tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "sweepTokenWithFee transaction reverted")
	s.T().Logf("Test21SweepTokenWithFee: sweepTokenWithFee confirmed in block %d", receipt.BlockNumber)

	expectedFee := new(big.Int).Div(new(big.Int).Mul(amountToSweep, feeBips), big.NewInt(10000))
	expectedToRecipient := new(big.Int).Sub(amountToSweep, expectedFee)

	var routerBalanceAfter, balanceAAfter, feeRecipientBalanceAfter *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.router, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		feeRecipientBalanceAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.feeRecipient, rpc.BlockTagLatest)
		return err == nil && routerBalanceAfter.Sign() == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after sweep")
	s.Require().Zero(expectedToRecipient.Cmp(new(big.Int).Sub(balanceAAfter, balanceABefore)), "recipient must receive amount minus the fee")
	s.Require().Zero(expectedFee.Cmp(new(big.Int).Sub(feeRecipientBalanceAfter, feeRecipientBalanceBefore)), "feeRecipient must receive exactly feeBips/10000 of the swept amount")
	s.T().Logf("Test21SweepTokenWithFee: router's tokenA balance now %s (fully swept), recipient got %s, feeRecipient got %s (expected fee %s)", routerBalanceAfter, new(big.Int).Sub(balanceAAfter, balanceABefore), new(big.Int).Sub(feeRecipientBalanceAfter, feeRecipientBalanceBefore), expectedFee)
}

func (s *RouterLifeCycleSuite) Test22UnwrapWETH9() {
	amountToUnwrap := new(big.Int).Div(s.wethLiquidity, big.NewInt(10))
	s.T().Logf("Test22UnwrapWETH9: wrapping %s fresh native into WETH9, sending it to router, then unwrapping it back", amountToUnwrap)

	depositFn := abi.NewFunction("deposit", nil, nil)
	wrapBuilder := tx.NewDynamicFeeBuilder()
	wrapBuilder.PriceBuffer(150)
	wrapBuilder.LimitBuffer(300)
	unsignedWrap, err := wrapBuilder.Build(s.from, s.weth9, amountToUnwrap, depositFn.SelectorBytes()).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsignedWrap.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedWrap.MaxFee())
	s.mu.Unlock()
	wrapRaw, err := unsignedWrap.EncodeRLP()
	s.Require().NoError(err)
	wrapHash, err := tx.Broadcast(s.ctx, s.client, wrapRaw)
	s.Require().NoError(err)
	wrapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, wrapHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(wrapReceipt.Status, "WETH9.deposit() transaction reverted")
	s.T().Logf("Test22UnwrapWETH9: wrapped %s into WETH9, confirmed in block %d", amountToUnwrap, wrapReceipt.BlockNumber)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedTransfer, err := s.erc20.CreateTransfer(s.ctx, builder, s.weth9, s.from, s.router, amountToUnwrap)
	s.Require().NoError(err)
	s.Require().NoError(unsignedTransfer.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedTransfer.MaxFee())
	s.mu.Unlock()
	transferHash, err := s.erc20.Transfer(s.ctx, unsignedTransfer)
	s.Require().NoError(err)
	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	var routerBalanceBefore *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceBefore, err = s.erc20.BalanceOf(s.ctx, s.weth9, s.router, rpc.BlockTagLatest)
		return err == nil && amountToUnwrap.Cmp(routerBalanceBefore) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "router's WETH9 balance not visible after transfer")
	s.T().Logf("Test22UnwrapWETH9: sent %s of WETH9 to router, confirmed in block %d, router's WETH9 balance now %s", amountToUnwrap, transferReceipt.BlockNumber, routerBalanceBefore)

	var nativeBalanceBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceBeforeHex)))
	nativeBalanceBefore, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceBeforeHex, "0x"), 16)
	s.Require().True(ok)

	unwrapBuilder := tx.NewDynamicFeeBuilder()
	unwrapBuilder.PriceBuffer(150)
	unwrapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateUnwrapWETH9(s.ctx, unwrapBuilder, s.router, s.from, amountToUnwrap, s.from)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.T().Logf("Test22UnwrapWETH9: signed unwrapWETH9 tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.UnwrapWETH9(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test22UnwrapWETH9: broadcast unwrapWETH9 tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "unwrapWETH9 transaction reverted")
	s.T().Logf("Test22UnwrapWETH9: unwrapWETH9 confirmed in block %d, gasUsed=%d", receipt.BlockNumber, receipt.GasUsed)

	effectiveGasPriceHex, ok := receipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	effectiveGasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(effectiveGasPriceHex, "0x"), 16)
	s.Require().True(ok)
	gasCost := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), effectiveGasPrice)
	if l1FeeHex, ok := receipt.Raw["l1Fee"].(string); ok {
		l1Fee, ok := new(big.Int).SetString(strings.TrimPrefix(l1FeeHex, "0x"), 16)
		s.Require().True(ok)
		gasCost.Add(gasCost, l1Fee)
	}
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, gasCost)
	s.mu.Unlock()
	s.T().Logf("Test22UnwrapWETH9: actual gas cost paid = %s (gasUsed=%d, effectiveGasPrice=%s, including any L1 data fee on OP-stack chains)", gasCost, receipt.GasUsed, effectiveGasPrice)

	var routerBalanceAfter, actualReceived *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceAfter, err = s.erc20.BalanceOf(s.ctx, s.weth9, s.router, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		var nativeBalanceAfterHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceAfterHex)); err != nil {
			return false
		}
		nativeBalanceAfter, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceAfterHex, "0x"), 16)
		if !ok {
			return false
		}
		actualReceived = new(big.Int).Add(new(big.Int).Sub(nativeBalanceAfter, nativeBalanceBefore), gasCost)
		return routerBalanceAfter.Sign() == 0 && amountToUnwrap.Cmp(actualReceived) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after unwrap")
	s.T().Logf("Test22UnwrapWETH9: router's WETH9 balance now %s (fully unwrapped), we actually received %s native %s (after backing out gas cost)", routerBalanceAfter, actualReceived, s.NativeSymbol)
}

func (s *RouterLifeCycleSuite) Test23UnwrapWETH9WithFee() {
	amountToUnwrap := new(big.Int).Div(s.wethLiquidity, big.NewInt(10))
	feeBips := big.NewInt(100)
	s.T().Logf("Test23UnwrapWETH9WithFee: wrapping %s fresh native into WETH9, sending it to router, unwrapping with feeBips=%s to feeRecipient=%s", amountToUnwrap, feeBips, s.feeRecipient)

	depositFn := abi.NewFunction("deposit", nil, nil)
	wrapBuilder := tx.NewDynamicFeeBuilder()
	wrapBuilder.PriceBuffer(150)
	wrapBuilder.LimitBuffer(300)
	unsignedWrap, err := wrapBuilder.Build(s.from, s.weth9, amountToUnwrap, depositFn.SelectorBytes()).Pack(s.ctx, s.client)
	s.Require().NoError(err)
	s.Require().NoError(unsignedWrap.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedWrap.MaxFee())
	s.mu.Unlock()
	wrapRaw, err := unsignedWrap.EncodeRLP()
	s.Require().NoError(err)
	wrapHash, err := tx.Broadcast(s.ctx, s.client, wrapRaw)
	s.Require().NoError(err)
	wrapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, wrapHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(wrapReceipt.Status, "WETH9.deposit() transaction reverted")
	s.T().Logf("Test23UnwrapWETH9WithFee: wrapped %s into WETH9, confirmed in block %d", amountToUnwrap, wrapReceipt.BlockNumber)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsignedTransfer, err := s.erc20.CreateTransfer(s.ctx, builder, s.weth9, s.from, s.router, amountToUnwrap)
	s.Require().NoError(err)
	s.Require().NoError(unsignedTransfer.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedTransfer.MaxFee())
	s.mu.Unlock()
	transferHash, err := s.erc20.Transfer(s.ctx, unsignedTransfer)
	s.Require().NoError(err)
	transferReceipt, err := tx.WaitForReceipt(s.ctx, s.client, transferHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(transferReceipt.Status, "transfer transaction reverted")
	var routerBalanceBefore *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceBefore, err = s.erc20.BalanceOf(s.ctx, s.weth9, s.router, rpc.BlockTagLatest)
		return err == nil && amountToUnwrap.Cmp(routerBalanceBefore) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "router's WETH9 balance not visible after transfer")
	s.T().Logf("Test23UnwrapWETH9WithFee: sent %s of WETH9 to router, confirmed in block %d", amountToUnwrap, transferReceipt.BlockNumber)

	var feeRecipientBalanceBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.feeRecipient.String(), rpc.BlockTagLatest, &feeRecipientBalanceBeforeHex)))
	feeRecipientBalanceBefore, ok := new(big.Int).SetString(strings.TrimPrefix(feeRecipientBalanceBeforeHex, "0x"), 16)
	s.Require().True(ok)

	unwrapBuilder := tx.NewDynamicFeeBuilder()
	unwrapBuilder.PriceBuffer(150)
	unwrapBuilder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateUnwrapWETH9WithFee(s.ctx, unwrapBuilder, s.router, s.from, amountToUnwrap, s.from, feeBips, s.feeRecipient)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.T().Logf("Test23UnwrapWETH9WithFee: signed unwrapWETH9WithFee tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.UnwrapWETH9WithFee(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test23UnwrapWETH9WithFee: broadcast unwrapWETH9WithFee tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "unwrapWETH9WithFee transaction reverted")
	s.T().Logf("Test23UnwrapWETH9WithFee: unwrapWETH9WithFee confirmed in block %d", receipt.BlockNumber)

	effectiveGasPriceHex, ok := receipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	effectiveGasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(effectiveGasPriceHex, "0x"), 16)
	s.Require().True(ok)
	gasCost := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), effectiveGasPrice)
	if l1FeeHex, ok := receipt.Raw["l1Fee"].(string); ok {
		l1Fee, ok := new(big.Int).SetString(strings.TrimPrefix(l1FeeHex, "0x"), 16)
		s.Require().True(ok)
		gasCost.Add(gasCost, l1Fee)
	}
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, gasCost)
	s.mu.Unlock()

	expectedFee := new(big.Int).Div(new(big.Int).Mul(amountToUnwrap, feeBips), big.NewInt(10000))
	expectedToRecipient := new(big.Int).Sub(amountToUnwrap, expectedFee)

	var routerBalanceAfter, feeRecipientBalanceAfter *big.Int
	s.Require().Eventually(func() bool {
		routerBalanceAfter, err = s.erc20.BalanceOf(s.ctx, s.weth9, s.router, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		var feeRecipientBalanceAfterHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.feeRecipient.String(), rpc.BlockTagLatest, &feeRecipientBalanceAfterHex)); err != nil {
			return false
		}
		feeRecipientBalanceAfter, ok = new(big.Int).SetString(strings.TrimPrefix(feeRecipientBalanceAfterHex, "0x"), 16)
		return ok && routerBalanceAfter.Sign() == 0 && expectedFee.Cmp(new(big.Int).Sub(feeRecipientBalanceAfter, feeRecipientBalanceBefore)) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after unwrap")
	s.T().Logf("Test23UnwrapWETH9WithFee: router's WETH9 balance now %s (fully unwrapped), feeRecipient received %s native %s (expected %s), recipient should have received %s", routerBalanceAfter, new(big.Int).Sub(feeRecipientBalanceAfter, feeRecipientBalanceBefore), s.NativeSymbol, expectedFee, expectedToRecipient)
}

func (s *RouterLifeCycleSuite) Test24RefundETH() {
	amountIn := new(big.Int).Div(s.wethLiquidity, big.NewInt(100))
	excess := new(big.Int).Div(amountIn, big.NewInt(2))
	value := new(big.Int).Add(amountIn, excess)
	s.T().Logf("Test24RefundETH: swapping with amountIn=%s but attaching value=%s (excess=%s should be left in the router)", amountIn, value, excess)

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)

	params := NewExactInputSingleParams(s.weth9, s.tokenA, s.Fee, s.from, amountIn, big.NewInt(0), big.NewInt(0))

	predictedAmountOut, err := s.v3Router.CallExactInputSingle(s.ctx, s.router, s.from, value, params, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.T().Logf("Test24RefundETH: CallExactInputSingle predicted amountOut=%s", predictedAmountOut)

	swapBuilder := tx.NewDynamicFeeBuilder()
	swapBuilder.PriceBuffer(150)
	swapBuilder.LimitBuffer(300)
	unsignedSwap, err := s.v3Router.CreateExactInputSingle(s.ctx, swapBuilder, s.router, s.from, value, params)
	s.Require().NoError(err)
	s.Require().NoError(unsignedSwap.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedSwap.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test24RefundETH: signed exactInputSingle tx (nonce %d, gasLimit %d, maxFee %s, value %s)", unsignedSwap.AddressNonce(), unsignedSwap.Gas(), unsignedSwap.MaxFee(), value)

	swapHash, err := s.v3Router.ExactInputSingle(s.ctx, unsignedSwap)
	s.Require().NoError(err)
	s.T().Logf("Test24RefundETH: broadcast exactInputSingle tx %s", swapHash)

	swapReceipt, err := tx.WaitForReceipt(s.ctx, s.client, swapHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(swapReceipt.Status, "exactInputSingle transaction reverted")
	s.T().Logf("Test24RefundETH: exactInputSingle confirmed in block %d", swapReceipt.BlockNumber)

	var routerNativeBalance *big.Int
	s.Require().Eventually(func() bool {
		var routerNativeBalanceHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.router.String(), rpc.BlockTagLatest, &routerNativeBalanceHex)); err != nil {
			return false
		}
		var ok bool
		routerNativeBalance, ok = new(big.Int).SetString(strings.TrimPrefix(routerNativeBalanceHex, "0x"), 16)
		return ok && excess.Cmp(routerNativeBalance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "router's native balance not visible after the overpaid swap")
	s.T().Logf("Test24RefundETH: router's native balance is now %s (the leftover excess), %d WETH9 wrapped and spent", routerNativeBalance, 1)

	var nativeBalanceBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceBeforeHex)))
	nativeBalanceBefore, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceBeforeHex, "0x"), 16)
	s.Require().True(ok)

	refundBuilder := tx.NewDynamicFeeBuilder()
	refundBuilder.PriceBuffer(150)
	refundBuilder.LimitBuffer(300)
	unsignedRefund, err := s.v3Router.CreateRefundETH(s.ctx, refundBuilder, s.router, s.from)
	s.Require().NoError(err)
	s.Require().NoError(unsignedRefund.Sign(s.signer))
	s.T().Logf("Test24RefundETH: signed refundETH tx (nonce %d, gasLimit %d, maxFee %s)", unsignedRefund.AddressNonce(), unsignedRefund.Gas(), unsignedRefund.MaxFee())

	refundHash, err := s.v3Router.RefundETH(s.ctx, unsignedRefund)
	s.Require().NoError(err)
	s.T().Logf("Test24RefundETH: broadcast refundETH tx %s", refundHash)

	refundReceipt, err := tx.WaitForReceipt(s.ctx, s.client, refundHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(refundReceipt.Status, "refundETH transaction reverted")
	s.T().Logf("Test24RefundETH: refundETH confirmed in block %d", refundReceipt.BlockNumber)

	effectiveGasPriceHex, ok := refundReceipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	effectiveGasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(effectiveGasPriceHex, "0x"), 16)
	s.Require().True(ok)
	gasCost := new(big.Int).Mul(new(big.Int).SetUint64(refundReceipt.GasUsed), effectiveGasPrice)
	if l1FeeHex, ok := refundReceipt.Raw["l1Fee"].(string); ok {
		l1Fee, ok := new(big.Int).SetString(strings.TrimPrefix(l1FeeHex, "0x"), 16)
		s.Require().True(ok)
		gasCost.Add(gasCost, l1Fee)
	}
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, gasCost)
	s.mu.Unlock()

	var routerNativeBalanceAfter, actualRefund *big.Int
	s.Require().Eventually(func() bool {
		var routerNativeBalanceAfterHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.router.String(), rpc.BlockTagLatest, &routerNativeBalanceAfterHex)); err != nil {
			return false
		}
		var ok bool
		routerNativeBalanceAfter, ok = new(big.Int).SetString(strings.TrimPrefix(routerNativeBalanceAfterHex, "0x"), 16)
		if !ok {
			return false
		}
		var nativeBalanceAfterHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceAfterHex)); err != nil {
			return false
		}
		nativeBalanceAfter, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceAfterHex, "0x"), 16)
		if !ok {
			return false
		}
		actualRefund = new(big.Int).Add(new(big.Int).Sub(nativeBalanceAfter, nativeBalanceBefore), gasCost)
		return routerNativeBalanceAfter.Sign() == 0 && excess.Cmp(actualRefund) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after refund")
	s.T().Logf("Test24RefundETH: router's native balance now %s (fully refunded), we actually received %s native %s back (after backing out gas cost), matching the excess", routerNativeBalanceAfter, actualRefund, s.NativeSymbol)

	var balanceAAfter *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		return err == nil && new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(predictedAmountOut) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "tokenA balance not visible after swap")
	s.T().Logf("Test24RefundETH: tokenA balance %s (%s -> %s), matching the swap's predicted amountOut regardless of the overpaid value", balanceAAfter, balanceABefore, balanceAAfter)
}

func (s *RouterLifeCycleSuite) Test25SelfPermit() {
	value := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	name, err := s.erc20.Name(s.ctx, s.tokenA, rpc.BlockTagLatest)
	s.Require().NoError(err)

	noncesFn := abi.NewFunction("nonces", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	noncesCalldata, err := noncesFn.EncodeCall(s.from)
	s.Require().NoError(err)
	var rawNonce string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHCall("call", map[string]any{"to": s.tokenA.String(), "data": encoding.Hex.EncodePrefixed(noncesCalldata)}, rpc.BlockTagLatest, &rawNonce)))
	nonceData, err := encoding.Hex.Decode(rawNonce)
	s.Require().NoError(err)
	nonce, err := noncesFn.DecodeSingleReturn[*big.Int](nonceData)
	s.Require().NoError(err)
	s.T().Logf("Test25SelfPermit: token=%s name=%s nonce=%s value=%s deadline=%s", s.tokenA, name, nonce, value, deadline)

	domain := eip712.NewDomain(name, "1", s.chainID, s.tokenA, nil)
	permitTypes := eip712.Types{
		"Permit": {
			eip712.NewField("owner", abi.Address.String()),
			eip712.NewField("spender", abi.Address.String()),
			eip712.NewField("value", abi.Uint256.String()),
			eip712.NewField("nonce", abi.Uint256.String()),
			eip712.NewField("deadline", abi.Uint256.String()),
		},
	}
	message := eip712.Message{
		"owner":    s.from,
		"spender":  s.router,
		"value":    value,
		"nonce":    nonce,
		"deadline": deadline,
	}
	digest, err := eip712.Hash(domain, "Permit", permitTypes, message)
	s.Require().NoError(err)
	sig, err := s.signer.Sign(digest)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateSelfPermit(s.ctx, builder, s.router, s.from, s.tokenA, value, deadline, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test25SelfPermit: signed selfPermit tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.SelfPermit(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test25SelfPermit: broadcast selfPermit tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "selfPermit transaction reverted")
	s.T().Logf("Test25SelfPermit: selfPermit confirmed in block %d", receipt.BlockNumber)

	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.tokenA, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && value.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after selfPermit")
	s.T().Logf("Test25SelfPermit: allowance(from, router) = %s, set purely from the holder's signature, no approve tx sent", allowance)
}

func (s *RouterLifeCycleSuite) Test26SelfPermitIfNecessary() {
	value1 := new(big.Int).Div(s.mintValue, big.NewInt(1000))
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	resetBuilder := tx.NewDynamicFeeBuilder()
	resetBuilder.PriceBuffer(150)
	resetBuilder.LimitBuffer(300)
	unsignedReset, err := s.erc20.CreateApprove(s.ctx, resetBuilder, s.tokenB, s.from, s.router, big.NewInt(0))
	s.Require().NoError(err)
	s.Require().NoError(unsignedReset.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsignedReset.MaxFee())
	s.mu.Unlock()
	resetHash, err := s.erc20.Approve(s.ctx, unsignedReset)
	s.Require().NoError(err)
	resetReceipt, err := tx.WaitForReceipt(s.ctx, s.client, resetHash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(resetReceipt.Status, "allowance-reset transaction reverted")
	s.Require().Eventually(func() bool {
		allowance, err := s.erc20.Allowance(s.ctx, s.tokenB, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && allowance.Sign() == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible as 0 after reset")
	s.T().Logf("Test26SelfPermitIfNecessary: reset tokenB's leftover router allowance (from Test05) back to 0, confirmed in block %d", resetReceipt.BlockNumber)

	name, err := s.erc20.Name(s.ctx, s.tokenB, rpc.BlockTagLatest)
	s.Require().NoError(err)

	noncesFn := abi.NewFunction("nonces", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	noncesCalldata, err := noncesFn.EncodeCall(s.from)
	s.Require().NoError(err)
	var rawNonce string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHCall("call", map[string]any{"to": s.tokenB.String(), "data": encoding.Hex.EncodePrefixed(noncesCalldata)}, rpc.BlockTagLatest, &rawNonce)))
	nonceData, err := encoding.Hex.Decode(rawNonce)
	s.Require().NoError(err)
	nonce, err := noncesFn.DecodeSingleReturn[*big.Int](nonceData)
	s.Require().NoError(err)
	s.T().Logf("Test26SelfPermitIfNecessary: token=%s name=%s nonce=%s value1=%s deadline=%s", s.tokenB, name, nonce, value1, deadline)

	domain := eip712.NewDomain(name, "1", s.chainID, s.tokenB, nil)
	permitTypes := eip712.Types{
		"Permit": {
			eip712.NewField("owner", abi.Address.String()),
			eip712.NewField("spender", abi.Address.String()),
			eip712.NewField("value", abi.Uint256.String()),
			eip712.NewField("nonce", abi.Uint256.String()),
			eip712.NewField("deadline", abi.Uint256.String()),
		},
	}
	message1 := eip712.Message{
		"owner":    s.from,
		"spender":  s.router,
		"value":    value1,
		"nonce":    nonce,
		"deadline": deadline,
	}
	digest1, err := eip712.Hash(domain, "Permit", permitTypes, message1)
	s.Require().NoError(err)
	sig1, err := s.signer.Sign(digest1)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned1, err := s.v3Router.CreateSelfPermitIfNecessary(s.ctx, builder, s.router, s.from, s.tokenB, value1, deadline, sig1)
	s.Require().NoError(err)
	s.Require().NoError(unsigned1.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned1.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test26SelfPermitIfNecessary: signed first selfPermitIfNecessary tx (nonce %d, maxFee %s)", unsigned1.AddressNonce(), unsigned1.MaxFee())

	hash1, err := s.v3Router.SelfPermitIfNecessary(s.ctx, unsigned1)
	s.Require().NoError(err)
	receipt1, err := tx.WaitForReceipt(s.ctx, s.client, hash1, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt1.Status, "first selfPermitIfNecessary transaction reverted")

	var allowanceAfterFirst *big.Int
	s.Require().Eventually(func() bool {
		allowanceAfterFirst, err = s.erc20.Allowance(s.ctx, s.tokenB, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && value1.Cmp(allowanceAfterFirst) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after first selfPermitIfNecessary")
	s.T().Logf("Test26SelfPermitIfNecessary: first call (allowance 0 < value1) executed the permit, allowance now %s, confirmed in block %d", allowanceAfterFirst, receipt1.BlockNumber)

	value2 := new(big.Int).Div(value1, big.NewInt(2))
	s.T().Logf("Test26SelfPermitIfNecessary: second call with value2=%s (< current allowance) and a stale/wrong signature -- must skip permit entirely and still succeed", value2)

	builder2 := tx.NewDynamicFeeBuilder()
	builder2.PriceBuffer(150)
	builder2.LimitBuffer(300)
	unsigned2, err := s.v3Router.CreateSelfPermitIfNecessary(s.ctx, builder2, s.router, s.from, s.tokenB, value2, deadline, sig1)
	s.Require().NoError(err)
	s.Require().NoError(unsigned2.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned2.MaxFee())
	s.mu.Unlock()

	hash2, err := s.v3Router.SelfPermitIfNecessary(s.ctx, unsigned2)
	s.Require().NoError(err)
	receipt2, err := tx.WaitForReceipt(s.ctx, s.client, hash2, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt2.Status, "second selfPermitIfNecessary transaction reverted -- it should have skipped permit and succeeded trivially")

	allowanceAfterSecond, err := s.erc20.Allowance(s.ctx, s.tokenB, s.from, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(allowanceAfterFirst.Cmp(allowanceAfterSecond), "allowance must be unchanged -- confirms permit (with the stale signature) was never actually invoked")
	s.T().Logf("Test26SelfPermitIfNecessary: second call succeeded despite the wrong signature, allowance still %s, confirmed in block %d", allowanceAfterSecond, receipt2.BlockNumber)
}

func (s *RouterLifeCycleSuite) Test27SelfPermitAllowed() {
	expiry := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	name, err := s.erc20.Name(s.ctx, s.tokenC, rpc.BlockTagLatest)
	s.Require().NoError(err)

	noncesFn := abi.NewFunction("nonces", abi.NewTypes(abi.Address), abi.NewTypes(abi.Uint256))
	noncesCalldata, err := noncesFn.EncodeCall(s.from)
	s.Require().NoError(err)
	var rawNonce string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHCall("call", map[string]any{"to": s.tokenC.String(), "data": encoding.Hex.EncodePrefixed(noncesCalldata)}, rpc.BlockTagLatest, &rawNonce)))
	nonceData, err := encoding.Hex.Decode(rawNonce)
	s.Require().NoError(err)
	nonce, err := noncesFn.DecodeSingleReturn[*big.Int](nonceData)
	s.Require().NoError(err)
	s.T().Logf("Test27SelfPermitAllowed: token=%s name=%s nonce=%s expiry=%s", s.tokenC, name, nonce, expiry)

	domain := eip712.NewDomain(name, "1", s.chainID, s.tokenC, nil)
	permitAllowedTypes := eip712.Types{
		"Permit": {
			eip712.NewField("holder", abi.Address.String()),
			eip712.NewField("spender", abi.Address.String()),
			eip712.NewField("nonce", abi.Uint256.String()),
			eip712.NewField("expiry", abi.Uint256.String()),
			eip712.NewField("allowed", abi.Bool.String()),
		},
	}
	message := eip712.Message{
		"holder":  s.from,
		"spender": s.router,
		"nonce":   nonce,
		"expiry":  expiry,
		"allowed": true,
	}
	digest, err := eip712.Hash(domain, "Permit", permitAllowedTypes, message)
	s.Require().NoError(err)
	sig, err := s.signer.Sign(digest)
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateSelfPermitAllowed(s.ctx, builder, s.router, s.from, s.tokenC, nonce, expiry, sig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test27SelfPermitAllowed: signed selfPermitAllowed tx (nonce %d, gasLimit %d, maxFee %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee())

	hash, err := s.v3Router.SelfPermitAllowed(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test27SelfPermitAllowed: broadcast selfPermitAllowed tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "selfPermitAllowed transaction reverted")
	s.T().Logf("Test27SelfPermitAllowed: selfPermitAllowed confirmed in block %d", receipt.BlockNumber)

	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	var allowance *big.Int
	s.Require().Eventually(func() bool {
		allowance, err = s.erc20.Allowance(s.ctx, s.tokenC, s.from, s.router, rpc.BlockTagLatest)
		return err == nil && maxUint256.Cmp(allowance) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "allowance not visible after selfPermitAllowed")
	s.T().Logf("Test27SelfPermitAllowed: allowance(from, router) = %s (max, DAI-style allowed=true), set purely from the holder's signature", allowance)
}

func (s *RouterLifeCycleSuite) Test28SelfPermitAllowedIfNecessary() {
	nextExpiry := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
	s.T().Logf("Test28SelfPermitAllowedIfNecessary: token=%s allowance should already be max from Test27, so a stale/wrong signature here must be skipped and still succeed", s.tokenC)

	wrongNonce := big.NewInt(999999)
	wrongSig, err := s.signer.Sign(types.NewHashFromBytes([]byte("deliberately the wrong 32-byte digest!!")))
	s.Require().NoError(err)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateSelfPermitAllowedIfNecessary(s.ctx, builder, s.router, s.from, s.tokenC, wrongNonce, nextExpiry, wrongSig)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test28SelfPermitAllowedIfNecessary: signed selfPermitAllowedIfNecessary tx (nonce %d, maxFee %s)", unsigned.AddressNonce(), unsigned.MaxFee())

	hash, err := s.v3Router.SelfPermitAllowedIfNecessary(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test28SelfPermitAllowedIfNecessary: broadcast tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "selfPermitAllowedIfNecessary transaction reverted -- it should have skipped permit (allowance already max) and succeeded trivially")
	s.T().Logf("Test28SelfPermitAllowedIfNecessary: succeeded despite the bogus nonce/signature, confirming allowance already being max skipped the permit call, confirmed in block %d", receipt.BlockNumber)

	maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	allowance, err := s.erc20.Allowance(s.ctx, s.tokenC, s.from, s.router, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Zero(maxUint256.Cmp(allowance), "allowance must remain max, unchanged")
	s.T().Logf("Test28SelfPermitAllowedIfNecessary: allowance(from, router) still %s", allowance)
}

func (s *RouterLifeCycleSuite) Test29Multicall() {
	amountIn := new(big.Int).Div(s.wethLiquidity, big.NewInt(100))
	excess := new(big.Int).Div(amountIn, big.NewInt(2))
	value := new(big.Int).Add(amountIn, excess)
	s.T().Logf("Test29Multicall: bundling exactInputSingle(amountIn=%s, value=%s) + refundETH into one multicall tx (excess=%s expected back)", amountIn, value, excess)

	params := NewExactInputSingleParams(s.weth9, s.tokenA, s.Fee, s.from, amountIn, big.NewInt(0), big.NewInt(0))
	swapCalldata, err := EncodeExactInputSingle(params.TokenIn, params.TokenOut, params.Fee, params.Recipient, params.AmountIn, params.AmountOutMinimum, params.SqrtPriceLimitX96)
	s.Require().NoError(err)
	refundCalldata := EncodeRouterRefundETH()
	calls := [][]byte{swapCalldata, refundCalldata}

	balanceABefore, err := s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
	s.Require().NoError(err)
	var nativeBalanceBeforeHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceBeforeHex)))
	nativeBalanceBefore, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceBeforeHex, "0x"), 16)
	s.Require().True(ok)

	results, err := s.v3Router.CallMulticall(s.ctx, s.router, s.from, value, calls, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(results, 2)
	predictedAmountOut, err := DecodeExactInputSingle(results[0])
	s.Require().NoError(err)
	s.Require().True(predictedAmountOut.Sign() > 0)
	s.Require().Empty(results[1], "refundETH has no return data")
	s.T().Logf("Test29Multicall: CallMulticall predicted amountOut=%s", predictedAmountOut)

	builder := tx.NewDynamicFeeBuilder()
	builder.PriceBuffer(150)
	builder.LimitBuffer(300)
	unsigned, err := s.v3Router.CreateMulticall(s.ctx, builder, s.router, s.from, value, calls)
	s.Require().NoError(err)
	s.Require().NoError(unsigned.Sign(s.signer))
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, unsigned.MaxFee())
	s.mu.Unlock()
	s.T().Logf("Test29Multicall: signed multicall tx (nonce %d, gasLimit %d, maxFee %s, value %s)", unsigned.AddressNonce(), unsigned.Gas(), unsigned.MaxFee(), value)

	hash, err := s.v3Router.Multicall(s.ctx, unsigned)
	s.Require().NoError(err)
	s.T().Logf("Test29Multicall: broadcast multicall tx %s", hash)

	receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, time.Duration(s.Interval)*time.Millisecond, time.Duration(s.Timeout)*time.Second)
	s.Require().NoError(err)
	s.Require().True(receipt.Status, "multicall transaction reverted")
	s.T().Logf("Test29Multicall: multicall confirmed in block %d", receipt.BlockNumber)

	swaps, err := ExtractSwap(receipt.Logs)
	s.Require().NoError(err)
	s.Require().Len(swaps, 1)
	s.Require().True(swaps[0].Contract.Equal(s.poolWETH))

	effectiveGasPriceHex, ok := receipt.Raw["effectiveGasPrice"].(string)
	s.Require().True(ok)
	effectiveGasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(effectiveGasPriceHex, "0x"), 16)
	s.Require().True(ok)
	gasCost := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), effectiveGasPrice)
	if l1FeeHex, ok := receipt.Raw["l1Fee"].(string); ok {
		l1Fee, ok := new(big.Int).SetString(strings.TrimPrefix(l1FeeHex, "0x"), 16)
		s.Require().True(ok)
		gasCost.Add(gasCost, l1Fee)
	}
	s.mu.Lock()
	s.totalFee.Add(s.totalFee, gasCost)
	s.mu.Unlock()

	var balanceAAfter, actualRefund *big.Int
	s.Require().Eventually(func() bool {
		balanceAAfter, err = s.erc20.BalanceOf(s.ctx, s.tokenA, s.from, rpc.BlockTagLatest)
		if err != nil {
			return false
		}
		var nativeBalanceAfterHex string
		if err = s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.from.String(), rpc.BlockTagLatest, &nativeBalanceAfterHex)); err != nil {
			return false
		}
		nativeBalanceAfter, ok := new(big.Int).SetString(strings.TrimPrefix(nativeBalanceAfterHex, "0x"), 16)
		if !ok {
			return false
		}
		actualRefund = new(big.Int).Add(new(big.Int).Sub(nativeBalanceAfter, nativeBalanceBefore), gasCost)
		return new(big.Int).Sub(balanceAAfter, balanceABefore).Cmp(predictedAmountOut) == 0 && excess.Cmp(actualRefund) == 0
	}, time.Duration(s.Timeout)*time.Second, time.Duration(s.Interval)*time.Millisecond, "balances not visible after multicall")

	var routerNativeBalanceAfterHex string
	s.Require().NoError(s.client.Call(s.ctx, rpc.ETHGetBalance("balance", s.router.String(), rpc.BlockTagLatest, &routerNativeBalanceAfterHex)))
	routerNativeBalanceAfter, ok := new(big.Int).SetString(strings.TrimPrefix(routerNativeBalanceAfterHex, "0x"), 16)
	s.Require().True(ok)
	s.Require().Zero(routerNativeBalanceAfter.Sign(), "router must not retain any native balance after the bundled refund")

	s.T().Logf("Test29Multicall: in one tx, tokenA went %s -> %s (swap) and we got %s native %s back (refund, after backing out gas), router left holding 0", balanceABefore, balanceAAfter, actualRefund, s.NativeSymbol)
}
