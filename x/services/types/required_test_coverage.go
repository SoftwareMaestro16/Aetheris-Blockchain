package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aethercore/types"
)

type ServiceRequiredTestKind string
type ServiceUnitTestCaseID string
type ServiceIntegrationTestCaseID string

const (
	ServiceRequiredTestUnit        ServiceRequiredTestKind = "unit"
	ServiceRequiredTestIntegration ServiceRequiredTestKind = "integration"

	ServiceUnitCallIDDerivation    ServiceUnitTestCaseID = "call_id_derivation"
	ServiceUnitDescriptorHash      ServiceUnitTestCaseID = "descriptor_hash_calculation"
	ServiceUnitIdempotencyKey      ServiceUnitTestCaseID = "idempotency_key_behavior"
	ServiceUnitInterfaceHash       ServiceUnitTestCaseID = "interface_hash_calculation"
	ServiceUnitMethodIDValidation  ServiceUnitTestCaseID = "method_id_validation"
	ServiceUnitNonceReplay         ServiceUnitTestCaseID = "nonce_replay_rejection"
	ServiceUnitPaymentModel        ServiceUnitTestCaseID = "payment_model_validation"
	ServiceUnitReceiptHash         ServiceUnitTestCaseID = "receipt_hash_calculation"
	ServiceUnitServiceIDValidation ServiceUnitTestCaseID = "service_id_validation"
	ServiceUnitTrustModel          ServiceUnitTestCaseID = "trust_model_validation"

	ServiceIntegrationAnchorOffChainResult     ServiceIntegrationTestCaseID = "anchor_off_chain_service_result"
	ServiceIntegrationChallengeMixedResult     ServiceIntegrationTestCaseID = "challenge_mixed_service_result"
	ServiceIntegrationExecuteOnChainCall       ServiceIntegrationTestCaseID = "execute_on_chain_service_call"
	ServiceIntegrationGenerateReceiptProof     ServiceIntegrationTestCaseID = "generate_and_verify_service_receipt_proof"
	ServiceIntegrationRegisterFogProvider      ServiceIntegrationTestCaseID = "register_provider_for_fog_market_service"
	ServiceIntegrationRegisterInterfaceBinding ServiceIntegrationTestCaseID = "register_interface_and_bind_to_service"
	ServiceIntegrationRegisterMixedService     ServiceIntegrationTestCaseID = "register_mixed_service"
	ServiceIntegrationRegisterOffChainAnchor   ServiceIntegrationTestCaseID = "register_off_chain_service_anchor"
	ServiceIntegrationRegisterOnChainService   ServiceIntegrationTestCaseID = "register_on_chain_service"
	ServiceIntegrationResolveAETBinding        ServiceIntegrationTestCaseID = "resolve_service_through_aet_binding"
	ServiceIntegrationSettleEscrowPayment      ServiceIntegrationTestCaseID = "settle_escrow_payment"
)

type ServiceRequiredTestCase struct {
	Kind         ServiceRequiredTestKind
	UnitCase     ServiceUnitTestCaseID
	Integration  ServiceIntegrationTestCaseID
	PackagePath  string
	TestName     string
	EvidenceHash string
	CaseHash     string
}

type ServiceRequiredTestCoverage struct {
	UnitTests        []ServiceRequiredTestCase
	IntegrationTests []ServiceRequiredTestCase
	CoverageHash     string
}

func DefaultServiceRequiredTestCoverage() (ServiceRequiredTestCoverage, error) {
	unit := []ServiceRequiredTestCase{
		newServiceUnitTestCase(ServiceUnitCallIDDerivation, "TestUnifiedCallIDDerivation", "ComputeUnifiedServiceCallID"),
		newServiceUnitTestCase(ServiceUnitDescriptorHash, "TestServiceDescriptorHashCalculation", "ComputeServiceDescriptorHash"),
		newServiceUnitTestCase(ServiceUnitIdempotencyKey, "TestServiceCallIdempotencyBehavior", "NewServiceCallIdempotencyRecord"),
		newServiceUnitTestCase(ServiceUnitInterfaceHash, "TestServiceInterfaceHashCalculation", "ComputeFormalServiceInterfaceHash"),
		newServiceUnitTestCase(ServiceUnitMethodIDValidation, "TestServiceMethodIDValidation", "ServiceInterfaceMethodSchema.Validate"),
		newServiceUnitTestCase(ServiceUnitNonceReplay, "TestServiceNonceReplayRejection", "ValidateServiceCallAnte"),
		newServiceUnitTestCase(ServiceUnitPaymentModel, "TestServicePaymentModelValidation", "ServicePaymentModel.Validate"),
		newServiceUnitTestCase(ServiceUnitReceiptHash, "TestServiceReceiptHashCalculation", "ComputeServiceCallReceiptHash"),
		newServiceUnitTestCase(ServiceUnitServiceIDValidation, "TestServiceIDValidation", "ServiceDescriptor.Validate"),
		newServiceUnitTestCase(ServiceUnitTrustModel, "TestServiceTrustModelValidation", "ServiceTrustModelSecurityRule.Validate"),
	}
	integration := []ServiceRequiredTestCase{
		newServiceIntegrationTestCase(ServiceIntegrationAnchorOffChainResult, "TestAnchorOffChainServiceResult", "MsgAnchorReceipt"),
		newServiceIntegrationTestCase(ServiceIntegrationChallengeMixedResult, "TestChallengeMixedServiceResult", "NewServiceChallengeFlow"),
		newServiceIntegrationTestCase(ServiceIntegrationExecuteOnChainCall, "TestExecuteOnChainServiceCall", "ValidateUnifiedServiceCallForDescriptor"),
		newServiceIntegrationTestCase(ServiceIntegrationGenerateReceiptProof, "TestGenerateAndVerifyServiceReceiptProof", "QueryReceiptProof"),
		newServiceIntegrationTestCase(ServiceIntegrationRegisterFogProvider, "TestRegisterProviderForFogMarketService", "MsgRegisterProvider"),
		newServiceIntegrationTestCase(ServiceIntegrationRegisterInterfaceBinding, "TestRegisterInterfaceAndBindToService", "MsgRegisterInterface/MsgUpdateService"),
		newServiceIntegrationTestCase(ServiceIntegrationRegisterMixedService, "TestRegisterMixedService", "MsgRegisterService"),
		newServiceIntegrationTestCase(ServiceIntegrationRegisterOffChainAnchor, "TestRegisterOffChainServiceAnchor", "ServiceAnchor"),
		newServiceIntegrationTestCase(ServiceIntegrationRegisterOnChainService, "TestRegisterOnChainService", "MsgRegisterService"),
		newServiceIntegrationTestCase(ServiceIntegrationResolveAETBinding, "TestResolveServiceThroughAETBinding", "IdentityServiceBinding"),
		newServiceIntegrationTestCase(ServiceIntegrationSettleEscrowPayment, "TestSettleEscrowPayment", "MsgSettleServiceEscrow"),
	}
	return NewServiceRequiredTestCoverage(unit, integration)
}

func NewServiceRequiredTestCoverage(unit []ServiceRequiredTestCase, integration []ServiceRequiredTestCase) (ServiceRequiredTestCoverage, error) {
	coverage := ServiceRequiredTestCoverage{
		UnitTests:        cloneServiceRequiredTestCases(unit),
		IntegrationTests: cloneServiceRequiredTestCases(integration),
	}
	sortServiceRequiredTestCases(coverage.UnitTests)
	sortServiceRequiredTestCases(coverage.IntegrationTests)
	if err := coverage.ValidateFormat(); err != nil {
		return ServiceRequiredTestCoverage{}, err
	}
	coverage.CoverageHash = ComputeServiceRequiredTestCoverageHash(coverage)
	return coverage, coverage.Validate()
}

func ValidateServiceRequiredTestEvidence(coverage ServiceRequiredTestCoverage) error {
	if err := coverage.Validate(); err != nil {
		return err
	}
	for _, testCase := range coverage.UnitTests {
		if testCase.EvidenceHash == "" {
			return fmt.Errorf("services required test %s missing evidence", testCase.UnitCase)
		}
	}
	for _, testCase := range coverage.IntegrationTests {
		if testCase.EvidenceHash == "" {
			return fmt.Errorf("services required test %s missing evidence", testCase.Integration)
		}
	}
	return nil
}

func (coverage ServiceRequiredTestCoverage) ValidateFormat() error {
	if err := validateServiceRequiredUnitTests(coverage.UnitTests); err != nil {
		return err
	}
	if err := validateServiceRequiredIntegrationTests(coverage.IntegrationTests); err != nil {
		return err
	}
	if coverage.CoverageHash != "" {
		return coretypes.ValidateHash("services required test coverage hash", coverage.CoverageHash)
	}
	return nil
}

func (coverage ServiceRequiredTestCoverage) Validate() error {
	if err := coverage.ValidateFormat(); err != nil {
		return err
	}
	if coverage.CoverageHash == "" {
		return errors.New("services required test coverage hash is required")
	}
	if expected := ComputeServiceRequiredTestCoverageHash(coverage); coverage.CoverageHash != expected {
		return fmt.Errorf("services required test coverage hash mismatch: expected %s", expected)
	}
	return nil
}

func (testCase ServiceRequiredTestCase) Validate() error {
	if !IsServiceRequiredTestKind(testCase.Kind) {
		return fmt.Errorf("services required test unknown kind %q", testCase.Kind)
	}
	switch testCase.Kind {
	case ServiceRequiredTestUnit:
		if !IsServiceUnitTestCaseID(testCase.UnitCase) {
			return fmt.Errorf("services required test unknown unit case %q", testCase.UnitCase)
		}
		if testCase.Integration != "" {
			return errors.New("services required unit test cannot set integration case")
		}
	case ServiceRequiredTestIntegration:
		if !IsServiceIntegrationTestCaseID(testCase.Integration) {
			return fmt.Errorf("services required test unknown integration case %q", testCase.Integration)
		}
		if testCase.UnitCase != "" {
			return errors.New("services required integration test cannot set unit case")
		}
	}
	if err := validateInterfaceToken("services required test package", testCase.PackagePath); err != nil {
		return err
	}
	if err := validateInterfaceToken("services required test name", testCase.TestName); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services required test evidence hash", testCase.EvidenceHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services required test case hash", testCase.CaseHash); err != nil {
		return err
	}
	if expected := ComputeServiceRequiredTestCaseHash(testCase); testCase.CaseHash != expected {
		return fmt.Errorf("services required test case hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeServiceRequiredTestCoverageHash(coverage ServiceRequiredTestCoverage) string {
	unit := cloneServiceRequiredTestCases(coverage.UnitTests)
	integration := cloneServiceRequiredTestCases(coverage.IntegrationTests)
	sortServiceRequiredTestCases(unit)
	sortServiceRequiredTestCases(integration)
	parts := []string{"aetheris-services-required-test-coverage-v1"}
	for _, testCase := range unit {
		parts = append(parts, "unit", testCase.CaseHash)
	}
	for _, testCase := range integration {
		parts = append(parts, "integration", testCase.CaseHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceRequiredTestCaseHash(testCase ServiceRequiredTestCase) string {
	return servicesHashParts(
		"aetheris-services-required-test-case-v1",
		string(testCase.Kind),
		string(testCase.UnitCase),
		string(testCase.Integration),
		testCase.PackagePath,
		testCase.TestName,
		testCase.EvidenceHash,
	)
}

func IsServiceRequiredTestKind(kind ServiceRequiredTestKind) bool {
	switch kind {
	case ServiceRequiredTestUnit, ServiceRequiredTestIntegration:
		return true
	default:
		return false
	}
}

func IsServiceUnitTestCaseID(caseID ServiceUnitTestCaseID) bool {
	switch caseID {
	case ServiceUnitCallIDDerivation, ServiceUnitDescriptorHash, ServiceUnitIdempotencyKey, ServiceUnitInterfaceHash,
		ServiceUnitMethodIDValidation, ServiceUnitNonceReplay, ServiceUnitPaymentModel, ServiceUnitReceiptHash,
		ServiceUnitServiceIDValidation, ServiceUnitTrustModel:
		return true
	default:
		return false
	}
}

func IsServiceIntegrationTestCaseID(caseID ServiceIntegrationTestCaseID) bool {
	switch caseID {
	case ServiceIntegrationAnchorOffChainResult, ServiceIntegrationChallengeMixedResult, ServiceIntegrationExecuteOnChainCall,
		ServiceIntegrationGenerateReceiptProof, ServiceIntegrationRegisterFogProvider, ServiceIntegrationRegisterInterfaceBinding,
		ServiceIntegrationRegisterMixedService, ServiceIntegrationRegisterOffChainAnchor, ServiceIntegrationRegisterOnChainService,
		ServiceIntegrationResolveAETBinding, ServiceIntegrationSettleEscrowPayment:
		return true
	default:
		return false
	}
}

func newServiceUnitTestCase(caseID ServiceUnitTestCaseID, testName, evidence string) ServiceRequiredTestCase {
	testCase := ServiceRequiredTestCase{
		Kind:         ServiceRequiredTestUnit,
		UnitCase:     caseID,
		PackagePath:  "x/services/types",
		TestName:     strings.TrimSpace(testName),
		EvidenceHash: servicesHashParts("aetheris-services-required-test-evidence-v1", string(caseID), evidence),
	}
	testCase.CaseHash = ComputeServiceRequiredTestCaseHash(testCase)
	return testCase
}

func newServiceIntegrationTestCase(caseID ServiceIntegrationTestCaseID, testName, evidence string) ServiceRequiredTestCase {
	testCase := ServiceRequiredTestCase{
		Kind:         ServiceRequiredTestIntegration,
		Integration:  caseID,
		PackagePath:  "x/services/types",
		TestName:     strings.TrimSpace(testName),
		EvidenceHash: servicesHashParts("aetheris-services-required-test-evidence-v1", string(caseID), evidence),
	}
	testCase.CaseHash = ComputeServiceRequiredTestCaseHash(testCase)
	return testCase
}

func validateServiceRequiredUnitTests(testCases []ServiceRequiredTestCase) error {
	required := []ServiceUnitTestCaseID{
		ServiceUnitCallIDDerivation,
		ServiceUnitDescriptorHash,
		ServiceUnitIdempotencyKey,
		ServiceUnitInterfaceHash,
		ServiceUnitMethodIDValidation,
		ServiceUnitNonceReplay,
		ServiceUnitPaymentModel,
		ServiceUnitReceiptHash,
		ServiceUnitServiceIDValidation,
		ServiceUnitTrustModel,
	}
	if len(testCases) != len(required) {
		return fmt.Errorf("services required test coverage expected %d unit tests", len(required))
	}
	seen := map[ServiceUnitTestCaseID]struct{}{}
	previous := ""
	for _, testCase := range testCases {
		if err := testCase.Validate(); err != nil {
			return err
		}
		if testCase.Kind != ServiceRequiredTestUnit {
			return errors.New("services required unit coverage includes non-unit test")
		}
		current := string(testCase.UnitCase)
		if previous != "" && previous >= current {
			return errors.New("services required unit tests must be sorted canonically")
		}
		previous = current
		seen[testCase.UnitCase] = struct{}{}
	}
	for _, caseID := range required {
		if _, found := seen[caseID]; !found {
			return fmt.Errorf("services required test coverage missing unit test %s", caseID)
		}
	}
	return nil
}

func validateServiceRequiredIntegrationTests(testCases []ServiceRequiredTestCase) error {
	required := []ServiceIntegrationTestCaseID{
		ServiceIntegrationAnchorOffChainResult,
		ServiceIntegrationChallengeMixedResult,
		ServiceIntegrationExecuteOnChainCall,
		ServiceIntegrationGenerateReceiptProof,
		ServiceIntegrationRegisterFogProvider,
		ServiceIntegrationRegisterInterfaceBinding,
		ServiceIntegrationRegisterMixedService,
		ServiceIntegrationRegisterOffChainAnchor,
		ServiceIntegrationRegisterOnChainService,
		ServiceIntegrationResolveAETBinding,
		ServiceIntegrationSettleEscrowPayment,
	}
	if len(testCases) != len(required) {
		return fmt.Errorf("services required test coverage expected %d integration tests", len(required))
	}
	seen := map[ServiceIntegrationTestCaseID]struct{}{}
	previous := ""
	for _, testCase := range testCases {
		if err := testCase.Validate(); err != nil {
			return err
		}
		if testCase.Kind != ServiceRequiredTestIntegration {
			return errors.New("services required integration coverage includes non-integration test")
		}
		current := string(testCase.Integration)
		if previous != "" && previous >= current {
			return errors.New("services required integration tests must be sorted canonically")
		}
		previous = current
		seen[testCase.Integration] = struct{}{}
	}
	for _, caseID := range required {
		if _, found := seen[caseID]; !found {
			return fmt.Errorf("services required test coverage missing integration test %s", caseID)
		}
	}
	return nil
}

func cloneServiceRequiredTestCases(testCases []ServiceRequiredTestCase) []ServiceRequiredTestCase {
	out := make([]ServiceRequiredTestCase, len(testCases))
	copy(out, testCases)
	return out
}

func sortServiceRequiredTestCases(testCases []ServiceRequiredTestCase) {
	sort.SliceStable(testCases, func(i, j int) bool {
		return serviceRequiredTestCaseSortKey(testCases[i]) < serviceRequiredTestCaseSortKey(testCases[j])
	})
}

func serviceRequiredTestCaseSortKey(testCase ServiceRequiredTestCase) string {
	if testCase.Kind == ServiceRequiredTestUnit {
		return string(testCase.UnitCase)
	}
	return string(testCase.Integration)
}
