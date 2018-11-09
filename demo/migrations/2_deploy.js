var FailContract = artifacts.require('FailContract');

module.exports = function (deployer) {
    deployer.then(function () {
        return deployer.deploy(FailContract)
    })
};
