package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func TestValidateTxMetadataOptionalAndUTF8(t *testing.T) {
	params := DefaultMemoParams()
	require.NoError(t, ValidateTxMetadata(TxMetadata{}, params))
	require.NoError(t, ValidateTxMetadata(TxMetadata{Memo: "hello Aetra", MemoVisible: true}, params))
	require.ErrorContains(t, ValidateTxMetadata(TxMetadata{Memo: string([]byte{0xff})}, params), "UTF-8")
}

func TestValidateTxMetadataBoundsAndControlChars(t *testing.T) {
	params := DefaultMemoParams()
	params.MaxMemoChars = 3
	params.MaxMemoBytes = 12

	require.NoError(t, ValidateTxMetadata(TxMetadata{Memo: "abc"}, params))
	require.ErrorContains(t, ValidateTxMetadata(TxMetadata{Memo: "abcd"}, params), "character count")

	params.MaxMemoChars = 10
	params.MaxMemoBytes = 8
	require.ErrorContains(t, ValidateTxMetadata(TxMetadata{Memo: "abcdefghi"}, params), "byte length")
	require.ErrorContains(t, ValidateTxMetadata(TxMetadata{Memo: "hello\n"}, DefaultMemoParams()), "control character")
}

func TestValidateTxMetadataHash(t *testing.T) {
	params := DefaultMemoParams()
	memo := "visible note"
	require.NoError(t, ValidateTxMetadata(TxMetadata{Memo: memo, MemoHash: MemoHash(memo), MemoVisible: true}, params))
	require.NoError(t, ValidateTxMetadata(TxMetadata{MemoHash: MemoHash("")}, params))
	require.ErrorContains(t, ValidateTxMetadata(TxMetadata{Memo: memo, MemoHash: []byte{1, 2, 3}}, params), "32 bytes")
	require.ErrorContains(t, ValidateTxMetadata(TxMetadata{Memo: memo, MemoHash: MemoHash("other")}, params), "does not match")
}

func TestValidateMemoParamsHardBounds(t *testing.T) {
	params := DefaultMemoParams()
	require.NoError(t, ValidateMemoParams(params))

	params.MaxMemoChars = HardMaxMemoChars + 1
	require.ErrorContains(t, ValidateMemoParams(params), "hard bound")

	params = DefaultMemoParams()
	params.MaxMemoBytes = HardMaxMemoBytes + 1
	require.ErrorContains(t, ValidateMemoParams(params), "hard bound")

	params = DefaultMemoParams()
	params.MemoByteFee = params.MemoByteFee.Neg()
	require.ErrorContains(t, ValidateMemoParams(params), "byte fee")
}

func TestMemoFee(t *testing.T) {
	params := DefaultMemoParams()
	params.MemoBaseFee = params.MemoBaseFee.AddRaw(10)
	params.MemoByteFee = params.MemoByteFee.AddRaw(1)

	fee, denom, err := MemoFee(TxMetadata{}, params, 0, DefaultCongestionBps)
	require.NoError(t, err)
	require.True(t, fee.IsZero())
	require.Equal(t, appparams.BaseDenom, denom)

	fee, denom, err = MemoFee(TxMetadata{Memo: strings.Repeat("a", 10)}, params, 50, DefaultCongestionBps)
	require.NoError(t, err)
	require.Equal(t, "30", fee.String())
	require.Equal(t, appparams.BaseDenom, denom)

	fee, _, err = MemoFee(TxMetadata{Memo: strings.Repeat("a", 10)}, params, 80, DefaultCongestionBps)
	require.NoError(t, err)
	require.Equal(t, "23", fee.String())

	fee, _, err = MemoFee(TxMetadata{Memo: strings.Repeat("a", 10)}, params, 10, 20_000)
	require.NoError(t, err)
	require.Equal(t, "180", fee.String())

	_, _, err = MemoFee(TxMetadata{Memo: "x"}, params, 50, 0)
	require.ErrorContains(t, err, "congestion multiplier")
}

func TestMemoMetadataDoesNotAffectExecution(t *testing.T) {
	require.False(t, MetadataAffectsExecution(TxMetadata{Memo: "note", MemoVisible: true}))
}
