pragma solidity ^0.4.24;

contract FailContract {
    function requireFunction() external {
        require(1 == 2);
    }
}