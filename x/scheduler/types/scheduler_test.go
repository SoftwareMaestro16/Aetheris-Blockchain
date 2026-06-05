package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSequentialPlanIsDeterministic(t *testing.T) {
	plan, err := PlanSequential([]Task{
		task("b", 1, 2, 0, nil, []string{"b"}),
		task("a", 1, 1, 0, nil, []string{"a"}),
	})
	require.NoError(t, err)
	require.Equal(t, ModeSequential, plan.Mode)
	require.Len(t, plan.Batches, 2)
	require.Equal(t, "a", plan.Batches[0].Tasks[0].ID)
	require.Equal(t, "b", plan.Batches[1].Tasks[0].ID)
}

func TestConflictDetectionAndOptimisticBatching(t *testing.T) {
	independentA := task("a", 1, 1, 0, []string{"x"}, []string{"y"})
	independentB := task("b", 1, 2, 0, []string{"m"}, []string{"n"})
	conflict := task("c", 1, 3, 0, []string{"y"}, []string{"z"})

	require.False(t, Conflicts(independentA, independentB))
	require.True(t, Conflicts(independentA, conflict))

	plan, err := PlanOptimistic([]Task{conflict, independentB, independentA})
	require.NoError(t, err)
	require.Equal(t, ModeOptimisticParallel, plan.Mode)
	require.Len(t, plan.Batches, 2)
	require.Len(t, plan.Batches[0].Tasks, 2)
	require.Len(t, plan.Batches[1].Tasks, 1)
}

func TestSchedulerValidation(t *testing.T) {
	_, err := PlanSequential([]Task{{ID: "empty"}})
	require.ErrorContains(t, err, "read or write")
	_, err = PlanSequential([]Task{task("a", 1, 0, 0, nil, []string{"x"}), task("a", 1, 1, 0, nil, []string{"y"})})
	require.ErrorContains(t, err, "duplicate")
}

func task(id string, height uint64, tx uint32, msg uint32, reads []string, writes []string) Task {
	return Task{ID: id, TxHeight: height, TxIndex: tx, MessageIndex: msg, Reads: reads, Writes: writes}
}
