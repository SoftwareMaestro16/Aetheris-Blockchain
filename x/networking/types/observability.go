package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkingObservableMetric string

const (
	ObservableMetricActivePeers                     NetworkingObservableMetric = "active_peers"
	ObservableMetricPeersByRole                     NetworkingObservableMetric = "peers_by_role"
	ObservableMetricActiveSessions                  NetworkingObservableMetric = "active_sessions"
	ObservableMetricStreamsByChannelType            NetworkingObservableMetric = "streams_by_channel_type"
	ObservableMetricPerChannelBandwidth             NetworkingObservableMetric = "per_channel_bandwidth"
	ObservableMetricPeerScore                       NetworkingObservableMetric = "peer_score"
	ObservableMetricOverlaySize                     NetworkingObservableMetric = "overlay_size"
	ObservableMetricOverlayChurn                    NetworkingObservableMetric = "overlay_churn"
	ObservableMetricDiscoveryQueryLatency           NetworkingObservableMetric = "discovery_query_latency"
	ObservableMetricBroadcastDedupHitRate           NetworkingObservableMetric = "broadcast_dedup_hit_rate"
	ObservableMetricRL2TransferThroughput           NetworkingObservableMetric = "rl2_transfer_throughput"
	ObservableMetricRL2ChunkRetryRate               NetworkingObservableMetric = "rl2_chunk_retry_rate"
	ObservableMetricBlockPropagationLatency         NetworkingObservableMetric = "block_propagation_latency"
	ObservableMetricCrossZoneMessageDeliveryLatency NetworkingObservableMetric = "cross_zone_message_delivery_latency"
	ObservableMetricServiceTrafficVolume            NetworkingObservableMetric = "service_traffic_volume"
	ObservableMetricRoutingFailureCount             NetworkingObservableMetric = "routing_failure_count"
)

type NetworkingObservableEvent string

const (
	ObservableEventNetworkNodeRegistered         NetworkingObservableEvent = "network_node_registered"
	ObservableEventNetworkSessionOpened          NetworkingObservableEvent = "network_session_opened"
	ObservableEventNetworkSessionClosed          NetworkingObservableEvent = "network_session_closed"
	ObservableEventNetworkPeerScoreUpdated       NetworkingObservableEvent = "network_peer_score_updated"
	ObservableEventNetworkOverlayJoined          NetworkingObservableEvent = "network_overlay_joined"
	ObservableEventNetworkOverlayLeft            NetworkingObservableEvent = "network_overlay_left"
	ObservableEventNetworkDiscoveryRecordStored  NetworkingObservableEvent = "network_discovery_record_stored"
	ObservableEventNetworkDiscoveryRecordExpired NetworkingObservableEvent = "network_discovery_record_expired"
	ObservableEventNetworkRL2TransferStarted     NetworkingObservableEvent = "network_rl2_transfer_started"
	ObservableEventNetworkRL2TransferCompleted   NetworkingObservableEvent = "network_rl2_transfer_completed"
	ObservableEventNetworkInvalidChunk           NetworkingObservableEvent = "network_invalid_chunk"
	ObservableEventNetworkBroadcastConflict      NetworkingObservableEvent = "network_broadcast_conflict"
	ObservableEventNetworkRouteFailed            NetworkingObservableEvent = "network_route_failed"
)

type NetworkingMetricSample struct {
	Metric NetworkingObservableMetric
	Labels []string
	Value  uint64
	Height uint64
}

type NetworkingEventRecord struct {
	Event        NetworkingObservableEvent
	NodeID       string
	OverlayID    string
	Channel      ChannelClass
	TransferID   string
	MessageID    string
	EvidenceHash string
	Height       uint64
	EventID      string
}

type NetworkingObservabilitySpec struct {
	Metrics  []NetworkingObservableMetric
	Events   []NetworkingObservableEvent
	SpecRoot string
}

type NetworkingObservabilityReport struct {
	Spec           NetworkingObservabilitySpec
	Metrics        []NetworkingMetricSample
	Events         []NetworkingEventRecord
	MissingMetrics []NetworkingObservableMetric
	MissingEvents  []NetworkingObservableEvent
	Ready          bool
	ReportHash     string
}

func DefaultNetworkingObservabilitySpec() NetworkingObservabilitySpec {
	spec := NetworkingObservabilitySpec{
		Metrics: []NetworkingObservableMetric{
			ObservableMetricActivePeers,
			ObservableMetricPeersByRole,
			ObservableMetricActiveSessions,
			ObservableMetricStreamsByChannelType,
			ObservableMetricPerChannelBandwidth,
			ObservableMetricPeerScore,
			ObservableMetricOverlaySize,
			ObservableMetricOverlayChurn,
			ObservableMetricDiscoveryQueryLatency,
			ObservableMetricBroadcastDedupHitRate,
			ObservableMetricRL2TransferThroughput,
			ObservableMetricRL2ChunkRetryRate,
			ObservableMetricBlockPropagationLatency,
			ObservableMetricCrossZoneMessageDeliveryLatency,
			ObservableMetricServiceTrafficVolume,
			ObservableMetricRoutingFailureCount,
		},
		Events: []NetworkingObservableEvent{
			ObservableEventNetworkNodeRegistered,
			ObservableEventNetworkSessionOpened,
			ObservableEventNetworkSessionClosed,
			ObservableEventNetworkPeerScoreUpdated,
			ObservableEventNetworkOverlayJoined,
			ObservableEventNetworkOverlayLeft,
			ObservableEventNetworkDiscoveryRecordStored,
			ObservableEventNetworkDiscoveryRecordExpired,
			ObservableEventNetworkRL2TransferStarted,
			ObservableEventNetworkRL2TransferCompleted,
			ObservableEventNetworkInvalidChunk,
			ObservableEventNetworkBroadcastConflict,
			ObservableEventNetworkRouteFailed,
		},
	}
	spec = NormalizeNetworkingObservabilitySpec(spec)
	spec.SpecRoot = ComputeNetworkingObservabilitySpecRoot(spec)
	return spec
}

func ValidateNetworkingObservabilitySpec(spec NetworkingObservabilitySpec) error {
	normalized := NormalizeNetworkingObservabilitySpec(spec)
	required := DefaultNetworkingObservabilitySpec()
	if len(normalized.Metrics) != len(required.Metrics) {
		return fmt.Errorf("networking observability spec must define %d metrics", len(required.Metrics))
	}
	if len(normalized.Events) != len(required.Events) {
		return fmt.Errorf("networking observability spec must define %d events", len(required.Events))
	}
	seenMetrics := make(map[NetworkingObservableMetric]struct{}, len(normalized.Metrics))
	for _, metric := range normalized.Metrics {
		if !IsNetworkingObservableMetric(metric) {
			return fmt.Errorf("unknown networking observability metric %q", metric)
		}
		if _, found := seenMetrics[metric]; found {
			return errors.New("networking observability duplicate metric")
		}
		seenMetrics[metric] = struct{}{}
	}
	for _, metric := range required.Metrics {
		if _, found := seenMetrics[metric]; !found {
			return fmt.Errorf("networking observability missing metric %s", metric)
		}
	}
	seenEvents := make(map[NetworkingObservableEvent]struct{}, len(normalized.Events))
	for _, event := range normalized.Events {
		if !IsNetworkingObservableEvent(event) {
			return fmt.Errorf("unknown networking observability event %q", event)
		}
		if _, found := seenEvents[event]; found {
			return errors.New("networking observability duplicate event")
		}
		seenEvents[event] = struct{}{}
	}
	for _, event := range required.Events {
		if _, found := seenEvents[event]; !found {
			return fmt.Errorf("networking observability missing event %s", event)
		}
	}
	if normalized.SpecRoot == "" {
		return errors.New("networking observability spec root is required")
	}
	if normalized.SpecRoot != ComputeNetworkingObservabilitySpecRoot(normalized) {
		return errors.New("networking observability spec root mismatch")
	}
	return nil
}

func BuildNetworkingObservabilityReport(spec NetworkingObservabilitySpec, metrics []NetworkingMetricSample, events []NetworkingEventRecord) (NetworkingObservabilityReport, error) {
	spec = NormalizeNetworkingObservabilitySpec(spec)
	if err := ValidateNetworkingObservabilitySpec(spec); err != nil {
		return NetworkingObservabilityReport{}, err
	}
	normalizedMetrics := NormalizeNetworkingMetricSamples(metrics)
	normalizedEvents := NormalizeNetworkingEventRecords(events)
	report := NetworkingObservabilityReport{
		Spec:    spec,
		Metrics: normalizedMetrics,
		Events:  normalizedEvents,
	}
	coveredMetrics := make(map[NetworkingObservableMetric]struct{}, len(normalizedMetrics))
	for _, sample := range normalizedMetrics {
		if err := sample.Validate(); err != nil {
			return NetworkingObservabilityReport{}, err
		}
		coveredMetrics[sample.Metric] = struct{}{}
	}
	for _, metric := range spec.Metrics {
		if _, found := coveredMetrics[metric]; !found {
			report.MissingMetrics = append(report.MissingMetrics, metric)
		}
	}
	coveredEvents := make(map[NetworkingObservableEvent]struct{}, len(normalizedEvents))
	for _, event := range normalizedEvents {
		if err := event.Validate(); err != nil {
			return NetworkingObservabilityReport{}, err
		}
		coveredEvents[event.Event] = struct{}{}
	}
	for _, event := range spec.Events {
		if _, found := coveredEvents[event]; !found {
			report.MissingEvents = append(report.MissingEvents, event)
		}
	}
	sortObservableMetrics(report.MissingMetrics)
	sortObservableEvents(report.MissingEvents)
	report.Ready = len(report.MissingMetrics) == 0 && len(report.MissingEvents) == 0
	report.ReportHash = ComputeNetworkingObservabilityReportHash(report)
	return report, nil
}

func NewNetworkingEventRecord(event NetworkingObservableEvent, nodeID, overlayID string, channel ChannelClass, transferID, messageID, evidenceHash string, height uint64) NetworkingEventRecord {
	record := NetworkingEventRecord{
		Event:        event,
		NodeID:       nodeID,
		OverlayID:    overlayID,
		Channel:      channel,
		TransferID:   transferID,
		MessageID:    messageID,
		EvidenceHash: evidenceHash,
		Height:       height,
	}
	record = NormalizeNetworkingEventRecord(record)
	record.EventID = ComputeNetworkingEventID(record)
	return record
}

func (sample NetworkingMetricSample) Validate() error {
	sample = NormalizeNetworkingMetricSample(sample)
	if !IsNetworkingObservableMetric(sample.Metric) {
		return fmt.Errorf("unknown networking observability metric %q", sample.Metric)
	}
	if sample.Height == 0 {
		return errors.New("networking observability metric height must be positive")
	}
	return nil
}

func (record NetworkingEventRecord) Validate() error {
	record = NormalizeNetworkingEventRecord(record)
	if !IsNetworkingObservableEvent(record.Event) {
		return fmt.Errorf("unknown networking observability event %q", record.Event)
	}
	if record.Height == 0 {
		return errors.New("networking observability event height must be positive")
	}
	if record.NodeID != "" {
		if err := ValidateHash("networking observability event node id", record.NodeID); err != nil {
			return err
		}
	}
	if record.OverlayID != "" {
		if err := ValidateHash("networking observability event overlay id", record.OverlayID); err != nil {
			return err
		}
	}
	if record.Channel != "" && !IsChannelClass(record.Channel) {
		return fmt.Errorf("unknown networking observability event channel %q", record.Channel)
	}
	if record.TransferID != "" {
		if err := ValidateHash("networking observability event transfer id", record.TransferID); err != nil {
			return err
		}
	}
	if record.MessageID != "" {
		if err := ValidateHash("networking observability event message id", record.MessageID); err != nil {
			return err
		}
	}
	if record.EvidenceHash != "" {
		if err := ValidateHash("networking observability event evidence hash", record.EvidenceHash); err != nil {
			return err
		}
	}
	if err := ValidateHash("networking observability event id", record.EventID); err != nil {
		return err
	}
	if record.EventID != ComputeNetworkingEventID(record) {
		return errors.New("networking observability event id mismatch")
	}
	return nil
}

func ComputeNetworkingObservabilitySpecRoot(spec NetworkingObservabilitySpec) string {
	spec = NormalizeNetworkingObservabilitySpec(spec)
	parts := []string{"networking-observability-spec"}
	for _, metric := range spec.Metrics {
		parts = append(parts, "metric", string(metric))
	}
	for _, event := range spec.Events {
		parts = append(parts, "event", string(event))
	}
	return HashParts(parts...)
}

func ComputeNetworkingObservabilityReportHash(report NetworkingObservabilityReport) string {
	parts := []string{"networking-observability-report", report.Spec.SpecRoot, fmt.Sprintf("%t", report.Ready)}
	for _, sample := range NormalizeNetworkingMetricSamples(report.Metrics) {
		parts = append(parts, "metric", string(sample.Metric), fmt.Sprintf("%d", sample.Value), fmt.Sprintf("%d", sample.Height))
		parts = append(parts, sample.Labels...)
	}
	for _, event := range NormalizeNetworkingEventRecords(report.Events) {
		parts = append(parts, "event", string(event.Event), event.EventID, event.NodeID, event.OverlayID, string(event.Channel), event.TransferID, event.MessageID, event.EvidenceHash, fmt.Sprintf("%d", event.Height))
	}
	for _, metric := range report.MissingMetrics {
		parts = append(parts, "missing_metric", string(metric))
	}
	for _, event := range report.MissingEvents {
		parts = append(parts, "missing_event", string(event))
	}
	return HashParts(parts...)
}

func ComputeNetworkingEventID(record NetworkingEventRecord) string {
	record = NormalizeNetworkingEventRecord(record)
	return HashParts(
		"networking-observability-event",
		string(record.Event),
		record.NodeID,
		record.OverlayID,
		string(record.Channel),
		record.TransferID,
		record.MessageID,
		record.EvidenceHash,
		fmt.Sprintf("%d", record.Height),
	)
}

func IsNetworkingObservableMetric(metric NetworkingObservableMetric) bool {
	switch metric {
	case ObservableMetricActivePeers,
		ObservableMetricPeersByRole,
		ObservableMetricActiveSessions,
		ObservableMetricStreamsByChannelType,
		ObservableMetricPerChannelBandwidth,
		ObservableMetricPeerScore,
		ObservableMetricOverlaySize,
		ObservableMetricOverlayChurn,
		ObservableMetricDiscoveryQueryLatency,
		ObservableMetricBroadcastDedupHitRate,
		ObservableMetricRL2TransferThroughput,
		ObservableMetricRL2ChunkRetryRate,
		ObservableMetricBlockPropagationLatency,
		ObservableMetricCrossZoneMessageDeliveryLatency,
		ObservableMetricServiceTrafficVolume,
		ObservableMetricRoutingFailureCount:
		return true
	default:
		return false
	}
}

func IsNetworkingObservableEvent(event NetworkingObservableEvent) bool {
	switch event {
	case ObservableEventNetworkNodeRegistered,
		ObservableEventNetworkSessionOpened,
		ObservableEventNetworkSessionClosed,
		ObservableEventNetworkPeerScoreUpdated,
		ObservableEventNetworkOverlayJoined,
		ObservableEventNetworkOverlayLeft,
		ObservableEventNetworkDiscoveryRecordStored,
		ObservableEventNetworkDiscoveryRecordExpired,
		ObservableEventNetworkRL2TransferStarted,
		ObservableEventNetworkRL2TransferCompleted,
		ObservableEventNetworkInvalidChunk,
		ObservableEventNetworkBroadcastConflict,
		ObservableEventNetworkRouteFailed:
		return true
	default:
		return false
	}
}

func NormalizeNetworkingObservabilitySpec(spec NetworkingObservabilitySpec) NetworkingObservabilitySpec {
	spec.Metrics = normalizeObservableMetrics(spec.Metrics)
	spec.Events = normalizeObservableEvents(spec.Events)
	spec.SpecRoot = normalizeHashText(spec.SpecRoot)
	return spec
}

func NormalizeNetworkingMetricSamples(samples []NetworkingMetricSample) []NetworkingMetricSample {
	out := make([]NetworkingMetricSample, 0, len(samples))
	for _, sample := range samples {
		sample = NormalizeNetworkingMetricSample(sample)
		if sample.Metric == "" {
			continue
		}
		out = append(out, sample)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Metric != out[j].Metric {
			return out[i].Metric < out[j].Metric
		}
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return strings.Join(out[i].Labels, "\x00") < strings.Join(out[j].Labels, "\x00")
	})
	return out
}

func NormalizeNetworkingMetricSample(sample NetworkingMetricSample) NetworkingMetricSample {
	sample.Metric = NetworkingObservableMetric(strings.ToLower(strings.TrimSpace(string(sample.Metric))))
	labels := make([]string, 0, len(sample.Labels))
	seen := make(map[string]struct{}, len(sample.Labels))
	for _, label := range sample.Labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		if _, found := seen[label]; found {
			continue
		}
		seen[label] = struct{}{}
		labels = append(labels, label)
	}
	sort.Strings(labels)
	sample.Labels = labels
	return sample
}

func NormalizeNetworkingEventRecords(records []NetworkingEventRecord) []NetworkingEventRecord {
	out := make([]NetworkingEventRecord, 0, len(records))
	for _, record := range records {
		record = NormalizeNetworkingEventRecord(record)
		if record.Event == "" {
			continue
		}
		out = append(out, record)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Event != out[j].Event {
			return out[i].Event < out[j].Event
		}
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return out[i].EventID < out[j].EventID
	})
	return out
}

func NormalizeNetworkingEventRecord(record NetworkingEventRecord) NetworkingEventRecord {
	record.Event = NetworkingObservableEvent(strings.ToLower(strings.TrimSpace(string(record.Event))))
	record.NodeID = normalizeHashText(record.NodeID)
	record.OverlayID = normalizeHashText(record.OverlayID)
	record.Channel = ChannelClass(strings.ToUpper(strings.TrimSpace(string(record.Channel))))
	record.TransferID = normalizeHashText(record.TransferID)
	record.MessageID = normalizeHashText(record.MessageID)
	record.EvidenceHash = normalizeHashText(record.EvidenceHash)
	record.EventID = normalizeHashText(record.EventID)
	return record
}

func normalizeObservableMetrics(metrics []NetworkingObservableMetric) []NetworkingObservableMetric {
	out := make([]NetworkingObservableMetric, 0, len(metrics))
	seen := make(map[NetworkingObservableMetric]struct{}, len(metrics))
	for _, metric := range metrics {
		metric = NetworkingObservableMetric(strings.ToLower(strings.TrimSpace(string(metric))))
		if metric == "" {
			continue
		}
		if _, found := seen[metric]; found {
			continue
		}
		seen[metric] = struct{}{}
		out = append(out, metric)
	}
	sortObservableMetrics(out)
	return out
}

func normalizeObservableEvents(events []NetworkingObservableEvent) []NetworkingObservableEvent {
	out := make([]NetworkingObservableEvent, 0, len(events))
	seen := make(map[NetworkingObservableEvent]struct{}, len(events))
	for _, event := range events {
		event = NetworkingObservableEvent(strings.ToLower(strings.TrimSpace(string(event))))
		if event == "" {
			continue
		}
		if _, found := seen[event]; found {
			continue
		}
		seen[event] = struct{}{}
		out = append(out, event)
	}
	sortObservableEvents(out)
	return out
}

func sortObservableMetrics(values []NetworkingObservableMetric) {
	sort.SliceStable(values, func(i, j int) bool { return values[i] < values[j] })
}

func sortObservableEvents(values []NetworkingObservableEvent) {
	sort.SliceStable(values, func(i, j int) bool { return values[i] < values[j] })
}
