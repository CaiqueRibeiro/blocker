# 💸 Blocker - Bitcoin blockchain simulator

Blocker is a simplified version of blockchain algorithms that works on UTXO transactions and connect peer to peer with nodes in chain.

## Specs

|     | Specs                                                                                                                |
| --- | -------------------------------------------------------------------------------------------------------------------- |
| 🚀  | **Go** Fast and easy language for performance apps.                   |
| 🧙🏼‍♀️  | **gRPC** modern open source high performance Remote Procedure Call (RPC) framework that can run in any environment.                                                                                         |
## How to run
The main.go file of project creates dummy blocks to simulate connection and UTXO transactions.
To run this simulation, execute de Makefile command `make run`.
```bash
make run
```
It will generate blocks locally and communicate with each other, logging its simulation in terminal.

```bash
INFO	node/node.go:113	node started...	{"port": ":3000"}
INFO	node/node.go:148	starting validator loop	{"pubkey": {}, "blockTime": "5s"}
INFO	node/node.go:113	node started...	{"port": ":4000"}
DEBUG	node/node.go:177	dialing remote node	{"we": ":4000", "remote": ":3000"}
DEBUG	node/node.go:253	new peer connected	{"we": ":3000", "remoteNode": ":4000", "height": 0}
DEBUG	node/node.go:253	new peer connected	{"we": ":4000", "remoteNode": ":3000", "height": 0}
INFO	node/node.go:113	node started...	{"port": ":6000"}
DEBUG	node/node.go:177	dialing remote node	{"we": ":6000", "remote": ":4000"}
DEBUG	node/node.go:253	new peer connected	{"we": ":4000", "remoteNode": ":6000", "height": 0}
DEBUG	node/node.go:253	new peer connected	{"we": ":6000", "remoteNode": ":4000", "height": 0}
DEBUG	node/node.go:177	dialing remote node	{"we": ":6000", "remote": ":3000"}
DEBUG	node/node.go:253	new peer connected	{"we": ":3000", "remoteNode": ":6000", "height": 0}
DEBUG	node/node.go:253	new peer connected	{"we": ":6000", "remoteNode": ":3000", "height": 0}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56845", "hash": "a3b5b92c2d7a03fcf96bdb30af8d4e5492cba4079363e9fc67263c1249616bac", "we": ":3000"}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56838", "hash": "a3b5b92c2d7a03fcf96bdb30af8d4e5492cba4079363e9fc67263c1249616bac", "we": ":4000"}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56840", "hash": "a3b5b92c2d7a03fcf96bdb30af8d4e5492cba4079363e9fc67263c1249616bac", "we": ":6000"}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56847", "hash": "792d2989114767038d6527768d157dc1aff2d58d5189b2f4873410f914f4c5e3", "we": ":3000"}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56838", "hash": "792d2989114767038d6527768d157dc1aff2d58d5189b2f4873410f914f4c5e3", "we": ":4000"}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56840", "hash": "792d2989114767038d6527768d157dc1aff2d58d5189b2f4873410f914f4c5e3", "we": ":6000"}
DEBUG	node/node.go:153	time to create a new block	{"lenTx": 2}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56848", "hash": "dc1b4b78ebf26ab0b66bb8a1eff68c3e59bc4b0fb992bfaf176a364d9eb43aa9", "we": ":3000"}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56838", "hash": "dc1b4b78ebf26ab0b66bb8a1eff68c3e59bc4b0fb992bfaf176a364d9eb43aa9", "we": ":4000"}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56842", "hash": "dc1b4b78ebf26ab0b66bb8a1eff68c3e59bc4b0fb992bfaf176a364d9eb43aa9", "we": ":6000"}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56849", "hash": "04c6a7e51d3d8fb1fe93f8fd5d293772cd0f8bb719b5f27b959964cd19236426", "we": ":3000"}
DEBUG	node/node.go:137	received tx	{"from": "127.0.0.1:56838", "hash": "04c6a7e51d3d8fb1fe93f8fd5d293772cd0f8bb719b5f27b959964cd19236426", "we": ":4000"}
```

## Tests
To run all the tests (unit and integration), execute the bash command `make test`.
```bash
make test

=== RUN   TestGeneratePrivateKey
--- PASS: TestGeneratePrivateKey (0.00s)
=== RUN   TestGeneratePrivateKeyFromString
--- PASS: TestGeneratePrivateKeyFromString (0.00s)
=== RUN   TestPrivateKeySign
--- PASS: TestPrivateKeySign (0.00s)
=== RUN   TestPublicKeyToAddress
--- PASS: TestPublicKeyToAddress (0.00s)
PASS
ok      github.com/CaiqueRibeiro/blocker/crypto     (cached)
?       github.com/CaiqueRibeiro/blocker/proto      [no test files]
?       github.com/CaiqueRibeiro/blocker/util       [no test files]
=== RUN   TestNewChain
--- PASS: TestNewChain (0.00s)
=== RUN   TestAddBlock
--- PASS: TestAddBlock (0.01s)
=== RUN   TestChainHeight
--- PASS: TestChainHeight (0.01s)
=== RUN   TestAddBlockWithTxInsufficientFunds
--- PASS: TestAddBlockWithTxInsufficientFunds (0.00s)
=== RUN   TestAddBlockWithTx
--- PASS: TestAddBlockWithTx (0.00s)
PASS
ok      github.com/CaiqueRibeiro/blocker/node       (cached)
=== RUN   TestCalculateRootHash
--- PASS: TestCalculateRootHash (0.00s)
=== RUN   TestHashBlock
--- PASS: TestHashBlock (0.00s)
=== RUN   TestSignVerifyBlock
--- PASS: TestSignVerifyBlock (0.00s)
=== RUN   TestNewtransaction
```
