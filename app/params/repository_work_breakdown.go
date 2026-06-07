package params

import (
	"fmt"
	"sort"
)

const (
	AetraRepoAreaProto = "proto/"
)

const (
	AetraRepoProtoTaskDefineMessages       = "define_protobuf_messages_for_new_modules"
	AetraRepoProtoTaskDefineQueryServices  = "define_query_services"
	AetraRepoProtoTaskDefineTxServices     = "define_tx_services"
	AetraRepoProtoTaskDefineGenesis        = "define_genesis_messages"
	AetraRepoProtoTaskDefineParams         = "define_params_messages"
	AetraRepoProtoTaskRunCodeGeneration    = "run_code_generation"
	AetraRepoProtoTaskBreakingChangeChecks = "add_proto_breaking_change_checks_if_available"
)

const (
	AetraRepoProtoTestGeneratedCodeCompiles = "generated_code_compiles"
	AetraRepoProtoTestLintPasses            = "proto_lint_passes_if_configured"
	AetraRepoProtoTestServiceRegistration   = "query_tx_service_registration_tested"
)

type AetraRepoWorkAreaEvidence struct {
	Area  string
	Tasks []string
	Tests []string
}

type AetraRepoWorkAreaReport struct {
	Area     string
	Required int
	Passed   int
	Failed   []string
	Ready    bool
}

func DefaultAetraRepoProtoWorkEvidence() AetraRepoWorkAreaEvidence {
	return AetraRepoWorkAreaEvidence{
		Area:  AetraRepoAreaProto,
		Tasks: RequiredAetraRepoProtoTasks(),
		Tests: RequiredAetraRepoProtoTests(),
	}
}

func ValidateAetraRepoProtoWork(evidence AetraRepoWorkAreaEvidence) error {
	report := BuildAetraRepoProtoWorkReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra repository proto work breakdown failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraRepoProtoWorkReport(evidence AetraRepoWorkAreaEvidence) AetraRepoWorkAreaReport {
	failed := make([]string, 0)
	if evidence.Area != AetraRepoAreaProto {
		failed = append(failed, "area_must_be_"+AetraRepoAreaProto)
	}
	passedTasks, failedTasks := validateRepoWorkCatalog("tasks", evidence.Tasks, RequiredAetraRepoProtoTasks())
	passedTests, failedTests := validateRepoWorkCatalog("tests", evidence.Tests, RequiredAetraRepoProtoTests())
	failed = append(failed, failedTasks...)
	failed = append(failed, failedTests...)

	sort.Strings(failed)
	return AetraRepoWorkAreaReport{
		Area:     evidence.Area,
		Required: len(RequiredAetraRepoProtoTasks()) + len(RequiredAetraRepoProtoTests()),
		Passed:   passedTasks + passedTests,
		Failed:   failed,
		Ready:    len(failed) == 0,
	}
}

func RequiredAetraRepoProtoTasks() []string {
	return []string{
		AetraRepoProtoTaskDefineMessages,
		AetraRepoProtoTaskDefineQueryServices,
		AetraRepoProtoTaskDefineTxServices,
		AetraRepoProtoTaskDefineGenesis,
		AetraRepoProtoTaskDefineParams,
		AetraRepoProtoTaskRunCodeGeneration,
		AetraRepoProtoTaskBreakingChangeChecks,
	}
}

func RequiredAetraRepoProtoTests() []string {
	return []string{
		AetraRepoProtoTestGeneratedCodeCompiles,
		AetraRepoProtoTestLintPasses,
		AetraRepoProtoTestServiceRegistration,
	}
}

func validateRepoWorkCatalog(kind string, actual []string, required []string) (int, []string) {
	requiredSet := map[string]bool{}
	actualCounts := map[string]int{}
	for _, item := range required {
		requiredSet[item] = true
	}
	for _, item := range actual {
		actualCounts[item]++
	}

	failed := make([]string, 0)
	passed := 0
	for _, item := range required {
		switch actualCounts[item] {
		case 0:
			failed = append(failed, kind+"."+item+":missing")
		case 1:
			passed++
		default:
			failed = append(failed, kind+"."+item+":duplicate")
		}
	}
	for item := range actualCounts {
		if !requiredSet[item] {
			failed = append(failed, kind+"."+item+":unexpected")
		}
	}
	return passed, failed
}
