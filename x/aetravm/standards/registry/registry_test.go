package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
)

func TestRegisterAFT44Standard(t *testing.T) {
	reg := newTestRegistry(t)
	template := aft44Template(t, "aft44-code")

	require.NoError(t, reg.RegisterStandardCode(DefaultParams().Authority, template))

	registered, found := reg.Standard(StandardAFT44, 1)
	require.True(t, found)
	require.Equal(t, template.Spec, registered)
	require.Len(t, reg.Standards(QueryStandardsRequest{}), 1)
	require.NotEmpty(t, reg.ExportState().Root)
}

func TestDeployTokenMasterFromAFT44Standard(t *testing.T) {
	reg := newTestRegistry(t)
	template := aft44Template(t, "aft44-code")
	require.NoError(t, reg.RegisterStandardCode(DefaultParams().Authority, template))
	deployer := &capturingDeployer{}

	response, err := reg.DeployStandardContract(DeployStandardRequest{
		Creator:        aeAddress(0x11),
		StandardID:     StandardAFT44,
		Version:        1,
		ChainID:        "aetra-test",
		Namespace:      "tokens",
		Salt:           "token-master",
		InitPayload:    []byte("token-init"),
		InitialBalance: 25,
		Admin:          aeAddress(0x12),
		Height:         7,
	}, deployer)
	require.NoError(t, err)
	require.Equal(t, "AEtokenmaster", response.ContractAddressUser)
	require.Len(t, deployer.messages, 1)

	msg := deployer.messages[0]
	require.Equal(t, template.Spec.CodeHash, msg.CodeID)
	require.Equal(t, []byte("token-init"), msg.InitPayload)
	require.Equal(t, uint64(25), msg.InitialBalance)
	require.Equal(t, uint64(7), msg.Height)
}

func TestUnknownStandardRejected(t *testing.T) {
	reg := newTestRegistry(t)
	template, err := DefaultStandardTemplate("UNKNOWN", 1, testHash("unknown"))
	require.ErrorContains(t, err, "unknown AVM contract standard")
	require.Empty(t, template.Spec.StandardID)

	_, err = reg.DeployStandardContract(DeployStandardRequest{
		Creator:    aeAddress(0x11),
		StandardID: "UNKNOWN",
		Version:    1,
		Height:     1,
	}, &capturingDeployer{})
	require.ErrorContains(t, err, "unknown AVM contract standard")
}

func TestDisabledStandardCannotDeploy(t *testing.T) {
	reg := newTestRegistry(t)
	template := aft44Template(t, "disabled-aft44")
	template.Spec.Enabled = false
	require.NoError(t, reg.RegisterStandardCode(DefaultParams().Authority, template))

	_, err := reg.DeployStandardContract(DeployStandardRequest{
		Creator:    aeAddress(0x11),
		StandardID: StandardAFT44,
		Version:    1,
		Height:     1,
	}, &capturingDeployer{})
	require.ErrorContains(t, err, "standard is disabled")
	require.Empty(t, reg.Standards(QueryStandardsRequest{}))
	require.Len(t, reg.Standards(QueryStandardsRequest{IncludeDisabled: true}), 1)
}

func TestDisabledRegistryCannotDeploy(t *testing.T) {
	params := DefaultParams()
	reg, err := NewRegistry(params)
	require.NoError(t, err)
	require.NoError(t, reg.RegisterStandardCode(params.Authority, aft44Template(t, "aft44-code")))
	reg.params.Enabled = false

	_, err = reg.DeployStandardContract(DeployStandardRequest{
		Creator:    aeAddress(0x11),
		StandardID: StandardAFT44,
		Version:    1,
		Height:     1,
	}, &capturingDeployer{})
	require.ErrorContains(t, err, "registry is disabled")
}

func TestStandardRegistryExportImportStable(t *testing.T) {
	reg := newTestRegistry(t)
	aftTemplate := aft44Template(t, "aft44-code")
	awTemplate, err := DefaultStandardTemplate(StandardAW5, 1, testHash("aw5-code"))
	require.NoError(t, err)
	require.NoError(t, reg.RegisterStandardCode(DefaultParams().Authority, awTemplate))
	require.NoError(t, reg.RegisterStandardCode(DefaultParams().Authority, aftTemplate))

	exported := reg.ExportState()
	require.Equal(t, StandardAFT44, exported.Standards[0].StandardID)
	require.Equal(t, StandardAW5, exported.Standards[1].StandardID)

	imported, err := ImportRegistry(DefaultParams(), exported)
	require.NoError(t, err)
	require.Equal(t, exported, imported.ExportState())
	require.Equal(t, exported.Root, ComputeRegistryRoot(exported))
}

func TestRegisterStandardVerifiesRequiredMetadata(t *testing.T) {
	reg := newTestRegistry(t)
	template := aft44Template(t, "aft44-code")
	template.Metadata.Opcodes = template.Metadata.Opcodes[:len(template.Metadata.Opcodes)-1]

	err := reg.RegisterStandardCode(DefaultParams().Authority, template)
	require.ErrorContains(t, err, "missing opcodes")
}

func TestRegisterRequiresGovernanceOrSystemAuthority(t *testing.T) {
	reg := newTestRegistry(t)
	template := aft44Template(t, "aft44-code")

	err := reg.RegisterStandardCode("4:0000000000000000000000000000000000000000000000000000000000000002", template)
	require.ErrorContains(t, err, "requires governance/system authority")
}

func TestPreventNativeTokenNFTDEXModules(t *testing.T) {
	require.NoError(t, ValidateNoNativeAssetModules([]string{
		"auth",
		"bank",
		"contracts",
		"x/aetravm/standards/aft",
		"x/aetravm/standards/anft",
	}))

	require.ErrorContains(t, ValidateNoNativeAssetModules([]string{"x/token"}), "native asset module")
	require.ErrorContains(t, ValidateNoNativeAssetModules([]string{"native-nft"}), "native asset module")
	require.ErrorContains(t, ValidateNoNativeAssetModules([]string{"dex"}), "native asset module")
}

func TestRegistryImportRejectsRootMismatch(t *testing.T) {
	reg := newTestRegistry(t)
	require.NoError(t, reg.RegisterStandardCode(DefaultParams().Authority, aft44Template(t, "aft44-code")))
	exported := reg.ExportState()
	exported.Root = testHash("wrong-root")

	_, err := ImportRegistry(DefaultParams(), exported)
	require.ErrorContains(t, err, "root mismatch")
}

type capturingDeployer struct {
	messages []contractstypes.MsgDeployContract
}

func (d *capturingDeployer) DeployContract(msg contractstypes.MsgDeployContract) (contractstypes.InstantiateContractResponse, error) {
	d.messages = append(d.messages, msg)
	return contractstypes.InstantiateContractResponse{
		ContractAddressUser: "AEtokenmaster",
		ContractAddressRaw:  "4:tokenmaster",
		Owner:               msg.Creator,
		Admin:               msg.Admin,
		Balance:             msg.InitialBalance,
	}, nil
}

func newTestRegistry(t *testing.T) *Registry {
	t.Helper()
	reg, err := NewRegistry(DefaultParams())
	require.NoError(t, err)
	return reg
}

func aft44Template(t *testing.T, seed string) StandardTemplate {
	t.Helper()
	template, err := DefaultStandardTemplate(StandardAFT44, 1, testHash(seed))
	require.NoError(t, err)
	return template
}

func aeAddress(fill byte) string {
	return addressing.FormatAccAddress(bytesOf(fill, 20))
}

func bytesOf(fill byte, count int) []byte {
	out := make([]byte, count)
	for i := range out {
		out[i] = fill
	}
	return out
}

func testHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}
