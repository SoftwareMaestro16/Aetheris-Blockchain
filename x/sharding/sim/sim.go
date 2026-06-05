package sim

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MasterchainID int32  = -1
	BaseWorkchain int32  = 0
	BaseShardID   string = ""
	FeeDenomNaet  string = "naet"
)

type Validator struct {
	Address string
	Power   int64
}

type WorkchainConfig struct {
	ID               int32
	AllowedVMs       []string
	FeeDenom         string
	AddressFormat    string
	GenesisStateHash string
	UpgradePolicy    string
}

type ShardID struct {
	WorkchainID int32
	Prefix      string
}

type ShardState struct {
	ID               ShardID
	Height           uint64
	StateRoot        string
	MessageQueueRoot string
	ReceiptRoot      string
	ValidatorSubset  []string
	Queue            []CrossShardMessage
	Receipts         map[string]Receipt
	Available        bool
}

type ShardHeader struct {
	ShardID          ShardID
	Height           uint64
	StateRoot        string
	MessageQueueRoot string
	ReceiptRoot      string
	ValidatorSubset  []string
	Available        bool
	Commitment       string
}

type CrossShardMessage struct {
	Source      ShardID
	Destination ShardID
	MessageID   string
	Nonce       uint64
	Payload     []byte
	Proof       string
	Timeout     uint64
	Bounce      bool
	Bounced     bool
}

type Receipt struct {
	MessageID   string
	Source      ShardID
	Destination ShardID
	Success     bool
	Height      uint64
	ResultCode  uint32
	Proof       string
}

type EquivocationEvidence struct {
	Validator string
	ShardID   ShardID
	Height    uint64
	LeftRoot  string
	RightRoot string
}

type MasterchainState struct {
	Height             uint64
	Validators         []Validator
	StakingSnapshot    map[string]int64
	Workchains         map[int32]WorkchainConfig
	Shards             map[string]ShardState
	Headers            map[string]ShardHeader
	CrossShardReceipts map[string]Receipt
	ConfigUpdates      []string
	Evidence           []EquivocationEvidence
	FinalityLag        uint64
	RandomnessSeed     string
}

type Simulator struct {
	state          MasterchainState
	processed      map[string]struct{}
	pendingReceipt map[string]CrossShardMessage
}

func New(validators []Validator, seed string) (*Simulator, error) {
	if len(validators) == 0 {
		return nil, errors.New("validator set must not be empty")
	}
	normalized := cloneValidators(validators)
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].Address < normalized[j].Address })
	staking := make(map[string]int64, len(normalized))
	for _, validator := range normalized {
		if strings.TrimSpace(validator.Address) == "" {
			return nil, errors.New("validator address must not be empty")
		}
		if validator.Power <= 0 {
			return nil, errors.New("validator power must be positive")
		}
		staking[validator.Address] = validator.Power
	}
	return &Simulator{
		state: MasterchainState{
			Validators:         normalized,
			StakingSnapshot:    staking,
			Workchains:         make(map[int32]WorkchainConfig),
			Shards:             make(map[string]ShardState),
			Headers:            make(map[string]ShardHeader),
			CrossShardReceipts: make(map[string]Receipt),
			FinalityLag:        2,
			RandomnessSeed:     seed,
		},
		processed:      make(map[string]struct{}),
		pendingReceipt: make(map[string]CrossShardMessage),
	}, nil
}

func (s *Simulator) AddWorkchain(config WorkchainConfig) error {
	if config.ID == MasterchainID {
		return errors.New("masterchain id is reserved")
	}
	if len(config.AllowedVMs) == 0 {
		return errors.New("workchain must allow at least one VM")
	}
	if config.FeeDenom != FeeDenomNaet {
		return errors.New("workchain fee policy must use naet")
	}
	if strings.TrimSpace(config.AddressFormat) == "" {
		return errors.New("workchain address format must be set")
	}
	if strings.TrimSpace(config.GenesisStateHash) == "" {
		return errors.New("workchain genesis state hash must be set")
	}
	if _, exists := s.state.Workchains[config.ID]; exists {
		return errors.New("workchain already registered")
	}
	config.AllowedVMs = append([]string(nil), config.AllowedVMs...)
	sort.Strings(config.AllowedVMs)
	s.state.Workchains[config.ID] = config
	s.state.ConfigUpdates = append(s.state.ConfigUpdates, fmt.Sprintf("add-workchain:%d", config.ID))
	return s.AddShard(ShardID{WorkchainID: config.ID, Prefix: BaseShardID})
}

func (s *Simulator) AddShard(id ShardID) error {
	if err := s.validateShardID(id); err != nil {
		return err
	}
	key := id.Key()
	if _, exists := s.state.Shards[key]; exists {
		return errors.New("shard already registered")
	}
	shard := ShardState{
		ID:              id,
		StateRoot:       HashParts("state", key, "0"),
		ValidatorSubset: s.AssignValidators(id, 0),
		Receipts:        make(map[string]Receipt),
		Available:       true,
	}
	shard.MessageQueueRoot = hashQueue(shard.Queue)
	shard.ReceiptRoot = hashReceipts(shard.Receipts)
	s.state.Shards[key] = shard
	s.commitHeader(shard)
	return nil
}

func (s *Simulator) AssignValidators(id ShardID, height uint64) []string {
	validators := cloneValidators(s.state.Validators)
	sort.Slice(validators, func(i, j int) bool {
		left := HashParts(s.state.RandomnessSeed, id.Key(), fmt.Sprint(height), validators[i].Address)
		right := HashParts(s.state.RandomnessSeed, id.Key(), fmt.Sprint(height), validators[j].Address)
		if left == right {
			return validators[i].Address < validators[j].Address
		}
		return left < right
	})
	limit := 3
	if len(validators) < limit {
		limit = len(validators)
	}
	out := make([]string, limit)
	for i := 0; i < limit; i++ {
		out[i] = validators[i].Address
	}
	sort.Strings(out)
	return out
}

func (s *Simulator) ReassignValidators(height uint64) {
	keys := sortedShardKeys(s.state.Shards)
	for _, key := range keys {
		shard := s.state.Shards[key]
		shard.ValidatorSubset = s.AssignValidators(shard.ID, height)
		s.state.Shards[key] = shard
		s.commitHeader(shard)
	}
	s.state.Height = max(s.state.Height, height)
}

func (s *Simulator) EnqueueMessage(msg CrossShardMessage) error {
	if err := s.validateMessage(msg); err != nil {
		return err
	}
	msg.MessageID = MessageID(msg.Source, msg.Destination, msg.Nonce, msg.Payload)
	source, ok := s.state.Shards[msg.Source.Key()]
	if !ok {
		return errors.New("source shard not registered")
	}
	header := s.state.Headers[msg.Source.Key()]
	msg.Proof = header.Commitment
	source.Queue = append(source.Queue, cloneMessage(msg))
	source.MessageQueueRoot = hashQueue(source.Queue)
	s.state.Shards[msg.Source.Key()] = source
	s.commitHeader(source)
	return nil
}

func (s *Simulator) ProcessNext(sourceID ShardID, height uint64) (Receipt, error) {
	source, ok := s.state.Shards[sourceID.Key()]
	if !ok {
		return Receipt{}, errors.New("source shard not registered")
	}
	if !source.Available {
		return Receipt{}, errors.New("source shard data unavailable")
	}
	if len(source.Queue) == 0 {
		return Receipt{}, errors.New("source shard queue is empty")
	}
	msg := source.Queue[0]
	source.Queue = source.Queue[1:]
	source.MessageQueueRoot = hashQueue(source.Queue)
	s.state.Shards[sourceID.Key()] = source
	s.commitHeader(source)
	return s.Deliver(msg, height)
}

func (s *Simulator) Deliver(msg CrossShardMessage, height uint64) (Receipt, error) {
	if _, exists := s.processed[msg.MessageID]; exists {
		return Receipt{}, errors.New("replayed cross-shard message")
	}
	if msg.Proof != s.state.Headers[msg.Source.Key()].Commitment {
		return Receipt{}, errors.New("invalid shard proof")
	}
	if height > msg.Timeout {
		if msg.Bounce && !msg.Bounced {
			bounced := msg
			bounced.Source, bounced.Destination = msg.Destination, msg.Source
			bounced.Bounced = true
			bounced.Nonce++
			bounced.MessageID = MessageID(bounced.Source, bounced.Destination, bounced.Nonce, bounced.Payload)
			_ = s.EnqueueMessage(bounced)
		}
		return Receipt{}, errors.New("cross-shard message timeout")
	}
	dest, ok := s.state.Shards[msg.Destination.Key()]
	if !ok {
		return Receipt{}, errors.New("destination shard not registered")
	}
	if !dest.Available {
		s.pendingReceipt[msg.MessageID] = cloneMessage(msg)
		return Receipt{}, errors.New("destination shard data unavailable")
	}
	receipt := Receipt{
		MessageID:   msg.MessageID,
		Source:      msg.Source,
		Destination: msg.Destination,
		Success:     true,
		Height:      height,
		Proof:       HashParts("receipt", msg.MessageID, fmt.Sprint(height), dest.StateRoot),
	}
	if err := s.CommitReceipt(receipt); err != nil {
		return Receipt{}, err
	}
	s.processed[msg.MessageID] = struct{}{}
	return receipt, nil
}

func (s *Simulator) CommitReceipt(receipt Receipt) error {
	if receipt.MessageID == "" {
		return errors.New("receipt message id must be set")
	}
	if _, exists := s.state.CrossShardReceipts[receipt.MessageID]; exists {
		return errors.New("duplicate cross-shard receipt")
	}
	if _, ok := s.state.Shards[receipt.Destination.Key()]; !ok {
		return errors.New("receipt destination shard not registered")
	}
	if receipt.Proof == "" {
		return errors.New("receipt proof must be set")
	}
	s.state.CrossShardReceipts[receipt.MessageID] = receipt
	dest := s.state.Shards[receipt.Destination.Key()]
	dest.Receipts[receipt.MessageID] = receipt
	dest.ReceiptRoot = hashReceipts(dest.Receipts)
	s.state.Shards[dest.ID.Key()] = dest
	s.commitHeader(dest)
	return nil
}

func (s *Simulator) RequireReceipt(messageID string) error {
	if _, ok := s.state.CrossShardReceipts[messageID]; !ok {
		return errors.New("missing cross-shard receipt")
	}
	return nil
}

func (s *Simulator) VerifyHeaderFresh(id ShardID, height uint64) error {
	header, ok := s.state.Headers[id.Key()]
	if !ok {
		return errors.New("shard header not found")
	}
	if height > header.Height+s.state.FinalityLag {
		return errors.New("stale shard header")
	}
	return nil
}

func (s *Simulator) SplitShard(id ShardID) error {
	parent, ok := s.state.Shards[id.Key()]
	if !ok {
		return errors.New("parent shard not registered")
	}
	if len(parent.ID.Prefix) >= 60 {
		return errors.New("shard prefix is already at max depth")
	}
	delete(s.state.Shards, id.Key())
	leftID := ShardID{WorkchainID: id.WorkchainID, Prefix: id.Prefix + "0"}
	rightID := ShardID{WorkchainID: id.WorkchainID, Prefix: id.Prefix + "1"}
	left := ShardState{ID: leftID, Height: parent.Height + 1, StateRoot: HashParts(parent.StateRoot, "split-left"), ValidatorSubset: s.AssignValidators(leftID, parent.Height+1), Receipts: make(map[string]Receipt), Available: parent.Available}
	right := ShardState{ID: rightID, Height: parent.Height + 1, StateRoot: HashParts(parent.StateRoot, "split-right"), ValidatorSubset: s.AssignValidators(rightID, parent.Height+1), Receipts: make(map[string]Receipt), Available: parent.Available}
	for _, msg := range parent.Queue {
		if strings.HasPrefix(msg.Destination.Prefix, rightID.Prefix) {
			right.Queue = append(right.Queue, msg)
		} else {
			left.Queue = append(left.Queue, msg)
		}
	}
	left.MessageQueueRoot = hashQueue(left.Queue)
	right.MessageQueueRoot = hashQueue(right.Queue)
	left.ReceiptRoot = hashReceipts(left.Receipts)
	right.ReceiptRoot = hashReceipts(right.Receipts)
	s.state.Shards[leftID.Key()] = left
	s.state.Shards[rightID.Key()] = right
	s.commitHeader(left)
	s.commitHeader(right)
	return nil
}

func (s *Simulator) MergeShards(leftID ShardID, rightID ShardID) error {
	if leftID.WorkchainID != rightID.WorkchainID {
		return errors.New("cannot merge shards from different workchains")
	}
	if len(leftID.Prefix) == 0 || len(rightID.Prefix) == 0 {
		return errors.New("cannot merge root shard")
	}
	if leftID.Prefix[:len(leftID.Prefix)-1] != rightID.Prefix[:len(rightID.Prefix)-1] {
		return errors.New("can only merge sibling shards")
	}
	left, ok := s.state.Shards[leftID.Key()]
	if !ok {
		return errors.New("left shard not registered")
	}
	right, ok := s.state.Shards[rightID.Key()]
	if !ok {
		return errors.New("right shard not registered")
	}
	parentID := ShardID{WorkchainID: leftID.WorkchainID, Prefix: leftID.Prefix[:len(leftID.Prefix)-1]}
	parent := ShardState{
		ID:              parentID,
		Height:          max(left.Height, right.Height) + 1,
		StateRoot:       HashParts(left.StateRoot, right.StateRoot, "merge"),
		ValidatorSubset: s.AssignValidators(parentID, max(left.Height, right.Height)+1),
		Queue:           append(cloneQueue(left.Queue), right.Queue...),
		Receipts:        mergeReceipts(left.Receipts, right.Receipts),
		Available:       left.Available && right.Available,
	}
	sort.Slice(parent.Queue, func(i, j int) bool { return parent.Queue[i].MessageID < parent.Queue[j].MessageID })
	parent.MessageQueueRoot = hashQueue(parent.Queue)
	parent.ReceiptRoot = hashReceipts(parent.Receipts)
	delete(s.state.Shards, leftID.Key())
	delete(s.state.Shards, rightID.Key())
	s.state.Shards[parentID.Key()] = parent
	s.commitHeader(parent)
	return nil
}

func (s *Simulator) MarkShardAvailability(id ShardID, available bool) error {
	shard, ok := s.state.Shards[id.Key()]
	if !ok {
		return errors.New("shard not registered")
	}
	shard.Available = available
	s.state.Shards[id.Key()] = shard
	s.commitHeader(shard)
	return nil
}

func (s *Simulator) SubmitEquivocation(e EquivocationEvidence) error {
	if strings.TrimSpace(e.Validator) == "" {
		return errors.New("equivocation validator must be set")
	}
	if e.LeftRoot == "" || e.RightRoot == "" || e.LeftRoot == e.RightRoot {
		return errors.New("equivocation must include conflicting roots")
	}
	if _, ok := s.state.Shards[e.ShardID.Key()]; !ok {
		return errors.New("equivocation shard not registered")
	}
	s.state.Evidence = append(s.state.Evidence, e)
	return nil
}

func (s *Simulator) Export() MasterchainState {
	out := s.state
	out.Validators = cloneValidators(s.state.Validators)
	out.StakingSnapshot = cloneIntMap(s.state.StakingSnapshot)
	out.Workchains = cloneWorkchains(s.state.Workchains)
	out.Shards = cloneShards(s.state.Shards)
	out.Headers = cloneHeaders(s.state.Headers)
	out.CrossShardReceipts = cloneReceipts(s.state.CrossShardReceipts)
	out.ConfigUpdates = append([]string(nil), s.state.ConfigUpdates...)
	out.Evidence = append([]EquivocationEvidence(nil), s.state.Evidence...)
	return out
}

func Import(state MasterchainState) (*Simulator, error) {
	sim, err := New(state.Validators, state.RandomnessSeed)
	if err != nil {
		return nil, err
	}
	if err := ValidateState(state); err != nil {
		return nil, err
	}
	sim.state = state
	sim.processed = make(map[string]struct{}, len(state.CrossShardReceipts))
	for id := range state.CrossShardReceipts {
		sim.processed[id] = struct{}{}
	}
	sim.pendingReceipt = make(map[string]CrossShardMessage)
	return sim, nil
}

func ValidateState(state MasterchainState) error {
	if len(state.Validators) == 0 {
		return errors.New("validator set must not be empty")
	}
	for id, wc := range state.Workchains {
		if id != wc.ID {
			return errors.New("workchain registry key mismatch")
		}
		if wc.FeeDenom != FeeDenomNaet {
			return errors.New("workchain fee policy must use naet")
		}
	}
	for key, shard := range state.Shards {
		if key != shard.ID.Key() {
			return errors.New("shard registry key mismatch")
		}
		header, ok := state.Headers[key]
		if !ok {
			return errors.New("missing shard header")
		}
		if header.Commitment != headerCommitment(shard) {
			return errors.New("invalid shard header commitment")
		}
	}
	for id, receipt := range state.CrossShardReceipts {
		if id != receipt.MessageID {
			return errors.New("receipt registry key mismatch")
		}
		if receipt.Proof == "" {
			return errors.New("receipt proof must be set")
		}
	}
	return nil
}

func MessageID(source ShardID, destination ShardID, nonce uint64, payload []byte) string {
	return HashParts("message", source.Key(), destination.Key(), fmt.Sprint(nonce), string(payload))
}

func HashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte{0})
		h.Write([]byte(part))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (id ShardID) Key() string {
	return fmt.Sprintf("%d:%s", id.WorkchainID, id.Prefix)
}

func (s *Simulator) validateShardID(id ShardID) error {
	if _, ok := s.state.Workchains[id.WorkchainID]; !ok {
		return errors.New("workchain is not registered")
	}
	if len(id.Prefix) > 60 {
		return errors.New("shard prefix length must be <= 60")
	}
	for _, r := range id.Prefix {
		if r != '0' && r != '1' {
			return errors.New("shard prefix must be binary")
		}
	}
	return nil
}

func (s *Simulator) validateMessage(msg CrossShardMessage) error {
	if msg.Source.Key() == msg.Destination.Key() {
		return errors.New("cross-shard message requires different shards")
	}
	if _, ok := s.state.Shards[msg.Destination.Key()]; !ok {
		return errors.New("destination shard not registered")
	}
	if msg.Timeout == 0 {
		return errors.New("cross-shard message timeout must be set")
	}
	return nil
}

func (s *Simulator) commitHeader(shard ShardState) {
	header := ShardHeader{
		ShardID:          shard.ID,
		Height:           shard.Height,
		StateRoot:        shard.StateRoot,
		MessageQueueRoot: shard.MessageQueueRoot,
		ReceiptRoot:      shard.ReceiptRoot,
		ValidatorSubset:  append([]string(nil), shard.ValidatorSubset...),
		Available:        shard.Available,
	}
	header.Commitment = headerCommitment(shard)
	s.state.Headers[shard.ID.Key()] = header
}

func headerCommitment(shard ShardState) string {
	return HashParts("header", shard.ID.Key(), fmt.Sprint(shard.Height), shard.StateRoot, shard.MessageQueueRoot, shard.ReceiptRoot, strings.Join(shard.ValidatorSubset, ","), fmt.Sprint(shard.Available))
}

func hashQueue(queue []CrossShardMessage) string {
	ids := make([]string, len(queue))
	for i, msg := range queue {
		ids[i] = msg.MessageID
	}
	sort.Strings(ids)
	return HashParts(append([]string{"queue"}, ids...)...)
}

func hashReceipts(receipts map[string]Receipt) string {
	ids := make([]string, 0, len(receipts))
	for id := range receipts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return HashParts(append([]string{"receipts"}, ids...)...)
}

func sortedShardKeys(shards map[string]ShardState) []string {
	keys := make([]string, 0, len(shards))
	for key := range shards {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func cloneValidators(in []Validator) []Validator {
	out := make([]Validator, len(in))
	copy(out, in)
	return out
}

func cloneMessage(msg CrossShardMessage) CrossShardMessage {
	msg.Payload = append([]byte(nil), msg.Payload...)
	return msg
}

func cloneQueue(in []CrossShardMessage) []CrossShardMessage {
	out := make([]CrossShardMessage, len(in))
	for i, msg := range in {
		out[i] = cloneMessage(msg)
	}
	return out
}

func cloneIntMap(in map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneWorkchains(in map[int32]WorkchainConfig) map[int32]WorkchainConfig {
	out := make(map[int32]WorkchainConfig, len(in))
	for key, value := range in {
		value.AllowedVMs = append([]string(nil), value.AllowedVMs...)
		out[key] = value
	}
	return out
}

func cloneShards(in map[string]ShardState) map[string]ShardState {
	out := make(map[string]ShardState, len(in))
	for key, value := range in {
		value.ValidatorSubset = append([]string(nil), value.ValidatorSubset...)
		value.Queue = cloneQueue(value.Queue)
		value.Receipts = cloneReceipts(value.Receipts)
		out[key] = value
	}
	return out
}

func cloneHeaders(in map[string]ShardHeader) map[string]ShardHeader {
	out := make(map[string]ShardHeader, len(in))
	for key, value := range in {
		value.ValidatorSubset = append([]string(nil), value.ValidatorSubset...)
		out[key] = value
	}
	return out
}

func cloneReceipts(in map[string]Receipt) map[string]Receipt {
	out := make(map[string]Receipt, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func mergeReceipts(left, right map[string]Receipt) map[string]Receipt {
	out := cloneReceipts(left)
	for key, value := range right {
		out[key] = value
	}
	return out
}

func max(left, right uint64) uint64 {
	if left > right {
		return left
	}
	return right
}
