# Aetra Slashing System

Slashing in Aetra is a core consensus security primitive, not an optional
module or governance convenience. Every full node must deterministically verify
valid slashing evidence and apply the same stake penalties, validator status
changes, and replay protections during block execution.

## 1. Role Of Slashing In AET L1

Slashing protects:

- validator honesty;
- chain finality safety;
- prevention of forks;
- prevention of double-signing attacks.

Proof-of-Stake security depends on validators having bonded AET collateral at
risk. If validators can sign conflicting blocks, equivocate in consensus votes,
or stay unavailable without penalty, rational validators can sell safety, cause
forks, or free-ride on honest validators. Without slashing:

- double-signing becomes a low-cost strategy during forks;
- finality safety depends on social coordination instead of protocol rules;
- delegators cannot distinguish secure validators from reckless validators;
- attackers can unbond after a violation and escape delayed punishment;
- governance pressure can replace cryptographic evidence with subjective
  punishment.

The Aetra rule is strict: objective, cryptographically verifiable evidence
causes automatic penalties. Subjective accusations do not.

## 2. Staking And Validator Model

Validator set structure:

- each validator has an operator address, consensus public key, staking
  account, voting power, commission params, status, jailed flag, tombstone
  flag, and signing-info record;
- active validator set is selected by bonded stake and bounded by
  governance-controlled `max_validators`;
- voting power is derived deterministically from bonded `naet` stake;
- consensus key changes are delayed and cannot erase past evidence.

Staking requirements:

- staking collateral is AET in base denom `naet`;
- minimum validator self-delegation must be positive and governance-bounded;
- a recommended public-testnet floor is `100 AET` self-delegation, expressed as
  `100_000_000_000naet` with exponent `9`;
- production minimum stake is a security parameter and must be set from
  validator-count, market-cap, and attack-cost assumptions;
- validator commission is bounded by protocol hard limits.

Delegation model:

- delegators can delegate `naet` to validators and receive validator shares;
- delegators earn rewards through the validator and pay validator commission;
- delegators share slashing exposure pro rata because delegated stake
  contributes to the validator's consensus power;
- delegators cannot avoid slashing by redelegating or unbonding after the
  validator violated consensus rules.

Bonding, unbonding, and stake locks:

- bonded stake is immediately slashable;
- unbonding stake remains slashable for evidence whose infraction height is
  inside the evidence age window;
- redelegated stake remains slashable for prior-validator infractions inside
  the evidence age window;
- unbonding period must be longer than the maximum evidence age;
- unbonding completion is delayed until all applicable evidence windows close;
- jailed or tombstoned validators cannot re-enter the active set until protocol
  conditions allow it, and tombstoned validators never return with the same
  consensus identity.

## 3. Slashable Conditions

Penalty rates are expressed in basis points of slashable stake. Governance can
adjust rates only inside hard software bounds.

| Violation | Formal Definition | Cryptographic Detection | Evidence Format | Severity | Penalty |
| --- | --- | --- | --- | --- | ---: |
| Double signing | Same validator signs two different block IDs for the same chain id, height, round, and vote type. | Verify both consensus signatures against the validator consensus public key and verify canonical sign bytes differ only by block id. | `DoubleSignEvidence{validator_consensus_address, chain_id, height, round, vote_type, vote_a, vote_b, signature_a, signature_b, pubkey}` | Critical | `5%` slash, jail, tombstone |
| Equivocation in consensus votes | Same validator signs conflicting prevote/precommit messages that can support incompatible commits at the same height/round. | Verify vote signatures, step/type, height, round, chain id, and incompatible block IDs. | `EquivocationEvidence{validator_consensus_address, height, round, step, vote_a, vote_b, signatures, commit_refs}` | Critical | `5%` slash, jail, tombstone |
| Conflicting block proposals | Same validator proposes two different blocks for the same chain id, height, and round. | Verify proposal signatures against validator consensus public key and block hashes differ. | `ConflictingProposalEvidence{validator_consensus_address, chain_id, height, round, proposal_a, proposal_b, signature_a, signature_b}` | Critical | `5%` slash, jail, tombstone |
| Prolonged downtime | Validator misses more than the allowed missed-block threshold in a signed window. | Deterministic signing-info bitmap updated from consensus begin/end block evidence and commit participation. | `DowntimeEvidence{validator_consensus_address, window_start_height, window_end_height, missed_count, signed_blocks_window_root}` or internal signing-info state transition. | Medium | `0.01%` slash, jail until downtime recovery period |
| Invalid votes beyond tolerance threshold | Validator signs malformed or invalid consensus votes above a configured threshold where invalidity is objectively checkable. | Verify signatures and deterministic vote validation failure reason, then count invalid votes in a bounded window. | `InvalidVoteEvidence{validator_consensus_address, height, round, vote_type, vote, signature, validation_error_code}` plus signing-info accumulator. | Medium | `0.1%` slash, jail if threshold crossed |
| Protocol signing rule violation | Validator signs consensus messages violating protocol signing-domain rules, key-rotation delay, or disabled vote type rules. | Verify signature is valid and signed message violates a deterministic active protocol rule. | `SigningRuleEvidence{validator_consensus_address, rule_id, height, signed_message, signature, active_rule_hash}` | Optional advanced | `0.5%` slash, jail |
| Censorship proof | Only enabled if Aetra has protocol-level inclusion lists or encrypted mempool commitments with objective proof. | Verify inclusion-list commitment, tx validity at commitment height, proposer duty, and absence from block without valid exclusion proof. | `CensorshipEvidence{proposer, height, inclusion_commitment, tx_hash, tx_proof, block_header, absence_proof}` | Optional advanced | `0.5%` slash, jail |

Censorship slashing is disabled until the protocol has objective inclusion
commitments. Mempool gossip complaints, screenshots, or RPC logs are not valid
slashing evidence.

## 4. Evidence And Proof System

Valid slashing evidence must be:

- cryptographically verifiable;
- tied to a chain id and consensus domain;
- tied to an infraction height;
- submitted before evidence expiry;
- signed by the accused validator's consensus key or proven by deterministic
  signing-info state;
- independent of off-chain trust assumptions;
- no off-chain trust assumptions;
- non-subjective.

Evidence structure:

```text
SlashingEvidence {
  evidence_id
  evidence_type
  chain_id
  validator_consensus_address
  validator_operator_address
  infraction_height
  infraction_time
  consensus_pubkey
  signed_messages[]
  signatures[]
  block_headers[]
  commit_or_proposal_proofs[]
  validation_error_code
}
```

Nodes independently verify:

1. evidence type is enabled at infraction height;
2. evidence id has not already been processed;
3. evidence is inside max evidence age;
4. validator existed and had slashable stake at infraction height;
5. consensus public key matches validator signing info at infraction height;
6. signatures verify against canonical consensus sign bytes;
7. messages are objectively conflicting or objectively invalid;
8. submitted proof roots match block headers or local canonical state;
9. penalty and status transition are deterministic.

No governance vote, validator multisig, RPC operator, block explorer, or
off-chain committee is trusted to decide whether evidence is valid.

## 5. Slashing Execution Flow

Lifecycle:

```text
violation occurs
  -> evidence is gossiped or included in a block
  -> proposer includes evidence or internal signing-info transition
  -> all nodes validate evidence deterministically
  -> consensus commits block containing valid evidence
  -> slashable stake is reduced automatically
  -> penalty is burned/allocated by deterministic split
  -> validator is jailed or tombstoned
  -> active validator set updates
  -> processed evidence id is stored forever or until safe pruning
```

Replay protection:

- evidence id is the hash of evidence type, chain id, validator consensus
  address, infraction height, round, vote type, and canonical conflicting
  message hashes;
- processed evidence ids are stored in slashing state;
- evidence for the same infraction can be submitted many times but applied
  once;
- duplicate evidence returns an idempotent no-op result.

Conflicting evidence resolution:

- if multiple valid evidence objects prove the same infraction, the first
  committed evidence applies the slash and later duplicates are no-ops;
- if evidence objects prove different infractions, each applies once;
- if penalties would exceed remaining slashable stake, slash amount is capped
  at remaining slashable stake;
- critical tombstone evidence dominates medium downtime evidence for status:
  tombstone is terminal.

Automatic execution:

- slashing is not a discretionary governance action;
- valid evidence must either be included by honest proposers or remain
  broadcast until inclusion;
- evidence spam is limited by evidence size bounds, per-block evidence limits,
  signature prechecks, and evidence submission fees/deposits refundable only
  for valid evidence.

## 6. Economic Security Model

Slashing prevents rational attacks by making the expected cost of violating
consensus exceed the expected profit:

```text
expected_attack_cost =
  slash_fraction * slashable_stake
  + lost_future_rewards
  + lost_commission
  + tombstone_reputation_loss

attack_is_irrational if expected_attack_cost > expected_attack_reward
```

Consensus safety targets:

- `< 1/3` malicious voting power cannot break safety;
- `1/3+` malicious voting power is a halt/liveness threat and must trigger
  monitoring;
- `2/3+` malicious voting power can finalize malicious state and must be
  economically infeasible.

Minimum stake:

- validator minimum self-delegation prevents zero-cost validator identities;
- total bonded value must be large enough that buying or bribing `1/3+` power
  is uneconomic;
- production security requires monitoring top-N concentration and bonded value
  against protected value.

Delegator impact:

- delegators are slashed pro rata because their stake contributes to validator
  power;
- rewards and unbonding claims are recalculated after slash;
- delegators are incentivized to choose reliable validators and diversify;
- delegators must be able to see validator uptime, commission, slashing
  history, concentration, and tombstone status.

Penalty distribution:

```text
critical_slash:
  burn:     80%
  treasury: 15%
  reporter: 5% capped by max_reporter_reward

medium_slash:
  burn:     90%
  treasury: 10%
  reporter: 0%
```

Reporter rewards are optional and capped to avoid evidence-spam incentives.
Critical evidence reporters receive rewards only after full evidence
verification and only once per evidence id.

## 7. Validator Lifecycle Impact

Text lifecycle diagram:

```text
candidate
  -> bonded validator
  -> active validator set
  -> missed blocks above threshold
  -> jailed
  -> downtime recovery period
  -> unjailed by valid tx
  -> active validator set

active validator set
  -> critical evidence confirmed
  -> slashed
  -> jailed
  -> tombstoned
  -> permanently removed for that consensus identity

active validator set
  -> unbond request
  -> unbonding stake locked
  -> evidence window remains open
  -> no valid evidence until expiry
  -> stake released

active validator set
  -> redelegation request
  -> redelegating stake remains slashable for source validator evidence window
  -> evidence window expires
  -> redelegation fully safe
```

Effects:

- active validator set removes jailed or tombstoned validators on the next
  deterministic validator update;
- jailed validators cannot sign as active validators;
- downtime jail can be cleared after recovery rules pass;
- tombstone is permanent for severe equivocation if enabled;
- unbonding and redelegating stake is slashable for infractions committed
  before the stake moved;
- there is no escape window for malicious validators who unbond after signing
  conflicting evidence.

## 8. Governance Constraints

Governance can adjust only bounded parameters:

- critical penalty rates inside `[3%, 10%]`;
- medium downtime penalty inside `[0.001%, 0.1%]`;
- invalid vote penalty inside `[0.01%, 1%]`;
- minimum validator self-delegation;
- validator set size;
- signed-block window and downtime threshold;
- evidence max age, only if it remains shorter than unbonding period;
- reporter reward cap;
- treasury/burn/reporter split inside hard bounds.

Governance cannot:

- reverse valid slashing events;
- override cryptographic proofs;
- selectively punish validators;
- exempt specific validators from evidence verification;
- shorten unbonding below evidence max age;
- disable double-sign/equivocation slashing without software upgrade;
- redirect slashed funds outside deterministic split for one validator.

Hard safety bounds require a software upgrade, not an ordinary parameter vote.

## 9. Security Model And Attack Resistance

Validator collusion:

- top-N concentration metrics and `1/3+` alerts expose cartel risk;
- slashing makes equivocation costly even for coordinated validators;
- delegation transparency lets delegators move away from correlated operators.

Long-range attacks:

- unbonding is longer than evidence max age;
- clients must use weak-subjectivity checkpoints and state-sync trust data;
- evidence for historical infractions remains valid during the evidence
  window even if stake starts unbonding.

Bribery attacks:

- rational validators compare bribe value against slash, lost future rewards,
  lost commission, tombstone, and reputational loss;
- minimum bonded value and concentration monitoring raise bribery cost;
- delegators bear slash exposure and therefore pressure validators toward
  operational safety.

Fake slashing evidence:

- invalid evidence fails signature, domain, height, public key, state-root, or
  conflict checks;
- evidence submission requires fees or deposits;
- invalid evidence is rejected before stake mutation;
- repeated invalid evidence is rate-limited and can harm sender reputation.

Evidence spam and griefing:

- evidence size is bounded;
- per-block evidence count is bounded;
- cheap prechecks reject malformed evidence before expensive verification;
- reporter reward is capped and only paid for valid non-duplicate evidence;
- duplicate valid evidence is idempotent and cannot multiply rewards.

The system remains secure under rational adversaries because profitable attacks
must overcome slashable stake, future reward loss, evidence verifiability,
unbonding locks, and social/market consequences of tombstone status.

## 10. Cryptographic Assumptions Summary

The slashing design assumes:

- consensus signatures are existentially unforgeable under chosen-message
  attack for the configured consensus key scheme;
- canonical vote/proposal sign bytes are domain-separated by chain id, height,
  round, step, and message type;
- block headers, commits, and evidence roots are deterministic and
  cryptographically committed by consensus;
- validator consensus key history is stored and export/import deterministic;
- hash functions used for block ids and evidence ids are collision resistant;
- all validators run the same evidence verification and slashing state
  transition rules.

## Slashable Event Table

| Event | Severity | Penalty | Status Impact | Reporter Reward |
| --- | --- | ---: | --- | ---: |
| Double signing | Critical | `5%` | jail + tombstone + removal | up to `5%` of slashed amount, capped |
| Consensus vote equivocation | Critical | `5%` | jail + tombstone + removal | up to `5%` of slashed amount, capped |
| Conflicting block proposals | Critical | `5%` | jail + tombstone + removal | up to `5%` of slashed amount, capped |
| Prolonged downtime | Medium | `0.01%` | jail until recovery | none |
| Invalid votes beyond threshold | Medium | `0.1%` | jail if threshold crossed | none |
| Protocol signing rule violation | Optional advanced | `0.5%` | jail | capped, only if enabled |
| Censorship proof | Optional advanced | `0.5%` | jail | capped, only if objective inclusion proofs exist |

## Exact Penalty Structure

Default public-testnet penalty parameters:

```text
double_sign_slash_fraction        = 0.05
equivocation_slash_fraction       = 0.05
conflicting_proposal_fraction     = 0.05
downtime_slash_fraction           = 0.0001
invalid_vote_slash_fraction       = 0.001
signing_rule_violation_fraction   = 0.005
censorship_slash_fraction         = 0.005

critical_distribution:
  burn_ratio      = 0.80
  treasury_ratio  = 0.15
  reporter_ratio  = 0.05

medium_distribution:
  burn_ratio      = 0.90
  treasury_ratio  = 0.10
  reporter_ratio  = 0.00
```

All arithmetic is integer `naet` arithmetic. Rounding dust goes to burn to keep
`slashed_amount == burned + treasury + reporter` deterministic.

Slashing in Aetra is a deterministic, protocol-level enforcement mechanism that guarantees economic security of consensus through enforceable stake penalties.
