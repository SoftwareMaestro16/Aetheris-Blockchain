# Security Risks And Controls

Track 9 records economic, spam, staking, and metadata risk controls that must
remain visible before public testnet and production. This file is a security
planning artifact; it does not replace keeper tests, adversarial tests,
state-export checks, or audit findings.

## Infinite Supply Risks

Risks:

- excessive inflation;
- validator reward dilution;
- weak token confidence;
- governance abuse.

Controls:

- hard inflation cap;
- target staking controller;
- public mint/burn telemetry;
- governance bounds;
- export/import supply invariants.

Production control notes:

- AET has uncapped PoS supply, but configured inflation remains bounded by hard
  protocol caps.
- Governance may tune soft mint parameters only inside software-defined bounds.
- Mint, burn, reward, slash, and fee accounting must be exported/imported
  deterministically.
- Public telemetry must expose current inflation, annual mint, annual burn,
  staking ratio, net issuance, and validator reward allocation.

## Deflation Risks

Risks:

- burn exceeds mint for too long;
- validator income falls;
- users hoard instead of transacting;
- tx fees become politically hard to lower.

Controls:

- burn ratio floor/ceiling;
- validator baseline mint rewards;
- fee caps;
- deflation guard.

Production control notes:

- Burn response is bounded and smoothed; activity cannot force unbounded
  deflation.
- Validator security budget must retain baseline mint rewards even when fee
  revenue is low.
- Dynamic transaction fees remain capped so validator income cannot depend on
  high user fees.
- The system balance controller must reduce burn pressure when sustained burn
  materially exceeds mint.

## Spam Risks

Risks:

- low-cost tx floods;
- memo spam;
- async queue flooding;
- contract deploy spam;
- domain auction spam.

Controls:

- reputation rate limits;
- per-account queue caps;
- per-contract queue caps;
- memo byte fees;
- deploy deposits;
- domain auction bid deposits;
- bounded block processing;
- scheduler fairness.

Production control notes:

- Spam prevention must rely on deterministic admission limits, queue caps,
  rate limits, and deposits instead of unbounded fee escalation.
- Memo text is bounded, byte-priced, and cannot affect consensus execution.
- Async processing must enforce max messages per tx, per block, per account,
  and per contract.
- Scheduler fairness must prevent wealthy users from fully starving normal
  users during congestion.

## Staking Attacks

Risks:

- stake concentration;
- validator cartel;
- long-range attack;
- validator downtime;
- double-sign.

Controls:

- unbonding period;
- slashing;
- tombstone for severe equivocation if enabled;
- validator concentration alerts;
- commission bounds;
- delegation transparency;
- snapshot/state-sync safety.

Production control notes:

- Unbonding and state-sync policy must make long-range attacks economically
  and operationally difficult.
- Downtime and double-sign evidence must produce deterministic slashing and
  jailing behavior.
- Validator concentration metrics must alert for top-N voting power and
  `1/3+` or `2/3+` risk thresholds.
- Snapshot and state-sync restore must preserve staking roots and must not
  bypass validator-set or slashing state.

## Economic Attacks

Risks:

- fee market manipulation;
- fake volume to trigger burn;
- reputation farming;
- domain squatting;
- token metadata spoofing;
- bridge asset spoofing.

Controls:

- fee multiplier smoothing;
- bounded burn response;
- deterministic reputation decay;
- auction pricing and renewal;
- native token metadata reservation;
- bridged asset namespace isolation.

Production control notes:

- Fee multiplier smoothing and hard fee caps prevent Ethereum-style fee
  auctions and unbounded escalation.
- Fake volume cannot drive unbounded burn because burn response is capped.
- Reputation changes are deterministic, decayed, bounded, and not directly
  purchasable.
- Domain auctions require pricing, renewal, anti-snipe bounds, and bid
  deposits before public launch.
- Native `AET`/`naet` metadata remains reserved across token, NFT, SBT, DEX,
  IBC, bridge, and resolver surfaces.
- Bridge assets must stay namespaced and cannot spoof native AET or pay
  Aetra protocol fees.

## Audit Gate

Public testnet cannot proceed while any high/critical finding in these classes
is untriaged:

- uncontrolled mint or burn behavior;
- fee cap bypass or non-`naet` fee payment;
- spam path that bypasses rate, queue, memo, deploy, or scheduler controls;
- staking/slashing condition that can corrupt validator set safety;
- metadata or bridge spoofing that can misrepresent native AET;
- state export/import path that drops supply, queue, or staking invariants.
