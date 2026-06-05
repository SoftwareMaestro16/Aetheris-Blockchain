package async

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
)

func DefaultParams() Params {
	return Params{
		MaxMessagesPerTx:           32,
		MaxMessagesPerBlock:        128,
		MaxRecursionDepth:          8,
		MaxBodySize:                4096,
		MaxStateSize:               64 * 1024,
		MaxContractDeploysPerTx:    4,
		MaxContractDeploysPerBlock: 16,
		MaxEmittedMessagesPerExec:  16,
		MaxStorageWritesPerExec:    64,
		ExecutionGasPerMessage:     10_000,
		StorageFeePerByte:          sdkmath.NewInt(1),
		ForwardingFee:              sdkmath.NewInt(1),
		ContractDeploymentCost:     sdkmath.NewInt(1_000),
	}
}

func (p Params) Validate() error {
	if p.MaxMessagesPerTx == 0 {
		return errors.New("max messages per tx must be positive")
	}
	if p.MaxMessagesPerBlock == 0 {
		return errors.New("max messages per block must be positive")
	}
	if p.MaxRecursionDepth == 0 {
		return errors.New("max recursion depth must be positive")
	}
	if p.MaxBodySize == 0 {
		return errors.New("max body size must be positive")
	}
	if p.MaxStateSize == 0 {
		return errors.New("max state size must be positive")
	}
	if p.MaxContractDeploysPerTx == 0 {
		return errors.New("max contract deploys per tx must be positive")
	}
	if p.MaxContractDeploysPerBlock == 0 {
		return errors.New("max contract deploys per block must be positive")
	}
	if p.MaxEmittedMessagesPerExec == 0 {
		return errors.New("max emitted messages per execution must be positive")
	}
	if p.MaxStorageWritesPerExec == 0 {
		return errors.New("max storage writes per execution must be positive")
	}
	if p.ExecutionGasPerMessage == 0 {
		return errors.New("execution gas per message must be positive")
	}
	for _, item := range []struct {
		name  string
		value sdkmath.Int
	}{
		{name: "storage fee per byte", value: p.StorageFeePerByte},
		{name: "forwarding fee", value: p.ForwardingFee},
		{name: "contract deployment cost", value: p.ContractDeploymentCost},
	} {
		if item.value.IsNil() || item.value.IsNegative() {
			return fmt.Errorf("%s must be non-negative", item.name)
		}
	}
	return nil
}
