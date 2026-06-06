package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAVMObservabilitySpecMatchesSection21(t *testing.T) {
	spec, err := DefaultAVMObservabilitySpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Equal(t, ComputeAVMObservabilitySpecHash(spec), spec.SpecHash)
	require.Len(t, spec.Metrics, 18)
	require.Len(t, spec.Events, 16)
	require.Len(t, spec.Alerts, 9)
	require.Contains(t, spec.Metrics, AVMMetricAsyncMessagesSubmitted)
	require.Contains(t, spec.Metrics, AVMMetricReceiptRootGenerationTime)
	require.Contains(t, spec.Events, AVMEventMessageSubmitted)
	require.Contains(t, spec.Events, AVMEventRuntimeUpgradeScheduled)
	require.Contains(t, spec.Alerts, AVMAlertDeadLetterSpike)
	require.Contains(t, spec.Alerts, AVMAlertQueueRootGenerationLatencyThreshold)
}

func TestAVMObservabilityMetricsMatchSection211(t *testing.T) {
	spec, err := DefaultAVMObservabilitySpec()
	require.NoError(t, err)
	require.ElementsMatch(t, []AVMObservabilityMetric{
		AVMMetricAsyncMessagesSubmitted,
		AVMMetricAsyncMessagesExecuted,
		AVMMetricAsyncMessagesExpired,
		AVMMetricAsyncMessagesBounced,
		AVMMetricDeadLetterCount,
		AVMMetricRetryQueueSize,
		AVMMetricDelayedQueueSize,
		AVMMetricActorCount,
		AVMMetricActorMailboxDepth,
		AVMMetricContinuationsActive,
		AVMMetricContinuationsExpired,
		AVMMetricContractExecutions,
		AVMMetricContractFailures,
		AVMMetricGasReserved,
		AVMMetricGasConsumed,
		AVMMetricZoneAsyncBudgetUsage,
		AVMMetricQueueDrainLatency,
		AVMMetricReceiptRootGenerationTime,
	}, spec.Metrics)
}

func TestAVMObservabilityEventsMatchSection212(t *testing.T) {
	spec, err := DefaultAVMObservabilitySpec()
	require.NoError(t, err)
	require.ElementsMatch(t, []AVMObservabilityEvent{
		AVMEventMessageSubmitted,
		AVMEventMessageScheduled,
		AVMEventMessageExecuted,
		AVMEventMessageFailed,
		AVMEventMessageRetried,
		AVMEventMessageExpired,
		AVMEventMessageBounced,
		AVMEventDeadLettered,
		AVMEventActorCreated,
		AVMEventActorMessageHandled,
		AVMEventContinuationCreated,
		AVMEventContinuationResumed,
		AVMEventContinuationExpired,
		AVMEventContractExecuted,
		AVMEventInterfaceRegistered,
		AVMEventRuntimeUpgradeScheduled,
	}, spec.Events)
}

func TestAVMObservabilityAlertsMatchSection213(t *testing.T) {
	spec, err := DefaultAVMObservabilitySpec()
	require.NoError(t, err)
	require.ElementsMatch(t, []AVMObservabilityAlert{
		AVMAlertDeadLetterSpike,
		AVMAlertRetryQueueBacklog,
		AVMAlertDelayedQueueBacklog,
		AVMAlertZoneAsyncBudgetSaturation,
		AVMAlertActorMailboxBacklog,
		AVMAlertContinuationExpirySpike,
		AVMAlertContractFailureSpike,
		AVMAlertReceiptGenerationLatencyThreshold,
		AVMAlertQueueRootGenerationLatencyThreshold,
	}, spec.Alerts)
}

func TestAVMObservabilityRejectsMissingDuplicateUnknownAndHashMismatch(t *testing.T) {
	spec, err := DefaultAVMObservabilitySpec()
	require.NoError(t, err)

	missingMetric := spec
	missingMetric.Metrics = missingMetric.Metrics[:1]
	missingMetric.SpecHash = ComputeAVMObservabilitySpecHash(missingMetric)
	require.ErrorContains(t, missingMetric.Validate(), "every section 21 metric")

	spec, err = DefaultAVMObservabilitySpec()
	require.NoError(t, err)
	duplicateMetric := spec
	duplicateMetric.Metrics[1] = duplicateMetric.Metrics[0]
	duplicateMetric.SpecHash = ComputeAVMObservabilitySpecHash(duplicateMetric)
	require.ErrorContains(t, duplicateMetric.Validate(), "duplicate")

	spec, err = DefaultAVMObservabilitySpec()
	require.NoError(t, err)
	unknownEvent := spec
	unknownEvent.Events[0] = AVMObservabilityEvent("avm_unknown")
	unknownEvent.SpecHash = ComputeAVMObservabilitySpecHash(unknownEvent)
	require.ErrorContains(t, unknownEvent.Validate(), "invalid")

	spec, err = DefaultAVMObservabilitySpec()
	require.NoError(t, err)
	unknownAlert := spec
	unknownAlert.Alerts[0] = AVMObservabilityAlert("unknown_alert")
	unknownAlert.SpecHash = ComputeAVMObservabilitySpecHash(unknownAlert)
	require.ErrorContains(t, unknownAlert.Validate(), "invalid")

	spec, err = DefaultAVMObservabilitySpec()
	require.NoError(t, err)
	hashMismatch := spec
	hashMismatch.SpecHash = "0000000000000000000000000000000000000000000000000000000000000000"
	require.ErrorContains(t, hashMismatch.Validate(), "hash mismatch")
}
