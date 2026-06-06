package types

import (
	"bytes"
	"errors"
	"fmt"
)

type SubdomainDelegationTypeV2 string

const (
	SubdomainDelegationOwnerControlledV2    SubdomainDelegationTypeV2 = "owner_controlled"
	SubdomainDelegationDelegateControlledV2 SubdomainDelegationTypeV2 = "delegate_controlled"
	SubdomainDelegationZoneManagedV2        SubdomainDelegationTypeV2 = "zone_managed"
	SubdomainDelegationDetachedPaidV2       SubdomainDelegationTypeV2 = "detached_paid"
	SubdomainDelegationEphemeralServiceV2   SubdomainDelegationTypeV2 = "ephemeral_service"

	IdentityPathCommitmentVersionV2 uint64 = 1
)

type SubdomainCreationPolicyV2 struct {
	ParentName            string
	Label                 string
	Actor                 []byte
	ChildOwner            []byte
	Height                uint64
	ChildExpiryHeight     uint64
	DelegationType        SubdomainDelegationTypeV2
	ParentControlsRecord  bool
	DetachedPaid          bool
	IndependentPayment    bool
	ParentAuthorization   bool
	Ephemeral             bool
	TimeLockedUntilHeight uint64
	Delegation            *DelegationRecordV2
}

type IdentityPathCommitmentV2 struct {
	CommitmentVersion uint64
	RootName          string
	TargetName        string
	PathLabels        []string
	PathHashes        []string
	PathHash          string
	SourceVersion     uint64
	ParentEpoch       uint64
	ChildEpoch        uint64
	CommitmentHash    string
}

type OptimizedRecursiveResolutionProofRequestV2 struct {
	State         IdentityState
	ChainID       string
	RootName      string
	TargetName    string
	Height        uint64
	TTL           uint64
	Cache         *ResolutionCacheRecordV2
	SourceVersion uint64
	ParentEpoch   uint64
	ChildEpoch    uint64
	LightClient   bool
	ProofVerified bool
}

func ValidateSubdomainCreationV2(state IdentityState, policy SubdomainCreationPolicyV2) (string, error) {
	state = normalizeIdentityStateParams(state)
	if policy.Height == 0 {
		return "", errors.New("identity v2 subdomain creation height is required")
	}
	parent, err := requireActiveDomain(state, policy.ParentName, policy.Height)
	if err != nil {
		return "", err
	}
	if err := validateDomainLabel(policy.Label); err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity v2 subdomain actor", policy.Actor); err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity v2 subdomain child owner", policy.ChildOwner); err != nil {
		return "", err
	}
	delegationType := policy.DelegationType
	if delegationType == "" {
		delegationType = SubdomainDelegationOwnerControlledV2
	}
	if err := validateSubdomainDelegationTypeV2(delegationType); err != nil {
		return "", err
	}
	childName, err := NormalizeAETDomain(policy.Label + "." + parent.Name)
	if err != nil {
		return "", err
	}
	if !IsDomainAvailable(state, childName, policy.Height) {
		return "", errors.New("identity v2 subdomain already exists")
	}
	if err := validateSubdomainAuthorizationForTypeV2(parent, policy, delegationType); err != nil {
		return "", err
	}
	childExpiry := policy.ChildExpiryHeight
	if childExpiry == 0 {
		childExpiry = parent.ExpiryHeight
	}
	if childExpiry <= policy.Height {
		return "", errors.New("identity v2 subdomain expiry must be after creation height")
	}
	if childExpiry > parent.ExpiryHeight && !policy.DetachedPaid {
		return "", errors.New("identity v2 child expiry cannot exceed parent expiry unless detached mode is enabled")
	}
	if policy.DetachedPaid {
		if delegationType != SubdomainDelegationDetachedPaidV2 {
			return "", errors.New("identity v2 detached mode requires detached_paid delegation type")
		}
		if !policy.IndependentPayment || !policy.ParentAuthorization {
			return "", errors.New("identity v2 detached subdomain requires independent payment and explicit parent authorization")
		}
	}
	if delegationType == SubdomainDelegationEphemeralServiceV2 && !policy.Ephemeral {
		return "", errors.New("identity v2 ephemeral service subdomain must be marked ephemeral")
	}
	if policy.TimeLockedUntilHeight != 0 && policy.TimeLockedUntilHeight >= childExpiry {
		return "", errors.New("identity v2 subdomain time lock must end before child expiry")
	}
	return childName, nil
}

func IssueSubdomainV2(state IdentityState, policy SubdomainCreationPolicyV2) (IdentityState, SubdomainRecord, error) {
	childName, err := ValidateSubdomainCreationV2(state, policy)
	if err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	parent, err := requireActiveDomain(state, policy.ParentName, policy.Height)
	if err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	childExpiry := policy.ChildExpiryHeight
	if childExpiry == 0 {
		childExpiry = parent.ExpiryHeight
	}
	delegationType := policy.DelegationType
	if delegationType == "" {
		delegationType = SubdomainDelegationOwnerControlledV2
	}
	nftID, err := DomainNFTID(childName)
	if err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	domain := Domain{Name: childName, Owner: cloneSpecAddress(policy.ChildOwner), NFTID: nftID, RegisteredHeight: policy.Height, ExpiryHeight: childExpiry, UpdatedHeight: policy.Height, ParentName: parent.Name, ParentControlsRecord: policy.ParentControlsRecord}
	nft := DomainNFT{ID: nftID, Domain: childName, Owner: cloneSpecAddress(policy.ChildOwner), MintHeight: policy.Height}
	record := SubdomainRecord{
		ParentName:            parent.Name,
		Name:                  childName,
		Owner:                 cloneSpecAddress(policy.ChildOwner),
		ParentControlsRecord:  policy.ParentControlsRecord,
		CreatedHeight:         policy.Height,
		DelegationType:        delegationType,
		Detached:              policy.DetachedPaid,
		Ephemeral:             policy.Ephemeral,
		ExpiryHeight:          childExpiry,
		TimeLockedUntilHeight: policy.TimeLockedUntilHeight,
		ParentAuthorized:      policy.ParentAuthorization || bytes.Equal(policy.Actor, parent.Owner),
	}
	next := state.Clone()
	next.Domains = upsertDomain(next.Domains, domain)
	next.DomainNFTs = upsertDomainNFT(next.DomainNFTs, nft)
	next.Subdomains = append(next.Subdomains, record)
	sortIdentityState(&next)
	return next, record, next.Validate()
}

func RevokeDelegationV2(records []DelegationRecordV2, name string, delegate []byte, scope DelegationScopeV2, actor []byte, owner []byte, height uint64) ([]DelegationRecordV2, bool, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return nil, false, err
	}
	nameHash, err := DomainRecordV2NameHash(normalized)
	if err != nil {
		return nil, false, err
	}
	if err := validateSpecAddress("identity v2 delegation revocation actor", actor); err != nil {
		return nil, false, err
	}
	if err := validateSpecAddress("identity v2 delegation revocation owner", owner); err != nil {
		return nil, false, err
	}
	if !bytes.Equal(actor, owner) {
		return nil, false, errors.New("identity v2 delegation revocation requires parent owner")
	}
	out := make([]DelegationRecordV2, 0, len(records))
	revoked := false
	for _, record := range records {
		if record.NameHash == nameHash && bytes.Equal(record.Delegate, delegate) && record.Scope == scope {
			if record.TimeLockedUntilHeight != 0 && height < record.TimeLockedUntilHeight {
				return nil, false, errors.New("identity v2 delegation is time-locked and cannot be revoked yet")
			}
			revoked = true
			continue
		}
		out = append(out, cloneDelegationRecordV2(record))
	}
	return out, revoked, nil
}

func BuildIdentityPathCommitmentV2(path DeterministicResolutionPathV2, sourceVersion uint64, parentEpoch uint64, childEpoch uint64) (IdentityPathCommitmentV2, error) {
	if len(path.Path) == 0 || len(path.PathHashes) == 0 || len(path.Path) != len(path.PathHashes) {
		return IdentityPathCommitmentV2{}, errors.New("identity v2 path commitment requires canonical path and hashes")
	}
	pathHash, err := ComputeResolutionPathHashV2(path.Path)
	if err != nil {
		return IdentityPathCommitmentV2{}, err
	}
	commitment := IdentityPathCommitmentV2{
		CommitmentVersion: IdentityPathCommitmentVersionV2,
		RootName:          path.Path[0],
		TargetName:        path.TargetName,
		PathLabels:        append([]string(nil), path.Labels...),
		PathHashes:        append([]string(nil), path.PathHashes...),
		PathHash:          pathHash,
		SourceVersion:     sourceVersion,
		ParentEpoch:       parentEpoch,
		ChildEpoch:        childEpoch,
	}
	commitment.CommitmentHash = ComputeIdentityPathCommitmentHashV2(commitment)
	return commitment, ValidateIdentityPathCommitmentV2(commitment)
}

func ValidateIdentityPathCommitmentV2(commitment IdentityPathCommitmentV2) error {
	if commitment.CommitmentVersion != IdentityPathCommitmentVersionV2 {
		return fmt.Errorf("unsupported identity v2 path commitment version %d", commitment.CommitmentVersion)
	}
	if _, err := NormalizeAETDomain(commitment.RootName); err != nil {
		return err
	}
	if _, err := NormalizeAETDomain(commitment.TargetName); err != nil {
		return err
	}
	if len(commitment.PathLabels) == 0 || len(commitment.PathHashes) == 0 {
		return errors.New("identity v2 path commitment labels and hashes are required")
	}
	for _, hash := range commitment.PathHashes {
		if err := validateHexHash("identity v2 path commitment path hash", hash); err != nil {
			return err
		}
	}
	if err := validateHexHash("identity v2 path commitment path_hash", commitment.PathHash); err != nil {
		return err
	}
	if commitment.SourceVersion == 0 {
		return errors.New("identity v2 path commitment source_version is required")
	}
	if commitment.CommitmentHash == "" || commitment.CommitmentHash != ComputeIdentityPathCommitmentHashV2(commitment) {
		return errors.New("identity v2 path commitment hash mismatch")
	}
	return nil
}

func ComputeIdentityPathCommitmentHashV2(commitment IdentityPathCommitmentV2) string {
	parts := []string{
		"identity-v2-path-commitment",
		fmt.Sprintf("%020d", commitment.CommitmentVersion),
		commitment.RootName,
		commitment.TargetName,
		fmt.Sprintf("%020d", len(commitment.PathLabels)),
	}
	parts = append(parts, commitment.PathLabels...)
	parts = append(parts, fmt.Sprintf("%020d", len(commitment.PathHashes)))
	parts = append(parts, commitment.PathHashes...)
	parts = append(parts,
		commitment.PathHash,
		fmt.Sprintf("%020d", commitment.SourceVersion),
		fmt.Sprintf("%020d", commitment.ParentEpoch),
		fmt.Sprintf("%020d", commitment.ChildEpoch),
	)
	return identityHash(parts...)
}

func InvalidateResolutionCacheRecordV2ForParentEpochChange(record ResolutionCacheRecordV2, parentEpoch uint64) ResolutionCacheRecordV2 {
	next := record
	next.ParentEpoch = parentEpoch
	next.ValidUntilHeight = 0
	return next
}

func BuildOptimizedRecursiveResolutionProofV2(request OptimizedRecursiveResolutionProofRequestV2) (RecursiveResolutionProofV2, IdentityPathCommitmentV2, error) {
	path, err := CanonicalResolutionPathV2(request.TargetName)
	if err != nil {
		return RecursiveResolutionProofV2{}, IdentityPathCommitmentV2{}, err
	}
	commitment, err := BuildIdentityPathCommitmentV2(path, request.SourceVersion, request.ParentEpoch, request.ChildEpoch)
	if err != nil {
		return RecursiveResolutionProofV2{}, IdentityPathCommitmentV2{}, err
	}
	if request.Cache != nil {
		if err := ValidateResolutionCacheRecordV2Use(*request.Cache, ResolutionCacheUseContextV2{
			Height:        request.Height,
			SourceVersion: request.SourceVersion,
			ParentEpoch:   request.ParentEpoch,
			ChildEpoch:    request.ChildEpoch,
			LightClient:   request.LightClient,
			ProofVerified: request.ProofVerified,
		}); err != nil {
			return RecursiveResolutionProofV2{}, IdentityPathCommitmentV2{}, err
		}
		if request.Cache.ResolutionPathHash != commitment.PathHash {
			return RecursiveResolutionProofV2{}, IdentityPathCommitmentV2{}, errors.New("identity v2 optimized recursive proof cache path commitment mismatch")
		}
	}
	proof, err := BuildRecursiveResolutionProofV2(request.State, request.ChainID, request.RootName, request.TargetName, request.Height, request.TTL, request.Cache)
	if err != nil {
		return RecursiveResolutionProofV2{}, IdentityPathCommitmentV2{}, err
	}
	return proof, commitment, nil
}

func validateSubdomainDelegationTypeV2(value SubdomainDelegationTypeV2) error {
	switch value {
	case SubdomainDelegationOwnerControlledV2,
		SubdomainDelegationDelegateControlledV2,
		SubdomainDelegationZoneManagedV2,
		SubdomainDelegationDetachedPaidV2,
		SubdomainDelegationEphemeralServiceV2:
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 subdomain delegation type %q", value)
	}
}

func validateSubdomainAuthorizationForTypeV2(parent Domain, policy SubdomainCreationPolicyV2, delegationType SubdomainDelegationTypeV2) error {
	if bytes.Equal(policy.Actor, parent.Owner) {
		return nil
	}
	if policy.Delegation == nil {
		return errors.New("identity v2 subdomain creation requires parent owner or scoped delegate")
	}
	parentRecord, err := NewDomainRecordV2FromDomain(parent, DomainRecordV2Active, 0, policy.Height)
	if err != nil {
		return err
	}
	switch delegationType {
	case SubdomainDelegationZoneManagedV2:
		if !bytes.Equal(policy.Actor, policy.Delegation.Delegate) {
			return errors.New("identity v2 zone-managed subdomain delegate mismatch")
		}
		if policy.Delegation.NameHash != parentRecord.NameHash {
			return errors.New("identity v2 zone-managed subdomain delegation name_hash mismatch")
		}
		return ValidateDelegationRecordV2Use(*policy.Delegation, DelegationScopeZoneAdmin, "create", policy.Label, 1, policy.Height)
	default:
		return ValidateSubdomainCreationAuthorizationV2(parentRecord, policy.Actor, policy.Delegation, policy.Label, 1, policy.Height)
	}
}
