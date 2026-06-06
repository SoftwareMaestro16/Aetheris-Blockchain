package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplicationZoneBoundaryAndSpecStateKeys(t *testing.T) {
	boundary := DefaultApplicationZoneBoundary()
	require.NoError(t, boundary.Validate())
	require.Contains(t, boundary.Messages, ApplicationMessageExecuteAutomation)
	require.Contains(t, boundary.ProofKinds, ApplicationProofAppReceipts)

	appKey, err := ApplicationAppKey("billing")
	require.NoError(t, err)
	require.Equal(t, "apps/app/billing", appKey)

	workflowKey, err := ApplicationWorkflowKey("workflow-1")
	require.NoError(t, err)
	require.Equal(t, "apps/workflow/workflow-1", workflowKey)

	task := applicationTask("task-1", "workflow-1", "billing", "hourly", 10, 3, 1, 100)
	taskKey, err := ApplicationSchedulerTaskKey(task)
	require.NoError(t, err)
	require.Equal(t, "apps/scheduler/hourly/task-1", taskKey)

	automationKey, err := ApplicationAutomationKey("daily-settlement")
	require.NoError(t, err)
	require.Equal(t, "apps/automation/daily-settlement", automationKey)

	permissionKey, err := ApplicationPermissionKey("billing", "alice")
	require.NoError(t, err)
	require.Equal(t, "apps/permissions/billing/alice", permissionKey)

	receiptKey, err := ApplicationReceiptKey("exec-1")
	require.NoError(t, err)
	require.Equal(t, "apps/receipts/exec-1", receiptKey)
}

func TestApplicationSchedulerExecutesReadyTasksDeterministicallyWithBounds(t *testing.T) {
	tasks := []ApplicationScheduledTask{
		applicationTask("late", "wf-2", "billing", "hourly", 11, 10, 3, 100),
		applicationTask("low", "wf-1", "billing", "hourly", 10, 1, 2, 100),
		applicationTask("high", "wf-1", "billing", "hourly", 10, 50, 1, 100),
		applicationTask("gas-left", "wf-3", "billing", "daily", 10, 1, 4, 950),
	}
	queue, err := NewApplicationSchedulerQueue(9, 3, tasks)
	require.NoError(t, err)
	require.Equal(t, "high", queue.Tasks[0].TaskID)
	require.Equal(t, "gas-left", queue.Tasks[1].TaskID)

	next, ready, err := ExecuteApplicationScheduledTasks(queue, 10, ApplicationWorkLimit{
		MaxTasksPerBlock:    2,
		MaxMessagesPerBlock: 2,
		MaxGasPerBlock:      200,
	})
	require.NoError(t, err)
	require.Len(t, ready, 2)
	require.Equal(t, []string{"high", "low"}, []string{ready[0].TaskID, ready[1].TaskID})
	require.Equal(t, ApplicationTaskExecuted, ready[0].Status)
	require.Len(t, next.Tasks, 2)
	require.Equal(t, "gas-left", next.Tasks[0].TaskID)
	require.Equal(t, ApplicationTaskDeferred, next.Tasks[0].Status)
	require.Equal(t, uint64(10), next.Height)

	rootA, err := ComputeApplicationSchedulerRoot(queue)
	require.NoError(t, err)
	shuffled, err := NewApplicationSchedulerQueue(9, 3, []ApplicationScheduledTask{tasks[2], tasks[0], tasks[3], tasks[1]})
	require.NoError(t, err)
	rootB, err := ComputeApplicationSchedulerRoot(shuffled)
	require.NoError(t, err)
	require.Equal(t, rootA, rootB)

	canceled, canceledTask, err := CancelApplicationScheduledTask(next, "hourly", "late")
	require.NoError(t, err)
	require.Equal(t, ApplicationTaskCanceled, canceledTask.Status)
	require.Len(t, canceled.Tasks, 1)
}

func TestApplicationWorkflowReceiptsAndStateRoot(t *testing.T) {
	state := ApplicationZoneState{
		Apps: []ApplicationRecord{
			{AppID: "billing", Owner: "alice", RuntimeID: "avm", Version: 1, Enabled: true, ConfigHash: hash("config"), UpdatedHeight: 1},
		},
		Automations: []ApplicationAutomation{
			{AutomationID: "daily-settlement", AppID: "billing", WorkflowID: "workflow-1", Enabled: true, TriggerHash: hash("trigger"), NextRunHeight: 11, UpdatedHeight: 1},
		},
		Permissions: []ApplicationPermission{
			{AppID: "billing", Address: "alice", Scope: ApplicationPermissionAdmin, ExpiresHeight: 100, GrantHash: hash("grant")},
		},
	}
	workflow := ApplicationWorkflowState{
		WorkflowID:    "workflow-1",
		AppID:         "billing",
		Owner:         "alice",
		Status:        ApplicationWorkflowPending,
		CurrentStep:   0,
		TotalSteps:    2,
		PayloadHash:   hash("workflow-payload"),
		UpdatedHeight: 1,
	}
	next, startReceipt, err := StartApplicationWorkflow(state, workflow, 10)
	require.NoError(t, err)
	require.Equal(t, ApplicationWorkflowRunning, next.Workflows[0].Status)
	require.Equal(t, "workflow-start", startReceipt.TaskID)
	require.NotEmpty(t, startReceipt.ReceiptHash)

	next, advanceReceipt, err := AdvanceApplicationWorkflow(next, "workflow-1", 11)
	require.NoError(t, err)
	require.Equal(t, ApplicationWorkflowSucceeded, next.Workflows[0].Status)
	require.Equal(t, "workflow-advance", advanceReceipt.TaskID)
	require.Len(t, next.Receipts, 2)
	require.NotEmpty(t, ComputeApplicationZoneStateRoot(next))
	require.NoError(t, next.Validate())
}

func TestApplicationMessagesReceiptsProofsAndRootValidate(t *testing.T) {
	queues, err := NewApplicationMessageQueues(
		[]ZoneMessage{testZoneMessage(ZoneIDApplication, "app.inbound", 2, 100), testZoneMessage(ZoneIDApplication, "app.inbound", 1, 100)},
		[]ZoneMessage{testZoneMessage(ZoneIDApplication, "app.outbound", 3, 100)},
	)
	require.NoError(t, err)
	require.Equal(t, uint64(1), queues.Inbox[0].Sequence)

	receipt, err := NewApplicationExecutionReceipt(ApplicationExecutionReceipt{
		ZoneID:         ZoneIDApplication,
		Height:         77,
		ExecutionID:    "exec-1",
		TaskID:         "task-1",
		WorkflowID:     "workflow-1",
		AppID:          "billing",
		Status:         ApplicationTaskExecuted,
		GasUsed:        500,
		OutputHash:     hash("app-output"),
		OutboxMessages: 1,
		Sequence:       9,
	})
	require.NoError(t, err)
	zoneReceipt, err := receipt.ZoneReceipt()
	require.NoError(t, err)
	require.Equal(t, ZoneReceiptStatusSuccess, zoneReceipt.Status)

	state := ApplicationZoneState{
		Apps: []ApplicationRecord{
			{AppID: "billing", Owner: "alice", RuntimeID: "avm", Version: 1, Enabled: true, ConfigHash: hash("config"), UpdatedHeight: 1},
		},
		Workflows: []ApplicationWorkflowState{
			{WorkflowID: "workflow-1", AppID: "billing", Owner: "alice", Status: ApplicationWorkflowRunning, CurrentStep: 1, TotalSteps: 2, PayloadHash: hash("workflow"), UpdatedHeight: 77},
		},
		Tasks:       []ApplicationScheduledTask{applicationTask("task-1", "workflow-1", "billing", "hourly", 77, 1, 1, 100)},
		Automations: []ApplicationAutomation{{AutomationID: "auto-1", AppID: "billing", WorkflowID: "workflow-1", Enabled: true, TriggerHash: hash("trigger"), NextRunHeight: 78, UpdatedHeight: 77}},
		Permissions: []ApplicationPermission{{AppID: "billing", Address: "alice", Scope: ApplicationPermissionExecute, ExpiresHeight: 100, GrantHash: hash("grant")}},
		Receipts:    []ApplicationExecutionReceipt{receipt},
	}
	root, err := BuildApplicationZoneRootFromState(77, state, queues, hash("proofs"))
	require.NoError(t, err)
	require.Equal(t, ZoneIDApplication, root.ZoneID)
	require.Equal(t, ComputeApplicationZoneStateRoot(state), root.ZoneStateRoot)

	req, err := ApplicationProofRequest(ApplicationProofWorkflow, "workflow-1", 77, root.RootHash, 4)
	require.NoError(t, err)
	require.Equal(t, "QueryWorkflow/workflow-1", req.Key)
}

func applicationTask(taskID, workflowID, appID, bucket string, height uint64, priority uint32, sequence uint64, gas uint64) ApplicationScheduledTask {
	return ApplicationScheduledTask{
		Bucket:          bucket,
		TaskID:          taskID,
		WorkflowID:      workflowID,
		AppID:           appID,
		ScheduledHeight: height,
		Priority:        priority,
		Sequence:        sequence,
		GasLimit:        gas,
		PayloadHash:     hash(taskID),
		Status:          ApplicationTaskPending,
	}
}
