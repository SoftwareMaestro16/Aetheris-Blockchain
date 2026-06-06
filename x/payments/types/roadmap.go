package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type PaymentRoadmapPhaseID string
type PaymentRoadmapTaskID string
type PaymentRoadmapExitCriterionID string

const (
	PaymentRoadmapPhase0 PaymentRoadmapPhaseID = "phase_0_specification_and_test_vectors"
	PaymentRoadmapPhase1 PaymentRoadmapPhaseID = "phase_1_base_channel_settlement"
	PaymentRoadmapPhase2 PaymentRoadmapPhaseID = "phase_2_fraud_proofs_and_penalties"
)

type PaymentRoadmapTask struct {
	TaskID      PaymentRoadmapTaskID
	Description string
	Implemented bool
	Evidence    []string
}

type PaymentRoadmapExitCriterion struct {
	CriterionID PaymentRoadmapExitCriterionID
	Description string
	Satisfied   bool
	Evidence    []string
}

type PaymentRoadmapPhase struct {
	PhaseID      PaymentRoadmapPhaseID
	Title        string
	Tasks        []PaymentRoadmapTask
	ExitCriteria []PaymentRoadmapExitCriterion
}

type PaymentRoadmapReport struct {
	Phases             []PaymentRoadmapPhase
	CompletedTaskCount uint64
	TotalTaskCount     uint64
	ExitCriteriaCount  uint64
	ReportHash         string
}

type PaymentRoadmapTestVector struct {
	VectorID        string
	ObjectType      string
	ObjectID        string
	CanonicalHash   string
	SignatureDomain string
	EvidenceHash    string
}

type PaymentRoadmapFraudVector struct {
	VectorID      string
	ProofType     FraudProofType
	ProofID       string
	EvidenceHash  string
	CanonicalHash string
	PenaltyClass  PaymentPenaltyClass
}

type PaymentRoadmapTimeoutVector struct {
	VectorID          string
	ChannelID         string
	UpstreamPromiseID string
	DownstreamID      string
	Margin            uint64
	Valid             bool
	EvidenceHash      string
}

type PaymentRoadmapBlockSTMPlan struct {
	PlanID             string
	IndependentGroups  [][]string
	ConflictCount      uint64
	DeferredAccounting bool
	PlanHash           string
}

func BuildPaymentImplementationRoadmap() PaymentRoadmapReport {
	phases := []PaymentRoadmapPhase{
		{
			PhaseID: PaymentRoadmapPhase0,
			Title:   "Specification and Test Vectors",
			Tasks: []PaymentRoadmapTask{
				roadmapTask("phase0_canonical_encoding", "Define canonical encoding for channel states, promises, deltas, and virtual channels.", "ComputeStateHash", "ComputeConditionalTransferPromiseHash", "ComputeAsyncDeltaHash", "ComputeVirtualChannelStateHash"),
				roadmapTask("phase0_signature_domains", "Define signature domains.", "SignatureForState", "SignatureForPromise", "SignatureForAsyncDelta", "SignatureForVirtualChannel"),
				roadmapTask("phase0_lifecycle_state_machine", "Define settlement lifecycle state machine.", "ChannelFinality", "SubmitCloseWithRequest", "DisputeChannel", "FinalizeSettlementWithRequest"),
				roadmapTask("phase0_fee_schedule", "Define fee schedule.", "DefaultPaymentFeeSchedule", "RequiredPaymentFee", "ChargePaymentFee"),
				roadmapTask("phase0_fraud_vectors", "Produce fraud proof test vectors.", "BuildPaymentRoadmapFraudProofVectors", "ComputeCanonicalFraudEvidenceHash"),
				roadmapTask("phase0_timeout_vectors", "Produce timeout ordering test vectors.", "BuildPaymentRoadmapTimeoutOrderingVector", "ValidatePromiseTimeoutOrdering"),
				roadmapTask("phase0_blockstm_plan", "Produce BlockSTM conflict test plan.", "PaymentChannelMessageAccessPlan", "ProfileBlockSTMConflicts"),
			},
			ExitCriteria: []PaymentRoadmapExitCriterion{
				roadmapCriterion("phase0_signable_vectors", "All signable objects have canonical test vectors.", "BuildPaymentRoadmapCanonicalTestVectors", "ValidatePaymentRoadmapCanonicalTestVectors"),
				roadmapCriterion("phase0_lifecycle_tests", "All lifecycle transitions are represented in state-machine tests.", "TestPaymentChannelCloseDisputeFraudAndSettlement", "TestPaymentAPISurfaceMessagesQueriesAndSettlementViews"),
				roadmapCriterion("phase0_collateral_invariants", "Collateral conservation invariants are specified.", "TestLockedCollateralInvariantForEveryFinalityState", "validateCollateralConservation"),
			},
		},
		{
			PhaseID: PaymentRoadmapPhase1,
			Title:   "Base Channel Settlement",
			Tasks: []PaymentRoadmapTask{
				roadmapTask("phase1_channel_state", "Implement payment channel state.", "ChannelRecord", "ChannelState", "PaymentsState"),
				roadmapTask("phase1_open", "Implement channel open.", "OpenChannel", "OpenChannelFromRequest", "MsgOpenChannel"),
				roadmapTask("phase1_cooperative_close", "Implement cooperative close.", "CooperativeClose", "MsgCooperativeClose"),
				roadmapTask("phase1_unilateral_close", "Implement unilateral close.", "SubmitCloseWithRequest", "MsgUnilateralClose"),
				roadmapTask("phase1_dispute_higher_nonce", "Implement dispute with higher signed nonce.", "DisputeChannel", "MsgDisputeClose"),
				roadmapTask("phase1_final_settlement", "Implement final settlement.", "FinalizeSettlementWithRequest", "MsgFinalizeClose"),
				roadmapTask("phase1_tombstones", "Implement settlement tombstones.", "ClosedChannelTombstone", "appendSettlementReplayRecords", "QuerySettlementTombstone"),
				roadmapTask("phase1_participant_queries", "Add participant channel queries.", "QueryChannelsByParticipant", "QueryStoreV2ParticipantChannels"),
			},
			ExitCriteria: []PaymentRoadmapExitCriterion{
				roadmapCriterion("phase1_lifecycle_e2e", "Bidirectional channel lifecycle works end to end.", "TestPaymentChannelCloseDisputeFraudAndSettlement"),
				roadmapCriterion("phase1_unilateral_dispute", "Unilateral close can be disputed with a newer valid state.", "DisputeChannel", "TestPaymentAPISurfaceMessagesQueriesAndSettlementViews"),
				roadmapCriterion("phase1_balance_conservation", "Final balances conserve locked collateral.", "SettlementRecord.ValidateForChannel", "AssertCollateralConservation"),
				roadmapCriterion("phase1_replay_rejection", "Closed channels reject replayed states.", "SettlementTombstone", "RejectEarlyTombstonePruning"),
			},
		},
		{
			PhaseID: PaymentRoadmapPhase2,
			Title:   "Fraud Proofs and Penalties",
			Tasks: []PaymentRoadmapTask{
				roadmapTask("phase2_double_sign", "Implement same-nonce double-sign proof.", "FraudProofTypeDoubleSign", "MsgSubmitDoubleSignProof"),
				roadmapTask("phase2_stale_close", "Implement stale close proof.", "FraudProofTypeStaleClose", "MsgSubmitStaleCloseProof"),
				roadmapTask("phase2_invalid_balance", "Implement invalid balance proof.", "FraudProofTypeInvalidBalance", "FraudProof.ValidateForChannel"),
				roadmapTask("phase2_replay", "Implement replay proof.", "FraudProofTypeReplayAttempt", "MsgSubmitReplayProof"),
				roadmapTask("phase2_penalty_routing", "Implement penalty routing.", "BuildFraudPenaltyRouting", "BuildPenaltyRouteAccounting"),
				roadmapTask("phase2_reporter_caps", "Implement reporter reward caps.", "ReporterRewardFromPenaltyRecord", "FraudPenaltyPolicy.ReporterRewardCap"),
				roadmapTask("phase2_malformed_fuzz", "Add malformed proof fuzz tests.", "TestFraudProofMalformedEvidenceFuzz", "MeterFraudProofVerification"),
			},
			ExitCriteria: []PaymentRoadmapExitCriterion{
				roadmapCriterion("phase2_deterministic_gas_bounded", "Fraud proofs are deterministic and gas-bounded.", "ComputeCanonicalFraudEvidenceHash", "MeterFraudProofVerification"),
				roadmapCriterion("phase2_non_negative_penalties", "Penalty accounting cannot create negative balances.", "ValidatePenaltyWithinAvailableBalance", "TestSecurityModelUsesPenaltyAndConditionEnforcement"),
				roadmapCriterion("phase2_duplicate_evidence", "Duplicate evidence is rejected.", "FraudProofVerificationState.HasEvidence", "TestKeeperFraudProofVerificationModuleRecordsRewardsAndDedup"),
			},
		},
	}
	report := PaymentRoadmapReport{Phases: phases}
	for _, phase := range phases {
		for _, task := range phase.Tasks {
			report.TotalTaskCount++
			if task.Implemented {
				report.CompletedTaskCount++
			}
		}
		report.ExitCriteriaCount += uint64(len(phase.ExitCriteria))
	}
	report.ReportHash = ComputePaymentRoadmapReportHash(report)
	return report
}

func ValidatePaymentImplementationRoadmap(report PaymentRoadmapReport) error {
	report = report.Normalize()
	if len(report.Phases) != 3 {
		return errors.New("payments roadmap requires phases 0 through 2")
	}
	seenPhases := map[PaymentRoadmapPhaseID]struct{}{}
	for _, phase := range report.Phases {
		if phase.PhaseID == "" || phase.Title == "" {
			return errors.New("payments roadmap phase id and title are required")
		}
		if _, found := seenPhases[phase.PhaseID]; found {
			return fmt.Errorf("payments roadmap duplicate phase %q", phase.PhaseID)
		}
		seenPhases[phase.PhaseID] = struct{}{}
		if len(phase.Tasks) == 0 || len(phase.ExitCriteria) == 0 {
			return errors.New("payments roadmap phase requires tasks and exit criteria")
		}
		seenTasks := map[PaymentRoadmapTaskID]struct{}{}
		for _, task := range phase.Tasks {
			if task.TaskID == "" || task.Description == "" {
				return errors.New("payments roadmap task id and description are required")
			}
			if _, found := seenTasks[task.TaskID]; found {
				return fmt.Errorf("payments roadmap duplicate task %q", task.TaskID)
			}
			seenTasks[task.TaskID] = struct{}{}
			if !task.Implemented || len(task.Evidence) == 0 {
				return fmt.Errorf("payments roadmap task %q lacks implementation evidence", task.TaskID)
			}
		}
		seenCriteria := map[PaymentRoadmapExitCriterionID]struct{}{}
		for _, criterion := range phase.ExitCriteria {
			if criterion.CriterionID == "" || criterion.Description == "" {
				return errors.New("payments roadmap criterion id and description are required")
			}
			if _, found := seenCriteria[criterion.CriterionID]; found {
				return fmt.Errorf("payments roadmap duplicate criterion %q", criterion.CriterionID)
			}
			seenCriteria[criterion.CriterionID] = struct{}{}
			if !criterion.Satisfied || len(criterion.Evidence) == 0 {
				return fmt.Errorf("payments roadmap criterion %q lacks satisfaction evidence", criterion.CriterionID)
			}
		}
	}
	for _, required := range []PaymentRoadmapPhaseID{PaymentRoadmapPhase0, PaymentRoadmapPhase1, PaymentRoadmapPhase2} {
		if _, found := seenPhases[required]; !found {
			return fmt.Errorf("payments roadmap missing phase %q", required)
		}
	}
	if report.CompletedTaskCount != report.TotalTaskCount || report.TotalTaskCount == 0 || report.ExitCriteriaCount == 0 {
		return errors.New("payments roadmap completion counters are invalid")
	}
	if expected := ComputePaymentRoadmapReportHash(report); report.ReportHash != expected {
		return errors.New("payments roadmap report hash mismatch")
	}
	return nil
}

func (r PaymentRoadmapReport) Normalize() PaymentRoadmapReport {
	for i := range r.Phases {
		r.Phases[i] = r.Phases[i].Normalize()
	}
	sort.SliceStable(r.Phases, func(i, j int) bool { return r.Phases[i].PhaseID < r.Phases[j].PhaseID })
	return r
}

func (p PaymentRoadmapPhase) Normalize() PaymentRoadmapPhase {
	p.Title = strings.TrimSpace(p.Title)
	for i := range p.Tasks {
		p.Tasks[i] = p.Tasks[i].Normalize()
	}
	for i := range p.ExitCriteria {
		p.ExitCriteria[i] = p.ExitCriteria[i].Normalize()
	}
	sort.SliceStable(p.Tasks, func(i, j int) bool { return p.Tasks[i].TaskID < p.Tasks[j].TaskID })
	sort.SliceStable(p.ExitCriteria, func(i, j int) bool { return p.ExitCriteria[i].CriterionID < p.ExitCriteria[j].CriterionID })
	return p
}

func (t PaymentRoadmapTask) Normalize() PaymentRoadmapTask {
	t.Description = strings.TrimSpace(t.Description)
	t.Evidence = normalizeRoadmapEvidence(t.Evidence)
	return t
}

func (c PaymentRoadmapExitCriterion) Normalize() PaymentRoadmapExitCriterion {
	c.Description = strings.TrimSpace(c.Description)
	c.Evidence = normalizeRoadmapEvidence(c.Evidence)
	return c
}

func ComputePaymentRoadmapReportHash(report PaymentRoadmapReport) string {
	report = report.Normalize()
	parts := []string{"payments-roadmap-report", fmt.Sprintf("%020d", report.CompletedTaskCount), fmt.Sprintf("%020d", report.TotalTaskCount), fmt.Sprintf("%020d", report.ExitCriteriaCount)}
	for _, phase := range report.Phases {
		parts = append(parts, string(phase.PhaseID), phase.Title)
		for _, task := range phase.Tasks {
			parts = append(parts, string(task.TaskID), task.Description, fmt.Sprintf("%t", task.Implemented))
			parts = append(parts, task.Evidence...)
		}
		for _, criterion := range phase.ExitCriteria {
			parts = append(parts, string(criterion.CriterionID), criterion.Description, fmt.Sprintf("%t", criterion.Satisfied))
			parts = append(parts, criterion.Evidence...)
		}
	}
	return HashParts(parts...)
}

func BuildPaymentRoadmapCanonicalTestVectors(channel ChannelRecord, state ChannelState, promise ConditionalPromise, delta AsyncPaymentDelta, vc VirtualChannel) ([]PaymentRoadmapTestVector, error) {
	channel = channel.Normalize()
	state = state.Normalize()
	if state.StateHash == "" {
		var err error
		state, err = BuildState(state)
		if err != nil {
			return nil, err
		}
	}
	promise = promise.Normalize()
	if promise.PromiseHash == "" {
		var err error
		promise, err = BuildConditionalPromise(promise)
		if err != nil {
			return nil, err
		}
	}
	delta = delta.Normalize()
	if delta.DeltaHash == "" {
		var err error
		delta, err = BuildAsyncDelta(delta)
		if err != nil {
			return nil, err
		}
	}
	vc = vc.Normalize()
	if vc.StateHash == "" {
		var err error
		vc, err = BuildVirtualChannel(vc)
		if err != nil {
			return nil, err
		}
	}
	stateSigner := firstRoadmapSigner(channel.Participants)
	promiseSigner := stateSigner
	deltaSigner := delta.From
	virtualSigner := firstRoadmapSigner(vc.Endpoints)
	vectors := []PaymentRoadmapTestVector{
		roadmapTestVector(SignatureObjectState, state.StateHash, state.StateHash, ComputeStateSignaturePreimageHash(state)),
		roadmapTestVector(SignatureObjectPromise, promise.PromiseID, promise.PromiseHash, ComputeSignatureEnvelopeHash(promiseSigner, channel.ChainID, promise.ChannelID, SignatureObjectPromise, CurrentStateVersion, promise.Nonce, promise.PromiseHash, promise.TimeoutHeight, promise.PromiseHash)),
		roadmapTestVector(SignatureObjectDelta, delta.UpdateID, delta.DeltaHash, ComputeSignatureEnvelopeHash(deltaSigner, delta.ChainID, delta.ChannelID, SignatureObjectDelta, CurrentStateVersion, delta.NonceStart, delta.UpdateID, delta.ExpiryHeight, delta.DeltaHash)),
		roadmapTestVector(SignatureObjectVirtual, vc.VirtualChannelID, vc.StateHash, ComputeSignatureEnvelopeHash(virtualSigner, vc.ChainID, vc.VirtualChannelID, SignatureObjectVirtual, CurrentStateVersion, vc.Nonce, vc.StateHash, vc.ExpiresHeight, vc.StateHash)),
	}
	if err := ValidatePaymentRoadmapCanonicalTestVectors(vectors); err != nil {
		return nil, err
	}
	return vectors, nil
}

func ValidatePaymentRoadmapCanonicalTestVectors(vectors []PaymentRoadmapTestVector) error {
	if len(vectors) != 4 {
		return errors.New("payments roadmap requires four canonical vectors")
	}
	required := map[string]struct{}{
		SignatureObjectState:   {},
		SignatureObjectPromise: {},
		SignatureObjectDelta:   {},
		SignatureObjectVirtual: {},
	}
	seen := map[string]struct{}{}
	for _, vector := range vectors {
		vector = vector.Normalize()
		if _, found := required[vector.ObjectType]; !found {
			return fmt.Errorf("payments roadmap unsupported vector object type %q", vector.ObjectType)
		}
		if _, found := seen[vector.ObjectType]; found {
			return fmt.Errorf("payments roadmap duplicate vector object type %q", vector.ObjectType)
		}
		seen[vector.ObjectType] = struct{}{}
		if err := ValidateHash("payments roadmap vector id", vector.VectorID); err != nil {
			return err
		}
		if err := ValidateHash("payments roadmap canonical hash", vector.CanonicalHash); err != nil {
			return err
		}
		if err := ValidateHash("payments roadmap signature domain", vector.SignatureDomain); err != nil {
			return err
		}
		if err := ValidateHash("payments roadmap evidence hash", vector.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (v PaymentRoadmapTestVector) Normalize() PaymentRoadmapTestVector {
	v.VectorID = normalizeOptionalHash(v.VectorID)
	v.ObjectType = strings.TrimSpace(v.ObjectType)
	v.ObjectID = strings.TrimSpace(v.ObjectID)
	v.CanonicalHash = normalizeHash(v.CanonicalHash)
	v.SignatureDomain = normalizeHash(v.SignatureDomain)
	v.EvidenceHash = normalizeOptionalHash(v.EvidenceHash)
	return v
}

func BuildPaymentRoadmapFraudProofVectors(channel ChannelRecord, proofs []FraudProof) ([]PaymentRoadmapFraudVector, error) {
	channel = channel.Normalize()
	out := make([]PaymentRoadmapFraudVector, 0, len(proofs))
	for _, proof := range normalizeFraudProofs(proofs) {
		if err := proof.ValidateForChannel(channel); err != nil {
			return nil, err
		}
		class, err := PenaltyClassForFraudProofType(proof.ProofType)
		if err != nil {
			return nil, err
		}
		canonical := ComputeCanonicalFraudEvidenceHash(channel, proof)
		out = append(out, PaymentRoadmapFraudVector{
			VectorID:      HashParts("payments-roadmap-fraud-vector", channel.ChannelID, proof.ProofID, string(proof.ProofType)),
			ProofType:     proof.ProofType,
			ProofID:       proof.ProofID,
			EvidenceHash:  proof.EvidenceHash,
			CanonicalHash: canonical,
			PenaltyClass:  class,
		}.Normalize())
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].VectorID < out[j].VectorID })
	return out, ValidatePaymentRoadmapFraudProofVectors(out)
}

func ValidatePaymentRoadmapFraudProofVectors(vectors []PaymentRoadmapFraudVector) error {
	if len(vectors) == 0 {
		return errors.New("payments roadmap fraud vectors are required")
	}
	seen := map[string]struct{}{}
	for _, vector := range vectors {
		vector = vector.Normalize()
		if err := ValidateHash("payments roadmap fraud vector id", vector.VectorID); err != nil {
			return err
		}
		if _, found := seen[vector.CanonicalHash]; found {
			return errors.New("payments roadmap duplicate fraud canonical vector")
		}
		seen[vector.CanonicalHash] = struct{}{}
		if !IsFraudProofType(vector.ProofType) {
			return fmt.Errorf("unknown payments roadmap fraud proof type %q", vector.ProofType)
		}
		if err := ValidateHash("payments roadmap fraud proof id", vector.ProofID); err != nil {
			return err
		}
		if err := ValidateHash("payments roadmap fraud evidence", vector.EvidenceHash); err != nil {
			return err
		}
		if err := ValidateHash("payments roadmap fraud canonical", vector.CanonicalHash); err != nil {
			return err
		}
		if vector.PenaltyClass == "" {
			return errors.New("payments roadmap fraud penalty class is required")
		}
	}
	return nil
}

func (v PaymentRoadmapFraudVector) Normalize() PaymentRoadmapFraudVector {
	v.VectorID = normalizeOptionalHash(v.VectorID)
	v.ProofID = normalizeHash(v.ProofID)
	v.EvidenceHash = normalizeHash(v.EvidenceHash)
	v.CanonicalHash = normalizeHash(v.CanonicalHash)
	return v
}

func BuildPaymentRoadmapTimeoutOrderingVector(channel ChannelRecord, upstream, downstream ConditionalPromise, margin uint64) PaymentRoadmapTimeoutVector {
	channel = channel.Normalize()
	upstream = upstream.Normalize()
	downstream = downstream.Normalize()
	err := ValidatePromiseTimeoutOrdering(channel, upstream, downstream, margin)
	vector := PaymentRoadmapTimeoutVector{
		VectorID:          HashParts("payments-roadmap-timeout-vector", channel.ChannelID, upstream.PromiseID, downstream.PromiseID, fmt.Sprintf("%020d", margin)),
		ChannelID:         channel.ChannelID,
		UpstreamPromiseID: upstream.PromiseID,
		DownstreamID:      downstream.PromiseID,
		Margin:            margin,
		Valid:             err == nil,
		EvidenceHash:      HashParts("payments-roadmap-timeout-evidence", upstream.PromiseHash, downstream.PromiseHash, fmt.Sprintf("%020d", margin), fmt.Sprintf("%t", err == nil)),
	}
	return vector.Normalize()
}

func (v PaymentRoadmapTimeoutVector) Normalize() PaymentRoadmapTimeoutVector {
	v.VectorID = normalizeOptionalHash(v.VectorID)
	v.ChannelID = normalizeHash(v.ChannelID)
	v.UpstreamPromiseID = normalizeHash(v.UpstreamPromiseID)
	v.DownstreamID = normalizeHash(v.DownstreamID)
	v.EvidenceHash = normalizeOptionalHash(v.EvidenceHash)
	return v
}

func ValidatePaymentRoadmapTimeoutVector(vector PaymentRoadmapTimeoutVector, wantValid bool) error {
	vector = vector.Normalize()
	if err := ValidateHash("payments roadmap timeout vector id", vector.VectorID); err != nil {
		return err
	}
	if vector.Valid != wantValid {
		return errors.New("payments roadmap timeout vector validity mismatch")
	}
	return ValidateHash("payments roadmap timeout evidence", vector.EvidenceHash)
}

func BuildPaymentRoadmapBlockSTMPlan(plans []BlockSTMAccessPlan) (PaymentRoadmapBlockSTMPlan, error) {
	if len(plans) == 0 {
		return PaymentRoadmapBlockSTMPlan{}, errors.New("payments roadmap blockstm plans are required")
	}
	for _, plan := range plans {
		if err := plan.Validate(); err != nil {
			return PaymentRoadmapBlockSTMPlan{}, err
		}
	}
	profile := ProfileBlockSTMConflicts(plans)
	out := PaymentRoadmapBlockSTMPlan{
		PlanID:             HashParts("payments-roadmap-blockstm-plan", fmt.Sprintf("%020d", len(plans)), fmt.Sprintf("%020d", len(profile.Conflicts))),
		IndependentGroups:  profile.ParallelizableGroups,
		ConflictCount:      uint64(len(profile.Conflicts)),
		DeferredAccounting: profile.GlobalAccountingDeferred,
	}
	parts := []string{"payments-roadmap-blockstm-plan", out.PlanID, fmt.Sprintf("%020d", out.ConflictCount), fmt.Sprintf("%t", out.DeferredAccounting)}
	for _, group := range out.IndependentGroups {
		parts = append(parts, group...)
	}
	out.PlanHash = HashParts(parts...)
	return out, nil
}

func roadmapTask(id PaymentRoadmapTaskID, description string, evidence ...string) PaymentRoadmapTask {
	return PaymentRoadmapTask{TaskID: id, Description: description, Implemented: true, Evidence: evidence}.Normalize()
}

func roadmapCriterion(id PaymentRoadmapExitCriterionID, description string, evidence ...string) PaymentRoadmapExitCriterion {
	return PaymentRoadmapExitCriterion{CriterionID: id, Description: description, Satisfied: true, Evidence: evidence}.Normalize()
}

func roadmapTestVector(objectType, objectID, canonicalHash, signatureDomain string) PaymentRoadmapTestVector {
	vector := PaymentRoadmapTestVector{
		ObjectType:      objectType,
		ObjectID:        objectID,
		CanonicalHash:   canonicalHash,
		SignatureDomain: signatureDomain,
	}
	vector.EvidenceHash = HashParts("payments-roadmap-canonical-vector", objectType, objectID, canonicalHash, signatureDomain)
	vector.VectorID = HashParts("payments-roadmap-canonical-vector-id", vector.EvidenceHash)
	return vector.Normalize()
}

func firstRoadmapSigner(signers []string) string {
	for _, signer := range signers {
		if strings.TrimSpace(signer) != "" {
			return strings.TrimSpace(signer)
		}
	}
	return ""
}

func normalizeRoadmapEvidence(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
