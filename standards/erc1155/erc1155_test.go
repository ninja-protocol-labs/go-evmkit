package erc1155

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

var interfaceIDERC1155 = [4]byte{0xd9, 0xb6, 0x7a, 0x26}

type ERC1155Suite struct {
	suite.Suite

	RPCURL         string
	PrivateKeyHex  string
	MulticallHex   string
	NativeSymbol   string
	NativeDecimals int
	Timeout        time.Duration
	ToHex          string
	TokenCount     int
	ByteCode       string
	Constructor    *abi.Function
	Args           []any
	Salt           *types.Hash

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
	c := testenv.MustLoad(t, testenv.UnknownNetworkType, testenv.UnknownNetwork)

	suite.Run(t, &ERC1155Suite{
		RPCURL:         c.RPCURL,
		PrivateKeyHex:  c.ActorPrivateKeyHex,
		MulticallHex:   c.Multicall,
		NativeSymbol:   c.NativeSymbol,
		NativeDecimals: c.NativeDecimals,
		Timeout:        10 * time.Minute,
		ToHex:          "",
		TokenCount:     5,
		ByteCode:       testenv.ERC1155ByteCodeHex,
		Constructor:    abi.NewFunction("constructor", abi.NewTypes(abi.String), nil),
		Args:           []any{"https://evmkit-fixtures.test/erc1155/{id}.json"},
		Salt:           core.Keccak256([]byte("hello")),
	})
}

func (s *ERC1155Suite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(s.T().Context(), s.Timeout)

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

	mintBatchFn := abi.NewFunction("mintBatch", abi.NewTypes(abi.Address, abi.Slice(abi.Uint256), abi.Slice(abi.Uint256)), nil)
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
	ctx, cancel := context.WithTimeout(s.T().Context(), s.Timeout)
	defer cancel()

	if s.contracts != nil {
		burnBatchFn := abi.NewFunction("burnBatch", abi.NewTypes(abi.Address, abi.Slice(abi.Uint256), abi.Slice(abi.Uint256)), nil)

		for _, contract := range s.contracts {
			for _, holder := range []*types.Address{s.from, s.to} {
				accounts := make([]*types.Address, len(s.ids))
				for i := range accounts {
					accounts[i] = holder
				}
				balances, err := s.erc1155.BalanceOfBatch(ctx, nil, NewBalanceOfBatchRequest(contract, accounts, s.ids))
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

				burnReceipt, err := tx.WaitForReceipt(ctx, s.client, burnHash, 500, 15*time.Second)
				s.Require().NoError(err)
				s.Require().True(burnReceipt.Status, "burnBatch transaction reverted")

				s.T().Logf("TearDownSuite: burned ids %v (amounts %v) of %s from %s (tx %s)", burnIds, burnAmounts, contract, holder, burnHash)
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

		s.T().Logf("ERC1155Suite total fee spent: %s %s", decimal, s.NativeSymbol)
	}

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ERC1155Suite) Test01AccountIdBalances() {
	id := s.ids[0]

	aib, err := s.erc1155.AccountIdBalances(s.ctx, nil, NewAccountIdBalancesRequest(s.multicall, s.from, id, s.contracts, false))
	s.Require().NoError(err)
	s.Require().True(aib.Account.Equal(s.from))
	s.Require().Zero(id.Cmp(aib.Id))
	s.Require().Len(aib.Contracts, len(s.contracts))

	for i, contract := range s.contracts {
		want, err := s.erc1155.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.from, id))
		s.Require().NoError(err)
		s.Require().True(aib.Contracts[i].Contract.Equal(contract))
		s.Require().Zero(want.Cmp(aib.Contracts[i].Amount))
	}

	s.T().Logf("ERC1155 Test01AccountIdBalances Result (account %s, id %s)", s.from, id)
	for _, cb := range aib.Contracts {
		s.T().Logf("\t%s: %s", cb.Contract, cb.Amount)
	}
}

func (s *ERC1155Suite) Test02AccountIdBalancesAllowFailure() {
	id := s.ids[0]

	aib, err := s.erc1155.AccountIdBalances(s.ctx, nil, NewAccountIdBalancesRequest(s.multicall, s.from, id, []*types.Address{s.multicall}, true))
	s.Require().NoError(err)
	s.Require().Len(aib.Contracts, 1)
	s.Require().Nil(aib.Contracts[0].Amount)

	s.T().Logf("ERC1155 Test02AccountIdBalancesAllowFailure Result")
	s.T().Logf("\tAmount: %v", aib.Contracts[0].Amount)
}

func (s *ERC1155Suite) Test03ApprovalForAllWithBalance() {
	id := s.ids[0]

	for _, contract := range s.contracts {
		afawb, err := s.erc1155.ApprovalForAllWithBalance(s.ctx, nil, NewApprovalForAllWithBalanceRequest(s.multicall, contract, s.from, s.to, id, false))
		s.Require().NoError(err)

		approved, err := s.erc1155.IsApprovedForAll(s.ctx, nil, NewIsApprovedForAllRequest(contract, s.from, s.to))
		s.Require().NoError(err)
		fromBal, err := s.erc1155.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.from, id))
		s.Require().NoError(err)
		toBal, err := s.erc1155.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.to, id))
		s.Require().NoError(err)

		s.Require().True(afawb.Account.Equal(s.from))
		s.Require().True(afawb.Operator.Equal(s.to))
		s.Require().Zero(id.Cmp(afawb.Id))
		s.Require().Equal(approved, afawb.Approved)
		s.Require().Zero(fromBal.Cmp(afawb.AccountBalance))
		s.Require().Zero(toBal.Cmp(afawb.OperatorBalance))

		s.T().Logf("ERC1155 %s Test03ApprovalForAllWithBalance Result", contract)
		s.T().Logf("\tApproved: %v", afawb.Approved)
		s.T().Logf("\tAccountBalance: %s", afawb.AccountBalance)
		s.T().Logf("\tOperatorBalance: %s", afawb.OperatorBalance)
	}
}

func (s *ERC1155Suite) Test04ApprovalForAllWithBalanceAllowFailure() {
	id := s.ids[0]

	afawb, err := s.erc1155.ApprovalForAllWithBalance(s.ctx, nil, NewApprovalForAllWithBalanceRequest(s.multicall, s.multicall, s.from, s.to, id, true))
	s.Require().NoError(err)
	s.Require().False(afawb.Approved)
	s.Require().Nil(afawb.AccountBalance)
	s.Require().Nil(afawb.OperatorBalance)

	s.T().Logf("ERC1155 %s Test04ApprovalForAllWithBalanceAllowFailure Result", s.multicall)
	s.T().Logf("\tApproved: %v", afawb.Approved)
	s.T().Logf("\tAccountBalance: %v", afawb.AccountBalance)
	s.T().Logf("\tOperatorBalance: %v", afawb.OperatorBalance)
}

func (s *ERC1155Suite) Test05IdBalancePairs() {
	for _, contract := range s.contracts {
		accounts := make([]*types.Address, len(s.ids))
		for i := range accounts {
			if i%2 == 0 {
				accounts[i] = s.from
			} else {
				accounts[i] = s.to
			}
		}

		pairs, err := s.erc1155.IdBalancePairs(s.ctx, nil, NewIdBalancePairsRequest(contract, accounts, s.ids))
		s.Require().NoError(err)
		s.Require().Len(pairs, len(s.ids))

		for i, id := range s.ids {
			want, err := s.erc1155.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, accounts[i], id))
			s.Require().NoError(err)
			s.Require().True(pairs[i].Account.Equal(accounts[i]))
			s.Require().Zero(id.Cmp(pairs[i].Id))
			s.Require().Zero(want.Cmp(pairs[i].Amount))
		}

		s.T().Logf("ERC1155 %s Test05IdBalancePairs Result", contract)
		for _, p := range pairs {
			s.T().Logf("\taccount %s id %s: %s", p.Account, p.Id, p.Amount)
		}
	}
}

func (s *ERC1155Suite) Test06IdBalances() {
	for _, contract := range s.contracts {
		balances, err := s.erc1155.IdBalances(s.ctx, nil, NewIdBalancesRequest(contract, s.from, s.ids))
		s.Require().NoError(err)
		s.Require().Len(balances, len(s.ids))

		for i, id := range s.ids {
			want, err := s.erc1155.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.from, id))
			s.Require().NoError(err)
			s.Require().Zero(id.Cmp(balances[i].Id))
			s.Require().Zero(want.Cmp(balances[i].Amount))
		}

		s.T().Logf("ERC1155 %s Test06IdBalances Result", contract)
		for _, b := range balances {
			s.T().Logf("\tid %s: %s", b.Id, b.Amount)
		}
	}
}

func (s *ERC1155Suite) Test07ReadMethods() {
	for _, contract := range s.contracts {
		supports, err := s.erc1155.SupportsInterface(s.ctx, nil, erc165.NewSupportsInterfaceRequest(contract, interfaceIDERC1155))
		s.Require().NoError(err)
		s.Require().True(supports)

		uris := make([]string, len(s.ids))
		fromBalances := make([]*big.Int, len(s.ids))
		toBalances := make([]*big.Int, len(s.ids))
		for i, id := range s.ids {
			uris[i], err = s.erc1155.URI(s.ctx, nil, NewURIRequest(contract, id))
			s.Require().NoError(err)

			fromBalances[i], err = s.erc1155.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.from, id))
			s.Require().NoError(err)
			s.Require().Zero(s.amount.Cmp(fromBalances[i]))

			toBalances[i], err = s.erc1155.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.to, id))
			s.Require().NoError(err)
			s.Require().Zero(big.NewInt(0).Cmp(toBalances[i]))
		}

		accounts := make([]*types.Address, len(s.ids))
		for i := range accounts {
			accounts[i] = s.from
		}
		balances, err := s.erc1155.BalanceOfBatch(s.ctx, nil, NewBalanceOfBatchRequest(contract, accounts, s.ids))
		s.Require().NoError(err)
		s.Require().Len(balances, len(s.ids))
		for _, bal := range balances {
			s.Require().Zero(s.amount.Cmp(bal))
		}

		isApprovedForAll, err := s.erc1155.IsApprovedForAll(s.ctx, nil, NewIsApprovedForAllRequest(contract, s.from, s.to))
		s.Require().NoError(err)
		s.Require().False(isApprovedForAll)

		s.T().Logf("ERC1155 %s Test07ReadMethods Result", contract)
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

func (s *ERC1155Suite) Test08SafeBatchTransferFrom() {
	ids := s.ids[3:5]
	amounts := make([]*big.Int, len(ids))
	for i := range amounts {
		amounts[i] = s.amount
	}

	txs := make([]core.Transaction, len(s.contracts))
	hashes := make([]*types.Hash, len(s.contracts))
	for i, contract := range s.contracts {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(txs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc1155.SafeBatchTransferFrom(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSafeBatchTransferFromRequest(contract, s.from, s.to, ids, amounts, nil))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		txs[i] = signed
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
			toBalances, err = s.erc1155.BalanceOfBatch(s.ctx, nil, NewBalanceOfBatchRequest(contract, toAccounts, ids))
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
		fromBalances, err := s.erc1155.BalanceOfBatch(s.ctx, nil, NewBalanceOfBatchRequest(contract, fromAccounts, ids))
		s.Require().NoError(err)
		for _, bal := range fromBalances {
			s.Require().Zero(big.NewInt(0).Cmp(bal))
		}

		s.T().Logf("ERC1155 %s Test08SafeBatchTransferFrom Result", contract)
		s.T().Logf("\tFrom: %s", s.from)
		s.T().Logf("\tTo: %s", s.to)
		s.T().Logf("\tIds: %v", ids)
		s.T().Logf("\tAmounts: %v", amounts)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", txs[i].Gas())
		s.T().Logf("\tMaxFee: %s", txs[i].MaxFee())
		s.T().Logf("\tBalanceOfBatch")
		s.T().Logf("\t\tFrom: %v", fromBalances)
		s.T().Logf("\t\tTo: %v", toBalances)
	}
}

func (s *ERC1155Suite) Test09SafeBatchTransferFromInvalidReceiver() {
	ids := s.ids[3:5]
	amounts := make([]*big.Int, len(ids))
	for i := range amounts {
		amounts[i] = s.amount
	}
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		signed, hash, err := s.erc1155.SafeBatchTransferFrom(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSafeBatchTransferFromRequest(contract, s.from, zero, ids, amounts, nil))
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidReceiver)
		s.Require().Nil(signed)
		s.Require().Nil(hash)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC1155 %s Test09SafeBatchTransferFromInvalidReceiver Result", contract)
		s.T().Logf("\tIds: %v", ids)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC1155Suite) Test10SafeTransferFrom() {
	for _, id := range s.ids[1:3] {
		txs := make([]core.Transaction, len(s.contracts))
		hashes := make([]*types.Hash, len(s.contracts))
		for i, contract := range s.contracts {
			builder := tx.NewDynamicFeeBuilder()
			if i > 0 {
				builder.Nonce(txs[0].AddressNonce() + uint64(i))
			}

			signed, hash, err := s.erc1155.SafeTransferFrom(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSafeTransferFromRequest(contract, s.from, s.to, id, s.amount, nil))
			s.Require().NoError(err)
			s.mu.Lock()
			s.totalFee.Add(s.totalFee, signed.MaxFee())
			s.mu.Unlock()

			txs[i] = signed
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
				toBalance, err = s.erc1155.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.to, id))
				return err == nil && s.amount.Cmp(toBalance) == 0
			}, 10*time.Second, 500*time.Millisecond, "balance not visible after safeTransferFrom")

			fromBalance, err := s.erc1155.BalanceOf(s.ctx, nil, NewBalanceOfRequest(contract, s.from, id))
			s.Require().NoError(err)
			s.Require().Zero(big.NewInt(0).Cmp(fromBalance))

			s.T().Logf("ERC1155 %s Test10SafeTransferFrom Result (id %s)", contract, id)
			s.T().Logf("\tFrom: %s", s.from)
			s.T().Logf("\tTo: %s", s.to)
			s.T().Logf("\tId: %s", id)
			s.T().Logf("\tAmount: %s", s.amount)
			s.T().Logf("\tTx: %s", hashes[i])
			s.T().Logf("\tGas: %d", txs[i].Gas())
			s.T().Logf("\tMaxFee: %s", txs[i].MaxFee())
			s.T().Logf("\tBalanceOf")
			s.T().Logf("\t\tFrom: %s", fromBalance)
			s.T().Logf("\t\tTo: %s", toBalance)
		}
	}
}

func (s *ERC1155Suite) Test11SafeTransferFromInvalidReceiver() {
	id := s.ids[4]
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		signed, hash, err := s.erc1155.SafeTransferFrom(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSafeTransferFromRequest(contract, s.from, zero, id, s.amount, nil))
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidReceiver)
		s.Require().Nil(signed)
		s.Require().Nil(hash)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC1155 %s Test11SafeTransferFromInvalidReceiver Result", contract)
		s.T().Logf("\tId: %s", id)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}

func (s *ERC1155Suite) Test12SetApprovalForAll() {
	txs := make([]core.Transaction, len(s.contracts))
	hashes := make([]*types.Hash, len(s.contracts))
	for i, contract := range s.contracts {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(txs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc1155.SetApprovalForAll(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSetApprovalForAllRequest(contract, s.to, true))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		txs[i] = signed
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
			after, err = s.erc1155.IsApprovedForAll(s.ctx, nil, NewIsApprovedForAllRequest(contract, s.from, s.to))
			return err == nil && after
		}, 10*time.Second, 500*time.Millisecond, "approval-for-all not visible after setApprovalForAll")

		s.T().Logf("ERC1155 %s Test12SetApprovalForAll Result", contract)
		s.T().Logf("\tAccount: %s", s.from)
		s.T().Logf("\tOperator: %s", s.to)
		s.T().Logf("\tTx: %s", hashes[i])
		s.T().Logf("\tGas: %d", txs[i].Gas())
		s.T().Logf("\tNonce: %d", txs[i].AddressNonce())
		s.T().Logf("\tMaxFee: %s", txs[i].MaxFee())
		s.T().Logf("\tIsApprovedForAll")
		s.T().Logf("\t\tAfter: %v", after)
	}

	revertTxs := make([]core.Transaction, len(s.contracts))
	revertHashes := make([]*types.Hash, len(s.contracts))
	for i, contract := range s.contracts {
		builder := tx.NewDynamicFeeBuilder()
		if i > 0 {
			builder.Nonce(revertTxs[0].AddressNonce() + uint64(i))
		}

		signed, hash, err := s.erc1155.SetApprovalForAll(s.ctx, bind.NewTransactOptions(s.signer, builder), NewSetApprovalForAllRequest(contract, s.to, false))
		s.Require().NoError(err)
		s.mu.Lock()
		s.totalFee.Add(s.totalFee, signed.MaxFee())
		s.mu.Unlock()

		revertTxs[i] = signed
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
			reverted, err = s.erc1155.IsApprovedForAll(s.ctx, nil, NewIsApprovedForAllRequest(contract, s.from, s.to))
			return err == nil && !reverted
		}, 10*time.Second, 500*time.Millisecond, "approval-for-all not reverted to false")

		s.T().Logf("ERC1155 %s Test12SetApprovalForAll Revert Result", contract)
		s.T().Logf("\tRevertTx: %s", revertHashes[i])
		s.T().Logf("\t\tMaxFee: %s", revertTxs[i].MaxFee())
		s.T().Logf("\t\tReverted: %v", reverted)
	}
}

func (s *ERC1155Suite) Test13SetApprovalForAllInvalidOperator() {
	zero := types.ZeroAddress()

	for _, contract := range s.contracts {
		signed, hash, err := s.erc1155.SetApprovalForAll(s.ctx, bind.NewTransactOptions(s.signer, tx.NewDynamicFeeBuilder()), NewSetApprovalForAllRequest(contract, zero, true))
		s.Require().Error(err)
		s.Require().ErrorIs(err, ErrInvalidOperator)
		s.Require().Nil(signed)
		s.Require().Nil(hash)

		var rpcErr *rpc.ResponseError
		s.Require().ErrorAs(err, &rpcErr)
		s.Require().NotEmpty(rpcErr.Data)

		s.T().Logf("ERC1155 %s Test13SetApprovalForAllInvalidOperator Result", contract)
		s.T().Logf("\tOperator (zero): %s", zero)
		s.T().Logf("\tError: %v", err)
		s.T().Logf("\tRevertData: %s", rpcErr.Data)
	}
}
