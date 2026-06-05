package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDAGPlanParallelizesIndependentTasks(t *testing.T) {
	plan, err := BuildDAGPlan([]Task{
		task("b", "actor-b", 1, 2, 0, nil, []string{"b"}, nil),
		task("a", "actor-a", 1, 1, 0, nil, []string{"a"}, nil),
	})
	require.NoError(t, err)
	require.Equal(t, PlanStatusReady, plan.Status)
	require.Len(t, plan.Batches, 1)
	require.Equal(t, []string{"a", "b"}, taskIDs(plan.Batches[0].Tasks))
	require.Len(t, plan.ReplayHash, 32)
}

func TestDAGPlanHonorsDependenciesAndConflicts(t *testing.T) {
	plan, err := BuildDAGPlan([]Task{
		task("c", "actor-c", 1, 3, 0, []string{"b"}, []string{"c"}, nil),
		task("b", "actor-b", 1, 2, 0, []string{"a"}, []string{"b"}, nil),
		task("a", "actor-a", 1, 1, 0, nil, []string{"a"}, []string{"shared"}),
		task("x", "actor-x", 1, 1, 1, nil, []string{"shared"}, nil),
	})
	require.NoError(t, err)
	require.Equal(t, PlanStatusConflictSerialized, plan.Status)
	require.Equal(t, []string{"a"}, taskIDs(plan.Batches[0].Tasks))
	require.Equal(t, []string{"x", "b"}, taskIDs(plan.Batches[1].Tasks))
	require.Equal(t, []string{"c"}, taskIDs(plan.Batches[2].Tasks))
}

func TestDAGPlanRejectsCyclesAndMissingReadWriteSet(t *testing.T) {
	_, err := BuildDAGPlan([]Task{
		task("a", "actor-a", 1, 1, 0, []string{"b"}, []string{"a"}, nil),
		task("b", "actor-b", 1, 2, 0, []string{"a"}, []string{"b"}, nil),
	})
	require.ErrorContains(t, err, "cycle")

	_, err = BuildDAGPlan([]Task{{ID: "a", Actor: "actor-a"}})
	require.ErrorContains(t, err, "read/write")
}

func TestDAGPlanReplayHashIndependentOfInputOrder(t *testing.T) {
	left, err := BuildDAGPlan([]Task{
		task("b", "actor-b", 1, 2, 0, nil, []string{"b"}, nil),
		task("a", "actor-a", 1, 1, 0, nil, []string{"a"}, nil),
	})
	require.NoError(t, err)
	right, err := BuildDAGPlan([]Task{
		task("a", "actor-a", 1, 1, 0, nil, []string{"a"}, nil),
		task("b", "actor-b", 1, 2, 0, nil, []string{"b"}, nil),
	})
	require.NoError(t, err)
	require.Equal(t, left.ReplayHash, right.ReplayHash)
}

func TestActorMailboxPlanIsDeterministic(t *testing.T) {
	mailboxes, err := BuildMailboxPlan([]Task{
		task("b2", "actor-b", 1, 3, 0, nil, []string{"b2"}, nil),
		task("a1", "actor-a", 1, 1, 0, nil, []string{"a1"}, nil),
		task("b1", "actor-b", 1, 2, 0, nil, []string{"b1"}, nil),
	})
	require.NoError(t, err)
	require.Len(t, mailboxes, 2)
	require.Equal(t, "actor-a", mailboxes[0].Actor)
	require.Equal(t, []string{"a1"}, taskIDs(mailboxes[0].Tasks))
	require.Equal(t, "actor-b", mailboxes[1].Actor)
	require.Equal(t, []string{"b1", "b2"}, taskIDs(mailboxes[1].Tasks))
}

func task(id, actor string, height uint64, tx uint32, msg uint32, deps []string, reads []string, writes []string) Task {
	return Task{ID: id, Actor: actor, TxHeight: height, TxIndex: tx, MessageIndex: msg, Dependencies: deps, Reads: reads, Writes: writes}
}

func taskIDs(tasks []Task) []string {
	out := make([]string, len(tasks))
	for i, task := range tasks {
		out[i] = task.ID
	}
	return out
}
