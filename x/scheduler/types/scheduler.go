package types

import (
	"errors"
	"fmt"
	"sort"
)

const (
	ModeSequential           = "sequential"
	ModeOptimisticParallel   = "optimistic_parallel"
	StatusReady              = "ready"
	StatusConflictSequential = "conflict_fallback_sequential"
)

type Task struct {
	ID           string
	TxHeight     uint64
	TxIndex      uint32
	MessageIndex uint32
	Reads        []string
	Writes       []string
	Payload      []byte
}

type Plan struct {
	Mode    string
	Batches []Batch
	Status  string
}

type Batch struct {
	Tasks []Task
}

func PlanSequential(tasks []Task) (Plan, error) {
	if err := ValidateTasks(tasks); err != nil {
		return Plan{}, err
	}
	ordered := cloneTasks(tasks)
	sort.SliceStable(ordered, func(i, j int) bool {
		return taskLess(ordered[i], ordered[j])
	})
	batches := make([]Batch, len(ordered))
	for i, task := range ordered {
		batches[i] = Batch{Tasks: []Task{task}}
	}
	return Plan{Mode: ModeSequential, Batches: batches, Status: StatusReady}, nil
}

func PlanOptimistic(tasks []Task) (Plan, error) {
	if err := ValidateTasks(tasks); err != nil {
		return Plan{}, err
	}
	ordered := cloneTasks(tasks)
	sort.SliceStable(ordered, func(i, j int) bool {
		return taskLess(ordered[i], ordered[j])
	})
	batches := make([]Batch, 0)
	for _, task := range ordered {
		placed := false
		for i := range batches {
			if !ConflictsWithBatch(task, batches[i]) {
				batches[i].Tasks = append(batches[i].Tasks, task)
				placed = true
				break
			}
		}
		if !placed {
			batches = append(batches, Batch{Tasks: []Task{task}})
		}
	}
	status := StatusReady
	if len(batches) == len(ordered) && HasConflicts(ordered) {
		status = StatusConflictSequential
	}
	return Plan{Mode: ModeOptimisticParallel, Batches: batches, Status: status}, nil
}

func HasConflicts(tasks []Task) bool {
	for i := range tasks {
		for j := i + 1; j < len(tasks); j++ {
			if Conflicts(tasks[i], tasks[j]) {
				return true
			}
		}
	}
	return false
}

func ConflictsWithBatch(task Task, batch Batch) bool {
	for _, other := range batch.Tasks {
		if Conflicts(task, other) {
			return true
		}
	}
	return false
}

func Conflicts(a, b Task) bool {
	aWrites := set(a.Writes)
	bWrites := set(b.Writes)
	for key := range aWrites {
		if _, ok := bWrites[key]; ok {
			return true
		}
	}
	for _, key := range a.Writes {
		if contains(b.Reads, key) {
			return true
		}
	}
	for _, key := range b.Writes {
		if contains(a.Reads, key) {
			return true
		}
	}
	return false
}

func ValidateTasks(tasks []Task) error {
	seen := make(map[string]struct{}, len(tasks))
	for _, task := range tasks {
		if task.ID == "" {
			return errors.New("scheduler task id is required")
		}
		if _, ok := seen[task.ID]; ok {
			return fmt.Errorf("duplicate scheduler task id %q", task.ID)
		}
		seen[task.ID] = struct{}{}
		if len(task.Writes) == 0 && len(task.Reads) == 0 {
			return fmt.Errorf("scheduler task %q must declare read or write set", task.ID)
		}
	}
	return nil
}

func taskLess(a, b Task) bool {
	if a.TxHeight != b.TxHeight {
		return a.TxHeight < b.TxHeight
	}
	if a.TxIndex != b.TxIndex {
		return a.TxIndex < b.TxIndex
	}
	if a.MessageIndex != b.MessageIndex {
		return a.MessageIndex < b.MessageIndex
	}
	return a.ID < b.ID
}

func set(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func cloneTasks(tasks []Task) []Task {
	out := make([]Task, len(tasks))
	for i, task := range tasks {
		out[i] = task
		out[i].Reads = append([]string(nil), task.Reads...)
		out[i].Writes = append([]string(nil), task.Writes...)
		out[i].Payload = append([]byte(nil), task.Payload...)
	}
	return out
}
