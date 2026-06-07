package observability

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultPublicMetricsCoverRequiredSection14Metrics(t *testing.T) {
	report := BuildPublicMetricsReadinessReport(nil, nil)

	require.True(t, report.Ready)
	require.Empty(t, report.Failed)
	require.Empty(t, report.PrometheusOnly)
	require.Equal(t, len(requiredPublicMetricIDs()), report.RequiredCount)
	require.Equal(t, report.RequiredCount, report.ReadyCount)
	require.Equal(t, len(requiredPublicSurfaceIDs()), report.SurfaceCount)
	require.Equal(t, report.SurfaceCount, report.SurfacesReady)
	require.NoError(t, ValidatePublicMetricsReadiness(nil, nil))
}

func TestPublicMetricsRejectMissingRequiredMetricSurface(t *testing.T) {
	metrics := DefaultPublicMetricSpecs()
	metrics[0].CLIQuery = false

	report := BuildPublicMetricsReadinessReport(metrics, nil)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, metrics[0].ID+":missing_required_surface")
	require.Error(t, ValidatePublicMetricsReadiness(metrics, nil))
}

func TestPublicMetricsRejectPrometheusMetricNotInRegistry(t *testing.T) {
	metrics := DefaultPublicMetricSpecs()
	metrics[0].PrometheusName = "aetra_missing_metric"

	report := BuildPublicMetricsReadinessReport(metrics, nil)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, metrics[0].ID+":missing_required_surface")
}

func TestPublicMetricsRejectMissingRequiredSurface(t *testing.T) {
	surfaces := DefaultPublicSurfaceSpecs()
	surfaces[0].Ready = false

	report := BuildPublicMetricsReadinessReport(nil, surfaces)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, surfaces[0].ID+":surface_not_ready")
}

func TestPublicMetricsRejectPrometheusOnlyExposure(t *testing.T) {
	metrics := DefaultPublicMetricSpecs()
	metrics[0].CLIQuery = false
	metrics[0].GRPCQuery = false
	metrics[0].RESTQuery = false
	metrics[0].IndexerEvent = false
	metrics[0].PublicDashboard = false

	report := BuildPublicMetricsReadinessReport(metrics, nil)
	require.False(t, report.Ready)
	require.Contains(t, report.PrometheusOnly, metrics[0].ID)
}
