package erc721

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
)

func TestNewTransferEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	tokenId := big.NewInt(42)

	tr := NewTransferEvent(contract, from, to, tokenId)
	require.Same(t, contract, tr.Contract)
	require.Same(t, from, tr.From)
	require.Same(t, to, tr.To)
	require.Same(t, tokenId, tr.TokenId)
}

func TestNewApprovalEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	approved := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	tokenId := big.NewInt(7)

	ap := NewApprovalEvent(contract, owner, approved, tokenId)
	require.Same(t, contract, ap.Contract)
	require.Same(t, owner, ap.Owner)
	require.Same(t, approved, ap.Approved)
	require.Same(t, tokenId, ap.TokenId)
}

func TestNewApprovalForAllEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	afa := NewApprovalForAllEvent(contract, owner, operator, true)
	require.Same(t, contract, afa.Contract)
	require.Same(t, owner, afa.Owner)
	require.Same(t, operator, afa.Operator)
	require.True(t, afa.Approved)
}

func TestNewMetadata(t *testing.T) {
	m := NewMetadata("Bored Ape Yacht Club", "BAYC")
	require.Equal(t, "Bored Ape Yacht Club", m.Name)
	require.Equal(t, "BAYC", m.Symbol)
}

func TestNewMetadataWithTokenID(t *testing.T) {
	m := NewMetadata("Bored Ape Yacht Club", "BAYC")
	tokenId := big.NewInt(1234)
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	mt := NewMetadataWithTokenID(m, tokenId, "ipfs://1234", owner)
	require.Equal(t, *m, mt.Metadata)
	require.Same(t, tokenId, mt.TokenId)
	require.Equal(t, "ipfs://1234", mt.TokenURI)
	require.Same(t, owner, mt.Owner)
}

func TestNewAllowanceWithTokenID(t *testing.T) {
	tokenId := big.NewInt(99)
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	awt := NewAllowanceWithTokenID(tokenId, owner, operator)
	require.Same(t, tokenId, awt.TokenId)
	require.Same(t, owner, awt.Owner)
	require.Same(t, operator, awt.Operator)
}

func TestNewApprovalForAllWithBalance(t *testing.T) {
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	ownerBalance := big.NewInt(3)
	operatorBalance := big.NewInt(4)

	afawb := NewApprovalForAllWithBalance(owner, operator, true, ownerBalance, operatorBalance)
	require.Same(t, owner, afawb.Owner)
	require.Same(t, operator, afawb.Operator)
	require.True(t, afawb.Approved)
	require.Same(t, ownerBalance, afawb.OwnerBalance)
	require.Same(t, operatorBalance, afawb.OperatorBalance)
}

func TestNewTokenOwner(t *testing.T) {
	tokenId := big.NewInt(5)
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")

	to := NewTokenOwner(tokenId, owner)
	require.Same(t, tokenId, to.TokenId)
	require.Same(t, owner, to.Owner)
}

func TestNewTokenOwners(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	tokens := []TokenOwner{
		*NewTokenOwner(big.NewInt(1), types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")),
		*NewTokenOwner(big.NewInt(2), types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")),
	}

	tos := NewTokenOwners(contract, tokens)
	require.Same(t, contract, tos.Contract)
	require.Equal(t, tokens, tos.Tokens)
}

func TestNewTokenOperator(t *testing.T) {
	tokenId := big.NewInt(6)
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	top := NewTokenOperator(tokenId, operator)
	require.Same(t, tokenId, top.TokenId)
	require.Same(t, operator, top.Operator)
}

func TestNewTokenOperators(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	tokens := []TokenOperator{
		*NewTokenOperator(big.NewInt(1), types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")),
		*NewTokenOperator(big.NewInt(2), types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")),
	}

	tops := NewTokenOperators(contract, tokens)
	require.Same(t, contract, tops.Contract)
	require.Equal(t, tokens, tops.Tokens)
}

func TestNewTokenBalance(t *testing.T) {
	token := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount := big.NewInt(4)

	tb := NewTokenBalance(token, amount)
	require.Same(t, token, tb.Token)
	require.Same(t, amount, tb.Amount)
}

func TestNewAddressBalances(t *testing.T) {
	address := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	tokens := []TokenBalance{
		*NewTokenBalance(types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"), big.NewInt(1)),
		*NewTokenBalance(types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8"), big.NewInt(2)),
	}

	ab := NewAddressBalances(address, tokens)
	require.Same(t, address, ab.Address)
	require.Equal(t, tokens, ab.Tokens)
}

func TestNewTokenOwnerPair(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	tokenId := big.NewInt(8)

	top := NewTokenOwnerPair(contract, owner, tokenId)
	require.Same(t, contract, top.Contract)
	require.Same(t, tokenId, top.TokenId)
	require.Same(t, owner, top.Owner)
}

func TestNewAddressBalance(t *testing.T) {
	address := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	amount := big.NewInt(5)

	ab := NewAddressBalance(address, amount)
	require.Same(t, address, ab.Address)
	require.Same(t, amount, ab.Amount)
}

func TestNewTokenBalances(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	addrs := []AddressBalance{
		*NewAddressBalance(types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f"), big.NewInt(1)),
		*NewAddressBalance(types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8"), big.NewInt(2)),
	}

	tbs := NewTokenBalances(contract, addrs)
	require.Same(t, contract, tbs.Contract)
	require.Equal(t, addrs, tbs.Addresses)
}

func TestScalarRequestsValidate(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	owner := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	operator := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	from := owner
	to := operator

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"balanceOf nil", (*BalanceOfRequest)(nil).Validate(), "invalid request: request is nil"},
		{"balanceOf empty", (&BalanceOfRequest{}).Validate(), "invalid request: contract is nil"},
		{"balanceOf nil contract", (&BalanceOfRequest{Owner: owner}).Validate(), "invalid request: contract is nil"},
		{"balanceOf nil owner", (&BalanceOfRequest{Contract: contract}).Validate(), "invalid request: owner is nil"},
		{"balanceOf valid", (&BalanceOfRequest{Contract: contract, Owner: owner}).Validate(), ""},
		{"ownerOf nil", (*OwnerOfRequest)(nil).Validate(), "invalid request: request is nil"},
		{"ownerOf empty", (&OwnerOfRequest{}).Validate(), "invalid request: contract is nil"},
		{"ownerOf nil contract", (&OwnerOfRequest{TokenID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"ownerOf nil tokenID", (&OwnerOfRequest{Contract: contract}).Validate(), "invalid request: tokenID is nil"},
		{"ownerOf valid", (&OwnerOfRequest{Contract: contract, TokenID: big.NewInt(7)}).Validate(), ""},
		{"ownerOf zero tokenID", (&OwnerOfRequest{Contract: contract, TokenID: big.NewInt(0)}).Validate(), ""},
		{"getApproved nil", (*GetApprovedRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getApproved empty", (&GetApprovedRequest{}).Validate(), "invalid request: contract is nil"},
		{"getApproved nil contract", (&GetApprovedRequest{TokenID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"getApproved nil tokenID", (&GetApprovedRequest{Contract: contract}).Validate(), "invalid request: tokenID is nil"},
		{"getApproved valid", (&GetApprovedRequest{Contract: contract, TokenID: big.NewInt(7)}).Validate(), ""},
		{"getApproved zero tokenID", (&GetApprovedRequest{Contract: contract, TokenID: big.NewInt(0)}).Validate(), ""},
		{"isApprovedForAll nil", (*IsApprovedForAllRequest)(nil).Validate(), "invalid request: request is nil"},
		{"isApprovedForAll empty", (&IsApprovedForAllRequest{}).Validate(), "invalid request: contract is nil"},
		{"isApprovedForAll nil contract", (&IsApprovedForAllRequest{Owner: owner, Operator: operator}).Validate(), "invalid request: contract is nil"},
		{"isApprovedForAll nil owner", (&IsApprovedForAllRequest{Contract: contract, Operator: operator}).Validate(), "invalid request: owner is nil"},
		{"isApprovedForAll nil operator", (&IsApprovedForAllRequest{Contract: contract, Owner: owner}).Validate(), "invalid request: operator is nil"},
		{"isApprovedForAll valid", (&IsApprovedForAllRequest{Contract: contract, Owner: owner, Operator: operator}).Validate(), ""},
		{"name nil", (*NameRequest)(nil).Validate(), "invalid request: request is nil"},
		{"name empty", (&NameRequest{}).Validate(), "invalid request: contract is nil"},
		{"name nil contract", (&NameRequest{}).Validate(), "invalid request: contract is nil"},
		{"name valid", (&NameRequest{Contract: contract}).Validate(), ""},
		{"symbol nil", (*SymbolRequest)(nil).Validate(), "invalid request: request is nil"},
		{"symbol empty", (&SymbolRequest{}).Validate(), "invalid request: contract is nil"},
		{"symbol nil contract", (&SymbolRequest{}).Validate(), "invalid request: contract is nil"},
		{"symbol valid", (&SymbolRequest{Contract: contract}).Validate(), ""},
		{"tokenURI nil", (*TokenURIRequest)(nil).Validate(), "invalid request: request is nil"},
		{"tokenURI empty", (&TokenURIRequest{}).Validate(), "invalid request: contract is nil"},
		{"tokenURI nil contract", (&TokenURIRequest{TokenID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"tokenURI nil tokenID", (&TokenURIRequest{Contract: contract}).Validate(), "invalid request: tokenID is nil"},
		{"tokenURI valid", (&TokenURIRequest{Contract: contract, TokenID: big.NewInt(7)}).Validate(), ""},
		{"tokenURI zero tokenID", (&TokenURIRequest{Contract: contract, TokenID: big.NewInt(0)}).Validate(), ""},
		{"approve nil", (*ApproveRequest)(nil).Validate(), "invalid request: request is nil"},
		{"approve empty", (&ApproveRequest{}).Validate(), "invalid request: contract is nil"},
		{"approve nil contract", (&ApproveRequest{To: to, TokenID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"approve nil to", (&ApproveRequest{Contract: contract, TokenID: big.NewInt(7)}).Validate(), "invalid request: to is nil"},
		{"approve nil tokenID", (&ApproveRequest{Contract: contract, To: to}).Validate(), "invalid request: tokenID is nil"},
		{"approve valid", (&ApproveRequest{Contract: contract, To: to, TokenID: big.NewInt(7)}).Validate(), ""},
		{"approve zero tokenID", (&ApproveRequest{Contract: contract, To: to, TokenID: big.NewInt(0)}).Validate(), ""},
		{"setApprovalForAll nil", (*SetApprovalForAllRequest)(nil).Validate(), "invalid request: request is nil"},
		{"setApprovalForAll empty", (&SetApprovalForAllRequest{}).Validate(), "invalid request: contract is nil"},
		{"setApprovalForAll nil contract", (&SetApprovalForAllRequest{Operator: operator, Approved: true}).Validate(), "invalid request: contract is nil"},
		{"setApprovalForAll nil operator", (&SetApprovalForAllRequest{Contract: contract, Approved: true}).Validate(), "invalid request: operator is nil"},
		{"setApprovalForAll valid", (&SetApprovalForAllRequest{Contract: contract, Operator: operator, Approved: true}).Validate(), ""},
		{"setApprovalForAll zero optional fields", (&SetApprovalForAllRequest{Contract: contract, Operator: operator}).Validate(), ""},
		{"transferFrom nil", (*TransferFromRequest)(nil).Validate(), "invalid request: request is nil"},
		{"transferFrom empty", (&TransferFromRequest{}).Validate(), "invalid request: contract is nil"},
		{"transferFrom nil contract", (&TransferFromRequest{From: from, To: to, TokenID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"transferFrom nil from", (&TransferFromRequest{Contract: contract, To: to, TokenID: big.NewInt(7)}).Validate(), "invalid request: from is nil"},
		{"transferFrom nil to", (&TransferFromRequest{Contract: contract, From: from, TokenID: big.NewInt(7)}).Validate(), "invalid request: to is nil"},
		{"transferFrom nil tokenID", (&TransferFromRequest{Contract: contract, From: from, To: to}).Validate(), "invalid request: tokenID is nil"},
		{"transferFrom valid", (&TransferFromRequest{Contract: contract, From: from, To: to, TokenID: big.NewInt(7)}).Validate(), ""},
		{"transferFrom zero tokenID", (&TransferFromRequest{Contract: contract, From: from, To: to, TokenID: big.NewInt(0)}).Validate(), ""},
		{"safeTransferFrom nil", (*SafeTransferFromRequest)(nil).Validate(), "invalid request: request is nil"},
		{"safeTransferFrom empty", (&SafeTransferFromRequest{}).Validate(), "invalid request: contract is nil"},
		{"safeTransferFrom nil contract", (&SafeTransferFromRequest{From: from, To: to, TokenID: big.NewInt(7)}).Validate(), "invalid request: contract is nil"},
		{"safeTransferFrom nil from", (&SafeTransferFromRequest{Contract: contract, To: to, TokenID: big.NewInt(7)}).Validate(), "invalid request: from is nil"},
		{"safeTransferFrom nil to", (&SafeTransferFromRequest{Contract: contract, From: from, TokenID: big.NewInt(7)}).Validate(), "invalid request: to is nil"},
		{"safeTransferFrom nil tokenID", (&SafeTransferFromRequest{Contract: contract, From: from, To: to}).Validate(), "invalid request: tokenID is nil"},
		{"safeTransferFrom valid", (&SafeTransferFromRequest{Contract: contract, From: from, To: to, TokenID: big.NewInt(7)}).Validate(), ""},
		{"safeTransferFrom zero tokenID", (&SafeTransferFromRequest{Contract: contract, From: from, To: to, TokenID: big.NewInt(0)}).Validate(), ""},
		{"safeTransferFromWithData nil", (*SafeTransferFromWithDataRequest)(nil).Validate(), "invalid request: request is nil"},
		{"safeTransferFromWithData empty", (&SafeTransferFromWithDataRequest{}).Validate(), "invalid request: contract is nil"},
		{"safeTransferFromWithData nil contract", (&SafeTransferFromWithDataRequest{From: from, To: to, TokenID: big.NewInt(7), Data: []byte{0x01}}).Validate(), "invalid request: contract is nil"},
		{"safeTransferFromWithData nil from", (&SafeTransferFromWithDataRequest{Contract: contract, To: to, TokenID: big.NewInt(7), Data: []byte{0x01}}).Validate(), "invalid request: from is nil"},
		{"safeTransferFromWithData nil to", (&SafeTransferFromWithDataRequest{Contract: contract, From: from, TokenID: big.NewInt(7), Data: []byte{0x01}}).Validate(), "invalid request: to is nil"},
		{"safeTransferFromWithData nil tokenID", (&SafeTransferFromWithDataRequest{Contract: contract, From: from, To: to, Data: []byte{0x01}}).Validate(), "invalid request: tokenID is nil"},
		{"safeTransferFromWithData valid", (&SafeTransferFromWithDataRequest{Contract: contract, From: from, To: to, TokenID: big.NewInt(7), Data: []byte{0x01}}).Validate(), ""},
		{"safeTransferFromWithData zero tokenID", (&SafeTransferFromWithDataRequest{Contract: contract, From: from, To: to, TokenID: big.NewInt(0), Data: []byte{0x01}}).Validate(), ""},
		{"safeTransferFromWithData zero optional fields", (&SafeTransferFromWithDataRequest{Contract: contract, From: from, To: to, TokenID: big.NewInt(7)}).Validate(), ""},
		{"metadata nil", (*MetadataRequest)(nil).Validate(), "invalid request: request is nil"},
		{"metadata empty", (&MetadataRequest{}).Validate(), "invalid request: multicall is nil"},
		{"metadata nil multicall", (&MetadataRequest{Contract: contract, AllowFailure: true}).Validate(), "invalid request: multicall is nil"},
		{"metadata nil contract", (&MetadataRequest{Multicall: multicall, AllowFailure: true}).Validate(), "invalid request: contract is nil"},
		{"metadata valid", (&MetadataRequest{Multicall: multicall, Contract: contract, AllowFailure: true}).Validate(), ""},
		{"metadata zero optional fields", (&MetadataRequest{Multicall: multicall, Contract: contract}).Validate(), ""},
		{"metadataWithTokenID nil", (*MetadataWithTokenIDRequest)(nil).Validate(), "invalid request: request is nil"},
		{"metadataWithTokenID empty", (&MetadataWithTokenIDRequest{}).Validate(), "invalid request: multicall is nil"},
		{"metadataWithTokenID nil multicall", (&MetadataWithTokenIDRequest{Contract: contract, TokenID: big.NewInt(7), AllowFailure: true}).Validate(), "invalid request: multicall is nil"},
		{"metadataWithTokenID nil contract", (&MetadataWithTokenIDRequest{Multicall: multicall, TokenID: big.NewInt(7), AllowFailure: true}).Validate(), "invalid request: contract is nil"},
		{"metadataWithTokenID nil tokenID", (&MetadataWithTokenIDRequest{Multicall: multicall, Contract: contract, AllowFailure: true}).Validate(), "invalid request: tokenID is nil"},
		{"metadataWithTokenID valid", (&MetadataWithTokenIDRequest{Multicall: multicall, Contract: contract, TokenID: big.NewInt(7), AllowFailure: true}).Validate(), ""},
		{"metadataWithTokenID zero tokenID", (&MetadataWithTokenIDRequest{Multicall: multicall, Contract: contract, TokenID: big.NewInt(0), AllowFailure: true}).Validate(), ""},
		{"metadataWithTokenID zero optional fields", (&MetadataWithTokenIDRequest{Multicall: multicall, Contract: contract, TokenID: big.NewInt(7)}).Validate(), ""},
		{"allowanceWithTokenID nil", (*AllowanceWithTokenIDRequest)(nil).Validate(), "invalid request: request is nil"},
		{"allowanceWithTokenID empty", (&AllowanceWithTokenIDRequest{}).Validate(), "invalid request: multicall is nil"},
		{"allowanceWithTokenID nil multicall", (&AllowanceWithTokenIDRequest{Contract: contract, TokenID: big.NewInt(7), AllowFailure: true}).Validate(), "invalid request: multicall is nil"},
		{"allowanceWithTokenID nil contract", (&AllowanceWithTokenIDRequest{Multicall: multicall, TokenID: big.NewInt(7), AllowFailure: true}).Validate(), "invalid request: contract is nil"},
		{"allowanceWithTokenID nil tokenID", (&AllowanceWithTokenIDRequest{Multicall: multicall, Contract: contract, AllowFailure: true}).Validate(), "invalid request: tokenID is nil"},
		{"allowanceWithTokenID valid", (&AllowanceWithTokenIDRequest{Multicall: multicall, Contract: contract, TokenID: big.NewInt(7), AllowFailure: true}).Validate(), ""},
		{"allowanceWithTokenID zero tokenID", (&AllowanceWithTokenIDRequest{Multicall: multicall, Contract: contract, TokenID: big.NewInt(0), AllowFailure: true}).Validate(), ""},
		{"allowanceWithTokenID zero optional fields", (&AllowanceWithTokenIDRequest{Multicall: multicall, Contract: contract, TokenID: big.NewInt(7)}).Validate(), ""},
		{"approvalForAllWithBalance nil", (*ApprovalForAllWithBalanceRequest)(nil).Validate(), "invalid request: request is nil"},
		{"approvalForAllWithBalance empty", (&ApprovalForAllWithBalanceRequest{}).Validate(), "invalid request: multicall is nil"},
		{"approvalForAllWithBalance nil multicall", (&ApprovalForAllWithBalanceRequest{Contract: contract, Owner: owner, Operator: operator, AllowFailure: true}).Validate(), "invalid request: multicall is nil"},
		{"approvalForAllWithBalance nil contract", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Owner: owner, Operator: operator, AllowFailure: true}).Validate(), "invalid request: contract is nil"},
		{"approvalForAllWithBalance nil owner", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Operator: operator, AllowFailure: true}).Validate(), "invalid request: owner is nil"},
		{"approvalForAllWithBalance nil operator", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Owner: owner, AllowFailure: true}).Validate(), "invalid request: operator is nil"},
		{"approvalForAllWithBalance valid", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Owner: owner, Operator: operator, AllowFailure: true}).Validate(), ""},
		{"approvalForAllWithBalance zero optional fields", (&ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Owner: owner, Operator: operator}).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestSafeTransferFromWithDataRequestAcceptsEmptyData(t *testing.T) {
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	from := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	to := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	require.NoError(t, NewSafeTransferFromWithDataRequest(contract, from, to, big.NewInt(1), nil).Validate())
	require.NoError(t, NewSafeTransferFromWithDataRequest(contract, from, to, big.NewInt(1), []byte{}).Validate())
	require.NoError(t, NewSafeTransferFromWithDataRequest(contract, from, to, big.NewInt(1), make([]byte, 1024)).Validate())
}

func TestTokenIDListRequestsValidate(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"tokenOwners nil", (*TokenOwnersRequest)(nil).Validate(), "invalid request: request is nil"},
		{"tokenOwners nil multicall", NewTokenOwnersRequest(nil, contract, []*big.Int{big.NewInt(1)}, false).Validate(), "invalid request: multicall is nil"},
		{"tokenOwners nil contract", NewTokenOwnersRequest(multicall, nil, []*big.Int{big.NewInt(1)}, false).Validate(), "invalid request: contract is nil"},
		{"tokenOwners nil list", NewTokenOwnersRequest(multicall, contract, nil, false).Validate(), "invalid request: tokenIDs is empty"},
		{"tokenOwners empty list", NewTokenOwnersRequest(multicall, contract, []*big.Int{}, false).Validate(), "invalid request: tokenIDs is empty"},
		{"tokenOwners nil first", NewTokenOwnersRequest(multicall, contract, []*big.Int{nil, big.NewInt(1)}, false).Validate(), "invalid request: tokenIDs[0] is nil"},
		{"tokenOwners nil last", NewTokenOwnersRequest(multicall, contract, []*big.Int{big.NewInt(0), big.NewInt(1), nil}, false).Validate(), "invalid request: tokenIDs[2] is nil"},
		{"tokenOwners valid", NewTokenOwnersRequest(multicall, contract, []*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(1)}, true).Validate(), ""},
		{"tokenOperators nil", (*TokenOperatorsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"tokenOperators nil multicall", NewTokenOperatorsRequest(nil, contract, []*big.Int{big.NewInt(1)}, false).Validate(), "invalid request: multicall is nil"},
		{"tokenOperators nil contract", NewTokenOperatorsRequest(multicall, nil, []*big.Int{big.NewInt(1)}, false).Validate(), "invalid request: contract is nil"},
		{"tokenOperators empty list", NewTokenOperatorsRequest(multicall, contract, []*big.Int{}, false).Validate(), "invalid request: tokenIDs is empty"},
		{"tokenOperators nil middle", NewTokenOperatorsRequest(multicall, contract, []*big.Int{big.NewInt(0), nil, big.NewInt(1)}, false).Validate(), "invalid request: tokenIDs[1] is nil"},
		{"tokenOperators valid", NewTokenOperatorsRequest(multicall, contract, []*big.Int{big.NewInt(9)}, false).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestAddressListRequestsValidate(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"addressBalances nil", (*AddressBalancesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"addressBalances nil multicall", NewAddressBalancesRequest(nil, a, []*types.Address{contract}, false).Validate(), "invalid request: multicall is nil"},
		{"addressBalances nil account", NewAddressBalancesRequest(multicall, nil, []*types.Address{contract}, false).Validate(), "invalid request: account is nil"},
		{"addressBalances nil list", NewAddressBalancesRequest(multicall, a, nil, false).Validate(), "invalid request: contracts is empty"},
		{"addressBalances nil first", NewAddressBalancesRequest(multicall, a, []*types.Address{nil, contract}, false).Validate(), "invalid request: contracts[0] is nil"},
		{"addressBalances nil last", NewAddressBalancesRequest(multicall, a, []*types.Address{contract, contract, nil}, false).Validate(), "invalid request: contracts[2] is nil"},
		{"addressBalances valid", NewAddressBalancesRequest(multicall, a, []*types.Address{contract, contract}, true).Validate(), ""},
		{"tokenBalances nil", (*TokenBalancesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"tokenBalances nil multicall", NewTokenBalancesRequest(nil, contract, []*types.Address{a}, false).Validate(), "invalid request: multicall is nil"},
		{"tokenBalances nil contract", NewTokenBalancesRequest(multicall, nil, []*types.Address{a}, false).Validate(), "invalid request: contract is nil"},
		{"tokenBalances empty list", NewTokenBalancesRequest(multicall, contract, []*types.Address{}, false).Validate(), "invalid request: accounts is empty"},
		{"tokenBalances nil middle", NewTokenBalancesRequest(multicall, contract, []*types.Address{a, nil, b}, false).Validate(), "invalid request: accounts[1] is nil"},
		{"tokenBalances valid", NewTokenBalancesRequest(multicall, contract, []*types.Address{a, b}, true).Validate(), ""},
		{"ownerPairs nil", (*OwnerPairsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"ownerPairs nil multicall", NewOwnerPairsRequest(nil, []*types.Address{contract}, []*big.Int{big.NewInt(1)}, false).Validate(), "invalid request: multicall is nil"},
		{"ownerPairs contracts longer", NewOwnerPairsRequest(multicall, []*types.Address{contract, contract}, []*big.Int{big.NewInt(1)}, false).Validate(), "invalid request: contracts and tokenIDs differ in length"},
		{"ownerPairs tokenIDs longer", NewOwnerPairsRequest(multicall, []*types.Address{contract}, []*big.Int{big.NewInt(1), big.NewInt(2)}, false).Validate(), "invalid request: contracts and tokenIDs differ in length"},
		{"ownerPairs mismatch before nil", NewOwnerPairsRequest(multicall, []*types.Address{nil}, nil, false).Validate(), "invalid request: contracts and tokenIDs differ in length"},
		{"ownerPairs both empty", NewOwnerPairsRequest(multicall, nil, nil, false).Validate(), "invalid request: contracts is empty"},
		{"ownerPairs nil contract", NewOwnerPairsRequest(multicall, []*types.Address{contract, nil}, []*big.Int{big.NewInt(1), big.NewInt(2)}, false).Validate(), "invalid request: contracts[1] is nil"},
		{"ownerPairs nil tokenID", NewOwnerPairsRequest(multicall, []*types.Address{contract, contract}, []*big.Int{nil, big.NewInt(2)}, false).Validate(), "invalid request: tokenIDs[0] is nil"},
		{"ownerPairs lower index reported first", NewOwnerPairsRequest(multicall, []*types.Address{contract, nil}, []*big.Int{nil, big.NewInt(2)}, false).Validate(), "invalid request: tokenIDs[0] is nil"},
		{"ownerPairs contract before tokenID at same index", NewOwnerPairsRequest(multicall, []*types.Address{contract, nil}, []*big.Int{big.NewInt(1), nil}, false).Validate(), "invalid request: contracts[1] is nil"},
		{"ownerPairs valid", NewOwnerPairsRequest(multicall, []*types.Address{contract, contract}, []*big.Int{big.NewInt(1), big.NewInt(1)}, true).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestNewRequestsSetFields(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	id := big.NewInt(42)
	ids := []*big.Int{big.NewInt(1), big.NewInt(2)}
	addrs := []*types.Address{a, b}

	require.Equal(t, &BalanceOfRequest{Contract: contract, Owner: a}, NewBalanceOfRequest(contract, a))
	require.Equal(t, &OwnerOfRequest{Contract: contract, TokenID: id}, NewOwnerOfRequest(contract, id))
	require.Equal(t, &GetApprovedRequest{Contract: contract, TokenID: id}, NewGetApprovedRequest(contract, id))
	require.Equal(t, &IsApprovedForAllRequest{Contract: contract, Owner: a, Operator: b}, NewIsApprovedForAllRequest(contract, a, b))
	require.Equal(t, &NameRequest{Contract: contract}, NewNameRequest(contract))
	require.Equal(t, &SymbolRequest{Contract: contract}, NewSymbolRequest(contract))
	require.Equal(t, &TokenURIRequest{Contract: contract, TokenID: id}, NewTokenURIRequest(contract, id))
	require.Equal(t, &ApproveRequest{Contract: contract, To: a, TokenID: id}, NewApproveRequest(contract, a, id))
	require.Equal(t, &SetApprovalForAllRequest{Contract: contract, Operator: b, Approved: true}, NewSetApprovalForAllRequest(contract, b, true))
	require.Equal(t, &TransferFromRequest{Contract: contract, From: a, To: b, TokenID: id}, NewTransferFromRequest(contract, a, b, id))
	require.Equal(t, &SafeTransferFromRequest{Contract: contract, From: a, To: b, TokenID: id}, NewSafeTransferFromRequest(contract, a, b, id))
	require.Equal(t, &SafeTransferFromWithDataRequest{Contract: contract, From: a, To: b, TokenID: id, Data: []byte{0xde, 0xad}}, NewSafeTransferFromWithDataRequest(contract, a, b, id, []byte{0xde, 0xad}))
	require.Equal(t, &MetadataRequest{Multicall: multicall, Contract: contract, AllowFailure: true}, NewMetadataRequest(multicall, contract, true))
	require.Equal(t, &MetadataWithTokenIDRequest{Multicall: multicall, Contract: contract, TokenID: id}, NewMetadataWithTokenIDRequest(multicall, contract, id, false))
	require.Equal(t, &AllowanceWithTokenIDRequest{Multicall: multicall, Contract: contract, TokenID: id, AllowFailure: true}, NewAllowanceWithTokenIDRequest(multicall, contract, id, true))
	require.Equal(t, &ApprovalForAllWithBalanceRequest{Multicall: multicall, Contract: contract, Owner: a, Operator: b}, NewApprovalForAllWithBalanceRequest(multicall, contract, a, b, false))
	require.Equal(t, &TokenOwnersRequest{Multicall: multicall, Contract: contract, TokenIDs: ids, AllowFailure: true}, NewTokenOwnersRequest(multicall, contract, ids, true))
	require.Equal(t, &TokenOperatorsRequest{Multicall: multicall, Contract: contract, TokenIDs: ids}, NewTokenOperatorsRequest(multicall, contract, ids, false))
	require.Equal(t, &AddressBalancesRequest{Multicall: multicall, Account: a, Contracts: addrs}, NewAddressBalancesRequest(multicall, a, addrs, false))
	require.Equal(t, &TokenBalancesRequest{Multicall: multicall, Contract: contract, Accounts: addrs, AllowFailure: true}, NewTokenBalancesRequest(multicall, contract, addrs, true))
	require.Equal(t, &OwnerPairsRequest{Multicall: multicall, Contracts: addrs, TokenIDs: ids}, NewOwnerPairsRequest(multicall, addrs, ids, false))
}

func TestClientMethodsWrapValidateError(t *testing.T) {
	multicall := types.MustNewAddressFromHex("0xca11bde05977b3631167028862be2a173976ca11")
	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	e := New(mocks.NewClient(t))
	ctx := t.Context()

	_, err := e.Name(ctx, nil, &NameRequest{})
	require.EqualError(t, err, "erc721: name: "+(&NameRequest{}).Validate().Error())
	_, err = e.Symbol(ctx, nil, &SymbolRequest{})
	require.EqualError(t, err, "erc721: symbol: "+(&SymbolRequest{}).Validate().Error())
	_, err = e.TokenURI(ctx, nil, NewTokenURIRequest(contract, nil))
	require.EqualError(t, err, "erc721: tokenURI: "+NewTokenURIRequest(contract, nil).Validate().Error())
	_, err = e.BalanceOf(ctx, nil, NewBalanceOfRequest(contract, nil))
	require.EqualError(t, err, "erc721: balanceOf: "+NewBalanceOfRequest(contract, nil).Validate().Error())
	_, err = e.OwnerOf(ctx, nil, NewOwnerOfRequest(nil, big.NewInt(1)))
	require.EqualError(t, err, "erc721: ownerOf: "+NewOwnerOfRequest(nil, big.NewInt(1)).Validate().Error())
	_, err = e.GetApproved(ctx, nil, NewGetApprovedRequest(contract, nil))
	require.EqualError(t, err, "erc721: getApproved: "+NewGetApprovedRequest(contract, nil).Validate().Error())
	_, err = e.IsApprovedForAll(ctx, nil, NewIsApprovedForAllRequest(contract, contract, nil))
	require.EqualError(t, err, "erc721: isApprovedForAll: "+NewIsApprovedForAllRequest(contract, contract, nil).Validate().Error())

	_, _, err = e.Approve(ctx, nil, NewApproveRequest(contract, nil, big.NewInt(1)))
	require.EqualError(t, err, "erc721: approve: "+NewApproveRequest(contract, nil, big.NewInt(1)).Validate().Error())
	_, _, err = e.SetApprovalForAll(ctx, nil, NewSetApprovalForAllRequest(contract, nil, true))
	require.EqualError(t, err, "erc721: setApprovalForAll: "+NewSetApprovalForAllRequest(contract, nil, true).Validate().Error())
	_, _, err = e.TransferFrom(ctx, nil, NewTransferFromRequest(contract, contract, contract, nil))
	require.EqualError(t, err, "erc721: transferFrom: "+NewTransferFromRequest(contract, contract, contract, nil).Validate().Error())
	_, _, err = e.SafeTransferFrom(ctx, nil, NewSafeTransferFromRequest(contract, nil, contract, big.NewInt(1)))
	require.EqualError(t, err, "erc721: safeTransferFrom: "+NewSafeTransferFromRequest(contract, nil, contract, big.NewInt(1)).Validate().Error())
	_, _, err = e.SafeTransferFromWithData(ctx, nil, NewSafeTransferFromWithDataRequest(contract, contract, nil, big.NewInt(1), nil))
	require.EqualError(t, err, "erc721: safeTransferFromWithData: "+NewSafeTransferFromWithDataRequest(contract, contract, nil, big.NewInt(1), nil).Validate().Error())

	_, err = e.Metadata(ctx, nil, NewMetadataRequest(nil, contract, false))
	require.EqualError(t, err, "erc721: metadata: "+NewMetadataRequest(nil, contract, false).Validate().Error())
	_, err = e.MetadataWithTokenID(ctx, nil, NewMetadataWithTokenIDRequest(multicall, contract, nil, false))
	require.EqualError(t, err, "erc721: metadataWithTokenID: "+NewMetadataWithTokenIDRequest(multicall, contract, nil, false).Validate().Error())
	_, err = e.AllowanceWithTokenID(ctx, nil, NewAllowanceWithTokenIDRequest(multicall, nil, big.NewInt(1), false))
	require.EqualError(t, err, "erc721: allowanceWithTokenID: "+NewAllowanceWithTokenIDRequest(multicall, nil, big.NewInt(1), false).Validate().Error())
	_, err = e.ApprovalForAllWithBalance(ctx, nil, NewApprovalForAllWithBalanceRequest(multicall, contract, contract, nil, false))
	require.EqualError(t, err, "erc721: approvalForAllWithBalance: "+NewApprovalForAllWithBalanceRequest(multicall, contract, contract, nil, false).Validate().Error())
	_, err = e.TokenOwners(ctx, nil, NewTokenOwnersRequest(multicall, contract, nil, false))
	require.EqualError(t, err, "erc721: tokenOwners: "+NewTokenOwnersRequest(multicall, contract, nil, false).Validate().Error())
	_, err = e.TokenOperators(ctx, nil, NewTokenOperatorsRequest(multicall, contract, []*big.Int{nil}, false))
	require.EqualError(t, err, "erc721: tokenOperators: "+NewTokenOperatorsRequest(multicall, contract, []*big.Int{nil}, false).Validate().Error())
	_, err = e.AddressBalances(ctx, nil, NewAddressBalancesRequest(multicall, nil, []*types.Address{contract}, false))
	require.EqualError(t, err, "erc721: addressBalances: "+NewAddressBalancesRequest(multicall, nil, []*types.Address{contract}, false).Validate().Error())
	_, err = e.TokenBalances(ctx, nil, NewTokenBalancesRequest(multicall, contract, []*types.Address{}, false))
	require.EqualError(t, err, "erc721: tokenBalances: "+NewTokenBalancesRequest(multicall, contract, []*types.Address{}, false).Validate().Error())
	_, err = e.OwnerPairs(ctx, nil, NewOwnerPairsRequest(multicall, []*types.Address{contract}, nil, false))
	require.EqualError(t, err, "erc721: ownerPairs: "+NewOwnerPairsRequest(multicall, []*types.Address{contract}, nil, false).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, _, err = e.Approve(ctx, nil, NewApproveRequest(contract, contract, big.NewInt(1)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "erc721: approve: bind: transact: invalid options: options is nil")
}
