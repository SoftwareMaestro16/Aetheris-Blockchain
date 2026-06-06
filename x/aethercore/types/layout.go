package types

import (
	"errors"
	"fmt"
	"sort"
)

type ShardDescriptor struct {
	ShardID          ShardID
	StatePrefix      string
	ParentShardID    ShardID
	ActivationHeight uint64
	ValidatorSetHash string
	Available        bool
}

type ShardLayout struct {
	ZoneID           ZoneID
	LayoutEpoch      uint64
	ActivationHeight uint64
	RoutingSeedHash  string
	ActiveShards     []ShardDescriptor
	LayoutHash       string
}

type RoutingZoneEntry struct {
	ZoneID       ZoneID
	LayoutEpoch  uint64
	ActiveShards uint32
	LayoutHash   string
}

type RoutingTableCommitment struct {
	RoutingEpoch uint64
	Height       uint64
	Entries      []RoutingZoneEntry
	TableHash    string
}

func NewShardLayout(zoneID ZoneID, layoutEpoch uint64, activationHeight uint64, routingSeedHash string, shards []ShardDescriptor) (ShardLayout, error) {
	layout := ShardLayout{
		ZoneID:           zoneID,
		LayoutEpoch:      layoutEpoch,
		ActivationHeight: activationHeight,
		RoutingSeedHash:  routingSeedHash,
		ActiveShards:     cloneShardDescriptors(shards),
	}
	sortShardDescriptors(layout.ActiveShards)
	if err := layout.ValidateFormat(); err != nil {
		return ShardLayout{}, err
	}
	layout.LayoutHash = ComputeShardLayoutHash(layout)
	return layout, nil
}

func NewRoutingTableCommitment(routingEpoch uint64, height uint64, entries []RoutingZoneEntry) (RoutingTableCommitment, error) {
	table := RoutingTableCommitment{
		RoutingEpoch: routingEpoch,
		Height:       height,
		Entries:      cloneRoutingZoneEntries(entries),
	}
	sortRoutingZoneEntries(table.Entries)
	if err := table.ValidateFormat(); err != nil {
		return RoutingTableCommitment{}, err
	}
	table.TableHash = ComputeRoutingTableHash(table)
	return table, nil
}

func BuildRoutingTableCommitment(routingEpoch uint64, height uint64, layouts []ShardLayout) (RoutingTableCommitment, error) {
	entries := make([]RoutingZoneEntry, len(layouts))
	for i, layout := range layouts {
		if err := layout.ValidateHash(); err != nil {
			return RoutingTableCommitment{}, err
		}
		entries[i] = RoutingZoneEntry{
			ZoneID:       layout.ZoneID,
			LayoutEpoch:  layout.LayoutEpoch,
			ActiveShards: uint32(len(layout.ActiveShards)),
			LayoutHash:   layout.LayoutHash,
		}
	}
	return NewRoutingTableCommitment(routingEpoch, height, entries)
}

func (d ShardDescriptor) Validate() error {
	if err := ValidateShardID(d.ShardID); err != nil {
		return err
	}
	if err := validateToken("aethercore shard state prefix", d.StatePrefix, MaxScopeLength); err != nil {
		return err
	}
	if d.ParentShardID != "" {
		if err := ValidateShardID(d.ParentShardID); err != nil {
			return err
		}
	}
	if d.ActivationHeight == 0 {
		return errors.New("aethercore shard activation height must be positive")
	}
	return ValidateHash("aethercore shard validator set hash", d.ValidatorSetHash)
}

func (l ShardLayout) ValidateFormat() error {
	if err := ValidateZoneID(l.ZoneID); err != nil {
		return err
	}
	if l.LayoutEpoch == 0 {
		return errors.New("aethercore shard layout epoch must be positive")
	}
	if l.ActivationHeight == 0 {
		return errors.New("aethercore shard layout activation height must be positive")
	}
	if err := ValidateHash("aethercore shard layout routing seed", l.RoutingSeedHash); err != nil {
		return err
	}
	if len(l.ActiveShards) == 0 {
		return errors.New("aethercore shard layout requires active shards")
	}
	if err := validateShardDescriptors(l.ActiveShards); err != nil {
		return err
	}
	if l.LayoutHash != "" {
		return ValidateHash("aethercore shard layout hash", l.LayoutHash)
	}
	return nil
}

func (l ShardLayout) ValidateHash() error {
	if err := l.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeShardLayoutHash(l)
	if l.LayoutHash != expected {
		return fmt.Errorf("aethercore shard layout hash mismatch: expected %s", expected)
	}
	return nil
}

func (e RoutingZoneEntry) Validate() error {
	if err := ValidateZoneID(e.ZoneID); err != nil {
		return err
	}
	if e.LayoutEpoch == 0 {
		return errors.New("aethercore routing entry layout epoch must be positive")
	}
	if e.ActiveShards == 0 {
		return errors.New("aethercore routing entry active shards must be positive")
	}
	return ValidateHash("aethercore routing entry layout hash", e.LayoutHash)
}

func (t RoutingTableCommitment) ValidateFormat() error {
	if t.RoutingEpoch == 0 {
		return errors.New("aethercore routing table epoch must be positive")
	}
	if t.Height == 0 {
		return errors.New("aethercore routing table height must be positive")
	}
	if len(t.Entries) == 0 {
		return errors.New("aethercore routing table requires entries")
	}
	if err := validateRoutingZoneEntries(t.Entries); err != nil {
		return err
	}
	if t.TableHash != "" {
		return ValidateHash("aethercore routing table hash", t.TableHash)
	}
	return nil
}

func (t RoutingTableCommitment) ValidateHash() error {
	if err := t.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeRoutingTableHash(t)
	if t.TableHash != expected {
		return fmt.Errorf("aethercore routing table hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeShardLayoutHash(layout ShardLayout) string {
	ordered := cloneShardDescriptors(layout.ActiveShards)
	sortShardDescriptors(ordered)
	parts := []string{
		"aetheris-aek-shard-layout-v1",
		string(layout.ZoneID),
		fmt.Sprint(layout.LayoutEpoch),
		fmt.Sprint(layout.ActivationHeight),
		layout.RoutingSeedHash,
		fmt.Sprint(len(ordered)),
	}
	for _, shard := range ordered {
		parts = append(parts,
			string(shard.ShardID),
			shard.StatePrefix,
			string(shard.ParentShardID),
			fmt.Sprint(shard.ActivationHeight),
			shard.ValidatorSetHash,
			fmt.Sprint(shard.Available),
		)
	}
	return hashParts(parts...)
}

func ComputeRoutingTableHash(table RoutingTableCommitment) string {
	ordered := cloneRoutingZoneEntries(table.Entries)
	sortRoutingZoneEntries(ordered)
	parts := []string{
		"aetheris-aek-routing-table-v1",
		fmt.Sprint(table.RoutingEpoch),
		fmt.Sprint(table.Height),
		fmt.Sprint(len(ordered)),
	}
	for _, entry := range ordered {
		parts = append(parts,
			string(entry.ZoneID),
			fmt.Sprint(entry.LayoutEpoch),
			fmt.Sprint(entry.ActiveShards),
			entry.LayoutHash,
		)
	}
	return hashParts(parts...)
}

func validateShardDescriptors(shards []ShardDescriptor) error {
	var previous ShardID
	seen := make(map[ShardID]struct{}, len(shards))
	for i, shard := range shards {
		if err := shard.Validate(); err != nil {
			return err
		}
		if _, found := seen[shard.ShardID]; found {
			return fmt.Errorf("duplicate aethercore shard id %s", shard.ShardID)
		}
		seen[shard.ShardID] = struct{}{}
		if i > 0 && previous >= shard.ShardID {
			return errors.New("aethercore shard descriptors must be sorted canonically")
		}
		previous = shard.ShardID
	}
	return nil
}

func validateRoutingZoneEntries(entries []RoutingZoneEntry) error {
	var previous ZoneID
	seen := make(map[ZoneID]struct{}, len(entries))
	for i, entry := range entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if _, found := seen[entry.ZoneID]; found {
			return fmt.Errorf("duplicate aethercore routing zone entry %s", entry.ZoneID)
		}
		seen[entry.ZoneID] = struct{}{}
		if i > 0 && previous >= entry.ZoneID {
			return errors.New("aethercore routing zone entries must be sorted canonically")
		}
		previous = entry.ZoneID
	}
	return nil
}

func sortShardDescriptors(shards []ShardDescriptor) {
	sort.SliceStable(shards, func(i, j int) bool {
		return shards[i].ShardID < shards[j].ShardID
	})
}

func sortShardLayouts(layouts []ShardLayout) {
	sort.SliceStable(layouts, func(i, j int) bool {
		if layouts[i].ZoneID == layouts[j].ZoneID {
			return layouts[i].LayoutEpoch < layouts[j].LayoutEpoch
		}
		return layouts[i].ZoneID < layouts[j].ZoneID
	})
}

func sortRoutingZoneEntries(entries []RoutingZoneEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].ZoneID < entries[j].ZoneID
	})
}

func sortRoutingTables(tables []RoutingTableCommitment) {
	sort.SliceStable(tables, func(i, j int) bool {
		return tables[i].RoutingEpoch < tables[j].RoutingEpoch
	})
}

func cloneShardDescriptors(shards []ShardDescriptor) []ShardDescriptor {
	out := make([]ShardDescriptor, len(shards))
	copy(out, shards)
	return out
}

func cloneShardLayouts(layouts []ShardLayout) []ShardLayout {
	out := make([]ShardLayout, len(layouts))
	for i, layout := range layouts {
		out[i] = layout
		out[i].ActiveShards = cloneShardDescriptors(layout.ActiveShards)
	}
	return out
}

func cloneRoutingZoneEntries(entries []RoutingZoneEntry) []RoutingZoneEntry {
	out := make([]RoutingZoneEntry, len(entries))
	copy(out, entries)
	return out
}

func cloneRoutingTables(tables []RoutingTableCommitment) []RoutingTableCommitment {
	out := make([]RoutingTableCommitment, len(tables))
	for i, table := range tables {
		out[i] = table
		out[i].Entries = cloneRoutingZoneEntries(table.Entries)
	}
	return out
}

func ExportShardLayouts(layouts []ShardLayout) []ShardLayout {
	out := cloneShardLayouts(layouts)
	sortShardLayouts(out)
	return out
}

func ExportRoutingTables(tables []RoutingTableCommitment) []RoutingTableCommitment {
	out := cloneRoutingTables(tables)
	sortRoutingTables(out)
	return out
}
