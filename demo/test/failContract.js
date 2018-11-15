var InternalContract = artifacts.require('InternalContract');
const FailContract = artifacts.require('./FailContract.sol');

contract('FailContract', function () {
    it('require should fail', async function () {
        const fail = await FailContract.new(0x0);
        await fail.requireFunction(1, 2)
    });

    it('assert should fail', async function () {
        const fail = await FailContract.new(0x0);
        await fail.assertFunction(1, 2)
    });

    it('revert should fail', async function () {
        const fail = await FailContract.new(0x0);
        await fail.revertFunction()
    });

    it('division should fail', async function () {
        const fail = await FailContract.new(0x0);
        await fail.divisionFunction(5, 0)
    });

    it('internal require should fail', async function () {
        const internal = await InternalContract.new();
        const fail = await FailContract.new(internal.address);
        await fail.internalFunction(1, 2)
    })
});