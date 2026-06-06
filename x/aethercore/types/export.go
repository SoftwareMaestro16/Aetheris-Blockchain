package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

type ExportManifest struct {
	Height                 uint64
	AppHash                string
	GlobalRoot             string
	ZonesRoot              string
	ServicesRoot           string
	IdentityRoot           string
	StorageRoot            string
	MessageRoot            string
	ReceiptsRoot           string
	PaymentsRoot           string
	VMRoot                 string
	ZoneCommitmentCount    uint64
	ServiceDescriptorCount uint64
	ManifestHash           string
}

func NewExportManifest(root GlobalStateRoot, appHash string, state AetherCoreState) (ExportManifest, error) {
	if err := state.Validate(); err != nil {
		return ExportManifest{}, err
	}
	if err := root.ValidateHash(); err != nil {
		return ExportManifest{}, err
	}
	if err := ValidateHash("aethercore export app hash", appHash); err != nil {
		return ExportManifest{}, err
	}
	manifest := ExportManifest{
		Height:                 root.Height,
		AppHash:                appHash,
		GlobalRoot:             root.GlobalRoot,
		ZonesRoot:              root.ZonesRoot,
		ServicesRoot:           root.ServicesRoot,
		IdentityRoot:           root.IdentityRoot,
		StorageRoot:            root.StorageRoot,
		MessageRoot:            root.MessageRoot,
		ReceiptsRoot:           root.ReceiptsRoot,
		PaymentsRoot:           root.PaymentsRoot,
		VMRoot:                 root.VMRoot,
		ZoneCommitmentCount:    uint64(len(state.CommitmentsAtHeight(root.Height))),
		ServiceDescriptorCount: uint64(len(state.ServiceDescriptors)),
	}
	manifest.ManifestHash = ComputeExportManifestHash(manifest)
	return manifest, nil
}

func (m ExportManifest) ValidateFormat() error {
	if m.Height == 0 {
		return errors.New("aethercore export manifest height must be positive")
	}
	if err := ValidateHash("aethercore export app hash", m.AppHash); err != nil {
		return err
	}
	if err := ValidateHash("aethercore export global root", m.GlobalRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore export zones root", m.ZonesRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore export services root", m.ServicesRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore export identity root", m.IdentityRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore export storage root", m.StorageRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore export message root", m.MessageRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore export receipts root", m.ReceiptsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore export payments root", m.PaymentsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aethercore export VM root", m.VMRoot); err != nil {
		return err
	}
	if m.ManifestHash != "" {
		if err := ValidateHash("aethercore export manifest hash", m.ManifestHash); err != nil {
			return err
		}
	}
	return nil
}

func (m ExportManifest) ValidateHash() error {
	if err := m.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeExportManifestHash(m)
	if m.ManifestHash != expected {
		return fmt.Errorf("aethercore export manifest hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeExportManifestHash(m ExportManifest) string {
	h := sha256.New()
	writePart(h, "aetheris-aek-export-manifest-v1")
	writeUint64(h, m.Height)
	writePart(h, m.AppHash)
	writePart(h, m.GlobalRoot)
	writePart(h, m.ZonesRoot)
	writePart(h, m.ServicesRoot)
	writePart(h, m.IdentityRoot)
	writePart(h, m.StorageRoot)
	writePart(h, m.MessageRoot)
	writePart(h, m.ReceiptsRoot)
	writePart(h, m.PaymentsRoot)
	writePart(h, m.VMRoot)
	writeUint64(h, m.ZoneCommitmentCount)
	writeUint64(h, m.ServiceDescriptorCount)
	return hex.EncodeToString(h.Sum(nil))
}
