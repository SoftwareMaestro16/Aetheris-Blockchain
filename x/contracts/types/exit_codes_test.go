package types

import (
	"testing"
)

func TestExitCodes(t *testing.T) {
	codes := CanonicalExitCodes()

	for _, spec := range codes {
		if spec.Code > 127 {
			t.Errorf("exit code %d (%s) exceeds domain limit of 127", spec.Code, spec.Name)
		}
	}

	// Verify specific requested codes exist
	required := map[uint32]string{
		14: "invalid_jump",
		15: "call_stack_overflow",
		16: "continuation_not_found",
		17: "recursion_limit_exceeded",
		25: "out_of_gas",
		26: "gas_reservation_failed",
		27: "execution_timeout",
		38: "message_routing_failed",
		39: "queue_overflow",
		40: "shard_unavailable",
		69: "state_corruption",
		70: "state_version_mismatch",
		71: "snapshot_failure",
		44: "explicit_abort",
		45: "assertion_failed",
	}

	for code, name := range required {
		if ExitCodeName(code) != name {
			t.Errorf("expected exit code %d to have name %s, got %s", code, name, ExitCodeName(code))
		}
	}

	if ExitCodeName(999) != "unknown" {
		t.Errorf("expected unknown for invalid code, got %s", ExitCodeName(999))
	}
}
