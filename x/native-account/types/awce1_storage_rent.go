package types

type StorageRentDebt struct {
	Account          string
	CurrentDebt      uint64
	LastChargeHeight uint64
	AccumulatedRent  uint64
	GenesisFrozen    bool
	GenesisDebt      uint64
}

func (d StorageRentDebt) IsActiveDebt() bool {
	return d.CurrentDebt > 0
}

func (d StorageRentDebt) IsFrozen() bool {
	return d.GenesisFrozen || d.CurrentDebt > 0
}

func NewStorageRentDebt(account string) StorageRentDebt {
	return StorageRentDebt{
		Account:          account,
		CurrentDebt:      0,
		LastChargeHeight: 0,
		AccumulatedRent:  0,
		GenesisFrozen:    false,
		GenesisDebt:      0,
	}
}
