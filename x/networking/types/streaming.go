package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultStreamParallelism = uint32(1)
	MaxStreamParallelism     = uint32(64)
)

type StreamingPayloadType string

const (
	StreamingPayloadStateSync            StreamingPayloadType = "state_sync"
	StreamingPayloadZoneSnapshot         StreamingPayloadType = "zone_snapshot"
	StreamingPayloadBlockPropagation     StreamingPayloadType = "block_propagation"
	StreamingPayloadExecutionReceipts    StreamingPayloadType = "execution_receipts"
	StreamingPayloadStorageObject        StreamingPayloadType = "storage_object"
	StreamingPayloadProofBundle          StreamingPayloadType = "proof_bundle"
	StreamingPayloadHistoricalQueryRange StreamingPayloadType = "historical_query_range"
)

type StreamSessionState string

const (
	StreamStateOpening  StreamSessionState = "opening"
	StreamStateActive   StreamSessionState = "active"
	StreamStatePaused   StreamSessionState = "paused"
	StreamStateDraining StreamSessionState = "draining"
	StreamStateClosed   StreamSessionState = "closed"
	StreamStateFailed   StreamSessionState = "failed"
)

type StreamSession struct {
	StreamID          string
	SessionID         string
	PayloadType       StreamingPayloadType
	Priority          uint32
	FlowControlWindow uint64
	ChunkSize         uint64
	Parallelism       uint32
	BytesSent         uint64
	BytesAcknowledged uint64
	State             StreamSessionState
}

type StreamWindowUpdate struct {
	StreamID          string
	BytesSent         uint64
	BytesAcknowledged uint64
	AvailableWindow   uint64
	Backpressure      bool
}

type StreamChunkPlan struct {
	StreamID          string
	NextOffset        uint64
	ChunkSize         uint64
	MaxInFlightChunks uint32
	Parallelism       uint32
	Backpressure      bool
}

func NewStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.StreamID == "" {
		stream.StreamID = ComputeStreamSessionID(stream)
	}
	if err := stream.Validate(); err != nil {
		return StreamSession{}, err
	}
	return stream, nil
}

func StreamSessionFromSpec(session SessionChannel, spec StreamSpec, payloadType StreamingPayloadType, chunkSize uint64, parallelism uint32) (StreamSession, error) {
	if err := session.Validate(); err != nil {
		return StreamSession{}, err
	}
	spec.StreamID = strings.TrimSpace(spec.StreamID)
	spec.EncryptionContext = strings.TrimSpace(spec.EncryptionContext)
	found := false
	for _, stream := range session.Streams {
		if strings.TrimSpace(stream.StreamID) == spec.StreamID {
			found = true
			spec = stream
			break
		}
	}
	if !found {
		return StreamSession{}, errors.New("networking stream spec is not part of session")
	}
	if err := spec.Validate(); err != nil {
		return StreamSession{}, err
	}
	if chunkSize == 0 {
		chunkSize = spec.MaxMessageBytes
		if chunkSize > MaxChunkBytes {
			chunkSize = MaxChunkBytes
		}
	}
	return NewStreamSession(StreamSession{
		SessionID:         session.SessionID,
		PayloadType:       payloadType,
		Priority:          spec.Priority,
		FlowControlWindow: spec.FlowControlWindow,
		ChunkSize:         chunkSize,
		Parallelism:       parallelism,
		State:             StreamStateOpening,
	})
}

func (s StreamSession) Normalize() StreamSession {
	s.StreamID = normalizeHashText(s.StreamID)
	s.SessionID = normalizeHashText(s.SessionID)
	s.State = normalizeStreamingState(s.State)
	if s.Parallelism == 0 {
		s.Parallelism = DefaultStreamParallelism
	}
	if s.State == "" {
		s.State = StreamStateOpening
	}
	return s
}

func (s StreamSession) Validate() error {
	stream := s.Normalize()
	if err := ValidateHash("networking stream session id", stream.StreamID); err != nil {
		return err
	}
	if ComputeStreamSessionID(stream) != stream.StreamID {
		return errors.New("networking stream session id mismatch")
	}
	if err := ValidateHash("networking parent session id", stream.SessionID); err != nil {
		return err
	}
	if !IsStreamingPayloadType(stream.PayloadType) {
		return fmt.Errorf("unknown networking streaming payload type %q", stream.PayloadType)
	}
	if stream.Priority > MaxRL2Priority {
		return fmt.Errorf("networking stream priority must be <= %d", MaxRL2Priority)
	}
	if stream.FlowControlWindow == 0 || stream.FlowControlWindow > MaxStreamMessageBytes*2 {
		return fmt.Errorf("networking stream flow control window must be between 1 and %d", uint64(MaxStreamMessageBytes*2))
	}
	if stream.ChunkSize == 0 || stream.ChunkSize > MaxChunkBytes {
		return fmt.Errorf("networking stream chunk size must be between 1 and %d", MaxChunkBytes)
	}
	if stream.ChunkSize > stream.FlowControlWindow {
		return errors.New("networking stream chunk size exceeds flow control window")
	}
	if stream.Parallelism == 0 || stream.Parallelism > MaxStreamParallelism {
		return fmt.Errorf("networking stream parallelism must be between 1 and %d", MaxStreamParallelism)
	}
	if stream.BytesAcknowledged > stream.BytesSent {
		return errors.New("networking stream acknowledged bytes exceed sent bytes")
	}
	if !IsStreamSessionState(stream.State) {
		return fmt.Errorf("unknown networking stream session state %q", stream.State)
	}
	if stream.State == StreamStateOpening && (stream.BytesSent != 0 || stream.BytesAcknowledged != 0) {
		return errors.New("networking opening stream session cannot have byte progress")
	}
	return nil
}

func ComputeStreamSessionID(stream StreamSession) string {
	stream.StreamID = ""
	stream = stream.Normalize()
	return HashParts(
		"stream-session",
		stream.SessionID,
		string(stream.PayloadType),
		fmt.Sprintf("%d", stream.Priority),
		fmt.Sprintf("%d", stream.FlowControlWindow),
		fmt.Sprintf("%d", stream.ChunkSize),
		fmt.Sprintf("%d", stream.Parallelism),
	)
}

func OpenStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State != StreamStateOpening {
		return StreamSession{}, errors.New("networking stream session can open only from opening state")
	}
	stream.State = StreamStateActive
	return stream, stream.Validate()
}

func PauseStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State != StreamStateActive {
		return StreamSession{}, errors.New("networking stream session can pause only from active state")
	}
	stream.State = StreamStatePaused
	return stream, stream.Validate()
}

func ResumeStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State != StreamStatePaused {
		return StreamSession{}, errors.New("networking stream session can resume only from paused state")
	}
	stream.State = StreamStateActive
	return stream, stream.Validate()
}

func DrainStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State != StreamStateActive && stream.State != StreamStatePaused {
		return StreamSession{}, errors.New("networking stream session can drain only from active or paused state")
	}
	stream.State = StreamStateDraining
	return stream, stream.Validate()
}

func CloseStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State != StreamStateDraining {
		return StreamSession{}, errors.New("networking stream session can close only from draining state")
	}
	if stream.BytesAcknowledged != stream.BytesSent {
		return StreamSession{}, errors.New("networking stream session cannot close with unacknowledged bytes")
	}
	stream.State = StreamStateClosed
	return stream, stream.Validate()
}

func FailStreamSession(stream StreamSession) (StreamSession, error) {
	stream = stream.Normalize()
	if stream.State == StreamStateClosed || stream.State == StreamStateFailed {
		return StreamSession{}, errors.New("networking terminal stream session cannot fail again")
	}
	stream.State = StreamStateFailed
	return stream, stream.Validate()
}

func RecordStreamBytesSent(stream StreamSession, bytes uint64) (StreamSession, StreamWindowUpdate, error) {
	stream = stream.Normalize()
	if stream.State != StreamStateActive && stream.State != StreamStateDraining {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream session must be active or draining to send bytes")
	}
	if bytes == 0 {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream sent byte increment must be positive")
	}
	available := StreamAvailableWindow(stream)
	if bytes > available {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream flow control window exceeded")
	}
	stream.BytesSent += bytes
	if err := stream.Validate(); err != nil {
		return StreamSession{}, StreamWindowUpdate{}, err
	}
	return stream, StreamWindow(stream), nil
}

func AcknowledgeStreamBytes(stream StreamSession, acknowledged uint64) (StreamSession, StreamWindowUpdate, error) {
	stream = stream.Normalize()
	if stream.State == StreamStateOpening || stream.State == StreamStateClosed || stream.State == StreamStateFailed {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream session cannot acknowledge bytes in current state")
	}
	if acknowledged < stream.BytesAcknowledged {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream acknowledged bytes cannot regress")
	}
	if acknowledged > stream.BytesSent {
		return StreamSession{}, StreamWindowUpdate{}, errors.New("networking stream acknowledged bytes exceed sent bytes")
	}
	stream.BytesAcknowledged = acknowledged
	if err := stream.Validate(); err != nil {
		return StreamSession{}, StreamWindowUpdate{}, err
	}
	return stream, StreamWindow(stream), nil
}

func StreamWindow(stream StreamSession) StreamWindowUpdate {
	stream = stream.Normalize()
	available := StreamAvailableWindow(stream)
	return StreamWindowUpdate{
		StreamID:          stream.StreamID,
		BytesSent:         stream.BytesSent,
		BytesAcknowledged: stream.BytesAcknowledged,
		AvailableWindow:   available,
		Backpressure:      available == 0,
	}
}

func StreamAvailableWindow(stream StreamSession) uint64 {
	stream = stream.Normalize()
	inFlight := stream.BytesSent - stream.BytesAcknowledged
	if inFlight >= stream.FlowControlWindow {
		return 0
	}
	return stream.FlowControlWindow - inFlight
}

func PlanStreamChunks(stream StreamSession, remainingBytes uint64) (StreamChunkPlan, error) {
	stream = stream.Normalize()
	if err := stream.Validate(); err != nil {
		return StreamChunkPlan{}, err
	}
	if stream.State != StreamStateActive && stream.State != StreamStateDraining {
		return StreamChunkPlan{}, errors.New("networking stream session must be active or draining to plan chunks")
	}
	if remainingBytes == 0 {
		return StreamChunkPlan{
			StreamID:     stream.StreamID,
			NextOffset:   stream.BytesSent,
			Parallelism:  stream.Parallelism,
			Backpressure: false,
		}, nil
	}
	available := StreamAvailableWindow(stream)
	if available == 0 {
		return StreamChunkPlan{
			StreamID:     stream.StreamID,
			NextOffset:   stream.BytesSent,
			Parallelism:  stream.Parallelism,
			Backpressure: true,
		}, nil
	}
	chunkSize := minStreamingUint64(stream.ChunkSize, remainingBytes)
	chunkSize = minStreamingUint64(chunkSize, available)
	maxChunks := available / chunkSize
	if maxChunks == 0 {
		maxChunks = 1
	}
	if maxChunks > uint64(stream.Parallelism) {
		maxChunks = uint64(stream.Parallelism)
	}
	return StreamChunkPlan{
		StreamID:          stream.StreamID,
		NextOffset:        stream.BytesSent,
		ChunkSize:         chunkSize,
		MaxInFlightChunks: uint32(maxChunks),
		Parallelism:       stream.Parallelism,
		Backpressure:      false,
	}, nil
}

func IsStreamingPayloadType(payloadType StreamingPayloadType) bool {
	switch payloadType {
	case StreamingPayloadStateSync,
		StreamingPayloadZoneSnapshot,
		StreamingPayloadBlockPropagation,
		StreamingPayloadExecutionReceipts,
		StreamingPayloadStorageObject,
		StreamingPayloadProofBundle,
		StreamingPayloadHistoricalQueryRange:
		return true
	default:
		return false
	}
}

func IsStreamSessionState(state StreamSessionState) bool {
	switch state {
	case StreamStateOpening,
		StreamStateActive,
		StreamStatePaused,
		StreamStateDraining,
		StreamStateClosed,
		StreamStateFailed:
		return true
	default:
		return false
	}
}

func StreamingPayloadRL2Type(payloadType StreamingPayloadType) (RL2PayloadType, error) {
	switch payloadType {
	case StreamingPayloadStateSync:
		return RL2PayloadStateSyncStream, nil
	case StreamingPayloadZoneSnapshot:
		return RL2PayloadZoneSnapshot, nil
	case StreamingPayloadBlockPropagation:
		return RL2PayloadLargeBlock, nil
	case StreamingPayloadExecutionReceipts:
		return RL2PayloadExecutionResult, nil
	case StreamingPayloadStorageObject, StreamingPayloadHistoricalQueryRange:
		return RL2PayloadStorageObject, nil
	case StreamingPayloadProofBundle:
		return RL2PayloadProofSet, nil
	default:
		return "", fmt.Errorf("unknown networking streaming payload type %q", payloadType)
	}
}

func StreamingPayloadChannel(payloadType StreamingPayloadType) (ChannelClass, error) {
	switch payloadType {
	case StreamingPayloadStateSync:
		return ChannelStateSync, nil
	case StreamingPayloadBlockPropagation:
		return ChannelBlock, nil
	case StreamingPayloadExecutionReceipts:
		return ChannelExecution, nil
	case StreamingPayloadZoneSnapshot,
		StreamingPayloadStorageObject,
		StreamingPayloadProofBundle,
		StreamingPayloadHistoricalQueryRange:
		return ChannelData, nil
	default:
		return "", fmt.Errorf("unknown networking streaming payload type %q", payloadType)
	}
}

func normalizeStreamingState(state StreamSessionState) StreamSessionState {
	return StreamSessionState(strings.ToLower(strings.TrimSpace(string(state))))
}

func minStreamingUint64(left, right uint64) uint64 {
	if left < right {
		return left
	}
	return right
}
