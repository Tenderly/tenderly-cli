const FailContract = artifacts.require('./FailContract.sol');

contract('FailContract', function () {
    it('should fail', async function () {
        const fail = await FailContract.new();
        await fail.requireFunction(1, 2)
    })
});