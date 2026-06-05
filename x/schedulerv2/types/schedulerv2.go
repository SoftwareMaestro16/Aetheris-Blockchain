package types

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
)

const (
	PlanStatusReady              = "ready"
	PlanStatusConflictSerialized = "conflict_serialized"
)

type Task struct {
	ID           string
	Actor        string
	TxHeight     uint64
	TxIndex      uint32
	MessageIndex uint32
	Dependencies []string
	Reads        []string
	Writes       []string
}

type Batch struct {
	Tasks []Task
}

type Plan struct {
	Batches    []Batch
	Status     string
	ReplayHash []byte
}

type MailboxPlan struct {
	Actor string
	Tasks []Task
}

func BuildDAGPlan(tasks []Task) (Plan, error) {
	if err := ValidateTasks(tasks); err != nil {
		return Plan{}, err
	}
	remaining := taskMap(tasks)
	done := make(map[string]struct{}, len(tasks))
	batches := make([]Batch, 0)
	serializedConflicts := false

	for len(remaining) > 0 {
		ready := readyTasks(remaining, done)
		if len(ready) == 0 {
			return Plan{}, errors.New("scheduler-v2 dependency cycle detected")
		}
		sortTasks(ready)
		batch := Batch{}
		for _, task := range ready {
			if ConflictsWithBatch(task, batch) {
				serializedConflicts = true
				continue
			}
			batch.Tasks = append(batch.Tasks, task.Clone())
		}
		if len(batch.Tasks) == 0 {
			task := ready[0]
			batch.Tasks = append(batch.Tasks, task.Clone())
			serializedConflicts = true
		}
		for _, task := range batch.Tasks {
			done[task.ID] = struct{}{}
			delete(remaining, task.ID)
		}
		batches = append(batches, batch)
	}

	status := PlanStatusReady
	if serializedConflicts {
		status = PlanStatusConflictSerialized
	}
	return Plan{Batches: batches, Status: status, ReplayHash: ReplayHash(batches)}, nil
}

func BuildMailboxPlan(tasks []Task) ([]MailboxPlan, error) {
	if err := ValidateTasks(tasks); err != nil {
		return nil, err
	}
	byActor := make(map[string][]Task)
	for _, task := range tasks {
		byActor[task.Actor] = append(byActor[task.Actor], task.Clone())
	}
	actors := make([]string, 0, len(byActor))
	for actor := range byActor {
		actors = append(actors, actor)
		sortTasks(byActor[actor])
	}
	sort.Strings(actors)
	out := make([]MailboxPlan, len(actors))
	for i, actor := range actors {
		out[i] = MailboxPlan{Actor: actor, Tasks: cloneTasks(byActor[actor])}
	}
	return out, nil
}

func ValidateTasks(tasks []Task) error {
	seen := make(map[string]struct{}, len(tasks))
	for _, task := range tasks {
		if task.ID == "" {
			return errors.New("scheduler-v2 task id is required")
		}
		if _, ok := seen[task.ID]; ok {
			return fmt.Errorf("duplicate scheduler-v2 task id %q", task.ID)
		}
		seen[task.ID] = struct{}{}
		if task.Actor == "" {
			return fmt.Errorf("scheduler-v2 task %q actor is required", task.ID)
		}
		if len(task.Reads) == 0 && len(task.Writes) == 0 {
			return fmt.Errorf("scheduler-v2 task %q must declare deterministic read/write set", task.ID)
		}
	}
	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if _, ok := seen[dep]; !ok {
				return fmt.Errorf("scheduler-v2 task %q dependency %q not found", task.ID, dep)
			}
		}
	}
	return nil
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

func ReplayHash(batches []Batch) []byte {
	h := sha256.New()
	for _, batch := range batches {
		h.Write([]byte("batch"))
		for _, task := range batch.Tasks {
			h.Write([]byte(task.ID))
			h.Write([]byte{0})
		}
	}
	return h.Sum(nil)
}

func (t Task) Clone() Task {
	out := t
	out.Dependencies = append([]string(nil), t.Dependencies...)
	out.Reads = append([]string(nil), t.Reads...)
	out.Writes = append([]string(nil), t.Writes...)
	return out
}

func taskMap(tasks []Task) map[string]Task {
	out := make(map[string]Task, len(tasks))
	for _, task := range tasks {
		out[task.ID] = task.Clone()
	}
	return out
}

func readyTasks(remaining map[string]Task, done map[string]struct{}) []Task {
	ready := make([]Task, 0)
	for _, task := range remaining {
		blocked := false
		for _, dep := range task.Dependencies {
			if _, ok := done[dep]; !ok {
				blocked = true
				break
			}
		}
		if !blocked {
			ready = append(ready, task.Clone())
		}
	}
	return ready
}

func sortTasks(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].TxHeight != tasks[j].TxHeight {
			return tasks[i].TxHeight < tasks[j].TxHeight
		}
		if tasks[i].TxIndex != tasks[j].TxIndex {
			return tasks[i].TxIndex < tasks[j].TxIndex
		}
		if tasks[i].MessageIndex != tasks[j].MessageIndex {
			return tasks[i].MessageIndex < tasks[j].MessageIndex
		}
		if tasks[i].Actor != tasks[j].Actor {
			return tasks[i].Actor < tasks[j].Actor
		}
		return tasks[i].ID < tasks[j].ID
	})
}

func cloneTasks(tasks []Task) []Task {
	out := make([]Task, len(tasks))
	for i, task := range tasks {
		out[i] = task.Clone()
	}
	return out
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
