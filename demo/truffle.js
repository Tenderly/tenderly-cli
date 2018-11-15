const HDWalletProvider = require('truffle-hdwallet-provider-privkey');

module.exports = {
    networks: {
        tenderly: {
            provider: function() {
                return new HDWalletProvider(["24c6587bde13b53654e689d9918d3391b832d5f88741b59166dc978b73557a85"], `http://127.0.0.1:9545`);
            },
            network_id: '5777',
            gasPrice: 2000000000 // 2 GWei
        },
    }
};
