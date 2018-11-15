var InternalContract = artifacts.require('InternalContract');
var FailContract = artifacts.require('FailContract');

module.exports = function (deployer) {
    deployer.then(function () {
        return deployer.deploy(InternalContract)
    }).then(function () {
        return deployer.deploy(FailContract, InternalContract.address)
    })
};
