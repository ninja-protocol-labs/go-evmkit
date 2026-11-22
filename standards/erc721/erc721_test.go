package erc721

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/standards/deploy"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/internal/testenv"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/standards/erc165"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var interfaceIDERC721 = [4]byte{0x80, 0xac, 0x58, 0xcd}

type ERC721Suite struct {
	suite.Suite

	RPCURL              string
	PrivateKeyHex       string
	MulticallHex        string
	NicksFactoryHex     string
	NativeSymbol        string
	NativeDecimals      int
	Timeout             time.Duration
	ReceiptPollInterval time.Duration
	ReceiptTimeout      time.Duration
	ToHex               string
	NFTCount            int
	ByteCode            string
	Constructor         *abi.Function
	Args                []any
	Salt                *types.Hash

	ctx    context.Context
	cancel context.CancelFunc

	client    rpc.Client
	signer    core.Signer
	deployer  *deploy.NicksDeployer
	from      *types.Address
	contracts []*types.Address
	to        *types.Address
	multicall *types.Address
	factory   *types.Address

	erc721 *ERC721

	mu       *sync.Mutex
	totalFee *big.Int
}

func TestERC721Suite(t *testing.T) {
	c := testenv.MustLoad(t, testenv.Testnet, testenv.UnknownNetwork)

	suite.Run(t, &ERC721Suite{
		RPCURL:              c.RPCURL,
		PrivateKeyHex:       c.ActorPrivateKeyHex,
		MulticallHex:        c.Addresses.Standard.Multicall,
		NicksFactoryHex:     c.Addresses.Standard.NicksFactory,
		NativeSymbol:        c.Constant.NativeSymbol,
		NativeDecimals:      c.Constant.NativeDecimals,
		Timeout:             10 * time.Minute,
		ReceiptPollInterval: c.Constant.ReceiptPollInterval,
		ReceiptTimeout:      c.Constant.ReceiptTimeout,
		ToHex:               "",
		NFTCount:            5,
		ByteCode:            testenv.ERC721ByteCodeHex,
		Constructor:         abi.NewFunction("constructor", abi.NewTypes(abi.String, abi.String), nil),
		Args:                []any{"evmkit test NFT", "EVKNFT"},
		Salt:                core.Keccak256([]byte("hello")),
	})
}

func (s *ERC721Suite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

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

	s.factory, err = types.NewAddressFromHex(s.NicksFactoryHex)
	s.Require().NoError(err)

	bytecode, err := encoding.Hex.Decode(s.ByteCode)
	s.Require().NoError(err)

	mintFn := abi.NewFunction("mint", abi.NewTypes(abi.Address, abi.Uint256), nil)
	runtime := fmt.Appendf(nil, "%s-%d", s.Salt, time.Now().UnixNano())

	s.contracts = make([]*types.Address, s.NFTCount)
	for i := 0; i < s.NFTCount; i++ {
		var salt [32]byte
		copy(salt[:], core.Keccak256(fmt.Appendf(nil, "%s-%d", runtime, i)).Bytes())

		deployment := deploy.NewDeterministicContractDeployment(
			deploy.NewContractDeployment(bytecode, s.Constructor.Inputs, s.Args),
			salt,
		)

		deployTx, hash, err := s.deployer.Deploy(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), deploy.NewCreate2Request(s.factory, deployment))
		s.Require().NoError(err)

		s.mu.Lock()
		s.totalFee.Add(s.totalFee, deployTx.MaxFee())
		s.mu.Unlock()

		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hash, s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "deploy transaction reverted")

		contract, err := s.deployer.DeployAddress(deploy.NewCreate2Request(s.factory, deployment))
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

		s.T().Logf("SetupSuite: deployed contract %d/%d %s (nonce %d, tx %s)", i+1, s.NFTCount, contract, deployTx.AddressNonce(), hash)

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

			mintReceipt, err := tx.WaitForReceipt(s.ctx, s.client, mintHash, s.ReceiptPollInterval, s.ReceiptTimeout)
			s.Require().NoError(err)
			s.Require().True(mintReceipt.Status, "mint transaction reverted")

			s.T().Logf("SetupSuite: minted tokenId %s of %s to %s (tx %s)", tokenId, contract, s.from, mintHash)
		}
	}
}

func (s *ERC721Suite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(s.T().Context(), s.Timeout)
	defer cancel()

	if s.contracts != nil {
		burnFn := abi.NewFunction("burn", abi.NewTypes(abi.Uint256), nil)

		for _, contract := range s.contracts {
			for _, tokenId := range []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)} {
				calldata, err := burnFn.EncodeCall(tokenId)
				s.Require().NoError(err)

				unsignedBurn, err := tx.NewDynamicFeeBuilder().Build(s.from, contract, big.NewInt(0), calldata).Pack(ctx, s.client)
				s.Require().NoError(err)
				s.Require().NoError(unsignedBurn.Sign(s.signer))
				s.mu.Lock()
				s.totalFee.Add(s.totalFee, unsignedBurn.MaxFee())
				s.mu.Unlock()

				raw, err := unsignedBurn.EncodeRLP()
				s.Require().NoError(err)
				burnHash, err := tx.Broadcast(ctx, s.client, raw)
				s.Require().NoError(err)

				burnReceipt, err := tx.WaitForReceipt(ctx, s.client, burnHash, s.ReceiptPollInterval, s.ReceiptTimeout)
				s.Require().NoError(err)
				s.Require().True(burnReceipt.Status, "burn transaction reverted")

				s.T().Logf("TearDownSuite: burned tokenId %s of %s (tx %s)", tokenId, contract, burnHash)
			}
		}
	}

	if s.totalFee != nil {
		base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.NativeDecimals)), nil)
		whole, frac := new(big.Int), new(big.Int)
		whole.DivMod(s.totalFee, base, frac)

		fracStr := strings.TrimRight(fmt.Sprintf("%0*s", s.NativeDecimals, frac.String()), "0")
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

func (s *ERC721Suite) Test01AddressBalances() {
	ab, err := s.erc721.AddressBalances(s.ctx, nil, NewAddressBalancesRequest(s.multicall, s.from, s.contracts, false))
	s.Require().NoError(err)
	s.Require().True(ab.Address.Equal(s.from))
	s.Require().Len(ab.Tokens, len(s.contracts))

	s.T().Logf("ERC721 %s Test01AddressBalances Result", s.from)
	for i, contract := range s.contracts {
		balance, err := s.erc721.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.from))
		s.Require().NoError(err)

		s.Require().True(ab.Tokens[i].Token.Equal(contract))
		s.Require().Zero(balance.Cmp(ab.Tokens[i].Amount))
		s.T().Logf("\t%s: %s", contract, ab.Tokens[i].Amount)
	}
}

func (s *ERC721Suite) Test02AllowanceWithTokenID() {
	tokenId := big.NewInt(0)

	for _, contract := range s.contracts {
		awt, err := s.erc721.AllowanceWithTokenID(s.ctx, nil, NewAllowanceWithTokenIDRequest(s.multicall, contract, tokenId, false))
		s.Require().NoError(err)

		owner, err := s.erc721.OwnerOf(s.ctx, nil, NewOwnerOfRequest(contract, tokenId))
		s.Require().NoError(err)
		operator, err := s.erc721.GetApproved(s.ctx, nil, NewGetApprovedRequest(contract, tokenId))
		s.Require().NoError(err)

		s.Require().Zero(tokenId.Cmp(awt.TokenId))
		s.Require().True(owner.Equal(awt.Owner))
		s.Require().True(operator.Equal(awt.Operator))

		s.T().Logf("ERC721 %s Test02AllowanceWithTokenID Result", contract)
		s.T().Logf("\tTokenId: %s", awt.TokenId)
		s.T().Logf("\tOwner: %s", awt.Owner)
		s.T().Logf("\tOperator: %s", awt.Operator)
	}
}

func (s *ERC721Suite) Test03AllowanceWithTokenIDAllowFailure() {
	tokenId := big.NewInt(0)

	awt, err := s.erc721.AllowanceWithTokenID(s.ctx, nil, NewAllowanceWithTokenIDRequest(s.multicall, s.multicall, tokenId, true))
	s.Require().NoError(err)
	s.Require().Nil(awt.Owner)
	s.Require().Nil(awt.Operator)

	s.T().Logf("ERC721 %s Test03AllowanceWithTokenIDAllowFailure Result", s.multicall)
	s.T().Logf("\tTokenId: %s", awt.TokenId)
	s.T().Logf("\tOwner: %v", awt.Owner)
	s.T().Logf("\tOperator: %v", awt.Operator)
}

func (s *ERC721Suite) Test04ApprovalForAllWithBalance() {
	for _, contract := range s.contracts {
		afawb, err := s.erc721.ApprovalForAllWithBalance(s.ctx, nil, NewApprovalForAllWithBalanceRequest(s.multicall, contract, s.from, s.to, false))
		s.Require().NoError(err)

		isApprovedForAll, err := s.erc721.IsApprovedForAll(s.ctx, nil, NewIsApprovedForAllRequest(contract, s.from, s.to))
		s.Require().NoError(err)
		ownerBalance, err := s.erc721.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.from))
		s.Require().NoError(err)
		operatorBalance, err := s.erc721.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.to))
		s.Require().NoError(err)

		s.Require().True(afawb.Owner.Equal(s.from))
		s.Require().True(afawb.Operator.Equal(s.to))
		s.Require().Equal(isApprovedForAll, afawb.Approved)
		s.Require().Zero(ownerBalance.Cmp(afawb.OwnerBalance))
		s.Require().Zero(operatorBalance.Cmp(afawb.OperatorBalance))

		s.T().Logf("ERC721 %s Test04ApprovalForAllWithBalance Result", contract)
		s.T().Logf("\tOwner: %s", afawb.Owner)
		s.T().Logf("\tOperator: %s", afawb.Operator)
		s.T().Logf("\tApproved: %v", afawb.Approved)
		s.T().Logf("\tOwnerBalance: %s", afawb.OwnerBalance)
		s.T().Logf("\tOperatorBalance: %s", afawb.OperatorBalance)
	}
}

func (s *ERC721Suite) Test05ApprovalForAllWithBalanceAllowFailure() {
	afawb, err := s.erc721.ApprovalForAllWithBalance(s.ctx, nil, NewApprovalForAllWithBalanceRequest(s.multicall, s.multicall, s.from, s.to, true))
	s.Require().NoError(err)
	s.Require().False(afawb.Approved)
	s.Require().Nil(afawb.OwnerBalance)
	s.Require().Nil(afawb.OperatorBalance)

	s.T().Logf("ERC721 %s Test05ApprovalForAllWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tApproved: %v", afawb.Approved)
	s.T().Logf("\tOwnerBalance: %v", afawb.OwnerBalance)
	s.T().Logf("\tOperatorBalance: %v", afawb.OperatorBalance)
}

func (s *ERC721Suite) Test06Approve() {
	tokenId := big.NewInt(0)

	txs := make([]core.Transaction, len(s.contracts))
	hashes := make([]*types.Hash, len(s.contracts))
	for i, contract := range s.contracts {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(txs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc721.Approve(s.ctx, bind.NewTransactOptions(s.signer, builder), NewApproveRequest(contract, s.to, tokenId))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		txs[i] = signed
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], s.ReceiptPollInterval, s.ReceiptTimeout)
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
			after, err = s.erc721.GetApproved(s.ctx, nil, NewGetApprovedRequest(contract, tokenId))
			return err == nil && after.Equal(s.to)
		}, 10*time.Second, 500*time.Millisecond, "approval not visible after approve")

		s.T().Logf("ERC721 %s Test06Approve Result", contract)
		s.T().Logf("\tOwner: %s", s.from)
		s.T().Logf("\tApproved: %s", s.to)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", txs[i].Gas())
		s.T().Logf("\tNonce: %d", txs[i].AddressNonce())
		s.T().Logf("\tMaxFee: %s", txs[i].MaxFee())
		s.T().Logf("\tGetApproved")
		s.T().Logf("\t\tAfter: %s", after)
	}
}

func (s *ERC721Suite) Test07ApproveInvalidApprover() {
	tokenId := big.NewInt(0)

	stranger, err := types.GeneratePrivateKey()
	s.Require().NoError(err)
	strangerAddr := core.PubkeyToAddress(stranger.PublicKey())

	for _, contract := range s.contracts {
		signed, hash, err := s.erc721.Approve(s.ctx, bind.NewTransactOptions(stranger, tx.NewDynamicFeeBuilder()), NewApproveRequest(contract, s.from, tokenId))
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidApprover)
		s.Require().Nil(signed)
		s.Require().Nil(hash)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC721 %s Test07ApproveInvalidApprover Result", contract)
		s.T().Logf("\tApprover (unauthorized): %s", strangerAddr)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC721Suite) Test08Metadata() {
	for _, contract := range s.contracts {
		m, err := s.erc721.Metadata(s.ctx, nil, NewMetadataRequest(s.multicall, contract, false))
		s.Require().NoError(err)

		name, err := s.erc721.Name(s.ctx, nil, NewNameRequest(contract))
		s.Require().NoError(err)
		symbol, err := s.erc721.Symbol(s.ctx, nil, NewSymbolRequest(contract))
		s.Require().NoError(err)

		s.Require().Equal(name, m.Name)
		s.Require().Equal(symbol, m.Symbol)

		s.T().Logf("ERC721 %s Test08Metadata Result", contract)
		s.T().Logf("\tName: %s", m.Name)
		s.T().Logf("\tSymbol: %s", m.Symbol)
	}
}

func (s *ERC721Suite) Test09MetadataAllowFailure() {
	m, err := s.erc721.Metadata(s.ctx, nil, NewMetadataRequest(s.multicall, s.multicall, true))
	s.Require().NoError(err)
	s.Require().Empty(m.Name)
	s.Require().Empty(m.Symbol)

	s.T().Logf("ERC721 %s Test09MetadataAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", m.Name)
	s.T().Logf("\tSymbol: %q", m.Symbol)
}

func (s *ERC721Suite) Test10MetadataWithTokenID() {
	tokenId := big.NewInt(0)

	for _, contract := range s.contracts {
		mt, err := s.erc721.MetadataWithTokenID(s.ctx, nil, NewMetadataWithTokenIDRequest(s.multicall, contract, tokenId, false))
		s.Require().NoError(err)

		name, err := s.erc721.Name(s.ctx, nil, NewNameRequest(contract))
		s.Require().NoError(err)
		symbol, err := s.erc721.Symbol(s.ctx, nil, NewSymbolRequest(contract))
		s.Require().NoError(err)
		tokenURI, err := s.erc721.TokenURI(s.ctx, nil, NewTokenURIRequest(contract, tokenId))
		s.Require().NoError(err)
		owner, err := s.erc721.OwnerOf(s.ctx, nil, NewOwnerOfRequest(contract, tokenId))
		s.Require().NoError(err)

		s.Require().Equal(name, mt.Name)
		s.Require().Equal(symbol, mt.Symbol)
		s.Require().Zero(tokenId.Cmp(mt.TokenId))
		s.Require().Equal(tokenURI, mt.TokenURI)
		s.Require().True(owner.Equal(mt.Owner))

		s.T().Logf("ERC721 %s Test10MetadataWithTokenID Result", contract)
		s.T().Logf("\tName: %s", mt.Name)
		s.T().Logf("\tSymbol: %s", mt.Symbol)
		s.T().Logf("\tTokenId: %s", mt.TokenId)
		s.T().Logf("\tTokenURI: %s", mt.TokenURI)
		s.T().Logf("\tOwner: %s", mt.Owner)
	}
}

func (s *ERC721Suite) Test11MetadataWithTokenIDAllowFailure() {
	tokenId := big.NewInt(0)

	mt, err := s.erc721.MetadataWithTokenID(s.ctx, nil, NewMetadataWithTokenIDRequest(s.multicall, s.multicall, tokenId, true))
	s.Require().NoError(err)
	s.Require().Empty(mt.Name)
	s.Require().Empty(mt.Symbol)
	s.Require().Empty(mt.TokenURI)
	s.Require().Nil(mt.Owner)

	s.T().Logf("ERC721 %s Test11MetadataWithTokenIDAllowFailure Result", s.multicall)
	s.T().Logf("\tName: %q", mt.Name)
	s.T().Logf("\tSymbol: %q", mt.Symbol)
	s.T().Logf("\tTokenURI: %q", mt.TokenURI)
	s.T().Logf("\tOwner: %v", mt.Owner)
}

func (s *ERC721Suite) Test12OwnerPairs() {
	tokenIds := make([]*big.Int, len(s.contracts))
	for i := range tokenIds {
		tokenIds[i] = big.NewInt(0)
	}

	pairs, err := s.erc721.OwnerPairs(s.ctx, nil, NewOwnerPairsRequest(s.multicall, s.contracts, tokenIds, false))
	s.Require().NoError(err)
	s.Require().Len(pairs, len(s.contracts))

	s.T().Logf("ERC721 Test12OwnerPairs Result")
	for i, contract := range s.contracts {
		owner, err := s.erc721.OwnerOf(s.ctx, nil, NewOwnerOfRequest(contract, tokenIds[i]))
		s.Require().NoError(err)

		s.Require().True(pairs[i].Contract.Equal(contract))
		s.Require().Zero(tokenIds[i].Cmp(pairs[i].TokenId))
		s.Require().True(pairs[i].Owner.Equal(owner))
		s.T().Logf("\t%s / %s: %s", pairs[i].Contract, pairs[i].TokenId, pairs[i].Owner)
	}
}

func (s *ERC721Suite) Test13ReadMethods() {
	tokenId := big.NewInt(0)

	for _, contract := range s.contracts {
		supports, err := s.erc721.SupportsInterface(s.ctx, nil, erc165.NewSupportsInterfaceRequest(contract, interfaceIDERC721))
		s.Require().NoError(err)
		s.Require().True(supports)

		name, err := s.erc721.Name(s.ctx, nil, NewNameRequest(contract))
		s.Require().NoError(err)

		symbol, err := s.erc721.Symbol(s.ctx, nil, NewSymbolRequest(contract))
		s.Require().NoError(err)

		tokenURI, err := s.erc721.TokenURI(s.ctx, nil, NewTokenURIRequest(contract, tokenId))
		s.Require().NoError(err)

		fromBalance, err := s.erc721.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.from))
		s.Require().NoError(err)

		toBalance, err := s.erc721.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.to))
		s.Require().NoError(err)

		owner, err := s.erc721.OwnerOf(s.ctx, nil, NewOwnerOfRequest(contract, tokenId))
		s.Require().NoError(err)
		s.Require().True(owner.Equal(s.from))

		approved, err := s.erc721.GetApproved(s.ctx, nil, NewGetApprovedRequest(contract, tokenId))
		s.Require().NoError(err)
		s.Require().True(approved.Equal(s.to))

		isApprovedForAll, err := s.erc721.IsApprovedForAll(s.ctx, nil, NewIsApprovedForAllRequest(contract, s.from, s.to))
		s.Require().NoError(err)

		s.T().Logf("ERC721 %s Test13ReadMethods Result", contract)
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

func (s *ERC721Suite) Test14SafeTransferFrom() {
	tokenId := big.NewInt(0)

	txs := make([]core.Transaction, len(s.contracts))
	hashes := make([]*types.Hash, len(s.contracts))
	for i, contract := range s.contracts {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(txs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc721.SafeTransferFrom(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSafeTransferFromRequest(contract, s.from, s.to, tokenId))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		txs[i] = signed
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], s.ReceiptPollInterval, s.ReceiptTimeout)
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
			owner, err = s.erc721.OwnerOf(s.ctx, nil, NewOwnerOfRequest(contract, tokenId))
			return err == nil && owner.Equal(s.to)
		}, 10*time.Second, 500*time.Millisecond, "owner not visible after safeTransferFrom")

		s.T().Logf("ERC721 %s Test14SafeTransferFrom Result", contract)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", txs[i].Gas())
		s.T().Logf("\tMaxFee: %s", txs[i].MaxFee())
		s.T().Logf("\tOwnerOf: %s", owner)
	}
}

func (s *ERC721Suite) Test15SafeTransferFromInvalidReceiver() {
	tokenId := big.NewInt(1)
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		signed, hash, err := s.erc721.SafeTransferFrom(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSafeTransferFromRequest(contract, s.from, zero, tokenId))
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidReceiver)
		s.Require().Nil(signed)
		s.Require().Nil(hash)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC721 %s Test15SafeTransferFromInvalidReceiver Result", contract)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC721Suite) Test16SafeTransferFromWithData() {
	tokenId := big.NewInt(1)
	var callData []byte

	txs := make([]core.Transaction, len(s.contracts))
	hashes := make([]*types.Hash, len(s.contracts))
	for i, contract := range s.contracts {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(txs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc721.SafeTransferFromWithData(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSafeTransferFromWithDataRequest(contract, s.from, s.to, tokenId, callData))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		txs[i] = signed
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(receipt.Status, "transaction reverted")

		transfers, err := ExtractTransfers(receipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(transfers, 1)
		s.Require().True(transfers[0].Contract.Equal(contract))
		s.Require().Zero(tokenId.Cmp(transfers[0].TokenId))

		s.T().Logf("ERC721 %s Test16SafeTransferFromWithData Result", contract)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", txs[i].Gas())
		s.T().Logf("\tMaxFee: %s", txs[i].MaxFee())
	}
}

func (s *ERC721Suite) Test17SetApprovalForAll() {
	txs := make([]core.Transaction, len(s.contracts))
	hashes := make([]*types.Hash, len(s.contracts))
	for i, contract := range s.contracts {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(txs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc721.SetApprovalForAll(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSetApprovalForAllRequest(contract, s.to, true))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		txs[i] = signed
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], s.ReceiptPollInterval, s.ReceiptTimeout)
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

	revertTxs := make([]core.Transaction, len(s.contracts))
	revertHashes := make([]*types.Hash, len(s.contracts))
	for i, contract := range s.contracts {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(revertTxs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc721.SetApprovalForAll(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSetApprovalForAllRequest(contract, s.to, false))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		revertTxs[i] = signed
		revertHashes[i] = hash
	}

	for i, contract := range s.contracts {
		revertReceipt, err := tx.WaitForReceipt(s.ctx, s.client, revertHashes[i], s.ReceiptPollInterval, s.ReceiptTimeout)
		s.Require().NoError(err)
		s.Require().True(revertReceipt.Status, "revert transaction reverted")

		revertApprovals, err := ExtractApprovalForAll(revertReceipt.Logs)
		s.Require().NoError(err)
		s.Require().Len(revertApprovals, 1)
		s.Require().False(revertApprovals[0].Approved)

		var reverted bool
		s.Require().Eventually(func() bool {
			reverted, err = s.erc721.IsApprovedForAll(s.ctx, nil, NewIsApprovedForAllRequest(contract, s.from, s.to))
			return err == nil && !reverted
		}, 10*time.Second, 500*time.Millisecond, "approval-for-all not reverted to false")

		s.T().Logf("ERC721 %s Test17SetApprovalForAll Result", contract)
		s.T().Logf("\tOwner: %s", s.from)
		s.T().Logf("\tOperator: %s", s.to)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", txs[i].Gas())
		s.T().Logf("\tMaxFee: %s", txs[i].MaxFee())
		s.T().Logf("\tRevertTx: %s", revertHashes[i])
		s.T().Logf("\t\tMaxFee: %s", revertTxs[i].MaxFee())
		s.T().Logf("\t\tReverted: %v", reverted)
	}
}

func (s *ERC721Suite) Test18SetApprovalForAllInvalidOperator() {
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		signed, hash, err := s.erc721.SetApprovalForAll(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSetApprovalForAllRequest(contract, zero, true))
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidOperator)
		s.Require().Nil(signed)
		s.Require().Nil(hash)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC721 %s Test18SetApprovalForAllInvalidOperator Result", contract)
		s.T().Logf("\tOperator (zero): %s", zero)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC721Suite) Test19TokenBalances() {
	addresses := []*types.Address{s.from, s.to}

	for _, contract := range s.contracts {
		tb, err := s.erc721.TokenBalances(s.ctx, nil, NewTokenBalancesRequest(s.multicall, contract, addresses, false))
		s.Require().NoError(err)
		s.Require().True(tb.Contract.Equal(contract))
		s.Require().Len(tb.Addresses, len(addresses))

		s.T().Logf("ERC721 %s Test19TokenBalances Result", contract)
		for i, addr := range addresses {
			balance, err := s.erc721.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, addr))
			s.Require().NoError(err)

			s.Require().True(tb.Addresses[i].Address.Equal(addr))
			s.Require().Zero(balance.Cmp(tb.Addresses[i].Amount))
			s.T().Logf("\t%s: %s", addr, tb.Addresses[i].Amount)
		}
	}
}

func (s *ERC721Suite) Test20TokenOperators() {
	tokenIds := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)}

	for _, contract := range s.contracts {
		tops, err := s.erc721.TokenOperators(s.ctx, nil, NewTokenOperatorsRequest(s.multicall, contract, tokenIds, false))
		s.Require().NoError(err)
		s.Require().True(tops.Contract.Equal(contract))
		s.Require().Len(tops.Tokens, len(tokenIds))

		s.T().Logf("ERC721 %s Test20TokenOperators Result", contract)
		for i, tokenId := range tokenIds {
			operator, err := s.erc721.GetApproved(s.ctx, nil, NewGetApprovedRequest(contract, tokenId))
			s.Require().NoError(err)

			s.Require().Zero(tokenId.Cmp(tops.Tokens[i].TokenId))
			s.Require().True(tops.Tokens[i].Operator.Equal(operator))
			s.T().Logf("\t%s: %s", tokenId, tops.Tokens[i].Operator)
		}
	}
}

func (s *ERC721Suite) Test21TokenOwners() {
	tokenIds := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)}

	for _, contract := range s.contracts {
		tos, err := s.erc721.TokenOwners(s.ctx, nil, NewTokenOwnersRequest(s.multicall, contract, tokenIds, false))
		s.Require().NoError(err)
		s.Require().True(tos.Contract.Equal(contract))
		s.Require().Len(tos.Tokens, len(tokenIds))

		s.T().Logf("ERC721 %s Test21TokenOwners Result", contract)
		for i, tokenId := range tokenIds {
			owner, err := s.erc721.OwnerOf(s.ctx, nil, NewOwnerOfRequest(contract, tokenId))
			s.Require().NoError(err)

			s.Require().Zero(tokenId.Cmp(tos.Tokens[i].TokenId))
			s.Require().True(tos.Tokens[i].Owner.Equal(owner))
			s.T().Logf("\t%s: %s", tokenId, tos.Tokens[i].Owner)
		}
	}
}

func (s *ERC721Suite) Test22TransferFrom() {
	tokenId := big.NewInt(2)

	txs := make([]core.Transaction, len(s.contracts))
	hashes := make([]*types.Hash, len(s.contracts))
	for i, contract := range s.contracts {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(txs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc721.TransferFrom(s.ctx, bind.NewTransactOptions(s.signer, builder), NewTransferFromRequest(contract, s.from, s.to, tokenId))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		txs[i] = signed
		hashes[i] = hash
	}

	for i, contract := range s.contracts {
		receipt, err := tx.WaitForReceipt(s.ctx, s.client, hashes[i], s.ReceiptPollInterval, s.ReceiptTimeout)
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
			owner, err = s.erc721.OwnerOf(s.ctx, nil, NewOwnerOfRequest(contract, tokenId))
			return err == nil && owner.Equal(s.to)
		}, 10*time.Second, 500*time.Millisecond, "owner not visible after transferFrom")

		s.T().Logf("ERC721 %s Test22TransferFrom Result", contract)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", txs[i].Gas())
		s.T().Logf("\tMaxFee: %s", txs[i].MaxFee())
		s.T().Logf("\tOwnerOf: %s", owner)
	}
}

func (s *ERC721Suite) Test23TransferFromInvalidReceiver() {
	tokenId := big.NewInt(2)
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		signed, hash, err := s.erc721.TransferFrom(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewTransferFromRequest(contract, s.from, zero, tokenId))
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidReceiver)
		s.Require().Nil(signed)
		s.Require().Nil(hash)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC721 %s Test23TransferFromInvalidReceiver Result", contract)
		s.T().Logf("\tTokenId: %s", tokenId)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}
