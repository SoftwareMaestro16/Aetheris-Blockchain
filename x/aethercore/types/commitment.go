package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
)

type ZoneCommitment struct {
	Height               uint64
	ZoneID               ZoneID
	StateRoot            string
	InboxRoot            string
	OutboxRoot           string
	ReceiptsRoot         string
	EventsRoot           string
	ParamsHash           string
	ExecutionSummaryHash string
	CommitmentHash       string
}

type GlobalStateRoot struct {
	Height       uint64
	ZonesRoot    string
	ServicesRoot string
	IdentityRoot string
	StorageRoot  string
	MessageRoot  string
	ReceiptsRoot string
	PaymentsRoot string
	VMRoot       string
	ParamsHash   string
	GlobalRoot   string
}

type RootContributions struct {
	IdentityRoot string
	StorageRoot  string
	MessageRoot  string
	ReceiptsRoot string
	PaymentsRoot string
	VMRoot       string
	ParamsHash   string
}

func NewZoneCommitment(
	height uint64,
	zoneID ZoneID,
	stateRoot string,
	inboxRoot string,
	outboxRoot string,
	receiptsRoot string,
	eventsRoot string,
	paramsHash string,
	executionSummaryHash string,
) (ZoneCommitment, error) {
	commitment := ZoneCommitment{
		Height:               height,
		ZoneID:               zoneID,
		StateRoot:            stateRoot,
		InboxRoot:            inboxRoot,
		OutboxRoot:           outboxRoot,
		ReceiptsRoot:         receiptsRoot,
		EventsRoot:           eventsRoot,
		ParamsHash:           paramsHash,
		ExecutionSummaryHash: executionSummaryHash,
	}
	if err := commitment.ValidateFormat(); err != nil {
		return ZoneCommitment{}, err
	}
	commitment.CommitmentHash = ComputeZoneCommitmentHash(commitment)
	return commitment, nil
}

func (c ZoneCommitment) ValidateFormat() error {
	if c.Height == 0 {
		return errors.New("aethercore zone commitment height must be positive")
	}
	if err := ValidateZoneID(c.ZoneID); err != nil {
		return err
	}
	if err := ValidateHash("aethercore zone state root", c.StateRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore zone inbox root", c.InboxRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore zone outbox root", c.OutboxRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore zone receipts root", c.ReceiptsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore zone events root", c.EventsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore zone params hash", c.ParamsHash); err != nil {
		return err
	}
	if err := ValidateHash("aethercore zone execution summary hash", c.ExecutionSummaryHash); err != nil {
		return err
	}
	if c.CommitmentHash != "" {
		if err := ValidateHash("aethercore zone commitment hash", c.CommitmentHash); err != nil {
			return err
		}
	}
	return nil
}

func (c ZoneCommitment) ValidateHash() error {
	if err := c.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeZoneCommitmentHash(c)
	if c.CommitmentHash != expected {
		return fmt.Errorf("aethercore zone commitment hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeZoneCommitmentHash(c ZoneCommitment) string {
	h := sha256.New()
	writePart(h, "aetheris-aek-zone-commitment-v1")
	writeUint64(h, c.Height)
	writePart(h, string(c.ZoneID))
	writePart(h, c.StateRoot)
	writePart(h, c.InboxRoot)
	writePart(h, c.OutboxRoot)
	writePart(h, c.ReceiptsRoot)
	writePart(h, c.EventsRoot)
	writePart(h, c.ParamsHash)
	writePart(h, c.ExecutionSummaryHash)
	return hex.EncodeToString(h.Sum(nil))
}

func NewGlobalStateRoot(height uint64, zonesRoot string, servicesRoot string, contributions RootContributions) (GlobalStateRoot, error) {
	root := GlobalStateRoot{
		Height:       height,
		ZonesRoot:    zonesRoot,
		ServicesRoot: servicesRoot,
		IdentityRoot: contributions.IdentityRoot,
		StorageRoot:  contributions.StorageRoot,
		MessageRoot:  contributions.MessageRoot,
		ReceiptsRoot: contributions.ReceiptsRoot,
		PaymentsRoot: contributions.PaymentsRoot,
		VMRoot:       contributions.VMRoot,
		ParamsHash:   contributions.ParamsHash,
	}
	if err := root.ValidateFormat(); err != nil {
		return GlobalStateRoot{}, err
	}
	root.GlobalRoot = ComputeGlobalStateRootHash(root)
	return root, nil
}

func (r RootContributions) Validate() error {
	if err := ValidateHash("aethercore identity root", r.IdentityRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore storage root", r.StorageRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore message root", r.MessageRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore receipts root", r.ReceiptsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore payments root", r.PaymentsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore VM root", r.VMRoot); err != nil {
		return err
	}
	return ValidateHash("aethercore params hash", r.ParamsHash)
}

func (r GlobalStateRoot) ValidateFormat() error {
	if r.Height == 0 {
		return errors.New("aethercore global root height must be positive")
	}
	if err := ValidateHash("aethercore zones root", r.ZonesRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore services root", r.ServicesRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore identity root", r.IdentityRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore storage root", r.StorageRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore message root", r.MessageRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore receipts root", r.ReceiptsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore payments root", r.PaymentsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore VM root", r.VMRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore params hash", r.ParamsHash); err != nil {
		return err
	}
	if r.GlobalRoot != "" {
		if err := ValidateHash("aethercore global root", r.GlobalRoot); err != nil {
			return err
		}
	}
	return nil
}

func (r GlobalStateRoot) ValidateHash() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeGlobalStateRootHash(r)
	if r.GlobalRoot != expected {
		return fmt.Errorf("aethercore global root mismatch: expected %s", expected)
	}
	return nil
}

func ComputeGlobalStateRootHash(r GlobalStateRoot) string {
	h := sha256.New()
	writePart(h, "aetheris-aek-global-state-root-v1")
	writeUint64(h, r.Height)
	writePart(h, r.ZonesRoot)
	writePart(h, r.ServicesRoot)
	writePart(h, r.IdentityRoot)
	writePart(h, r.StorageRoot)
	writePart(h, r.MessageRoot)
	writePart(h, r.ReceiptsRoot)
	writePart(h, r.PaymentsRoot)
	writePart(h, r.VMRoot)
	writePart(h, r.ParamsHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeZoneCommitmentsRoot(height uint64, commitments []ZoneCommitment) (string, error) {
	if height == 0 {
		return "", errors.New("aethercore zone commitments root height must be positive")
	}
	ordered := append([]ZoneCommitment(nil), commitments...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].ZoneID == ordered[j].ZoneID {
			return ordered[i].Height < ordered[j].Height
		}
		return ordered[i].ZoneID < ordered[j].ZoneID
	})
	h := sha256.New()
	writePart(h, "aetheris-aek-zone-commitments-root-v1")
	writeUint64(h, height)
	writeUint64(h, uint64(len(ordered)))
	var previous ZoneCommitment
	for i, commitment := range ordered {
		if commitment.Height != height {
			return "", errors.New("aethercore zone commitments root contains different height")
		}
		if err := commitment.ValidateHash(); err != nil {
			return "", err
		}
		if i > 0 && compareZoneCommitments(previous, commitment) >= 0 {
			return "", errors.New("aethercore zone commitments must be sorted canonically")
		}
		writePart(h, commitment.CommitmentHash)
		previous = commitment
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func BuildGlobalStateRoot(height uint64, state AetherCoreState, contributions RootContributions) (GlobalStateRoot, error) {
	if err := state.Validate(); err != nil {
		return GlobalStateRoot{}, err
	}
	if err := contributions.Validate(); err != nil {
		return GlobalStateRoot{}, err
	}
	zonesRoot, err := ComputeZoneCommitmentsRoot(height, state.CommitmentsAtHeight(height))
	if err != nil {
		return GlobalStateRoot{}, err
	}
	servicesRoot, err := ComputeServiceRoot(state.ServiceDescriptors)
	if err != nil {
		return GlobalStateRoot{}, err
	}
	return NewGlobalStateRoot(height, zonesRoot, servicesRoot, contributions)
}

func compareZoneCommitments(left, right ZoneCommitment) int {
	if left.Height < right.Height {
		return -1
	}
	if left.Height > right.Height {
		return 1
	}
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	return 0
}
