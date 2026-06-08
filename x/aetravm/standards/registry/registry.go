package registry

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	"github.com/sovereign-l1/l1/x/aetravm/standards/aft"
	"github.com/sovereign-l1/l1/x/aetravm/standards/anft"
	"github.com/sovereign-l1/l1/x/aetravm/standards/aw"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const (
	StandardAFT44  = "AFT-44"
	StandardANFT66 = "ANFT-66"
	StandardAW5    = "AW-5"
	StandardADEX1  = "ADEX-1"
	StandardADNS1  = "ADNS-1"

	CompatibilityCompatible   = "compatible"
	CompatibilityDeprecated   = "deprecated"
	CompatibilityIncompatible = "incompatible"

	registryRootDomain = "aetra-avm-standards-registry-v1"
	schemaHashDomain   = "aetra-avm-standard-schema-v1"
)

type Params struct {
	Authority string
	Enabled   bool
}

type StandardSpec struct {
	StandardID          string
	Version             uint32
	CodeHash            string
	ABISchemaHash       string
	RequiredOpcodes     []uint32
	RequiredGetters     []string
	RequiredEvents      []string
	StorageSchemaHash   string
	CompatibilityStatus string
	Enabled             bool
}

type StandardMetadata struct {
	CodeHash          string
	ABISchemaHash     string
	Opcodes           []uint32
	Getters           []string
	Events            []string
	StorageSchemaHash string
}

type StandardTemplate struct {
	Spec     StandardSpec
	Metadata StandardMetadata
}

type RegistryState struct {
	Standards []StandardSpec
	Root      string
}

type Registry struct {
	params Params
	state  RegistryState
}

type QueryStandardsRequest struct {
	IncludeDisabled bool
}

type DeployStandardRequest struct {
	Creator        string
	StandardID     string
	Version        uint32
	ChainID        string
	Namespace      string
	Salt           string
	InitPayload    []byte
	InitialBalance uint64
	Admin          string
	Height         uint64
}

type TemplateDeployer interface {
	DeployContract(contractstypes.MsgDeployContract) (contractstypes.InstantiateContractResponse, error)
}

func DefaultParams() Params {
	return Params{
		Authority: prototype.DefaultAuthority,
		Enabled:   true,
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("AVM standards registry authority", p.Authority); err != nil {
		return err
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("AVM standards registry update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("AVM standards registry update requires governance/system authority")
	}
	return nil
}

func NewRegistry(params Params) (*Registry, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	state := RegistryState{}.Normalize()
	return &Registry{params: params, state: state}, nil
}

func ImportRegistry(params Params, state RegistryState) (*Registry, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	normalized, err := validateAndNormalizeState(state)
	if err != nil {
		return nil, err
	}
	if state.Root != "" && normalized.Root != state.Root {
		return nil, errors.New("AVM standards registry root mismatch")
	}
	return &Registry{params: params, state: normalized}, nil
}

func (r *Registry) RegisterStandardCode(authority string, template StandardTemplate) error {
	if r == nil {
		return errors.New("AVM standards registry is nil")
	}
	if !r.params.Enabled {
		return errors.New("AVM standards registry is disabled")
	}
	if err := r.params.Authorize(authority); err != nil {
		return err
	}
	spec := template.Spec.Normalize()
	metadata := template.Metadata.Normalize()
	if err := spec.Validate(); err != nil {
		return err
	}
	if err := VerifyStandardMetadata(spec, metadata); err != nil {
		return err
	}
	next := r.state.Normalize()
	next.Standards = upsertStandard(next.Standards, spec)
	normalized, err := validateAndNormalizeState(next)
	if err != nil {
		return err
	}
	r.state = normalized
	return nil
}

func (r *Registry) Standards(req QueryStandardsRequest) []StandardSpec {
	if r == nil {
		return nil
	}
	standards := r.state.Normalize().Standards
	out := make([]StandardSpec, 0, len(standards))
	for _, standard := range standards {
		if !req.IncludeDisabled && !standard.Enabled {
			continue
		}
		out = append(out, standard)
	}
	return out
}

func (r *Registry) Standard(standardID string, version uint32) (StandardSpec, bool) {
	if r == nil {
		return StandardSpec{}, false
	}
	keyID := strings.TrimSpace(standardID)
	for _, standard := range r.state.Normalize().Standards {
		if standard.StandardID == keyID && standard.Version == version {
			return standard, true
		}
	}
	return StandardSpec{}, false
}

func (r *Registry) DeployStandardContract(req DeployStandardRequest, deployer TemplateDeployer) (contractstypes.InstantiateContractResponse, error) {
	if r == nil {
		return contractstypes.InstantiateContractResponse{}, errors.New("AVM standards registry is nil")
	}
	if !r.params.Enabled {
		return contractstypes.InstantiateContractResponse{}, errors.New("AVM standards registry is disabled")
	}
	if deployer == nil {
		return contractstypes.InstantiateContractResponse{}, errors.New("AVM standards registry deployer is required")
	}
	version := req.Version
	if version == 0 {
		version = 1
	}
	standard, found := r.Standard(req.StandardID, version)
	if !found {
		return contractstypes.InstantiateContractResponse{}, errors.New("unknown AVM contract standard")
	}
	if !standard.Enabled {
		return contractstypes.InstantiateContractResponse{}, errors.New("AVM contract standard is disabled")
	}
	if standard.CompatibilityStatus == CompatibilityIncompatible {
		return contractstypes.InstantiateContractResponse{}, errors.New("AVM contract standard is incompatible")
	}
	if strings.TrimSpace(req.Creator) == "" {
		return contractstypes.InstantiateContractResponse{}, errors.New("standard deploy creator is required")
	}
	return deployer.DeployContract(contractstypes.MsgDeployContract{
		Creator:        req.Creator,
		CodeID:         standard.CodeHash,
		ChainID:        req.ChainID,
		Namespace:      req.Namespace,
		Salt:           req.Salt,
		InitPayload:    append([]byte(nil), req.InitPayload...),
		InitialBalance: req.InitialBalance,
		Admin:          req.Admin,
		Height:         req.Height,
	})
}

func (r *Registry) ExportState() RegistryState {
	if r == nil {
		return RegistryState{}.Normalize()
	}
	return r.state.Normalize()
}

func DefaultStandardTemplate(standardID string, version uint32, codeHash string) (StandardTemplate, error) {
	requirements, err := DefaultRequirements(standardID)
	if err != nil {
		return StandardTemplate{}, err
	}
	if version == 0 {
		version = 1
	}
	spec := StandardSpec{
		StandardID:          requirements.StandardID,
		Version:             version,
		CodeHash:            strings.TrimSpace(codeHash),
		ABISchemaHash:       ComputeSchemaHash(requirements.StandardID, version, "abi", requirements.RequiredGetters, requirements.RequiredEvents, requirements.RequiredOpcodes),
		RequiredOpcodes:     append([]uint32(nil), requirements.RequiredOpcodes...),
		RequiredGetters:     append([]string(nil), requirements.RequiredGetters...),
		RequiredEvents:      append([]string(nil), requirements.RequiredEvents...),
		StorageSchemaHash:   ComputeSchemaHash(requirements.StandardID, version, "storage", requirements.RequiredGetters, requirements.RequiredEvents, requirements.RequiredOpcodes),
		CompatibilityStatus: CompatibilityCompatible,
		Enabled:             true,
	}.Normalize()
	metadata := StandardMetadata{
		CodeHash:          spec.CodeHash,
		ABISchemaHash:     spec.ABISchemaHash,
		Opcodes:           append([]uint32(nil), spec.RequiredOpcodes...),
		Getters:           append([]string(nil), spec.RequiredGetters...),
		Events:            append([]string(nil), spec.RequiredEvents...),
		StorageSchemaHash: spec.StorageSchemaHash,
	}.Normalize()
	return StandardTemplate{Spec: spec, Metadata: metadata}, nil
}

func DefaultRequirements(standardID string) (StandardSpec, error) {
	switch strings.TrimSpace(standardID) {
	case StandardAFT44:
		return requirementSpec(StandardAFT44,
			[]uint32{aft.OpcodeMint, aft.OpcodeTransfer, aft.OpcodeBurn, aft.OpcodeChangeAdmin, aft.OpcodeRenounce, aft.OpcodeMetadata},
			[]string{"token_info", "total_supply", "wallet_address", "balance_of"},
			[]string{"TokenMinted", "TokenTransferred", "TokenBurned", "AdminChanged", "MetadataChanged"}), nil
	case StandardANFT66:
		return requirementSpec(StandardANFT66,
			[]uint32{anft.OpcodeMintNFT, anft.OpcodeTransfer, anft.OpcodeMintSBT, anft.OpcodeRevokeSBT, anft.OpcodeProveSBT},
			[]string{"collection_info", "item_info", "owner_of", "royalty_policy"},
			[]string{"CollectionCreated", "ItemMinted", "ItemTransferred", "SoulboundRevoked"}), nil
	case StandardAW5:
		return requirementSpec(StandardAW5,
			[]uint32{aw.OpcodeSignedExternal},
			[]string{"wallet_state", "seqno", "extensions", "recovery_policy"},
			[]string{"WalletDeployed", "ExternalCommandAccepted", "ExtensionInstalled", "RecoveryUpdated"}), nil
	case StandardADEX1:
		return requirementSpec(StandardADEX1,
			[]uint32{0x4445_5801, 0x4445_5802, 0x4445_5803, 0x4445_5804, 0x4445_5805},
			[]string{"factory_info", "pool_info", "reserves", "quote"},
			[]string{"FactoryDeployed", "PoolCreated", "LiquidityAdded", "LiquidityRemoved", "SwapExecuted"}), nil
	case StandardADNS1:
		return requirementSpec(StandardADNS1,
			[]uint32{0x444e_5301, 0x444e_5302, 0x444e_5303, 0x444e_5304},
			[]string{"domain_record", "resolver", "owner", "expiry"},
			[]string{"DomainMinted", "ResolverUpdated", "DomainTransferred", "DomainExpired"}), nil
	default:
		return StandardSpec{}, errors.New("unknown AVM contract standard")
	}
}

func VerifyStandardMetadata(spec StandardSpec, metadata StandardMetadata) error {
	spec = spec.Normalize()
	metadata = metadata.Normalize()
	if spec.CodeHash != metadata.CodeHash {
		return errors.New("AVM standard metadata code hash mismatch")
	}
	if spec.ABISchemaHash != metadata.ABISchemaHash {
		return errors.New("AVM standard metadata ABI schema hash mismatch")
	}
	if spec.StorageSchemaHash != metadata.StorageSchemaHash {
		return errors.New("AVM standard metadata storage schema hash mismatch")
	}
	if missing := missingOpcodes(spec.RequiredOpcodes, metadata.Opcodes); len(missing) > 0 {
		return fmt.Errorf("AVM standard metadata missing opcodes %v", missing)
	}
	if missing := missingStrings(spec.RequiredGetters, metadata.Getters); len(missing) > 0 {
		return fmt.Errorf("AVM standard metadata missing getters %v", missing)
	}
	if missing := missingStrings(spec.RequiredEvents, metadata.Events); len(missing) > 0 {
		return fmt.Errorf("AVM standard metadata missing events %v", missing)
	}
	return nil
}

func ValidateNoNativeAssetModules(moduleNames []string) error {
	for _, name := range moduleNames {
		normalized := normalizeModuleName(name)
		if _, forbidden := forbiddenNativeAssetModules()[normalized]; forbidden {
			return fmt.Errorf("native asset module %q is forbidden; use AVM contract standards", name)
		}
	}
	return nil
}

func (s StandardSpec) Validate() error {
	s = s.Normalize()
	if _, err := DefaultRequirements(s.StandardID); err != nil {
		return err
	}
	if s.Version == 0 {
		return errors.New("AVM standard version must be positive")
	}
	if err := coretypes.ValidateHash("AVM standard code hash", s.CodeHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("AVM standard ABI schema hash", s.ABISchemaHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("AVM standard storage schema hash", s.StorageSchemaHash); err != nil {
		return err
	}
	if len(s.RequiredOpcodes) == 0 {
		return errors.New("AVM standard required opcodes must not be empty")
	}
	if len(s.RequiredGetters) == 0 {
		return errors.New("AVM standard required getters must not be empty")
	}
	if len(s.RequiredEvents) == 0 {
		return errors.New("AVM standard required events must not be empty")
	}
	switch s.CompatibilityStatus {
	case CompatibilityCompatible, CompatibilityDeprecated, CompatibilityIncompatible:
	default:
		return errors.New("AVM standard compatibility status is invalid")
	}
	return nil
}

func (s StandardSpec) Normalize() StandardSpec {
	s.StandardID = strings.TrimSpace(s.StandardID)
	s.CodeHash = strings.TrimSpace(s.CodeHash)
	s.ABISchemaHash = strings.TrimSpace(s.ABISchemaHash)
	s.StorageSchemaHash = strings.TrimSpace(s.StorageSchemaHash)
	s.CompatibilityStatus = strings.TrimSpace(s.CompatibilityStatus)
	if s.CompatibilityStatus == "" {
		s.CompatibilityStatus = CompatibilityCompatible
	}
	s.RequiredOpcodes = normalizeOpcodes(s.RequiredOpcodes)
	s.RequiredGetters = normalizeStrings(s.RequiredGetters)
	s.RequiredEvents = normalizeStrings(s.RequiredEvents)
	return s
}

func (m StandardMetadata) Normalize() StandardMetadata {
	m.CodeHash = strings.TrimSpace(m.CodeHash)
	m.ABISchemaHash = strings.TrimSpace(m.ABISchemaHash)
	m.StorageSchemaHash = strings.TrimSpace(m.StorageSchemaHash)
	m.Opcodes = normalizeOpcodes(m.Opcodes)
	m.Getters = normalizeStrings(m.Getters)
	m.Events = normalizeStrings(m.Events)
	return m
}

func (s RegistryState) Normalize() RegistryState {
	out := RegistryState{Standards: make([]StandardSpec, len(s.Standards))}
	for i, standard := range s.Standards {
		out.Standards[i] = standard.Normalize()
	}
	sort.SliceStable(out.Standards, func(i, j int) bool {
		if out.Standards[i].StandardID != out.Standards[j].StandardID {
			return out.Standards[i].StandardID < out.Standards[j].StandardID
		}
		return out.Standards[i].Version < out.Standards[j].Version
	})
	out.Root = ComputeRegistryRoot(out)
	return out
}

func ComputeRegistryRoot(state RegistryState) string {
	state = RegistryState{Standards: append([]StandardSpec(nil), state.Standards...)}
	sort.SliceStable(state.Standards, func(i, j int) bool {
		if state.Standards[i].StandardID != state.Standards[j].StandardID {
			return state.Standards[i].StandardID < state.Standards[j].StandardID
		}
		return state.Standards[i].Version < state.Standards[j].Version
	})
	buf := bytes.NewBuffer(nil)
	writeString(buf, registryRootDomain)
	writeU32(buf, uint32(len(state.Standards)))
	for _, standard := range state.Standards {
		writeStandard(buf, standard.Normalize())
	}
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:])
}

func ComputeSchemaHash(standardID string, version uint32, kind string, getters []string, events []string, opcodes []uint32) string {
	buf := bytes.NewBuffer(nil)
	writeString(buf, schemaHashDomain)
	writeString(buf, strings.TrimSpace(standardID))
	writeU32(buf, version)
	writeString(buf, strings.TrimSpace(kind))
	for _, opcode := range normalizeOpcodes(opcodes) {
		writeU32(buf, opcode)
	}
	for _, getter := range normalizeStrings(getters) {
		writeString(buf, getter)
	}
	for _, event := range normalizeStrings(events) {
		writeString(buf, event)
	}
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:])
}

func validateAndNormalizeState(state RegistryState) (RegistryState, error) {
	normalized := state.Normalize()
	seen := make(map[string]struct{}, len(normalized.Standards))
	for _, standard := range normalized.Standards {
		if err := standard.Validate(); err != nil {
			return RegistryState{}, err
		}
		key := standardKey(standard.StandardID, standard.Version)
		if _, exists := seen[key]; exists {
			return RegistryState{}, fmt.Errorf("duplicate AVM standard registration %s", key)
		}
		seen[key] = struct{}{}
	}
	return normalized, nil
}

func upsertStandard(standards []StandardSpec, standard StandardSpec) []StandardSpec {
	out := append([]StandardSpec(nil), standards...)
	for i := range out {
		if out[i].StandardID == standard.StandardID && out[i].Version == standard.Version {
			out[i] = standard
			return out
		}
	}
	return append(out, standard)
}

func requirementSpec(id string, opcodes []uint32, getters []string, events []string) StandardSpec {
	return StandardSpec{
		StandardID:      id,
		RequiredOpcodes: normalizeOpcodes(opcodes),
		RequiredGetters: normalizeStrings(getters),
		RequiredEvents:  normalizeStrings(events),
	}
}

func writeStandard(buf *bytes.Buffer, standard StandardSpec) {
	writeString(buf, standard.StandardID)
	writeU32(buf, standard.Version)
	writeString(buf, standard.CodeHash)
	writeString(buf, standard.ABISchemaHash)
	writeU32(buf, uint32(len(standard.RequiredOpcodes)))
	for _, opcode := range standard.RequiredOpcodes {
		writeU32(buf, opcode)
	}
	writeU32(buf, uint32(len(standard.RequiredGetters)))
	for _, getter := range standard.RequiredGetters {
		writeString(buf, getter)
	}
	writeU32(buf, uint32(len(standard.RequiredEvents)))
	for _, event := range standard.RequiredEvents {
		writeString(buf, event)
	}
	writeString(buf, standard.StorageSchemaHash)
	writeString(buf, standard.CompatibilityStatus)
	writeBool(buf, standard.Enabled)
}

func missingOpcodes(required []uint32, available []uint32) []uint32 {
	have := make(map[uint32]struct{}, len(available))
	for _, opcode := range available {
		have[opcode] = struct{}{}
	}
	var missing []uint32
	for _, opcode := range required {
		if _, ok := have[opcode]; !ok {
			missing = append(missing, opcode)
		}
	}
	return missing
}

func missingStrings(required []string, available []string) []string {
	have := make(map[string]struct{}, len(available))
	for _, value := range available {
		have[value] = struct{}{}
	}
	var missing []string
	for _, value := range required {
		if _, ok := have[value]; !ok {
			missing = append(missing, value)
		}
	}
	return missing
}

func normalizeOpcodes(values []uint32) []uint32 {
	out := append([]uint32(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return compactOpcodes(out)
}

func compactOpcodes(values []uint32) []uint32 {
	if len(values) < 2 {
		return values
	}
	out := values[:1]
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

func standardKey(id string, version uint32) string {
	return fmt.Sprintf("%s/%d", strings.TrimSpace(id), version)
}

func forbiddenNativeAssetModules() map[string]struct{} {
	return map[string]struct{}{
		"token":        {},
		"tokens":       {},
		"nft":          {},
		"nfts":         {},
		"dex":          {},
		"amm":          {},
		"asset":        {},
		"assets":       {},
		"native-token": {},
		"native-nft":   {},
		"native-dex":   {},
		"x/token":      {},
		"x/tokens":     {},
		"x/nft":        {},
		"x/nfts":       {},
		"x/dex":        {},
		"x/amm":        {},
	}
}

func normalizeModuleName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(name, "\\", "/")))
	normalized = strings.TrimPrefix(normalized, "./")
	return normalized
}

func writeString(buf *bytes.Buffer, value string) {
	bz := []byte(value)
	writeU32(buf, uint32(len(bz)))
	buf.Write(bz)
}

func writeU32(buf *bytes.Buffer, value uint32) {
	var bz [4]byte
	binary.BigEndian.PutUint32(bz[:], value)
	buf.Write(bz[:])
}

func writeBool(buf *bytes.Buffer, value bool) {
	if value {
		buf.WriteByte(1)
		return
	}
	buf.WriteByte(0)
}
