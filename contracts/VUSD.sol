// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title VUSD
 * @notice Vesper USD stablecoin ERC-20 representation on the EVM side.
 *
 * On the Cosmos side, the stablecoin is issued as "uvusd" by the x/collateral
 * module when users deposit collateral.  When uvusd is bridged to the EVM
 * chain via IBC-ERC20 (cosmos/evm x/erc20), this contract holds the
 * corresponding ERC-20 balance.
 *
 * A designated minter (the Vault Precompile or an IBC bridge contract) may
 * mint and burn tokens.
 *
 * Decimals: 6  (matching the Cosmos uvusd denomination)
 */
contract VUSD is ERC20, ERC20Burnable, Ownable {
    /// @notice Address authorised to mint new tokens (vault precompile or bridge).
    address public minter;

    event MinterChanged(address indexed oldMinter, address indexed newMinter);

    modifier onlyMinter() {
        require(msg.sender == minter, "VUSD: caller is not the minter");
        _;
    }

    constructor(address _minter) ERC20("Vesper USD", "VUSD") Ownable(msg.sender) {
        require(_minter != address(0), "VUSD: zero minter address");
        minter = _minter;
    }

    /// @inheritdoc ERC20
    function decimals() public pure override returns (uint8) {
        return 6;
    }

    /// @notice Mint `amount` tokens to `to`.  Called by the vault precompile.
    function mint(address to, uint256 amount) external onlyMinter {
        _mint(to, amount);
    }

    /// @notice Burn `amount` tokens from `from`.  Called by the vault precompile.
    function burnFrom(address from, uint256 amount) public override onlyMinter {
        _burn(from, amount);
    }

    /// @notice Transfer minting authority to a new address.
    function setMinter(address newMinter) external onlyOwner {
        require(newMinter != address(0), "VUSD: zero minter address");
        emit MinterChanged(minter, newMinter);
        minter = newMinter;
    }
}
