package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMInterfaceExecutionSync      AVMInterfaceExecutionMode = "sync"
	AVMInterfaceExecutionAsync     AVMInterfaceExecutionMode = "async"
	AVMInterfaceExecutionScheduled AVMInterfaceExecutionMode = "scheduled"
	AVMInterfaceExecutionGet       AVMInterfaceExecutionMode = "get"

	AVMInterfaceTargetContract     AVMInterfaceTargetType = "contract"
	AVMInterfaceTargetService      AVMInterfaceTargetType = "service"
	AVMInterfaceTargetNativeModule AVMInterfaceTargetType = "native_module"
	AVMInterfaceTargetWASM         AVMInterfaceTargetType = "wasm_contract"
	AVMInterfaceTargetActor        AVMInterfaceTargetType = "actor_contract"

	AVMInterfaceSchemaJSONSchema AVMInterfaceSchemaEncoding = "json_schema"
	AVMInterfaceSchemaProtobuf   AVMInterfaceSchemaEncoding = "protobuf"
	AVMInterfaceSchemaTLB        AVMInterfaceSchemaEncoding = "tlb"
	AVMInterfaceSchemaBinary     AVMInterfaceSchemaEncoding = "binary"

	MaxAVMInterfaceTokenLength   = 128
	MaxAVMInterfaceVersionLength = 32
	MaxAVMInterfaceDescriptors   = 512
)

type AVMInterfaceExecutionMode string
type AVMInterfaceTargetType string
type AVMInterfaceSchemaEncoding string

type AVMMethodDescriptor struct {
	MethodID                   string
	Name                       string
	InputSchemaHash            string
	OutputSchemaHash           string
	ExecutionMode              AVMInterfaceExecutionMode
	GasHint                    uint64
	PaymentRequirementOptional string
	ProofRequirementOptional   string
}

type AVMEventDescriptor struct {
	EventID    string
	Name       string
	SchemaHash string
}

type AVMAsyncHandlerDescriptor struct {
	HandlerID           string
	Name                string
	InputSchemaHash     string
	OutputSchemaHash    string
	GasHint             uint64
	RetryPolicyOptional string
}

type AVMGetMethodDescriptor struct {
	MethodID         string
	Name             string
	InputSchemaHash  string
	OutputSchemaHash string
	GasHint          uint64
}

type AVMInterfaceDescriptor struct {
	InterfaceHash           string
	InterfaceVersion        string
	Owner                   string
	TargetType              AVMInterfaceTargetType
	MethodDescriptors       []AVMMethodDescriptor
	EventDescriptors        []AVMEventDescriptor
	AsyncHandlerDescriptors []AVMAsyncHandlerDescriptor
	GetMethodDescriptors    []AVMGetMethodDescriptor
	SchemaEncoding          AVMInterfaceSchemaEncoding
	MetadataHashOptional    string
}

type AVMInterfaceRegistry struct {
	Interfaces []AVMInterfaceDescriptor
	Root       string
}

func NewAVMInterfaceDescriptor(descriptor AVMInterfaceDescriptor) (AVMInterfaceDescriptor, error) {
	descriptor = canonicalAVMInterfaceDescriptor(descriptor)
	descriptor.InterfaceHash = ComputeAVMInterfaceHash(descriptor)
	return descriptor, descriptor.Validate()
}

func NewAVMInterfaceRegistry(registry AVMInterfaceRegistry) (AVMInterfaceRegistry, error) {
	registry = canonicalAVMInterfaceRegistry(registry)
	registry.Root = ComputeAVMInterfaceRegistryRoot(registry)
	return registry, registry.Validate()
}

func (d AVMMethodDescriptor) Validate() error {
	d = canonicalAVMMethodDescriptor(d)
	if err := validateInterfaceToken("AVM method id", d.MethodID); err != nil {
		return err
	}
	if err := validateInterfaceToken("AVM method name", d.Name); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM method input schema hash", d.InputSchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM method output schema hash", d.OutputSchemaHash); err != nil {
		return err
	}
	if !IsAVMInterfaceExecutionMode(d.ExecutionMode) {
		return fmt.Errorf("invalid AVM method execution mode %q", d.ExecutionMode)
	}
	if d.GasHint == 0 {
		return errors.New("AVM method gas hint must be positive")
	}
	if err := validateRouterOptionalToken("AVM method payment requirement", d.PaymentRequirementOptional, MaxAVMInterfaceTokenLength); err != nil {
		return err
	}
	return validateRouterOptionalToken("AVM method proof requirement", d.ProofRequirementOptional, MaxAVMInterfaceTokenLength)
}

func (d AVMEventDescriptor) Validate() error {
	d = canonicalAVMEventDescriptor(d)
	if err := validateInterfaceToken("AVM event id", d.EventID); err != nil {
		return err
	}
	if err := validateInterfaceToken("AVM event name", d.Name); err != nil {
		return err
	}
	return zonestypes.ValidateHash("AVM event schema hash", d.SchemaHash)
}

func (d AVMAsyncHandlerDescriptor) Validate() error {
	d = canonicalAVMAsyncHandlerDescriptor(d)
	if err := validateInterfaceToken("AVM async handler id", d.HandlerID); err != nil {
		return err
	}
	if err := validateInterfaceToken("AVM async handler name", d.Name); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM async handler input schema hash", d.InputSchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM async handler output schema hash", d.OutputSchemaHash); err != nil {
		return err
	}
	if d.GasHint == 0 {
		return errors.New("AVM async handler gas hint must be positive")
	}
	return validateRouterOptionalToken("AVM async handler retry policy", d.RetryPolicyOptional, MaxAVMInterfaceTokenLength)
}

func (d AVMGetMethodDescriptor) Validate() error {
	d = canonicalAVMGetMethodDescriptor(d)
	if err := validateInterfaceToken("AVM get method id", d.MethodID); err != nil {
		return err
	}
	if err := validateInterfaceToken("AVM get method name", d.Name); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM get method input schema hash", d.InputSchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM get method output schema hash", d.OutputSchemaHash); err != nil {
		return err
	}
	if d.GasHint == 0 {
		return errors.New("AVM get method gas hint must be positive")
	}
	return nil
}

func (d AVMInterfaceDescriptor) Validate() error {
	d = canonicalAVMInterfaceDescriptor(d)
	if err := zonestypes.ValidateHash("AVM interface hash", d.InterfaceHash); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM interface version", d.InterfaceVersion, MaxAVMInterfaceVersionLength); err != nil {
		return err
	}
	if d.InterfaceVersion == "" {
		return errors.New("AVM interface version is required")
	}
	if err := validateInterfaceToken("AVM interface owner", d.Owner); err != nil {
		return err
	}
	if !IsAVMInterfaceTargetType(d.TargetType) {
		return fmt.Errorf("invalid AVM interface target type %q", d.TargetType)
	}
	if !IsAVMInterfaceSchemaEncoding(d.SchemaEncoding) {
		return fmt.Errorf("invalid AVM interface schema encoding %q", d.SchemaEncoding)
	}
	if d.MetadataHashOptional != "" {
		if err := zonestypes.ValidateHash("AVM interface metadata hash", d.MetadataHashOptional); err != nil {
			return err
		}
	}
	total := len(d.MethodDescriptors) + len(d.EventDescriptors) + len(d.AsyncHandlerDescriptors) + len(d.GetMethodDescriptors)
	if total == 0 {
		return errors.New("AVM interface descriptor must expose at least one descriptor")
	}
	if total > MaxAVMInterfaceDescriptors {
		return fmt.Errorf("AVM interface descriptor entries must be <= %d", MaxAVMInterfaceDescriptors)
	}
	if err := validateAVMMethods(d.MethodDescriptors); err != nil {
		return err
	}
	if err := validateAVMEvents(d.EventDescriptors); err != nil {
		return err
	}
	if err := validateAVMAsyncHandlers(d.AsyncHandlerDescriptors); err != nil {
		return err
	}
	if err := validateAVMGetMethods(d.GetMethodDescriptors); err != nil {
		return err
	}
	if d.InterfaceHash != ComputeAVMInterfaceHash(d) {
		return errors.New("AVM interface hash mismatch")
	}
	return nil
}

func (r AVMInterfaceRegistry) Validate() error {
	r = canonicalAVMInterfaceRegistry(r)
	if len(r.Interfaces) == 0 {
		return errors.New("AVM interface registry must contain interfaces")
	}
	seen := make(map[string]struct{}, len(r.Interfaces))
	for i, descriptor := range r.Interfaces {
		if err := descriptor.Validate(); err != nil {
			return err
		}
		if _, found := seen[descriptor.InterfaceHash]; found {
			return fmt.Errorf("duplicate AVM interface hash %q", descriptor.InterfaceHash)
		}
		seen[descriptor.InterfaceHash] = struct{}{}
		if i > 0 && r.Interfaces[i-1].InterfaceHash >= descriptor.InterfaceHash {
			return errors.New("AVM interface registry entries must be sorted canonically")
		}
	}
	if err := zonestypes.ValidateHash("AVM interface registry root", r.Root); err != nil {
		return err
	}
	if r.Root != ComputeAVMInterfaceRegistryRoot(r) {
		return errors.New("AVM interface registry root mismatch")
	}
	return nil
}

func IsAVMInterfaceExecutionMode(mode AVMInterfaceExecutionMode) bool {
	switch mode {
	case AVMInterfaceExecutionSync, AVMInterfaceExecutionAsync, AVMInterfaceExecutionScheduled, AVMInterfaceExecutionGet:
		return true
	default:
		return false
	}
}

func IsAVMInterfaceTargetType(target AVMInterfaceTargetType) bool {
	switch target {
	case AVMInterfaceTargetContract, AVMInterfaceTargetService, AVMInterfaceTargetNativeModule, AVMInterfaceTargetWASM, AVMInterfaceTargetActor:
		return true
	default:
		return false
	}
}

func IsAVMInterfaceSchemaEncoding(encoding AVMInterfaceSchemaEncoding) bool {
	switch encoding {
	case AVMInterfaceSchemaJSONSchema, AVMInterfaceSchemaProtobuf, AVMInterfaceSchemaTLB, AVMInterfaceSchemaBinary:
		return true
	default:
		return false
	}
}

func ComputeAVMInterfaceHash(descriptor AVMInterfaceDescriptor) string {
	descriptor = canonicalAVMInterfaceDescriptor(descriptor)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm-interface-descriptor-v1")
	writeEnginePart(h, descriptor.InterfaceVersion)
	writeEnginePart(h, descriptor.Owner)
	writeEnginePart(h, string(descriptor.TargetType))
	writeEngineUint64(h, uint64(len(descriptor.MethodDescriptors)))
	for _, method := range descriptor.MethodDescriptors {
		writeAVMMethodDescriptor(h, method)
	}
	writeEngineUint64(h, uint64(len(descriptor.EventDescriptors)))
	for _, event := range descriptor.EventDescriptors {
		writeEnginePart(h, event.EventID)
		writeEnginePart(h, event.Name)
		writeEnginePart(h, event.SchemaHash)
	}
	writeEngineUint64(h, uint64(len(descriptor.AsyncHandlerDescriptors)))
	for _, handler := range descriptor.AsyncHandlerDescriptors {
		writeEnginePart(h, handler.HandlerID)
		writeEnginePart(h, handler.Name)
		writeEnginePart(h, handler.InputSchemaHash)
		writeEnginePart(h, handler.OutputSchemaHash)
		writeEngineUint64(h, handler.GasHint)
		writeEnginePart(h, handler.RetryPolicyOptional)
	}
	writeEngineUint64(h, uint64(len(descriptor.GetMethodDescriptors)))
	for _, method := range descriptor.GetMethodDescriptors {
		writeEnginePart(h, method.MethodID)
		writeEnginePart(h, method.Name)
		writeEnginePart(h, method.InputSchemaHash)
		writeEnginePart(h, method.OutputSchemaHash)
		writeEngineUint64(h, method.GasHint)
	}
	writeEnginePart(h, string(descriptor.SchemaEncoding))
	writeEnginePart(h, descriptor.MetadataHashOptional)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMInterfaceRegistryRoot(registry AVMInterfaceRegistry) string {
	registry = canonicalAVMInterfaceRegistry(registry)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm-interface-registry-v1")
	writeEngineUint64(h, uint64(len(registry.Interfaces)))
	for _, descriptor := range registry.Interfaces {
		writeEnginePart(h, descriptor.InterfaceHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMInterfaceDescriptor(descriptor AVMInterfaceDescriptor) AVMInterfaceDescriptor {
	descriptor.InterfaceHash = strings.TrimSpace(descriptor.InterfaceHash)
	descriptor.InterfaceVersion = strings.TrimSpace(descriptor.InterfaceVersion)
	descriptor.Owner = strings.TrimSpace(descriptor.Owner)
	descriptor.MetadataHashOptional = strings.TrimSpace(descriptor.MetadataHashOptional)
	descriptor.MethodDescriptors = append([]AVMMethodDescriptor(nil), descriptor.MethodDescriptors...)
	for i := range descriptor.MethodDescriptors {
		descriptor.MethodDescriptors[i] = canonicalAVMMethodDescriptor(descriptor.MethodDescriptors[i])
	}
	sort.SliceStable(descriptor.MethodDescriptors, func(i, j int) bool {
		return descriptor.MethodDescriptors[i].MethodID < descriptor.MethodDescriptors[j].MethodID
	})
	descriptor.EventDescriptors = append([]AVMEventDescriptor(nil), descriptor.EventDescriptors...)
	for i := range descriptor.EventDescriptors {
		descriptor.EventDescriptors[i] = canonicalAVMEventDescriptor(descriptor.EventDescriptors[i])
	}
	sort.SliceStable(descriptor.EventDescriptors, func(i, j int) bool {
		return descriptor.EventDescriptors[i].EventID < descriptor.EventDescriptors[j].EventID
	})
	descriptor.AsyncHandlerDescriptors = append([]AVMAsyncHandlerDescriptor(nil), descriptor.AsyncHandlerDescriptors...)
	for i := range descriptor.AsyncHandlerDescriptors {
		descriptor.AsyncHandlerDescriptors[i] = canonicalAVMAsyncHandlerDescriptor(descriptor.AsyncHandlerDescriptors[i])
	}
	sort.SliceStable(descriptor.AsyncHandlerDescriptors, func(i, j int) bool {
		return descriptor.AsyncHandlerDescriptors[i].HandlerID < descriptor.AsyncHandlerDescriptors[j].HandlerID
	})
	descriptor.GetMethodDescriptors = append([]AVMGetMethodDescriptor(nil), descriptor.GetMethodDescriptors...)
	for i := range descriptor.GetMethodDescriptors {
		descriptor.GetMethodDescriptors[i] = canonicalAVMGetMethodDescriptor(descriptor.GetMethodDescriptors[i])
	}
	sort.SliceStable(descriptor.GetMethodDescriptors, func(i, j int) bool {
		return descriptor.GetMethodDescriptors[i].MethodID < descriptor.GetMethodDescriptors[j].MethodID
	})
	return descriptor
}

func canonicalAVMMethodDescriptor(descriptor AVMMethodDescriptor) AVMMethodDescriptor {
	descriptor.MethodID = strings.TrimSpace(descriptor.MethodID)
	descriptor.Name = strings.TrimSpace(descriptor.Name)
	descriptor.InputSchemaHash = strings.TrimSpace(descriptor.InputSchemaHash)
	descriptor.OutputSchemaHash = strings.TrimSpace(descriptor.OutputSchemaHash)
	descriptor.PaymentRequirementOptional = strings.TrimSpace(descriptor.PaymentRequirementOptional)
	descriptor.ProofRequirementOptional = strings.TrimSpace(descriptor.ProofRequirementOptional)
	return descriptor
}

func canonicalAVMEventDescriptor(descriptor AVMEventDescriptor) AVMEventDescriptor {
	descriptor.EventID = strings.TrimSpace(descriptor.EventID)
	descriptor.Name = strings.TrimSpace(descriptor.Name)
	descriptor.SchemaHash = strings.TrimSpace(descriptor.SchemaHash)
	return descriptor
}

func canonicalAVMAsyncHandlerDescriptor(descriptor AVMAsyncHandlerDescriptor) AVMAsyncHandlerDescriptor {
	descriptor.HandlerID = strings.TrimSpace(descriptor.HandlerID)
	descriptor.Name = strings.TrimSpace(descriptor.Name)
	descriptor.InputSchemaHash = strings.TrimSpace(descriptor.InputSchemaHash)
	descriptor.OutputSchemaHash = strings.TrimSpace(descriptor.OutputSchemaHash)
	descriptor.RetryPolicyOptional = strings.TrimSpace(descriptor.RetryPolicyOptional)
	return descriptor
}

func canonicalAVMGetMethodDescriptor(descriptor AVMGetMethodDescriptor) AVMGetMethodDescriptor {
	descriptor.MethodID = strings.TrimSpace(descriptor.MethodID)
	descriptor.Name = strings.TrimSpace(descriptor.Name)
	descriptor.InputSchemaHash = strings.TrimSpace(descriptor.InputSchemaHash)
	descriptor.OutputSchemaHash = strings.TrimSpace(descriptor.OutputSchemaHash)
	return descriptor
}

func canonicalAVMInterfaceRegistry(registry AVMInterfaceRegistry) AVMInterfaceRegistry {
	registry.Root = strings.TrimSpace(registry.Root)
	registry.Interfaces = append([]AVMInterfaceDescriptor(nil), registry.Interfaces...)
	for i := range registry.Interfaces {
		registry.Interfaces[i] = canonicalAVMInterfaceDescriptor(registry.Interfaces[i])
	}
	sort.SliceStable(registry.Interfaces, func(i, j int) bool {
		return registry.Interfaces[i].InterfaceHash < registry.Interfaces[j].InterfaceHash
	})
	return registry
}

func validateAVMMethods(methods []AVMMethodDescriptor) error {
	seen := make(map[string]struct{}, len(methods))
	for i, method := range methods {
		if err := method.Validate(); err != nil {
			return err
		}
		if _, found := seen[method.MethodID]; found {
			return fmt.Errorf("duplicate AVM method id %q", method.MethodID)
		}
		seen[method.MethodID] = struct{}{}
		if i > 0 && methods[i-1].MethodID >= method.MethodID {
			return errors.New("AVM method descriptors must be sorted canonically")
		}
	}
	return nil
}

func validateAVMEvents(events []AVMEventDescriptor) error {
	seen := make(map[string]struct{}, len(events))
	for i, event := range events {
		if err := event.Validate(); err != nil {
			return err
		}
		if _, found := seen[event.EventID]; found {
			return fmt.Errorf("duplicate AVM event id %q", event.EventID)
		}
		seen[event.EventID] = struct{}{}
		if i > 0 && events[i-1].EventID >= event.EventID {
			return errors.New("AVM event descriptors must be sorted canonically")
		}
	}
	return nil
}

func validateAVMAsyncHandlers(handlers []AVMAsyncHandlerDescriptor) error {
	seen := make(map[string]struct{}, len(handlers))
	for i, handler := range handlers {
		if err := handler.Validate(); err != nil {
			return err
		}
		if _, found := seen[handler.HandlerID]; found {
			return fmt.Errorf("duplicate AVM async handler id %q", handler.HandlerID)
		}
		seen[handler.HandlerID] = struct{}{}
		if i > 0 && handlers[i-1].HandlerID >= handler.HandlerID {
			return errors.New("AVM async handler descriptors must be sorted canonically")
		}
	}
	return nil
}

func validateAVMGetMethods(methods []AVMGetMethodDescriptor) error {
	seen := make(map[string]struct{}, len(methods))
	for i, method := range methods {
		if err := method.Validate(); err != nil {
			return err
		}
		if _, found := seen[method.MethodID]; found {
			return fmt.Errorf("duplicate AVM get method id %q", method.MethodID)
		}
		seen[method.MethodID] = struct{}{}
		if i > 0 && methods[i-1].MethodID >= method.MethodID {
			return errors.New("AVM get method descriptors must be sorted canonically")
		}
	}
	return nil
}

func validateInterfaceToken(fieldName, value string) error {
	return validateEngineToken(fieldName, value, MaxAVMInterfaceTokenLength)
}

func writeAVMMethodDescriptor(w engineByteWriter, method AVMMethodDescriptor) {
	writeEnginePart(w, method.MethodID)
	writeEnginePart(w, method.Name)
	writeEnginePart(w, method.InputSchemaHash)
	writeEnginePart(w, method.OutputSchemaHash)
	writeEnginePart(w, string(method.ExecutionMode))
	writeEngineUint64(w, method.GasHint)
	writeEnginePart(w, method.PaymentRequirementOptional)
	writeEnginePart(w, method.ProofRequirementOptional)
}
