package docs

import (
	"os"
	"strings"
	"testing"
)

func TestArchitectureDocumentsAVMFirstCosmWasmGatedPolicy(t *testing.T) {
	bz, err := os.ReadFile("../architecture.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(bz)
	for _, required := range []string{
		"AVM is the default and production-target smart-contract runtime",
		"CosmWasm is optional compatibility research only",
		"default app wiring must not add the CosmWasm store key",
		"token, NFT, and DEX behavior remains contract-routed",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("architecture.md must mention %q", required)
		}
	}
}
