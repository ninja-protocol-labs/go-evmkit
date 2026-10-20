package permit2

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

var testPermit2 = DefaultPermit2Address

func TestExtractNonceInvalidationsSampleLog(t *testing.T) {
	topic0, ok := nonceInvalidationEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint48, abi.Uint48}, big.NewInt(2), big.NewInt(1))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPermit2,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
			types.NewHashFromBytes(testToken.Bytes()),
			types.NewHashFromBytes(testSpender.Bytes()),
		},
		Data: data,
	}

	got, err := ExtractNonceInvalidations([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.True(t, got[0].Contract.Equal(testPermit2))
	require.True(t, got[0].Owner.Equal(testOwner))
	require.True(t, got[0].Token.Equal(testToken))
	require.True(t, got[0].Spender.Equal(testSpender))
	require.Zero(t, big.NewInt(2).Cmp(got[0].NewNonce))
	require.Zero(t, big.NewInt(1).Cmp(got[0].OldNonce))
}

func TestExtractApprovalsSampleLog(t *testing.T) {
	topic0, ok := approvalEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint160, abi.Uint48}, big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPermit2,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
			types.NewHashFromBytes(testToken.Bytes()),
			types.NewHashFromBytes(testSpender.Bytes()),
		},
		Data: data,
	}

	got, err := ExtractApprovals([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.True(t, got[0].Contract.Equal(testPermit2))
	require.True(t, got[0].Owner.Equal(testOwner))
	require.True(t, got[0].Token.Equal(testToken))
	require.True(t, got[0].Spender.Equal(testSpender))
	require.Zero(t, big.NewInt(100).Cmp(got[0].Amount))
	require.Zero(t, big.NewInt(200).Cmp(got[0].Expiration))
}

func TestExtractPermitsSampleLog(t *testing.T) {
	topic0, ok := permitEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint160, abi.Uint48, abi.Uint48}, big.NewInt(100), big.NewInt(200), big.NewInt(1))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPermit2,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
			types.NewHashFromBytes(testToken.Bytes()),
			types.NewHashFromBytes(testSpender.Bytes()),
		},
		Data: data,
	}

	got, err := ExtractPermits([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.True(t, got[0].Contract.Equal(testPermit2))
	require.True(t, got[0].Owner.Equal(testOwner))
	require.True(t, got[0].Token.Equal(testToken))
	require.True(t, got[0].Spender.Equal(testSpender))
	require.Zero(t, big.NewInt(100).Cmp(got[0].Amount))
	require.Zero(t, big.NewInt(200).Cmp(got[0].Expiration))
	require.Zero(t, big.NewInt(1).Cmp(got[0].Nonce))
}

func TestExtractLockdownsSampleLog(t *testing.T) {
	topic0, ok := lockdownEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Address, abi.Address}, testToken, testSpender)
	require.NoError(t, err)

	log := tx.Log{
		Address: testPermit2,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
		},
		Data: data,
	}

	got, err := ExtractLockdowns([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.True(t, got[0].Contract.Equal(testPermit2))
	require.True(t, got[0].Owner.Equal(testOwner))
	require.True(t, got[0].Token.Equal(testToken))
	require.True(t, got[0].Spender.Equal(testSpender))
}

func TestExtractNonceInvalidationsEmptyLogs(t *testing.T) {
	got, err := ExtractNonceInvalidations(nil)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestExtractApprovalsUnrelatedEventSkipped(t *testing.T) {
	unrelated := types.NewHashFromBytes([]byte(strings.Repeat("\x11", 32)))
	logs := []tx.Log{{
		Address: testPermit2,
		Topics:  []*types.Hash{unrelated},
		Data:    []byte{0xde, 0xad},
	}}
	got, err := ExtractApprovals(logs)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestExtractPermitsRejectsNilTopic(t *testing.T) {
	logs := []tx.Log{{
		Address: testPermit2,
		Topics:  []*types.Hash{nil},
		Data:    nil,
	}}
	got, err := ExtractPermits(logs)
	require.NoError(t, err)
	require.Empty(t, got)

	topic0, ok := permitEvent.Topic0()
	require.True(t, ok)
	logs = []tx.Log{{
		Address: testPermit2,
		Topics:  []*types.Hash{types.NewHashFromBytes(topic0[:]), nil, nil, nil},
		Data:    nil,
	}}
	_, err = ExtractPermits(logs)
	require.Error(t, err)
}

func TestExtractLockdownsRejectsMalformedData(t *testing.T) {
	topic0, ok := lockdownEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPermit2,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
		},
		Data: []byte{0xde, 0xad},
	}}
	_, err := ExtractLockdowns(logs)
	require.Error(t, err)
}

func TestExtractNonceInvalidationsRejectsMalformedTopicCount(t *testing.T) {
	topic0, ok := nonceInvalidationEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPermit2,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
		},
		Data: nil,
	}}
	_, err := ExtractNonceInvalidations(logs)
	require.Error(t, err)
}

func TestExtractUnorderedNonceInvalidationsSampleLog(t *testing.T) {
	topic0, ok := unorderedNonceInvalidationEvent.Topic0()
	require.True(t, ok)

	data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256}, big.NewInt(1), big.NewInt(2))
	require.NoError(t, err)

	log := tx.Log{
		Address: testPermit2,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
		},
		Data: data,
	}

	got, err := ExtractUnorderedNonceInvalidations([]tx.Log{log})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.True(t, got[0].Contract.Equal(testPermit2))
	require.True(t, got[0].Owner.Equal(testOwner))
	require.Zero(t, big.NewInt(1).Cmp(got[0].Word))
	require.Zero(t, big.NewInt(2).Cmp(got[0].Mask))
}

func TestExtractUnorderedNonceInvalidationsEmptyLogs(t *testing.T) {
	got, err := ExtractUnorderedNonceInvalidations(nil)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestExtractUnorderedNonceInvalidationsRejectsMalformedData(t *testing.T) {
	topic0, ok := unorderedNonceInvalidationEvent.Topic0()
	require.True(t, ok)
	logs := []tx.Log{{
		Address: testPermit2,
		Topics: []*types.Hash{
			types.NewHashFromBytes(topic0[:]),
			types.NewHashFromBytes(testOwner.Bytes()),
		},
		Data: []byte{0xde, 0xad},
	}}
	_, err := ExtractUnorderedNonceInvalidations(logs)
	require.Error(t, err)
}
