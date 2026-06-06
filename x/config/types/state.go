package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const (
	MaxConfigEntriesV1    = uint32(128)
	MaxConfigKeyBytesV1   = uint32(96)
	MaxConfigValueBytesV1 = uint32(4096)
)

type Params struct {
	Authority     string
	MaxEntries    uint32
	MaxKeyBytes   uint32
	MaxValueBytes uint32
}

type ConfigEntry struct {
	Key           string
	Value         string
	Owner         string
	Version       uint64
	UpdatedHeight int64
}

type ConfigState struct {
	Entries []ConfigEntry
}

func DefaultParams() Params {
	return Params{
		Authority:     prototype.DefaultAuthority,
		MaxEntries:    MaxConfigEntriesV1,
		MaxKeyBytes:   MaxConfigKeyBytesV1,
		MaxValueBytes: MaxConfigValueBytesV1,
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("config authority", p.Authority); err != nil {
		return err
	}
	if p.MaxEntries == 0 || p.MaxEntries > MaxConfigEntriesV1 {
		return fmt.Errorf("config max entries must be between 1 and %d", MaxConfigEntriesV1)
	}
	if p.MaxKeyBytes == 0 || p.MaxKeyBytes > MaxConfigKeyBytesV1 {
		return fmt.Errorf("config max key bytes must be between 1 and %d", MaxConfigKeyBytesV1)
	}
	if p.MaxValueBytes == 0 || p.MaxValueBytes > MaxConfigValueBytesV1 {
		return fmt.Errorf("config max value bytes must be between 1 and %d", MaxConfigValueBytesV1)
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("config update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("config update requires governance authority")
	}
	return nil
}

func (e ConfigEntry) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	key := strings.TrimSpace(e.Key)
	if key == "" {
		return errors.New("config entry key must be set")
	}
	if key != e.Key {
		return errors.New("config entry key must be canonical")
	}
	if uint32(len(e.Key)) > params.MaxKeyBytes {
		return fmt.Errorf("config entry key exceeds %d bytes", params.MaxKeyBytes)
	}
	if uint32(len(e.Value)) > params.MaxValueBytes {
		return fmt.Errorf("config entry value exceeds %d bytes", params.MaxValueBytes)
	}
	if err := addressing.ValidateAuthorityAddress("config entry owner", e.Owner); err != nil {
		return err
	}
	if e.Version == 0 {
		return errors.New("config entry version must be positive")
	}
	if e.UpdatedHeight < 0 {
		return errors.New("config entry updated height must be non-negative")
	}
	return nil
}

func (s ConfigState) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint32(len(s.Entries)) > params.MaxEntries {
		return fmt.Errorf("config entries exceed limit %d", params.MaxEntries)
	}
	var previous string
	for i, entry := range s.Entries {
		if err := entry.Validate(params); err != nil {
			return err
		}
		if i > 0 {
			if previous == entry.Key {
				return fmt.Errorf("config entry key %s is duplicated", entry.Key)
			}
			if previous > entry.Key {
				return errors.New("config entries must be sorted by key")
			}
		}
		previous = entry.Key
	}
	return nil
}

func CloneState(state ConfigState) ConfigState {
	out := ConfigState{Entries: make([]ConfigEntry, len(state.Entries))}
	copy(out.Entries, state.Entries)
	return out
}

func SortedEntries(entries []ConfigEntry) []ConfigEntry {
	out := make([]ConfigEntry, len(entries))
	copy(out, entries)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}
