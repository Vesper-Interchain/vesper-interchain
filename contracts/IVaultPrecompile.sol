// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title IVaultPrecompile
 * @notice Solidity interface for the Vesper Vault stateful precompile.
 *
 * The precompile lives at VAULT_PRECOMPILE_ADDRESS and wraps the Cosmos
 * x/collateral module, allowing EVM callers to manage collateral vaults
 * directly from Solidity.
 *
 * Addresses are Cosmos bech32 addresses ABI-encoded as bytes32 using the
 * cosmos/evm bech32 precompile convention (raw 20-byte address).
 *
 * All amounts are in the smallest denomination (uatom for collateral,
 * uvusd for debt) as uint256.
 */
interface IVaultPrecompile {
    // -----------------------------------------------------------------------
    // Events
    // -----------------------------------------------------------------------

    event DepositCollateral(address indexed owner, uint256 amount);
    event WithdrawCollateral(address indexed owner, uint256 amount);
    event MintStablecoin(address indexed owner, uint256 amount, string collateralRatio);
    event RepayDebt(address indexed owner, uint256 amount, string remainingDebt);
    event LiquidatePosition(address indexed owner, address indexed liquidator);
    event ClaimRewards(address indexed owner, uint256 amount);

    // -----------------------------------------------------------------------
    // Transactions
    // -----------------------------------------------------------------------

    /**
     * @notice Deposit `amount` uatom collateral into the caller's vault.
     * @param amount Collateral amount in uatom (uint256).
     * @return collateralAmount Updated total collateral after deposit.
     */
    function depositCollateral(uint256 amount) external returns (uint256 collateralAmount);

    /**
     * @notice Withdraw `amount` uatom collateral from the caller's vault.
     * @param amount Collateral amount to withdraw.
     * @return collateralRatio Updated collateral ratio as a scaled integer (×1e18).
     */
    function withdrawCollateral(uint256 amount) external returns (uint256 collateralRatio);

    /**
     * @notice Mint `amount` uvusd against the caller's collateral.
     * @param amount uvusd amount to mint.
     * @return mintedAmount Amount minted.
     * @return collateralRatio Updated collateral ratio (×1e18).
     */
    function mintStablecoin(uint256 amount) external returns (uint256 mintedAmount, uint256 collateralRatio);

    /**
     * @notice Repay `amount` uvusd debt in the caller's vault.
     * @param amount uvusd amount to repay.
     * @return remainingDebt Remaining debt after repayment.
     * @return collateralRatio Updated collateral ratio (×1e18).
     */
    function repayDebt(uint256 amount) external returns (uint256 remainingDebt, uint256 collateralRatio);

    /**
     * @notice Liquidate the vault of `owner`.  Caller must hold sufficient
     *         uvusd to cover the debt.
     * @param owner Address of the vault owner.
     */
    function liquidatePosition(address owner) external;

    /**
     * @notice Claim pending rewards for the caller.
     * @return claimed Amount of uvusd claimed as rewards.
     */
    function claimRewards() external returns (uint256 claimed);

    // -----------------------------------------------------------------------
    // Queries
    // -----------------------------------------------------------------------

    /**
     * @notice Returns the vault position of `owner`.
     * @param owner Address to query.
     * @return collateralAmount Collateral held (uatom).
     * @return debtAmount Outstanding debt (uvusd).
     * @return collateralRatio Current collateral ratio (×1e18, 0 if no debt).
     */
    function getPosition(address owner)
        external
        view
        returns (
            uint256 collateralAmount,
            uint256 debtAmount,
            uint256 collateralRatio
        );

    /**
     * @notice Returns pending claimable rewards for `owner`.
     * @param owner Address to query.
     * @return pendingRewards uvusd rewards not yet claimed.
     */
    function getPendingRewards(address owner) external view returns (uint256 pendingRewards);
}

/// @notice Well-known address of the Vault Precompile on Vesper chain.
address constant VAULT_PRECOMPILE_ADDRESS = address(0x0000000000000000000000000000000000000900);
