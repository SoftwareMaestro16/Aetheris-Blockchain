package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkingState struct {
	ChannelPolicies []ChannelPolicy
	NodeRecords     []NodeRecord
	Sessions        []SessionChannel
}

func EmptyState() NetworkingState {
	return NetworkingState{
		ChannelPolicies: DefaultChannelPolicies(),
		NodeRecords:     []NodeRecord{},
		Sessions:        []SessionChannel{},
	}
}

func RegisterNodeRecord(state NetworkingState, record NodeRecord, networkSalt []byte, currentHeight uint64) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	record = NormalizeNodeRecord(record)
	if err := record.Validate(networkSalt, currentHeight); err != nil {
		return NetworkingState{}, err
	}
	next := state.Clone()
	replaced := false
	for i, existing := range next.NodeRecords {
		if existing.NodeID == record.NodeID {
			next.NodeRecords[i] = record
			replaced = true
			break
		}
	}
	if !replaced {
		next.NodeRecords = append(next.NodeRecords, record)
	}
	sortNodeRecords(next.NodeRecords)
	return next, next.Validate()
}

func OpenSession(state NetworkingState, session SessionChannel, currentHeight uint64) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	if currentHeight == 0 {
		return NetworkingState{}, errors.New("networking current height must be positive")
	}
	if err := session.Validate(); err != nil {
		return NetworkingState{}, err
	}
	if currentHeight > session.ExpiresHeight {
		return NetworkingState{}, errors.New("networking session is expired")
	}
	if !state.hasNode(session.LocalNodeID) || !state.hasNode(session.RemoteNodeID) {
		return NetworkingState{}, errors.New("networking session requires registered endpoints")
	}
	for _, existing := range state.Sessions {
		if existing.SessionID == session.SessionID {
			return NetworkingState{}, errors.New("networking session already exists")
		}
	}
	next := state.Clone()
	next.Sessions = append(next.Sessions, session)
	sortSessions(next.Sessions)
	return next, next.Validate()
}

func PruneExpired(state NetworkingState, currentHeight uint64) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	next := NetworkingState{ChannelPolicies: cloneChannelPolicies(state.ChannelPolicies)}
	for _, record := range state.NodeRecords {
		if currentHeight == 0 || currentHeight <= record.ExpiresHeight {
			next.NodeRecords = append(next.NodeRecords, record)
		}
	}
	for _, session := range state.Sessions {
		if currentHeight == 0 || currentHeight <= session.ExpiresHeight {
			if containsNode(next.NodeRecords, session.LocalNodeID) && containsNode(next.NodeRecords, session.RemoteNodeID) {
				next.Sessions = append(next.Sessions, session)
			}
		}
	}
	sortNodeRecords(next.NodeRecords)
	sortSessions(next.Sessions)
	return next, next.Validate()
}

func ImportState(state NetworkingState) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	return state, nil
}

func (s NetworkingState) Export() NetworkingState {
	out := s.Clone()
	if len(out.ChannelPolicies) == 0 {
		out.ChannelPolicies = DefaultChannelPolicies()
	}
	sortChannelPolicies(out.ChannelPolicies)
	sortNodeRecords(out.NodeRecords)
	sortSessions(out.Sessions)
	return out
}

func (s NetworkingState) Clone() NetworkingState {
	out := NetworkingState{
		ChannelPolicies: cloneChannelPolicies(s.ChannelPolicies),
		NodeRecords:     make([]NodeRecord, len(s.NodeRecords)),
		Sessions:        make([]SessionChannel, len(s.Sessions)),
	}
	for i, record := range s.NodeRecords {
		out.NodeRecords[i] = NormalizeNodeRecord(record)
	}
	for i, session := range s.Sessions {
		out.Sessions[i] = cloneSession(session)
	}
	return out
}

func (s NetworkingState) Validate() error {
	if err := ValidateChannelPolicies(s.ChannelPolicies); err != nil {
		return err
	}
	if err := validateNodeRecords(s.NodeRecords); err != nil {
		return err
	}
	return validateSessions(s.NodeRecords, s.Sessions)
}

func (s NetworkingState) hasNode(nodeID string) bool {
	return containsNode(s.NodeRecords, nodeID)
}

func validateNodeRecords(records []NodeRecord) error {
	seen := make(map[string]struct{}, len(records))
	var previous string
	for i, record := range records {
		normalized := NormalizeNodeRecord(record)
		if err := normalized.ValidateBasic(); err != nil {
			return err
		}
		if _, found := seen[normalized.NodeID]; found {
			return errors.New("networking duplicate node record")
		}
		seen[normalized.NodeID] = struct{}{}
		if i > 0 && previous >= normalized.NodeID {
			return errors.New("networking node records must be sorted canonically")
		}
		previous = normalized.NodeID
	}
	return nil
}

func validateSessions(records []NodeRecord, sessions []SessionChannel) error {
	seen := make(map[string]struct{}, len(sessions))
	var previous string
	for i, session := range sessions {
		session.SessionID = normalizeHashText(session.SessionID)
		if err := session.Validate(); err != nil {
			return err
		}
		if !containsNode(records, session.LocalNodeID) || !containsNode(records, session.RemoteNodeID) {
			return errors.New("networking session references unknown node")
		}
		if _, found := seen[session.SessionID]; found {
			return errors.New("networking duplicate session")
		}
		seen[session.SessionID] = struct{}{}
		if i > 0 && previous >= session.SessionID {
			return errors.New("networking sessions must be sorted canonically")
		}
		previous = session.SessionID
	}
	return nil
}

func containsNode(records []NodeRecord, nodeID string) bool {
	needle := normalizeHashText(nodeID)
	for _, record := range records {
		if NormalizeNodeRecord(record).NodeID == needle {
			return true
		}
	}
	return false
}

func cloneChannelPolicies(policies []ChannelPolicy) []ChannelPolicy {
	out := make([]ChannelPolicy, len(policies))
	copy(out, policies)
	return out
}

func cloneSession(session SessionChannel) SessionChannel {
	session.LocalNodeID = normalizeHashText(session.LocalNodeID)
	session.RemoteNodeID = normalizeHashText(session.RemoteNodeID)
	session.SessionID = normalizeHashText(session.SessionID)
	session.ProtocolVersions = append([]string(nil), session.ProtocolVersions...)
	sortStrings(session.ProtocolVersions)
	session.Streams = append([]StreamSpec(nil), session.Streams...)
	sort.SliceStable(session.Streams, func(i, j int) bool {
		if session.Streams[i].Priority != session.Streams[j].Priority {
			return session.Streams[i].Priority < session.Streams[j].Priority
		}
		return session.Streams[i].StreamID < session.Streams[j].StreamID
	})
	return session
}

func sortChannelPolicies(policies []ChannelPolicy) {
	sort.SliceStable(policies, func(i, j int) bool {
		if policies[i].Priority != policies[j].Priority {
			return policies[i].Priority < policies[j].Priority
		}
		return policies[i].Channel < policies[j].Channel
	})
}

func sortNodeRecords(records []NodeRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		return NormalizeNodeRecord(records[i]).NodeID < NormalizeNodeRecord(records[j]).NodeID
	})
}

func sortSessions(sessions []SessionChannel) {
	sort.SliceStable(sessions, func(i, j int) bool {
		return sessions[i].SessionID < sessions[j].SessionID
	})
}

func normalizeHashText(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (s NetworkingState) DebugString() string {
	return fmt.Sprintf("networking nodes=%d sessions=%d channels=%d", len(s.NodeRecords), len(s.Sessions), len(s.ChannelPolicies))
}
