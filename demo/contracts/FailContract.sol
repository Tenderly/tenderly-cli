pragma solidity ^0.4.24;

contract FailContract {
    function requireFunction(uint a, uint b) external {
        require(a == b);
    }
}