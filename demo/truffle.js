const HDWalletProvider = require('truffle-hdwallet-provider-privkey');

module.exports = {
    networks: {
        kovan: {
            provider: function() {
                return new HDWalletProvider(["470cb723777b71d9afca2f4148d60ae6e5744abcc324fd48991c7fa691e7f7db"], `https://kovan.tenderly.app/`);
            },
            network_id: '42',
            gasPrice: 2000000000 // 2 GWei
        },
        local: {
            network_id: '5777',
            host: 'localhost',
            port: 9545,
            gasPrice: 10000000000 // 10 GWei
        },
        geth: {
            host: "127.0.0.1",
            port: 8545,
            network_id: "*",
            gas: 4600000,
            gasPrice: 1000,
        },
        tenderly: {
            host: "127.0.0.1",
            port: 9545,
            network_id: "*",
            gasPrice: 0
        },
        ganache: {
            host: "127.0.0.1",
            port: 7545,
            network_id: "*",
            gasPrice: 0
        }
    }
};
