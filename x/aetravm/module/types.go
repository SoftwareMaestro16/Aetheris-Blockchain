package module

type TrustLevel uint8

const (
	Untrusted TrustLevel = iota
	Verified
	Canonical
)

type VerificationResult struct {
	ModuleHash         []byte
	VerifierVersion    uint32
	Passed             bool
	ErrorCode          uint32
	AnalyzedStackBound uint32
	CFGHash            []byte
}

type AVMModule struct {
	Magic            uint32
	Version          uint32
	ABIVersion       uint32
	ImportTable      []string
	ExportTable      []string
	MetadataHash     []byte
	Instructions     []byte
	DependencyHashes [][]byte
}
