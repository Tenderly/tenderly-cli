const FailContract = artifacts.require('FailContract');

module.exports = function (callback) {
    FailContract.deployed().then(function (c) {
        return c.requireFunction();
    }).then(function (res) {
        console.error("Did not fail transaction!", res);
    }).catch(function (e) {
        console.log("Successfully thrown error!", e)
    }).finally(callback);
};

