# Vesper Interchain

**Vesper Interchain** is a Cosmos SDK blockchain that implements a collateral vault and stablecoin protocol with full EVM compatibility. Users deposit collateral, mint the `uvusd` stablecoin against it, earn per-block rewards, and are subject to automated liquidation if their position's health factor falls below the protocol threshold.

The chain exposes its vault logic through both the standard Cosmos transaction interface and a **stateful EVM precompile** at address `0x0000000000000000000000000000000000000900`, enabling Solidity contracts to interact with the protocol natively.

---

## Table of Contents

- [Architecture](#architecture)
- [Modules](#modules)
- [EVM Integration](#evm-integration)
- [Build & Install](#build--install)
- [Running from Scratch](#running-vesper-interchain-from-scratch)
- [Development](#development)
- [Testing](#testing)
- [API Reference](#api-reference)

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Vesper Interchain                  │
│                                                     │
│  ┌──────────┐  ┌──────────┐  ┌────────────────┐    │
│  │ x/oracle │  │x/stablecoin│ │  x/collateral  │    │
│  │          │  │          │  │                │    │
│  │ Price    │─▶│ uvusd    │◀─│ Positions LTV  │    │
│  │ feeds    │  │ mint/burn│  │ Liquidation    │    │
│  └──────────┘  └──────────┘  └───────┬────────┘    │
│                                       │             │
│  ┌──────────┐  ┌──────────┐          │             │
│  │x/liquidation│ │x/rewards│◀─────────┘            │
│  │          │  │          │                         │
│  │ Queue    │  │MasterChef│                         │
│  │ bots     │  │ per-block│                         │
│  └──────────┘  └──────────┘                         │
│                                                     │
│  ┌──────────────────────────────────────────────┐   │
│  │         EVM (cosmos/evm)                     │   │
│  │  Vault Precompile @ 0x...0900  |  ERC-20     │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  IBC / ICS-20 ── CometBFT consensus                 │
└─────────────────────────────────────────────────────┘
```

| Property | Value |
|---|---|
| Cosmos SDK | v0.53.6 |
| CometBFT | v0.38.21 |
| cosmos/evm | v1.0.0-rc2 |
| IBC | ibc-go v10.4.0 |
| Go | 1.25+ |
| Binary | `vesper-interchaind` |
| Native denom | `stake` |
| Stablecoin | `uvusd` |
| Account prefix | `vesper` |

---

## Modules

### `x/collateral`

Core vault management. Users open positions by depositing collateral (`uatom`) and borrow `uvusd` up to the configured LTV ratio. Positions are tracked individually and evaluated against oracle prices for liquidation eligibility.

**Messages**

| Message | Description |
|---|---|
| `MsgDepositCollateral` | Deposit collateral and open or increase a position |
| `MsgWithdrawCollateral` | Withdraw collateral if the health ratio remains safe |
| `MsgMintStablecoin` | Borrow `uvusd` up to the max LTV |
| `MsgRepayDebt` | Burn `uvusd` to reduce outstanding debt |
| `MsgLiquidatePosition` | Liquidate an undercollateralised position |
| `MsgUpdateParams` | Governance-only parameter update |

**Key Parameters**

| Parameter | Description |
|---|---|
| `liquidation_ratio` | Minimum collateral-to-debt ratio (e.g. `1.5` = 150%) |
| `max_loan_to_value` | Maximum borrow as a fraction of collateral value (e.g. `0.8`) |
| `liquidation_penalty` | Bonus paid to liquidators (e.g. `0.05` = 5%) |
| `supported_collateral_denom` | Accepted collateral token |
| `min_collateral_amount` | Minimum deposit to open a position |
| `min_debt_amount` | Prevents dust positions |
| `oracle_price_stale_seconds` | Maximum age of an oracle price before it is rejected |

---

### `x/stablecoin`

Thin mint/burn layer for the `uvusd` stablecoin. The collateral module decides when to mint or burn; this module executes those token operations and maintains a `SupplyState` (total minted, total burned) for auditability.

**Queries**

```bash
vesper-interchaind query stablecoin params
vesper-interchaind query stablecoin total-minted
vesper-interchaind query stablecoin total-burned
```

---

### `x/oracle`

On-chain price feed. A single authorized oracle address posts asset prices with timestamps and source attribution. The collateral module reads these prices to evaluate position health and trigger liquidations. The oracle address is upgradeable via governance.

**Messages**

| Message | Description |
|---|---|
| `MsgUpdatePrice` | Authorized oracle submits a new price |
| `MsgUpdateParams` | Governance updates the oracle address |

**Queries**

```bash
vesper-interchaind query oracle params
vesper-interchaind query oracle prices
```

---

### `x/liquidation`

Queue-based liquidation coordination. When the collateral module identifies an unhealthy position, it enqueues a `LiquidationQueue` entry. Anyone can then call `MsgExecuteLiquidation` to execute it, enabling competitive bot participation. Completed liquidations are recorded in `LiquidationRecord` for on-chain history.

**Messages**

| Message | Description |
|---|---|
| `MsgExecuteLiquidation` | Execute a queued liquidation by position ID |
| `MsgUpdateParams` | Governance parameter update |

---

### `x/rewards`

Per-block `uvusd` reward distribution using a MasterChef accumulator pattern. A global `RewardAccumulator` increases by `rewardRate / totalShares` each block (in the EndBlocker), keeping cost **O(1) per block** regardless of the number of depositors. Rewards are settled on deposit, withdrawal, and liquidation to prevent gaming.

**Key mechanics**
- `UpdateShares()` — called by the collateral keeper on position changes
- `ClaimRewards()` — mints and sends pending `uvusd` to the caller
- `GetPendingRewards()` — read-only query of claimable amount
- `AccumulateRewards()` — EndBlocker accumulation step

---

## EVM Integration

Vesper Interchain embeds the `cosmos/evm` module, exposing a Prague-fork EVM alongside the standard Cosmos interface.

### Vault Precompile

**Address:** `0x0000000000000000000000000000000000000900`

A stateful precompile that bridges the EVM directly to the collateral vault. EVM addresses (20 bytes) map 1:1 to Cosmos `AccAddress` using raw bytes, so the same key controls both interfaces.

| Method | Type | Description |
|---|---|---|
| `depositCollateral(uint256)` | TX | Deposit collateral |
| `withdrawCollateral(uint256)` | TX | Withdraw collateral |
| `mintStablecoin(uint256)` | TX | Borrow `uvusd` |
| `repayDebt(uint256)` | TX | Repay `uvusd` debt |
| `liquidatePosition(address)` | TX | Liquidate a position by owner |
| `claimRewards()` | TX | Claim accumulated `uvusd` rewards |
| `getPosition(address)` | Query | Read position state |
| `getPendingRewards(address)` | Query | Read claimable rewards |

Ratio return values are scaled by `1e18`. Cosmos gas consumption is propagated to the EVM gas budget.

### Other Precompiles

| Precompile | Description |
|---|---|
| `bech32` | Cosmos/EVM address format conversions |
| `p256` | ECDSA secp256r1 (EIP-7212) |
| Standard Ethereum | All Prague-fork precompiles |

### Solidity Interfaces

The `contracts/` directory contains:
- `IVaultPrecompile.sol` — Solidity interface for the vault precompile (with events)
- `VesperVault.sol` — Sample EVM contract implementation
- `VUSD.sol` — Stablecoin token contract

---

## Build & Install

**Prerequisites:** Go 1.25+, `make`

```bash
git clone https://github.com/Vesper-Interchain/vesper-interchain.git
cd vesper-interchain
make install
```

This installs `vesper-interchaind` to `~/go/bin/`. Add it to your PATH:

```bash
export PATH=$PATH:$HOME/go/bin
# Make permanent
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc   # or ~/.zshrc
```

---

## Running Vesper Interchain from Scratch

### Step 0 — Clean up any previous installation

```bash
rm -rf vesper-interchain       # remove the local source directory
rm -rf ~/.vesper-interchain    # remove the chain's home and all data
```

### Step 1 — Clone and install

```bash
git clone https://github.com/Vesper-Interchain/vesper-interchain.git
cd vesper-interchain
make install
```

### Step 2 — Initialize the chain

```bash
vesper-interchaind init mynode --chain-id test-chain-x8xCNe
```

This automatically configures both `app.toml` (`minimum-gas-prices = "0stake"`) and `genesis.json` (`feemarket.no_base_fee = true`).

### Step 3 — Create keys

```bash
vesper-interchaind keys add validator --algo eth_secp256k1 --keyring-backend test
vesper-interchaind keys add bob --algo eth_secp256k1 --keyring-backend test
```

### Step 4 — Fund genesis accounts

```bash
# Validator gets 2x so half stays spendable after bonding
vesper-interchaind genesis add-genesis-account \
  $(vesper-interchaind keys show validator -a --keyring-backend test) \
  2000000000000000000stake

vesper-interchaind genesis add-genesis-account \
  $(vesper-interchaind keys show bob -a --keyring-backend test) \
  1000000000000000stake
```

### Step 5 — Create genesis validator

```bash
# Bond only half so validator keeps 10^18 stake spendable
vesper-interchaind genesis gentx validator 1000000000000000000stake \
  --chain-id test-chain-x8xCNe \
  --keyring-backend test
```

### Step 6 — Finalize and validate genesis

```bash
vesper-interchaind genesis collect-gentxs
vesper-interchaind genesis validate-genesis
```

Expected: `File at .../genesis.json is a valid genesis file`

### Step 7 — Start the chain

```bash
vesper-interchaind start
```

You should see blocks being produced every ~5 seconds.

### Step 8 — Verify the chain (new terminal)

```bash
# Node status
vesper-interchaind status

# Validator list
vesper-interchaind query staking validators

# Validator info
vesper-interchaind query staking validator \
  $(vesper-interchaind keys show validator --bech val -a --keyring-backend test)

# Account
vesper-interchaind query auth account \
  $(vesper-interchaind keys show validator -a --keyring-backend test)
```

### Step 9 — Test transactions

```bash
# Send tokens to bob
vesper-interchaind tx bank send validator \
  $(vesper-interchaind keys show bob -a --keyring-backend test) \
  1000000stake \
  --chain-id test-chain-x8xCNe \
  --keyring-backend test \
  --fees 1000stake \
  -y

# Check bob's balance
vesper-interchaind query bank balances \
  $(vesper-interchaind keys show bob -a --keyring-backend test)

# Delegate
vesper-interchaind tx staking delegate \
  $(vesper-interchaind keys show validator --bech val -a --keyring-backend test) \
  1000000stake \
  --from validator \
  --chain-id test-chain-x8xCNe \
  --keyring-backend test \
  --fees 1000stake \
  -y

# Check delegations
vesper-interchaind query staking delegations \
  $(vesper-interchaind keys show validator -a --keyring-backend test)

# Check distribution rewards
vesper-interchaind query distribution rewards \
  $(vesper-interchaind keys show validator -a --keyring-backend test)
```

### Step 10 — Query all modules

```bash
vesper-interchaind query staking params
vesper-interchaind query slashing params
vesper-interchaind query mint params
vesper-interchaind query gov params
vesper-interchaind query consensus params
vesper-interchaind query oracle params
vesper-interchaind query oracle prices
vesper-interchaind query stablecoin params
vesper-interchaind query stablecoin total-minted
vesper-interchaind query stablecoin total-burned
vesper-interchaind query collateral params
vesper-interchaind query upgrade module_versions
```

> **Note**: If you have an old `vesper-interchaind` binary at `/usr/local/bin/`, replace it after a fresh build:
> ```bash
> sudo cp ~/go/bin/vesper-interchaind /usr/local/bin/vesper-interchaind
> ```

---

## Development

```bash
# Run local development chain (Ignite)
ignite chain serve

# Generate / update protobuf
make proto-gen

# Lint
make lint

# Lint and auto-fix
make lint-fix

# Go vet
make govet

# Vulnerability scan
make govulncheck
```

---

## Testing

```bash
# Unit tests + go vet + vulnerability check
make test

# Unit tests only
make test-unit

# Race condition detection
make test-race

# Coverage report (generates HTML)
make test-cover

# Benchmarks
make bench
```

---

## API Reference

The REST API is available at `http://localhost:1317` when the chain is running. Interactive Swagger documentation is served at the same address.

**Module query endpoints:**

```bash
# Standard Cosmos modules
vesper-interchaind query staking params
vesper-interchaind query slashing params
vesper-interchaind query mint params
vesper-interchaind query gov params
vesper-interchaind query consensus params
vesper-interchaind query upgrade module_versions

# Vesper custom modules
vesper-interchaind query oracle params
vesper-interchaind query oracle prices
vesper-interchaind query stablecoin params
vesper-interchaind query stablecoin total-minted
vesper-interchaind query stablecoin total-burned
vesper-interchaind query collateral params
vesper-interchaind query liquidation params
```

---

## Learn More

- [Cosmos SDK](https://docs.cosmos.network)
- [CometBFT](https://docs.cometbft.com)
- [IBC Protocol](https://ibc.cosmos.network)
- [Ignite CLI](https://ignite.com/cli)
- [cosmos/evm](https://github.com/cosmos/evm)
