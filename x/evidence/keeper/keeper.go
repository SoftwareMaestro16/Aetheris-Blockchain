package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/evidence/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version uint64
	Params  types.Params
	State   types.State
}

type Keeper struct {
	genesis      GenesisState
	storeService corestore.KVStoreService
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	params := types.DefaultParams()
	return GenesisState{
		Version: prototype.CurrentGenesisVersion,
		Params:  params,
		State:   types.State{}.Normalize(params),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("native evidence unsupported genesis version")
	}
	return gs.State.Validate(gs.Params)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	if k.storeService == nil {
		return nil
	}
	bz, err := json.Marshal(cloneGenesis(gs))
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	if !reflect.DeepEqual(k.genesis, DefaultGenesis()) {
		return k.ExportGenesis(), nil
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return GenesisState{}, err
	}
	if len(bz) == 0 {
		return DefaultGenesis(), nil
	}
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) SubmitEvidence(msg types.MsgSubmitEvidence) (types.EvidenceRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.EvidenceRecord{}, err
	}
	if msg.Height == 0 {
		return types.EvidenceRecord{}, errors.New("native evidence submission height must be positive")
	}
	record := types.EvidenceRecord{
		EvidenceID:       msg.EvidenceID,
		Status:           types.StatusPending,
		EvidenceType:     msg.EvidenceType,
		AccusedValidator: msg.AccusedValidator,
		Reporter:         msg.Reporter,
		ProofPayloadHash: msg.ProofPayloadHash,
		PayloadSizeBytes: msg.PayloadSizeBytes,
		SlashDecision: types.SlashDecision{
			FractionBps: types.CanonicalSlashFraction(k.genesis.Params, msg.EvidenceType, msg.SlashFractionBps),
			Tombstone:   types.IsCriticalEvidenceType(msg.EvidenceType),
		},
		RewardDecision: types.RewardDecision{
			Reporter:   msg.Reporter,
			AmountNaet: msg.RewardNaet,
		},
		SubmittedHeight:  msg.Height,
		UpdatedHeight:    msg.Height,
		ExpirationHeight: msg.Height + k.genesis.Params.EvidenceTTLBlocks,
		RequiresReview:   msg.RequiresReview,
	}
	if record.RewardDecision.AmountNaet == 0 {
		record.RewardDecision.AmountNaet = k.genesis.Params.MaxReporterRewardNaet
	}
	if err := record.Validate(k.genesis.Params); err != nil {
		return types.EvidenceRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	if _, _, found := findEvidence(next.State.Evidence, record.EvidenceID); found {
		return types.EvidenceRecord{}, errors.New("native evidence duplicate evidence id")
	}
	if _, found := findEvidenceByHash(next.State.Evidence, record.ProofPayloadHash); found {
		return types.EvidenceRecord{}, errors.New("native evidence duplicate proof payload hash")
	}
	next.State.Evidence = append(next.State.Evidence, record)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) VoteEvidence(msg types.MsgVoteEvidence) (types.EvidenceRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.EvidenceRecord{}, err
	}
	idx, record, found := findEvidence(k.genesis.State.Evidence, msg.EvidenceID)
	if !found {
		return types.EvidenceRecord{}, errors.New("native evidence record not found")
	}
	if record.Status != types.StatusPending {
		return types.EvidenceRecord{}, errors.New("native evidence vote requires pending evidence")
	}
	if msg.Height == 0 || msg.Height > record.ExpirationHeight {
		return types.EvidenceRecord{}, errors.New("native evidence vote height is outside active evidence window")
	}
	support := types.VoteSupportReject
	if msg.Accept {
		support = types.VoteSupportAccept
	}
	vote := types.EvidenceVote{
		Voter:          msg.Voter,
		Support:        support,
		VotingPowerBps: msg.VotingPowerBps,
		Height:         msg.Height,
	}
	if err := vote.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	for _, existing := range record.Votes {
		if existing.Voter == vote.Voter {
			return types.EvidenceRecord{}, errors.New("native evidence duplicate vote")
		}
	}
	record.Votes = append(record.Votes, vote)
	record.UpdatedHeight = msg.Height
	next := cloneGenesis(k.genesis)
	next.State.Evidence[idx] = record
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) FinalizeEvidence(msg types.MsgFinalizeEvidence) (types.EvidenceRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.EvidenceRecord{}, err
	}
	idx, record, found := findEvidence(k.genesis.State.Evidence, msg.EvidenceID)
	if !found {
		return types.EvidenceRecord{}, errors.New("native evidence record not found")
	}
	if record.Status != types.StatusPending {
		return types.EvidenceRecord{}, errors.New("native evidence can only be finalized once")
	}
	if msg.Height == 0 {
		return types.EvidenceRecord{}, errors.New("native evidence finalization height must be positive")
	}
	if msg.Height > record.ExpirationHeight {
		record.Status = types.StatusExpired
		record.UpdatedHeight = msg.Height
		record.FinalizedHeight = msg.Height
		next := cloneGenesis(k.genesis)
		next.State.Evidence[idx] = record
		next.State = next.State.Normalize(next.Params)
		if err := next.Validate(); err != nil {
			return types.EvidenceRecord{}, err
		}
		k.genesis = next
		return record, nil
	}
	if record.RequiresReview && types.AcceptedVotingPowerBps(record.Votes) < k.genesis.Params.ReviewQuorumBps {
		record.Status = types.StatusRejected
		record.RejectionReason = "review quorum not reached"
		record.UpdatedHeight = msg.Height
		record.FinalizedHeight = msg.Height
		next := cloneGenesis(k.genesis)
		next.State.Evidence[idx] = record
		next.State = next.State.Normalize(next.Params)
		if err := next.Validate(); err != nil {
			return types.EvidenceRecord{}, err
		}
		k.genesis = next
		return record, nil
	}
	record.Status = types.StatusAccepted
	record.SlashDecision.Applied = true
	record.RewardDecision.Paid = true
	record.UpdatedHeight = msg.Height
	record.FinalizedHeight = msg.Height

	next := cloneGenesis(k.genesis)
	next.State.Evidence[idx] = record
	next.State.SlashEvents = append(next.State.SlashEvents, types.SlashEvent{
		EvidenceID:       record.EvidenceID,
		ValidatorAddress: record.AccusedValidator,
		FractionBps:      record.SlashDecision.FractionBps,
		Tombstone:        record.SlashDecision.Tombstone,
		Height:           msg.Height,
	})
	next.State.ReporterRewards = append(next.State.ReporterRewards, types.ReporterReward{
		EvidenceID: record.EvidenceID,
		Reporter:   record.Reporter,
		AmountNaet: record.RewardDecision.AmountNaet,
		Paid:       true,
		Height:     msg.Height,
	})
	status := types.RegistryStatusJailed
	if record.SlashDecision.Tombstone {
		status = types.RegistryStatusTombstoned
		next.State.TombstonedValidators = append(next.State.TombstonedValidators, record.AccusedValidator)
	}
	next.State.RegistryUpdates = append(next.State.RegistryUpdates, types.RegistryUpdate{
		EvidenceID:       record.EvidenceID,
		ValidatorAddress: record.AccusedValidator,
		Status:           status,
		Height:           msg.Height,
	})
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) CancelExpiredEvidence(msg types.MsgCancelExpiredEvidence) (types.EvidenceRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.EvidenceRecord{}, err
	}
	idx, record, found := findEvidence(k.genesis.State.Evidence, msg.EvidenceID)
	if !found {
		return types.EvidenceRecord{}, errors.New("native evidence record not found")
	}
	if record.Status != types.StatusPending {
		return types.EvidenceRecord{}, errors.New("native evidence cancel requires pending evidence")
	}
	if msg.Height == 0 || msg.Height <= record.ExpirationHeight {
		return types.EvidenceRecord{}, errors.New("native evidence has not expired")
	}
	record.Status = types.StatusExpired
	record.UpdatedHeight = msg.Height
	record.FinalizedHeight = msg.Height
	next := cloneGenesis(k.genesis)
	next.State.Evidence[idx] = record
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k Keeper) Evidence(evidenceID string) (types.EvidenceRecord, bool) {
	_, record, found := findEvidence(k.genesis.State.Evidence, evidenceID)
	return record, found
}

func (k Keeper) EvidenceByValidator(validator string) []types.EvidenceRecord {
	out := []types.EvidenceRecord{}
	for _, record := range types.SortEvidence(k.genesis.State.Evidence) {
		if record.AccusedValidator == validator {
			out = append(out, record)
		}
	}
	return out
}

func (k Keeper) EvidenceByReporter(reporter string) []types.EvidenceRecord {
	out := []types.EvidenceRecord{}
	for _, record := range types.SortEvidence(k.genesis.State.Evidence) {
		if record.Reporter == reporter {
			out = append(out, record)
		}
	}
	return out
}

func (k Keeper) PendingEvidence() []types.EvidenceRecord {
	out := []types.EvidenceRecord{}
	for _, record := range types.SortEvidence(k.genesis.State.Evidence) {
		if record.Status == types.StatusPending {
			out = append(out, record)
		}
	}
	return out
}

func (k Keeper) EvidenceParams() types.Params {
	return k.genesis.Params
}

func (k Keeper) SlashEvents() []types.SlashEvent {
	return types.SortSlashEvents(k.genesis.State.SlashEvents)
}

func (k Keeper) ReporterRewards() []types.ReporterReward {
	return types.SortReporterRewards(k.genesis.State.ReporterRewards)
}

func (k Keeper) RegistryUpdates() []types.RegistryUpdate {
	return types.SortRegistryUpdates(k.genesis.State.RegistryUpdates)
}

func (k Keeper) TombstonedValidators() []string {
	return append([]string(nil), k.genesis.State.TombstonedValidators...)
}

type Migrator struct{ keeper *Keeper }

func NewMigrator(k *Keeper) Migrator  { return Migrator{keeper: k} }
func (m Migrator) Migrate1to2() error { return m.keeper.ExportGenesis().Validate() }
func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Normalize(gs.Params)
	return gs
}

func findEvidence(records []types.EvidenceRecord, evidenceID string) (int, types.EvidenceRecord, bool) {
	for idx, record := range records {
		if record.EvidenceID == evidenceID {
			return idx, record, true
		}
	}
	return -1, types.EvidenceRecord{}, false
}

func findEvidenceByHash(records []types.EvidenceRecord, hash string) (types.EvidenceRecord, bool) {
	for _, record := range records {
		if record.ProofPayloadHash == hash {
			return record, true
		}
	}
	return types.EvidenceRecord{}, false
}
