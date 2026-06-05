package aw

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
)

func NewState(wallet WalletState) (*State, error) {
	if wallet.Extensions == nil {
		wallet.Extensions = make(map[string]ExtensionState)
	}
	if err := wallet.Validate(); err != nil {
		return nil, err
	}
	return &State{Wallet: wallet}, nil
}

func (w WalletState) Validate() error {
	if err := aetherisaddress.RejectZeroAddress("wallet", w.Address); err != nil {
		return err
	}
	if w.WalletID < DefaultWalletIDBase {
		return errors.New("wallet_id must be positive")
	}
	if len(w.PublicKey) != PublicKeyLength {
		return fmt.Errorf("public key must be %d bytes", PublicKeyLength)
	}
	if len(w.Owner) > 0 {
		if err := aetherisaddress.RejectZeroAddress("wallet owner", w.Owner); err != nil {
			return err
		}
	}
	if len(w.RecoveryAuthority) > 0 {
		if err := aetherisaddress.RejectZeroAddress("recovery authority", w.RecoveryAuthority); err != nil {
			return err
		}
	}
	if len(w.Extensions) > MaxExtensions {
		return fmt.Errorf("extensions count must be <= %d", MaxExtensions)
	}
	for _, extension := range w.Extensions {
		if err := extension.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (e ExtensionState) Validate() error {
	if err := aetherisaddress.RejectZeroAddress("wallet extension", e.Address); err != nil {
		return err
	}
	if !e.Installed {
		return errors.New("wallet extension must be installed")
	}
	return nil
}

func (s *State) ApplyExternalCommand(cmd ExternalCommand, now int64, fees sdk.Coins) error {
	if err := s.VerifyExternalCommand(cmd, now); err != nil {
		return err
	}
	if err := ValidateProtocolFees(fees); err != nil {
		return err
	}
	checkpoint := s.snapshot()
	if err := s.applyVerifiedExternalCommand(cmd); err != nil {
		s.restore(checkpoint)
		return err
	}
	s.Wallet.Seqno++
	return s.Wallet.Validate()
}

func (s *State) ApplyRelayedCommand(cmd RelayedCommand, now int64) error {
	if err := aetherisaddress.RejectZeroAddress("relayer", cmd.Relayer); err != nil {
		return err
	}
	return s.ApplyExternalCommand(cmd.Command, now, cmd.Fees)
}

func (s *State) ApplyExtensionCommand(cmd InternalExtensionCommand) error {
	if err := aetherisaddress.RejectZeroAddress("wallet extension", cmd.Extension); err != nil {
		return err
	}
	if _, ok := s.Wallet.Extensions[string(cmd.Extension)]; !ok {
		return errors.New("unauthorized wallet extension")
	}
	if err := ValidateOutboundMessages(cmd.Messages, true); err != nil {
		return err
	}
	s.SentMessages = append(s.SentMessages, cloneMessages(cmd.Messages)...)
	return nil
}

func (s *State) QueryWalletState() WalletState {
	wallet := s.Wallet
	wallet.PublicKey = append(ed25519.PublicKey(nil), s.Wallet.PublicKey...)
	wallet.Owner = append(sdk.AccAddress(nil), s.Wallet.Owner...)
	wallet.RecoveryAuthority = append(sdk.AccAddress(nil), s.Wallet.RecoveryAuthority...)
	wallet.Extensions = make(map[string]ExtensionState, len(s.Wallet.Extensions))
	for key, value := range s.Wallet.Extensions {
		value.Address = append(sdk.AccAddress(nil), value.Address...)
		wallet.Extensions[key] = value
	}
	return wallet
}

func (s *State) applyVerifiedExternalCommand(cmd ExternalCommand) error {
	switch cmd.Kind {
	case CommandSend:
		s.SentMessages = append(s.SentMessages, cloneMessages(cmd.Messages)...)
	case CommandUpdateSignatureAllowed:
		s.Wallet.SignatureAllowed = cmd.SignatureAllowed
	case CommandInstallExtension:
		if len(s.Wallet.Extensions) >= MaxExtensions {
			return fmt.Errorf("extensions count must be <= %d", MaxExtensions)
		}
		s.Wallet.Extensions[string(cmd.Extension)] = ExtensionState{
			Address:   append(sdk.AccAddress(nil), cmd.Extension...),
			Installed: true,
		}
	case CommandRemoveExtension:
		if _, ok := s.Wallet.Extensions[string(cmd.Extension)]; !ok {
			return errors.New("wallet extension is not installed")
		}
		delete(s.Wallet.Extensions, string(cmd.Extension))
	default:
		return errors.New("unknown wallet command kind")
	}
	return nil
}

func (s *State) snapshot() State {
	checkpoint := State{
		Wallet:       s.QueryWalletState(),
		SentMessages: cloneMessages(s.SentMessages),
	}
	return checkpoint
}

func (s *State) restore(checkpoint State) {
	s.Wallet = checkpoint.Wallet
	s.SentMessages = checkpoint.SentMessages
}

func cloneMessages(messages []OutboundMessage) []OutboundMessage {
	if len(messages) == 0 {
		return nil
	}
	out := make([]OutboundMessage, len(messages))
	for i, msg := range messages {
		out[i] = OutboundMessage{
			To:      append(sdk.AccAddress(nil), msg.To...),
			Amount:  msg.Amount.Sort(),
			Payload: append([]byte(nil), msg.Payload...),
		}
	}
	return out
}
