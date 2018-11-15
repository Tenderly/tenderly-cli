pragma solidity ^0.4.24;

contract InternalContract {
    function internalRequireFunction(uint a, uint b) {
        require(a == b);
    }
}

contract FailContract {
    uint c;
    InternalContract internalContract;

    constructor(address _internalContract) public {
        internalContract = InternalContract(_internalContract);
    }

    function requireFunction(uint a, uint b) external {
        require(a == b);
    }

    function assertFunction(uint a, uint b) external {
        assert(a == b);
    }

    function revertFunction() external {
        revert();
    }

    function divisionFunction(uint a, uint b) external {
        c = a / b;
    }

    function internalFunction(uint a, uint b) external {
        internalContract.internalRequireFunction(a, b);
    }
}