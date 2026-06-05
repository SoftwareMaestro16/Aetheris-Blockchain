package addressing_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

func BenchmarkParseRawAddress(b *testing.B) {
	text := addressing.FormatAccAddress(sdk.AccAddress(bytes20(0x11)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := addressing.ParseAccAddress(text); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseUserFriendlyAddress(b *testing.B) {
	text, err := addressing.FormatUserFriendly(sdk.AccAddress(bytes20(0x22)))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := addressing.ParseAccAddress(text); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateUserAddress(b *testing.B) {
	text := addressing.FormatAccAddress(sdk.AccAddress(bytes20(0x33)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := addressing.ValidateUserAddress("recipient", text); err != nil {
			b.Fatal(err)
		}
	}
}
