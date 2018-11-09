.PHONY: setup, check-env

setup:
	npm install -g ganache-cli truffle

migrate: check-env
	truffle migrate --reset --network=${ENV}

fail-require: check-env
	truffle exec ./scripts/fail-require.js --network=${ENV}

fail-assert: check-env
	truffle exec ./scripts/fail-assert.js --network=${ENV}

fail-revert: check-env
	truffle exec ./scripts/fail-revert.js --network=${ENV}

fail-multi-contract-1: check-env
	truffle exec ./scripts/fail-multi-contract-1.js --network=${ENV}

fail-multi-contract-2: check-env
	truffle exec ./scripts/fail-multi-contract-2.js --network=${ENV}

fail-division: check-env
	truffle exec ./scripts/fail-division.js --network=${ENV}

ganache:
	mkdir -p ${HOME}/.tenderly/data/ganache
	ganache-cli --port=7545 --db=${HOME}/.tenderly/data/ganache --gasPrice=0

geth:
	geth --datadir ~/.tenderly/data/geth --dev --dev.period 3 --rpc --rpcapi 'eth,net,web3,admin,debug,personal,miner,txpool' \
	    --ws --wsapi 'eth,net,web3,admin,debug,personal,miner,txpool'

parity:
	parity --chain=dev --jsonrpc-apis=all --ws-apis=all --ipc-apis=all  --jsonrpc-port=9545 --ws-port=9546 \
		--tracing=on --pruning=archive --base-path=~/.tenderly/data/parity --db-path=~/.tenderly/data/parity/db \
		--reseal-min-period 5000 \
		--unlock=0x00a329c0648769a73afac7f9381e08fb43dbea72 --password=pwd.txt

check-env:
ifeq (,$(findstring $(ENV),ganache_parity_geth))
	$(error ENV should be one of [ganache, parity, geth])
endif
