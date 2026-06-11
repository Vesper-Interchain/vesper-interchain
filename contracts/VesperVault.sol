// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./IVaultPrecompile.sol";
import "./VUSD.sol";

/**
 * @title VesperVault
 * @notice High-level vault wrapper for EVM users.
 *
 * EVM users interact with this contract instead of calling the low-level
 * precompile directly.  The contract:
 *   1. Accepts VUSD tokens for debt repayment (burns them via the precompile).
 *   2. Emits structured events for off-chain indexers.
 *   3. Provides a unified entry point for DeFi composability.
 *
 * IMPORTANT: collateral (uatom) lives on the Cosmos side.  Depositing /
 * withdrawing collateral from EVM requires that the user first bridge uatom
 * via IBC.  The precompile then instructs the Cosmos keeper to move the
 * cosmos-side funds.
 *
 * Usage:
 *   // Deposit 1 ATOM (1_000_000 uatom) of collateral
 *   IVaultPrecompile(VAULT_PRECOMPILE_ADDRESS).depositCollateral(1_000_000);
 *
 *   // Mint 500_000 uvusd against it
 *   IVaultPrecompile(VAULT_PRECOMPILE_ADDRESS).mintStablecoin(500_000);
 *
 *   // Repay via this wrapper (handles VUSD approve → repay flow)
 *   vusd.approve(address(vaultWrapper), 500_000);
 *   vaultWrapper.repayDebt(500_000);
 */
contract VesperVault {
    IVaultPrecompile public constant vault = IVaultPrecompile(VAULT_PRECOMPILE_ADDRESS);
    VUSD public immutable vusdToken;

    event Deposited(address indexed user, uint256 amount);
    event Withdrawn(address indexed user, uint256 amount);
    event Minted(address indexed user, uint256 amount);
    event Repaid(address indexed user, uint256 amount, uint256 remainingDebt);
    event Liquidated(address indexed owner, address indexed liquidator);
    event RewardsClaimed(address indexed user, uint256 amount);

    constructor(address _vusd) {
        require(_vusd != address(0), "VesperVault: zero VUSD address");
        vusdToken = VUSD(_vusd);
    }

    // -----------------------------------------------------------------------
    // Collateral operations — delegate to precompile
    // -----------------------------------------------------------------------

    function depositCollateral(uint256 amount) external returns (uint256 collateralAmount) {
        collateralAmount = vault.depositCollateral(amount);
        emit Deposited(msg.sender, amount);
    }

    function withdrawCollateral(uint256 amount) external returns (uint256 collateralRatio) {
        collateralRatio = vault.withdrawCollateral(amount);
        emit Withdrawn(msg.sender, amount);
    }

    function mintStablecoin(uint256 amount) external returns (uint256 minted, uint256 ratio) {
        (minted, ratio) = vault.mintStablecoin(amount);
        emit Minted(msg.sender, minted);
    }

    /**
     * @notice Repay `amount` uvusd debt.
     *
     * The caller must have approved this contract to spend `amount` VUSD
     * before calling.  The tokens are transferred to this contract and
     * then the precompile burns them on the Cosmos side.
     */
    function repayDebt(uint256 amount) external returns (uint256 remaining, uint256 ratio) {
        vusdToken.transferFrom(msg.sender, address(this), amount);
        vusdToken.approve(VAULT_PRECOMPILE_ADDRESS, amount);
        (remaining, ratio) = vault.repayDebt(amount);
        emit Repaid(msg.sender, amount, remaining);
    }

    function liquidatePosition(address owner) external {
        vault.liquidatePosition(owner);
        emit Liquidated(owner, msg.sender);
    }

    function claimRewards() external returns (uint256 claimed) {
        claimed = vault.claimRewards();
        emit RewardsClaimed(msg.sender, claimed);
    }

    // -----------------------------------------------------------------------
    // View helpers
    // -----------------------------------------------------------------------

    function getPosition(address owner)
        external
        view
        returns (
            uint256 collateralAmount,
            uint256 debtAmount,
            uint256 collateralRatio
        )
    {
        return vault.getPosition(owner);
    }

    function getPendingRewards(address owner) external view returns (uint256) {
        return vault.getPendingRewards(owner);
    }

    /**
     * @notice Returns true if a position is eligible for liquidation.
     *         Collateral ratio < 150 % (1.5e18).
     */
    function isLiquidatable(address owner) external view returns (bool) {
        (, , uint256 cr) = vault.getPosition(owner);
        if (cr == 0) return false; // no debt
        return cr < 1.5e18;
    }
}
