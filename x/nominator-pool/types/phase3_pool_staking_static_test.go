package types

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPhase31DepositMessageSurfaceHasNoValidatorOrRawTargetFields(t *testing.T) {
	msgType := reflect.TypeOf(MsgDepositToStakingPool{})
	surface := strings.ToLower(msgType.String())
	for idx := 0; idx < msgType.NumField(); idx++ {
		field := msgType.Field(idx)
		surface += "\n" + strings.ToLower(field.Name)
		surface += "\n" + strings.ToLower(string(field.Tag))
	}
	for _, forbidden := range []string{
		"validator",
		"operator_address",
		"consensus_address",
		"raw",
		"target",
	} {
		require.NotContains(t, surface, forbidden)
	}
	require.Contains(t, surface, "pool_id")
	require.Contains(t, surface, "wallet_address")
	require.Contains(t, surface, "official_contract")
	require.Contains(t, surface, "amount")
}

func TestPhase31DepositProtoSurfaceHasNoValidatorOrRawTargetFields(t *testing.T) {
	proto := readRepoFile(t, "proto", "l1", "nominatorpool", "v1", "tx.proto")
	block := extractMessageBlock(t, proto, "MsgDepositToStakingPool")
	for _, forbidden := range []string{
		"validator",
		"operator_address",
		"consensus_address",
		"raw",
		"target",
	} {
		require.NotContains(t, strings.ToLower(block), forbidden)
	}
	require.Contains(t, block, "string pool_id")
	require.Contains(t, block, "string wallet_address")
	require.Contains(t, block, "string official_contract")
	require.Contains(t, block, "uint64 amount")
}

func TestPhase31DocsDoNotShowUserFacingValidatorChoiceExamples(t *testing.T) {
	for _, path := range [][]string{
		{"docs", "official-liquid-staking.md"},
		{"docs", "native-account-staking-reputation.md"},
		{"docs", "test-matrix.md"},
	} {
		text := readRepoFile(t, path...)
		for lineNo, line := range strings.Split(text, "\n") {
			lower := strings.ToLower(strings.TrimSpace(line))
			if lower == "" {
				continue
			}
			if !mentionsUserFacingValidatorChoice(lower) {
				continue
			}
			require.Truef(t, negatesValidatorChoice(lower), "%s:%d contains user-facing validator choice example: %s", filepath.Join(path...), lineNo+1, line)
		}
	}
}

func mentionsUserFacingValidatorChoice(line string) bool {
	return strings.Contains(line, "choose a validator") ||
		strings.Contains(line, "select a validator") ||
		strings.Contains(line, "validator selection") ||
		strings.Contains(line, "validator address") ||
		strings.Contains(line, "msgdelegate") ||
		strings.Contains(line, "staking delegate") ||
		strings.Contains(line, "legacy validator address")
}

func negatesValidatorChoice(line string) bool {
	allowedPhrases := []string{
		"does not",
		"do not",
		"not ",
		"no ",
		"without",
		"disabled",
		"not used",
		"must not",
		"are always",
		"reject",
		"rejected",
	}
	for _, phrase := range allowedPhrases {
		if strings.Contains(line, phrase) {
			return true
		}
	}
	return false
}

func extractMessageBlock(t *testing.T, proto string, name string) string {
	t.Helper()
	startMarker := "message " + name + " {"
	start := strings.Index(proto, startMarker)
	require.NotEqual(t, -1, start, "message %s not found", name)
	rest := proto[start:]
	end := strings.Index(rest, "\n}")
	require.NotEqual(t, -1, end, "message %s is not closed", name)
	return rest[:end]
}

func readRepoFile(t *testing.T, parts ...string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
	path := filepath.Join(append([]string{root}, parts...)...)
	bz, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(bz)
}
