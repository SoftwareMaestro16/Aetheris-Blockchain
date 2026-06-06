package types

import (
	"testing"
)

func BenchmarkStoreV2DirectBalanceRead(b *testing.B) {
	state := benchmarkStoreV2State(b)
	prefix := "storev2/object/account-zone/shard-a/account/account/alice"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entries, err := StoreV2BoundedRangeScan(state, prefix, "", 1)
		if err != nil || len(entries) != 1 {
			b.Fatalf("balance read failed: %v", err)
		}
	}
}

func BenchmarkStoreV2DirectIdentityResolution(b *testing.B) {
	state := benchmarkStoreV2State(b)
	prefix := "storev2/kv/account-zone/shard-a/identity/alice.aet/resolver/address"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entries, err := StoreV2BoundedRangeScan(state, prefix, "", 1)
		if err != nil || len(entries) != 1 {
			b.Fatalf("identity resolution failed: %v", err)
		}
	}
}

func BenchmarkStoreV2RecursiveIdentityResolution(b *testing.B) {
	state := benchmarkStoreV2State(b)
	prefixes := []string{
		"storev2/kv/account-zone/shard-a/identity/alice.aet/resolver/address",
		"storev2/kv/account-zone/shard-a/identity/alice.aet/resolver/parent",
		"storev2/object/account-zone/shard-a/domain/identity/alice.aet",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, prefix := range prefixes {
			entries, err := StoreV2BoundedRangeScan(state, prefix, "", 1)
			if err != nil || len(entries) != 1 {
				b.Fatalf("recursive identity resolution failed: %v", err)
			}
		}
	}
}

func BenchmarkStoreV2ContractStorageReadWrite(b *testing.B) {
	state := benchmarkStoreV2State(b)
	prefix := "storev2/kv/account-zone/shard-a/contract/escrow/storage/slot-1"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entries, err := StoreV2BoundedRangeScan(state, prefix, "", 1)
		if err != nil || len(entries) != 1 {
			b.Fatalf("contract storage read failed: %v", err)
		}
		field := state.KVFields[2]
		field.Version++
		field.ValueHash = addStoreV2Amount(field.ValueHash, uint64(i+1))
		field.FieldHash = ComputeStoreV2FieldHash(field)
		state.KVFields[2] = field
		state.RootHash = ComputeStoreV2ShardRoot(state)
	}
}

func BenchmarkStoreV2MessageEnqueueDequeue(b *testing.B) {
	limits := MempoolSeparationLimits{MaxPerSender: 1024, MaxPerTargetObject: 1024}
	tx := sampleMempoolTx("bench-message", "alice", "contract", "shard-a", "contract/escrow", MempoolClassMessage, "5", 500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.TxID = "bench-message-" + hashStrings("n", string(rune(i)))[0:8]
		tx.TxHash = ComputeSeparatedMempoolTxHash(tx)
		snapshot, err := BuildSeparatedMempoolSnapshot(100, []SeparatedMempoolTx{tx}, limits)
		if err != nil || len(snapshot.Lanes) != 1 || len(snapshot.Lanes[0].Transactions) != 1 {
			b.Fatalf("message enqueue failed: %v", err)
		}
	}
}

func BenchmarkStoreV2PaymentChannelSettle(b *testing.B) {
	state := benchmarkStoreV2State(b)
	channelIndex := 3
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := state.Records[channelIndex]
		record.Version++
		record.ValueHash = hashStrings("channel-settled", record.ValueHash)
		record.RecordHash = ComputeStoreV2RecordHash(record)
		state.Records[channelIndex] = record
		state.RootHash = ComputeStoreV2ShardRoot(state)
	}
}

func BenchmarkStoreV2DEXPoolUpdate(b *testing.B) {
	state := benchmarkStoreV2State(b)
	poolIndex := 4
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := state.Records[poolIndex]
		record.Version++
		record.ValueHash = addStoreV2Amount(record.ValueHash, uint64(i+1))
		record.RecordHash = ComputeStoreV2RecordHash(record)
		state.Records[poolIndex] = record
		state.RootHash = ComputeStoreV2ShardRoot(state)
	}
}

func BenchmarkStoreV2ProofGeneration(b *testing.B) {
	state := benchmarkStoreV2State(b)
	prefix := "storev2/kv/account-zone/shard-a/identity/alice.aet"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proof, err := GenerateStoreV2PrefixProof(state, prefix, "", 8)
		if err != nil || len(proof.Entries) == 0 {
			b.Fatalf("proof generation failed: %v", err)
		}
	}
}

func benchmarkStoreV2State(b *testing.B) StoreV2ShardState {
	b.Helper()
	records := []StoreV2ObjectRecord{
		storeV2Record("account-zone", "shard-a", StoreV2RecordAccount, "account/alice", "balance:100", 1),
		storeV2Record("account-zone", "shard-a", StoreV2RecordDomain, "identity/alice.aet", "owner:alice", 1),
		storeV2Record("account-zone", "shard-a", StoreV2RecordContract, "contract/escrow", "code:escrow", 1),
		storeV2Record("account-zone", "shard-a", StoreV2RecordChannel, "channel/alice-bob", "settled:false", 1),
		storeV2Record("account-zone", "shard-a", StoreV2RecordPool, "pool/aet-usd", "x:100:y:200", 1),
	}
	fields := []StoreV2KVField{
		storeV2Field("account-zone", "shard-a", "identity/alice.aet", "resolver/address", "aet1alice", 1),
		storeV2Field("account-zone", "shard-a", "identity/alice.aet", "resolver/parent", "root.aet", 1),
		storeV2Field("account-zone", "shard-a", "contract/escrow", "storage/slot-1", "slot-value", 1),
	}
	state, err := BuildStoreV2ShardState("account-zone", "shard-a", records, fields)
	if err != nil {
		b.Fatal(err)
	}
	return state
}
