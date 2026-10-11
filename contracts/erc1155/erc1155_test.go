package erc1155

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

var interfaceIDERC1155 = [4]byte{0xd9, 0xb6, 0x7a, 0x26}

// erc1155ByteCodeHex is OpenZeppelin's ERC1155 plus mint/mintBatch/burn/burnBatch, constructor(uri).
const erc1155ByteCodeHex = "0x608060405234801561000f575f5ffd5b5060405161272f38038061272f833981810160405281019061003191906101a8565b806100418161004860201b60201c565b50506104df565b80600290816100579190610410565b5050565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6100ba82610074565b810181811067ffffffffffffffff821117156100d9576100d8610084565b5b80604052505050565b5f6100eb61005b565b90506100f782826100b1565b919050565b5f67ffffffffffffffff82111561011657610115610084565b5b61011f82610074565b9050602081019050919050565b8281835e5f83830152505050565b5f61014c610147846100fc565b6100e2565b90508281526020810184848401111561016857610167610070565b5b61017384828561012c565b509392505050565b5f82601f83011261018f5761018e61006c565b5b815161019f84826020860161013a565b91505092915050565b5f602082840312156101bd576101bc610064565b5b5f82015167ffffffffffffffff8111156101da576101d9610068565b5b6101e68482850161017b565b91505092915050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061023d57607f821691505b6020821081036102505761024f6101f9565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f600883026102b27fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610277565b6102bc8683610277565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f6103006102fb6102f6846102d4565b6102dd565b6102d4565b9050919050565b5f819050919050565b610319836102e6565b61032d61032582610307565b848454610283565b825550505050565b5f5f905090565b610344610335565b61034f818484610310565b505050565b5f5b828110156103755761036a5f82840161033c565b600181019050610356565b505050565b601f8211156103c857828211156103c75761039481610256565b61039d83610268565b6103a685610268565b60208610156103b3575f90505b8083016103c282840382610354565b505050505b5b505050565b5f82821c905092915050565b5f6103e85f19846008026103cd565b1980831691505092915050565b5f61040083836103d9565b9150826002028217905092915050565b610419826101ef565b67ffffffffffffffff81111561043257610431610084565b5b61043c8254610226565b61044782828561037a565b5f60209050601f831160018114610478575f8415610466578287015190505b61047085826103f5565b8655506104d7565b601f19841661048686610256565b5f5b828110156104ad57848901518255600182019150602085019450602081019050610488565b868310156104ca57848901516104c6601f8916826103d9565b8355505b6001600288020188555050505b505050505050565b612243806104ec5f395ff3fe608060405234801561000f575f5ffd5b50600436106100b1575f3560e01c80636b20c4541161006f5780636b20c454146101ad578063a22cb465146101c9578063d81d0a15146101e5578063e985e9c514610201578063f242432a14610231578063f5298aca1461024d576100b1565b8062fdd58e146100b557806301ffc9a7146100e55780630e89341c14610115578063156e29f6146101455780632eb2c2d6146101615780634e1273f41461017d575b5f5ffd5b6100cf60048036038101906100ca9190611639565b610269565b6040516100dc9190611686565b60405180910390f35b6100ff60048036038101906100fa91906116f4565b6102be565b60405161010c9190611739565b60405180910390f35b61012f600480360381019061012a9190611752565b61039f565b60405161013c91906117ed565b60405180910390f35b61015f600480360381019061015a919061180d565b610431565b005b61017b60048036038101906101769190611a4d565b610450565b005b61019760048036038101906101929190611bd8565b6104f7565b6040516101a49190611d05565b60405180910390f35b6101c760048036038101906101c29190611d25565b610601565b005b6101e360048036038101906101de9190611dd7565b610611565b005b6101ff60048036038101906101fa9190611d25565b610627565b005b61021b60048036038101906102169190611e15565b610646565b6040516102289190611739565b60405180910390f35b61024b60048036038101906102469190611e53565b6106d4565b005b6102676004803603810190610262919061180d565b61077b565b005b5f5f5f8381526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f7fd9b67a26000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916148061038857507f0e89341c000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916145b8061039857506103978261078b565b5b9050919050565b6060600280546103ae90611f13565b80601f01602080910402602001604051908101604052809291908181526020018280546103da90611f13565b80156104255780601f106103fc57610100808354040283529160200191610425565b820191905f5260205f20905b81548152906001019060200180831161040857829003601f168201915b50505050509050919050565b61044b83838360405180602001604052805f8152506107f4565b505050565b5f610459610889565b90508073ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff161415801561049e575061049c8682610646565b155b156104e25780866040517fe237d9220000000000000000000000000000000000000000000000000000000081526004016104d9929190611f52565b60405180910390fd5b6104ef8686868686610890565b505050505050565b6060815183511461054357815183516040517f5b05999100000000000000000000000000000000000000000000000000000000815260040161053a929190611f79565b60405180910390fd5b5f835167ffffffffffffffff81111561055f5761055e611861565b5b60405190808252806020026020018201604052801561058d5781602001602082028036833780820191505090505b5090505f5f90505b84518110156105f6576105cc6105b4828761098490919063ffffffff16565b6105c7838761099790919063ffffffff16565b610269565b8282815181106105df576105de611fa0565b5b602002602001018181525050806001019050610595565b508091505092915050565b61060c8383836109aa565b505050565b61062361061c610889565b8383610a3a565b5050565b61064183838360405180602001604052805f815250610ba3565b505050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff16905092915050565b5f6106dd610889565b90508073ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff161415801561072257506107208682610646565b155b156107665780866040517fe237d92200000000000000000000000000000000000000000000000000000000815260040161075d929190611f52565b60405180910390fd5b6107738686868686610c26565b505050505050565b610786838383610d2c565b505050565b5f7f01ffc9a7000000000000000000000000000000000000000000000000000000007bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916149050919050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610864575f6040517f57f447ce00000000000000000000000000000000000000000000000000000000815260040161085b9190611fcd565b60405180910390fd5b5f5f6108708585610dce565b915091506108815f87848487610dfe565b505050505050565b5f33905090565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610900575f6040517f57f447ce0000000000000000000000000000000000000000000000000000000081526004016108f79190611fcd565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff1603610970575f6040517f01a835140000000000000000000000000000000000000000000000000000000081526004016109679190611fcd565b60405180910390fd5b61097d8585858585610dfe565b5050505050565b5f60208202602084010151905092915050565b5f60208202602084010151905092915050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610a1a575f6040517f01a83514000000000000000000000000000000000000000000000000000000008152600401610a119190611fcd565b60405180910390fd5b610a35835f848460405180602001604052805f815250610dfe565b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610aaa575f6040517fced3e100000000000000000000000000000000000000000000000000000000008152600401610aa19190611fcd565b60405180910390fd5b8060015f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff0219169083151502179055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c3183604051610b969190611739565b60405180910390a3505050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610c13575f6040517f57f447ce000000000000000000000000000000000000000000000000000000008152600401610c0a9190611fcd565b60405180910390fd5b610c205f85858585610dfe565b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610c96575f6040517f57f447ce000000000000000000000000000000000000000000000000000000008152600401610c8d9190611fcd565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff1603610d06575f6040517f01a83514000000000000000000000000000000000000000000000000000000008152600401610cfd9190611fcd565b60405180910390fd5b5f5f610d128585610dce565b91509150610d238787848487610dfe565b50505050505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610d9c575f6040517f01a83514000000000000000000000000000000000000000000000000000000008152600401610d939190611fcd565b60405180910390fd5b5f5f610da88484610dce565b91509150610dc7855f848460405180602001604052805f815250610dfe565b5050505050565b60608060405191506001825283602083015260408201905060018152826020820152604081016040529250929050565b610e0a85858585610eaa565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1614610ea3575f610e46610889565b90506001845103610e92575f610e655f8661099790919063ffffffff16565b90505f610e7b5f8661099790919063ffffffff16565b9050610e8b83898985858961123d565b5050610ea1565b610ea08187878787876113ec565b5b505b5050505050565b8051825114610ef457815181516040517f5b059991000000000000000000000000000000000000000000000000000000008152600401610eeb929190611f79565b60405180910390fd5b5f610efd610889565b90505f5f90505b83518110156110fc575f610f21828661099790919063ffffffff16565b90505f610f37838661099790919063ffffffff16565b90505f73ffffffffffffffffffffffffffffffffffffffff168873ffffffffffffffffffffffffffffffffffffffff161461105a575f5f5f8481526020019081526020015f205f8a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205490508181101561100657888183856040517f03dee4c5000000000000000000000000000000000000000000000000000000008152600401610ffd9493929190611fe6565b60405180910390fd5b8181035f5f8581526020019081526020015f205f8b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff16146110ef57805f5f8481526020019081526020015f205f8973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8282546110e79190612056565b925050819055505b5050806001019050610f04565b5060018351036111b7575f61111a5f8561099790919063ffffffff16565b90505f6111305f8561099790919063ffffffff16565b90508573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f6285856040516111a8929190611f79565b60405180910390a45050611236565b8373ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff167f4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb868660405161122d929190612089565b60405180910390a45b5050505050565b5f8473ffffffffffffffffffffffffffffffffffffffff163b11156113e4578373ffffffffffffffffffffffffffffffffffffffff1663f23a6e6187878686866040518663ffffffff1660e01b815260040161129d959493929190612110565b6020604051808303815f875af19250505080156112d857506040513d601f19601f820116820180604052508101906112d5919061217c565b60015b611359573d805f8114611306576040519150601f19603f3d011682016040523d82523d5f602084013e61130b565b606091505b505f81510361135157846040517f57f447ce0000000000000000000000000000000000000000000000000000000081526004016113489190611fcd565b60405180910390fd5b805160208201fd5b63f23a6e6160e01b7bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916817bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916146113e257846040517f57f447ce0000000000000000000000000000000000000000000000000000000081526004016113d99190611fcd565b60405180910390fd5b505b505050505050565b5f8473ffffffffffffffffffffffffffffffffffffffff163b1115611593578373ffffffffffffffffffffffffffffffffffffffff1663bc197c8187878686866040518663ffffffff1660e01b815260040161144c9594939291906121a7565b6020604051808303815f875af192505050801561148757506040513d601f19601f82011682018060405250810190611484919061217c565b60015b611508573d805f81146114b5576040519150601f19603f3d011682016040523d82523d5f602084013e6114ba565b606091505b505f81510361150057846040517f57f447ce0000000000000000000000000000000000000000000000000000000081526004016114f79190611fcd565b60405180910390fd5b805160208201fd5b63bc197c8160e01b7bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916817bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19161461159157846040517f57f447ce0000000000000000000000000000000000000000000000000000000081526004016115889190611fcd565b60405180910390fd5b505b505050505050565b5f604051905090565b5f5ffd5b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6115d5826115ac565b9050919050565b6115e5816115cb565b81146115ef575f5ffd5b50565b5f81359050611600816115dc565b92915050565b5f819050919050565b61161881611606565b8114611622575f5ffd5b50565b5f813590506116338161160f565b92915050565b5f5f6040838503121561164f5761164e6115a4565b5b5f61165c858286016115f2565b925050602061166d85828601611625565b9150509250929050565b61168081611606565b82525050565b5f6020820190506116995f830184611677565b92915050565b5f7fffffffff0000000000000000000000000000000000000000000000000000000082169050919050565b6116d38161169f565b81146116dd575f5ffd5b50565b5f813590506116ee816116ca565b92915050565b5f60208284031215611709576117086115a4565b5b5f611716848285016116e0565b91505092915050565b5f8115159050919050565b6117338161171f565b82525050565b5f60208201905061174c5f83018461172a565b92915050565b5f60208284031215611767576117666115a4565b5b5f61177484828501611625565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f6117bf8261177d565b6117c98185611787565b93506117d9818560208601611797565b6117e2816117a5565b840191505092915050565b5f6020820190508181035f83015261180581846117b5565b905092915050565b5f5f5f60608486031215611824576118236115a4565b5b5f611831868287016115f2565b935050602061184286828701611625565b925050604061185386828701611625565b9150509250925092565b5f5ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b611897826117a5565b810181811067ffffffffffffffff821117156118b6576118b5611861565b5b80604052505050565b5f6118c861159b565b90506118d4828261188e565b919050565b5f67ffffffffffffffff8211156118f3576118f2611861565b5b602082029050602081019050919050565b5f5ffd5b5f61191a611915846118d9565b6118bf565b9050808382526020820190506020840283018581111561193d5761193c611904565b5b835b8181101561196657806119528882611625565b84526020840193505060208101905061193f565b5050509392505050565b5f82601f8301126119845761198361185d565b5b8135611994848260208601611908565b91505092915050565b5f5ffd5b5f67ffffffffffffffff8211156119bb576119ba611861565b5b6119c4826117a5565b9050602081019050919050565b828183375f83830152505050565b5f6119f16119ec846119a1565b6118bf565b905082815260208101848484011115611a0d57611a0c61199d565b5b611a188482856119d1565b509392505050565b5f82601f830112611a3457611a3361185d565b5b8135611a448482602086016119df565b91505092915050565b5f5f5f5f5f60a08688031215611a6657611a656115a4565b5b5f611a73888289016115f2565b9550506020611a84888289016115f2565b945050604086013567ffffffffffffffff811115611aa557611aa46115a8565b5b611ab188828901611970565b935050606086013567ffffffffffffffff811115611ad257611ad16115a8565b5b611ade88828901611970565b925050608086013567ffffffffffffffff811115611aff57611afe6115a8565b5b611b0b88828901611a20565b9150509295509295909350565b5f67ffffffffffffffff821115611b3257611b31611861565b5b602082029050602081019050919050565b5f611b55611b5084611b18565b6118bf565b90508083825260208201905060208402830185811115611b7857611b77611904565b5b835b81811015611ba15780611b8d88826115f2565b845260208401935050602081019050611b7a565b5050509392505050565b5f82601f830112611bbf57611bbe61185d565b5b8135611bcf848260208601611b43565b91505092915050565b5f5f60408385031215611bee57611bed6115a4565b5b5f83013567ffffffffffffffff811115611c0b57611c0a6115a8565b5b611c1785828601611bab565b925050602083013567ffffffffffffffff811115611c3857611c376115a8565b5b611c4485828601611970565b9150509250929050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b611c8081611606565b82525050565b5f611c918383611c77565b60208301905092915050565b5f602082019050919050565b5f611cb382611c4e565b611cbd8185611c58565b9350611cc883611c68565b805f5b83811015611cf8578151611cdf8882611c86565b9750611cea83611c9d565b925050600181019050611ccb565b5085935050505092915050565b5f6020820190508181035f830152611d1d8184611ca9565b905092915050565b5f5f5f60608486031215611d3c57611d3b6115a4565b5b5f611d49868287016115f2565b935050602084013567ffffffffffffffff811115611d6a57611d696115a8565b5b611d7686828701611970565b925050604084013567ffffffffffffffff811115611d9757611d966115a8565b5b611da386828701611970565b9150509250925092565b611db68161171f565b8114611dc0575f5ffd5b50565b5f81359050611dd181611dad565b92915050565b5f5f60408385031215611ded57611dec6115a4565b5b5f611dfa858286016115f2565b9250506020611e0b85828601611dc3565b9150509250929050565b5f5f60408385031215611e2b57611e2a6115a4565b5b5f611e38858286016115f2565b9250506020611e49858286016115f2565b9150509250929050565b5f5f5f5f5f60a08688031215611e6c57611e6b6115a4565b5b5f611e79888289016115f2565b9550506020611e8a888289016115f2565b9450506040611e9b88828901611625565b9350506060611eac88828901611625565b925050608086013567ffffffffffffffff811115611ecd57611ecc6115a8565b5b611ed988828901611a20565b9150509295509295909350565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680611f2a57607f821691505b602082108103611f3d57611f3c611ee6565b5b50919050565b611f4c816115cb565b82525050565b5f604082019050611f655f830185611f43565b611f726020830184611f43565b9392505050565b5f604082019050611f8c5f830185611677565b611f996020830184611677565b9392505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f602082019050611fe05f830184611f43565b92915050565b5f608082019050611ff95f830187611f43565b6120066020830186611677565b6120136040830185611677565b6120206060830184611677565b95945050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61206082611606565b915061206b83611606565b925082820190508082111561208357612082612029565b5b92915050565b5f6040820190508181035f8301526120a18185611ca9565b905081810360208301526120b58184611ca9565b90509392505050565b5f81519050919050565b5f82825260208201905092915050565b5f6120e2826120be565b6120ec81856120c8565b93506120fc818560208601611797565b612105816117a5565b840191505092915050565b5f60a0820190506121235f830188611f43565b6121306020830187611f43565b61213d6040830186611677565b61214a6060830185611677565b818103608083015261215c81846120d8565b90509695505050505050565b5f81519050612176816116ca565b92915050565b5f60208284031215612191576121906115a4565b5b5f61219e84828501612168565b91505092915050565b5f60a0820190506121ba5f830188611f43565b6121c76020830187611f43565b81810360408301526121d98186611ca9565b905081810360608301526121ed8185611ca9565b9050818103608083015261220181846120d8565b9050969550505050505056fea264697066735822122092d466ed5172d8b5e4e6079fa99f5289a55ec7abb2f9d6d7287bc60161a41d4564736f6c63430008230033"

type ERC1155Suite struct {
	suite.Suite

	Enable        bool
	RPCURL        string
	PrivateKeyHex string
	ToHex         string
	MulticallHex  string
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
	contracts []*types.Address
	to        *types.Address
	multicall *types.Address

	erc1155 *ERC1155

	ids    []*big.Int
	amount *big.Int

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestERC1155Suite(t *testing.T) {
	suite.Run(t, &ERC1155Suite{
		Enable:        false,
		RPCURL:        "",
		PrivateKeyHex: "",
		ToHex:         "",
		MulticallHex:  multicall3.DefaultMulticall3Address.String(),
		NativeSymbol:  "",
		TokenCount:    5,
		ByteCode:      erc1155ByteCodeHex,
		Constructor:   abi.NewFunction("constructor", abi.Types{abi.String}, nil),
		Args:          []any{"https://evmkit-fixtures.test/erc1155/{id}.json"},
		Salt:          core.Keccak256([]byte("hello")),
	})
}

func (s *ERC1155Suite) SetupSuite() {
	if !s.Enable {
		s.T().Skip("disabled")
	}

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	s.client = rpc.NewClient(s.RPCURL, rpc.WithRetry(5, rpc.ExponentialBackoff(500*time.Millisecond, 10*time.Second)))
	s.deployer = deploy.NewNicksDeployer(s.client)
	s.erc1155 = New(s.client)

	s.mu = &sync.Mutex{}
	s.totalFee = big.NewInt(0)

	s.ids = []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4)}
	s.amount = big.NewInt(100)

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

	mintBatchFn := abi.NewFunction("mintBatch", abi.Types{abi.Address, abi.Slice(abi.Uint256), abi.Slice(abi.Uint256)}, nil)
	amounts := make([]*big.Int, len(s.ids))
	for i := range amounts {
		amounts[i] = s.amount
	}
	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	s.contracts = make([]*types.Address, s.TokenCount)
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

		s.T().Logf("SetupSuite: deployed contract %d/%d %s (nonce %d, tx %s)", i+1, s.TokenCount, contract, unsigned.AddressNonce(), hash)

		calldata, err := mintBatchFn.EncodeCall(s.from, s.ids, amounts)
		s.Require().NoError(err)

		var unsignedMint core.Transaction
		s.Require().Eventually(func() bool {
			u, err := tx.NewDynamicFeeBuilder().Build(s.from, contract, big.NewInt(0), calldata).Pack(s.ctx, s.client)
			if err != nil || u.Gas() < 40000 {
				return false
			}
			unsignedMint = u
			return true
		}, 10*time.Second, 500*time.Millisecond, "mintBatch gas estimate looked like a no-code call")
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
		s.Require().True(mintReceipt.Status, "mintBatch transaction reverted")

		s.T().Logf("SetupSuite: minted ids %v (amount %s each) of %s to %s (tx %s)", s.ids, s.amount, contract, s.from, mintHash)
	}
}

func (s *ERC1155Suite) TearDownSuite() {
	if s.contracts != nil {
		burnBatchFn := abi.NewFunction("burnBatch", abi.Types{abi.Address, abi.Slice(abi.Uint256), abi.Slice(abi.Uint256)}, nil)

		for _, contract := range s.contracts {
			// Tests move balances between s.from and s.to (safeTransferFrom,
			// safeBatchTransferFrom), so burn whatever each actually holds
			// now rather than assuming s.from still has the original mint.
			for _, holder := range []*types.Address{s.from, s.to} {
				accounts := make([]*types.Address, len(s.ids))
				for i := range accounts {
					accounts[i] = holder
				}
				balances, err := s.erc1155.BalanceOfBatch(s.ctx, contract, accounts, s.ids, rpc.BlockTagLatest)
				s.Require().NoError(err)

				var burnIds, burnAmounts []*big.Int
				for i, bal := range balances {
					if bal.Sign() > 0 {
						burnIds = append(burnIds, s.ids[i])
						burnAmounts = append(burnAmounts, bal)
					}
				}
				if len(burnIds) == 0 {
					continue
				}

				calldata, err := burnBatchFn.EncodeCall(holder, burnIds, burnAmounts)
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
				s.Require().True(burnReceipt.Status, "burnBatch transaction reverted")

				s.T().Logf("TearDownSuite: burned ids %v (amounts %v) of %s from %s (tx %s)", burnIds, burnAmounts, contract, holder, burnHash)
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

		s.T().Logf("ERC1155Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC1155Suite) TestReadMethods() {
	for _, contract := range s.contracts {
		supports, err := s.erc1155.SupportsInterface(s.ctx, contract, interfaceIDERC1155, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(supports)

		uris := make([]string, len(s.ids))
		fromBalances := make([]*big.Int, len(s.ids))
		toBalances := make([]*big.Int, len(s.ids))
		for i, id := range s.ids {
			uris[i], err = s.erc1155.URI(s.ctx, contract, id, rpc.BlockTagLatest)
			s.Require().NoError(err)

			fromBalances[i], err = s.erc1155.BalanceOf(s.ctx, contract, s.from, id, rpc.BlockTagLatest)
			s.Require().NoError(err)
			s.Require().Zero(s.amount.Cmp(fromBalances[i]))

			toBalances[i], err = s.erc1155.BalanceOf(s.ctx, contract, s.to, id, rpc.BlockTagLatest)
			s.Require().NoError(err)
			s.Require().Zero(big.NewInt(0).Cmp(toBalances[i]))
		}

		accounts := make([]*types.Address, len(s.ids))
		for i := range accounts {
			accounts[i] = s.from
		}
		balances, err := s.erc1155.BalanceOfBatch(s.ctx, contract, accounts, s.ids, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Len(balances, len(s.ids))
		for _, bal := range balances {
			s.Require().Zero(s.amount.Cmp(bal))
		}

		isApprovedForAll, err := s.erc1155.IsApprovedForAll(s.ctx, contract, s.from, s.to, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().False(isApprovedForAll)

		s.T().Logf("ERC1155 %s TestReadMethods Result", contract)
		s.T().Logf("\tSupports: %v", supports)
		for i, id := range s.ids {
			s.T().Logf("\tURI(%s): %s", id, uris[i])
			s.T().Logf("\tBalanceOf(id %s)", id)
			s.T().Logf("\t\tFrom(%s): %s", s.from, fromBalances[i])
			s.T().Logf("\t\tTo(%s): %s", s.to, toBalances[i])
		}
		s.T().Logf("\tBalanceOfBatch(from, %v): %v", s.ids, balances)
		s.T().Logf("\tIsApprovedForAll(from, to): %v", isApprovedForAll)
	}
}

func (s *ERC1155Suite) TestSetApprovalForAll() {
	first, err := s.erc1155.CreateSetApprovalForAll(s.ctx, tx.NewDynamicFeeBuilder(), s.contracts[0], s.from, s.to, true)
	s.Require().NoError(err)
	baseNonce := first.AddressNonce()

	unsigned := make([]core.Transaction, len(s.contracts))
	unsigned[0] = first
	for i := 1; i < len(s.contracts); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(baseNonce + uint64(i))

		u, err := s.erc1155.CreateSetApprovalForAll(s.ctx, builder, s.contracts[i], s.from, s.to, true)
		s.Require().NoError(err)
		unsigned[i] = u
	}

	hashes := make([]*types.Hash, len(s.contracts))
	for i, u := range unsigned {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc1155.SetApprovalForAll(s.ctx, u)
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
		s.Require().True(approvals[0].Account.Equal(s.from))
		s.Require().True(approvals[0].Operator.Equal(s.to))
		s.Require().True(approvals[0].Approved)

		var after bool
		s.Require().Eventually(func() bool {
			after, err = s.erc1155.IsApprovedForAll(s.ctx, contract, s.from, s.to, rpc.BlockTagLatest)
			return err == nil && after
		}, 10*time.Second, 500*time.Millisecond, "approval-for-all not visible after setApprovalForAll")

		s.T().Logf("ERC1155 %s TestSetApprovalForAll Result", contract)
		s.T().Logf("\tAccount: %s", s.from)
		s.T().Logf("\tOperator: %s", s.to)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", unsigned[i].Gas())
		s.T().Logf("\tNonce: %d", unsigned[i].AddressNonce())
		s.T().Logf("\tMaxFee: %s", unsigned[i].MaxFee())
		s.T().Logf("\tIsApprovedForAll")
		s.T().Logf("\t\tAfter: %v", after)
	}

	firstRevert, err := s.erc1155.CreateSetApprovalForAll(s.ctx, tx.NewDynamicFeeBuilder(), s.contracts[0], s.from, s.to, false)
	s.Require().NoError(err)
	revertBaseNonce := firstRevert.AddressNonce()

	unsignedRevert := make([]core.Transaction, len(s.contracts))
	unsignedRevert[0] = firstRevert
	for i := 1; i < len(s.contracts); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(revertBaseNonce + uint64(i))

		u, err := s.erc1155.CreateSetApprovalForAll(s.ctx, builder, s.contracts[i], s.from, s.to, false)
		s.Require().NoError(err)
		unsignedRevert[i] = u
	}

	revertHashes := make([]*types.Hash, len(s.contracts))
	for i, u := range unsignedRevert {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc1155.SetApprovalForAll(s.ctx, u)
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
			reverted, err = s.erc1155.IsApprovedForAll(s.ctx, contract, s.from, s.to, rpc.BlockTagLatest)
			return err == nil && !reverted
		}, 10*time.Second, 500*time.Millisecond, "approval-for-all not reverted to false")

		s.T().Logf("ERC1155 %s TestSetApprovalForAll Revert Result", contract)
		s.T().Logf("\tRevertTx: %s", revertHashes[i])
		s.T().Logf("\t\tMaxFee: %s", unsignedRevert[i].MaxFee())
		s.T().Logf("\t\tReverted: %v", reverted)
	}
}

func (s *ERC1155Suite) TestSetApprovalForAllInvalidOperator() {
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		_, err := s.erc1155.CreateSetApprovalForAll(s.ctx, tx.NewDynamicFeeBuilder(), contract, s.from, zero, true)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidOperator)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC1155 %s TestSetApprovalForAllInvalidOperator Result", contract)
		s.T().Logf("\tOperator (zero): %s", zero)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC1155Suite) TestSafeTransferFrom() {
	for _, id := range s.ids[1:3] {
		first, err := s.erc1155.CreateSafeTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.contracts[0], s.from, s.from, s.to, id, s.amount, nil)
		s.Require().NoError(err)
		baseNonce := first.AddressNonce()

		unsigned := make([]core.Transaction, len(s.contracts))
		unsigned[0] = first
		for i := 1; i < len(s.contracts); i++ {
			builder := tx.NewDynamicFeeBuilder()
			builder.Nonce(baseNonce + uint64(i))

			u, err := s.erc1155.CreateSafeTransferFrom(s.ctx, builder, s.contracts[i], s.from, s.from, s.to, id, s.amount, nil)
			s.Require().NoError(err)
			unsigned[i] = u
		}

		hashes := make([]*types.Hash, len(s.contracts))
		for i, u := range unsigned {
			s.Require().NoError(u.Sign(s.signer))
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, u.MaxFee())
			s.mu.Unlock()

			hash, err := s.erc1155.SafeTransferFrom(s.ctx, u)
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
			s.Require().Zero(id.Cmp(transfers[0].Id))
			s.Require().Zero(s.amount.Cmp(transfers[0].Value))

			var toBalance *big.Int
			s.Require().Eventually(func() bool {
				toBalance, err = s.erc1155.BalanceOf(s.ctx, contract, s.to, id, rpc.BlockTagLatest)
				return err == nil && s.amount.Cmp(toBalance) == 0
			}, 10*time.Second, 500*time.Millisecond, "balance not visible after safeTransferFrom")

			fromBalance, err := s.erc1155.BalanceOf(s.ctx, contract, s.from, id, rpc.BlockTagLatest)
			s.Require().NoError(err)
			s.Require().Zero(big.NewInt(0).Cmp(fromBalance))

			s.T().Logf("ERC1155 %s TestSafeTransferFrom Result (id %s)", contract, id)
			s.T().Logf("\tFrom: %s", s.from)
			s.T().Logf("\tTo: %s", s.to)
			s.T().Logf("\tId: %s", id)
			s.T().Logf("\tAmount: %s", s.amount)
			s.T().Logf("\tTx: %s", hashes[i])
			s.T().Logf("\tGas: %d", unsigned[i].Gas())
			s.T().Logf("\tMaxFee: %s", unsigned[i].MaxFee())
			s.T().Logf("\tBalanceOf")
			s.T().Logf("\t\tFrom: %s", fromBalance)
			s.T().Logf("\t\tTo: %s", toBalance)
		}
	}
}

func (s *ERC1155Suite) TestSafeTransferFromInvalidReceiver() {
	id := s.ids[4]
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		_, err := s.erc1155.CreateSafeTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), contract, s.from, s.from, zero, id, s.amount, nil)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidReceiver)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC1155 %s TestSafeTransferFromInvalidReceiver Result", contract)
		s.T().Logf("\tId: %s", id)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC1155Suite) TestSafeBatchTransferFrom() {
	ids := s.ids[3:5]
	amounts := make([]*big.Int, len(ids))
	for i := range amounts {
		amounts[i] = s.amount
	}

	first, err := s.erc1155.CreateSafeBatchTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), s.contracts[0], s.from, s.from, s.to, ids, amounts, nil)
	s.Require().NoError(err)
	baseNonce := first.AddressNonce()

	unsigned := make([]core.Transaction, len(s.contracts))
	unsigned[0] = first
	for i := 1; i < len(s.contracts); i++ {
		builder := tx.NewDynamicFeeBuilder()
		builder.Nonce(baseNonce + uint64(i))

		u, err := s.erc1155.CreateSafeBatchTransferFrom(s.ctx, builder, s.contracts[i], s.from, s.from, s.to, ids, amounts, nil)
		s.Require().NoError(err)
		unsigned[i] = u
	}

	hashes := make([]*types.Hash, len(s.contracts))
	for i, u := range unsigned {
		s.Require().NoError(u.Sign(s.signer))
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, u.MaxFee())
		s.mu.Unlock()

		hash, err := s.erc1155.SafeBatchTransferFrom(s.ctx, u)
		s.Require().NoError(err)
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], 0, 0)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		batches, err := ExtractTransferBatches(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(batches, 1)
		s.Require().True(batches[0].Contract.Equal(contract))
		s.Require().True(batches[0].From.Equal(s.from))
		s.Require().True(batches[0].To.Equal(s.to))
		s.Require().Len(batches[0].Ids, len(ids))
		for j, id := range ids {
			s.Require().Zero(id.Cmp(batches[0].Ids[j]))
			s.Require().Zero(amounts[j].Cmp(batches[0].Values[j]))
		}

		toAccounts := make([]*types.Address, len(ids))
		for j := range toAccounts {
			toAccounts[j] = s.to
		}
		var toBalances []*big.Int
		s.Require().Eventually(func() bool {
			toBalances, err = s.erc1155.BalanceOfBatch(s.ctx, contract, toAccounts, ids, rpc.BlockTagLatest)
			if err != nil {
				return false
			}
			for j, bal := range toBalances {
				if amounts[j].Cmp(bal) != 0 {
					return false
				}
			}
			return true
		}, 10*time.Second, 500*time.Millisecond, "balances not visible after safeBatchTransferFrom")

		fromAccounts := make([]*types.Address, len(ids))
		for j := range fromAccounts {
			fromAccounts[j] = s.from
		}
		fromBalances, err := s.erc1155.BalanceOfBatch(s.ctx, contract, fromAccounts, ids, rpc.BlockTagLatest)
		s.Require().NoError(err)
		for _, bal := range fromBalances {
			s.Require().Zero(big.NewInt(0).Cmp(bal))
		}

		s.T().Logf("ERC1155 %s TestSafeBatchTransferFrom Result", contract)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tIds: %v", ids)
		s.T().Logf("\tAmounts: %v", amounts)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", unsigned[i].Gas())
		s.T().Logf("\tMaxFee: %s", unsigned[i].MaxFee())
		s.T().Logf("\tBalanceOfBatch")
		s.T().Logf("\t\tFrom: %v", fromBalances)
		s.T().Logf("\t\tTo: %v", toBalances)
	}
}

func (s *ERC1155Suite) TestSafeBatchTransferFromInvalidReceiver() {
	ids := s.ids[3:5]
	amounts := make([]*big.Int, len(ids))
	for i := range amounts {
		amounts[i] = s.amount
	}
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		_, err := s.erc1155.CreateSafeBatchTransferFrom(s.ctx, tx.NewDynamicFeeBuilder(), contract, s.from, s.from, zero, ids, amounts, nil)
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidReceiver)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC1155 %s TestSafeBatchTransferFromInvalidReceiver Result", contract)
		s.T().Logf("\tIds: %v", ids)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC1155Suite) TestIdBalances() {
	for _, contract := range s.contracts {
		balances, err := s.erc1155.IdBalances(s.ctx, contract, s.from, s.ids, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Len(balances, len(s.ids))

		for i, id := range s.ids {
			want, err := s.erc1155.BalanceOf(s.ctx, contract, s.from, id, rpc.BlockTagLatest)
			s.Require().NoError(err)
			s.Require().Zero(id.Cmp(balances[i].Id))
			s.Require().Zero(want.Cmp(balances[i].Amount))
		}

		s.T().Logf("ERC1155 %s TestIdBalances Result", contract)
		for _, b := range balances {
			s.T().Logf("\tid %s: %s", b.Id, b.Amount)
		}
	}
}

func (s *ERC1155Suite) TestIdBalancePairs() {
	for _, contract := range s.contracts {
		accounts := make([]*types.Address, len(s.ids))
		for i := range accounts {
			if i%2 == 0 {
				accounts[i] = s.from
			} else {
				accounts[i] = s.to
			}
		}

		pairs, err := s.erc1155.IdBalancePairs(s.ctx, contract, accounts, s.ids, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().Len(pairs, len(s.ids))

		for i, id := range s.ids {
			want, err := s.erc1155.BalanceOf(s.ctx, contract, accounts[i], id, rpc.BlockTagLatest)
			s.Require().NoError(err)
			s.Require().True(pairs[i].Account.Equal(accounts[i]))
			s.Require().Zero(id.Cmp(pairs[i].Id))
			s.Require().Zero(want.Cmp(pairs[i].Amount))
		}

		s.T().Logf("ERC1155 %s TestIdBalancePairs Result", contract)
		for _, p := range pairs {
			s.T().Logf("\taccount %s id %s: %s", p.Account, p.Id, p.Amount)
		}
	}
}

func (s *ERC1155Suite) TestAccountIdBalances() {
	id := s.ids[0]

	aib, err := s.erc1155.AccountIdBalances(s.ctx, s.multicall, s.from, id, s.contracts, false, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().True(aib.Account.Equal(s.from))
	s.Require().Zero(id.Cmp(aib.Id))
	s.Require().Len(aib.Contracts, len(s.contracts))

	for i, contract := range s.contracts {
		want, err := s.erc1155.BalanceOf(s.ctx, contract, s.from, id, rpc.BlockTagLatest)
		s.Require().NoError(err)
		s.Require().True(aib.Contracts[i].Contract.Equal(contract))
		s.Require().Zero(want.Cmp(aib.Contracts[i].Amount))
	}

	s.T().Logf("ERC1155 TestAccountIdBalances Result (account %s, id %s)", s.from, id)
	for _, cb := range aib.Contracts {
		s.T().Logf("\t%s: %s", cb.Contract, cb.Amount)
	}
}

func (s *ERC1155Suite) TestAccountIdBalancesAllowFailure() {
	id := s.ids[0]

	aib, err := s.erc1155.AccountIdBalances(s.ctx, s.multicall, s.from, id, []*types.Address{s.multicall}, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().Len(aib.Contracts, 1)
	s.Require().Nil(aib.Contracts[0].Amount)

	s.T().Logf("ERC1155 TestAccountIdBalancesAllowFailure Result")
	s.T().Logf("\tAmount: %v", aib.Contracts[0].Amount)
}

func (s *ERC1155Suite) TestApprovalForAllWithBalance() {
	id := s.ids[0]

	for _, contract := range s.contracts {
		afawb, err := s.erc1155.ApprovalForAllWithBalance(s.ctx, s.multicall, contract, s.from, s.to, id, false, rpc.BlockTagLatest)
		s.Require().NoError(err)

		approved, err := s.erc1155.IsApprovedForAll(s.ctx, contract, s.from, s.to, rpc.BlockTagLatest)
		s.Require().NoError(err)
		fromBal, err := s.erc1155.BalanceOf(s.ctx, contract, s.from, id, rpc.BlockTagLatest)
		s.Require().NoError(err)
		toBal, err := s.erc1155.BalanceOf(s.ctx, contract, s.to, id, rpc.BlockTagLatest)
		s.Require().NoError(err)

		s.Require().True(afawb.Account.Equal(s.from))
		s.Require().True(afawb.Operator.Equal(s.to))
		s.Require().Zero(id.Cmp(afawb.Id))
		s.Require().Equal(approved, afawb.Approved)
		s.Require().Zero(fromBal.Cmp(afawb.AccountBalance))
		s.Require().Zero(toBal.Cmp(afawb.OperatorBalance))

		s.T().Logf("ERC1155 %s TestApprovalForAllWithBalance Result", contract)
		s.T().Logf("\tApproved: %v", afawb.Approved)
		s.T().Logf("\tAccountBalance: %s", afawb.AccountBalance)
		s.T().Logf("\tOperatorBalance: %s", afawb.OperatorBalance)
	}
}

func (s *ERC1155Suite) TestApprovalForAllWithBalanceAllowFailure() {
	id := s.ids[0]

	afawb, err := s.erc1155.ApprovalForAllWithBalance(s.ctx, s.multicall, s.multicall, s.from, s.to, id, true, rpc.BlockTagLatest)
	s.Require().NoError(err)
	s.Require().False(afawb.Approved)
	s.Require().Nil(afawb.AccountBalance)
	s.Require().Nil(afawb.OperatorBalance)

	s.T().Logf("ERC1155 %s TestApprovalForAllWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tApproved: %v", afawb.Approved)
	s.T().Logf("\tAccountBalance: %v", afawb.AccountBalance)
	s.T().Logf("\tOperatorBalance: %v", afawb.OperatorBalance)
}
