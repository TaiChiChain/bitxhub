# Development Tips

# Configuration Loading Logic

The sequence for loading node configurations is as follows:

1. First, construct a Config instance using the default hardcoded configurations in the code.
2. Read the configurations from config.toml and load them into the Config instance.
3. Scan the environment variables and load the read variables into the Config instance.

Therefore, the priority and sequence of loading node configurations are opposite: **Environment Variables > Configuration File > Hardcoded Default Configurations in the Code**. The higher priority will override the lower priority.

**Note: Before executing config generate for node initialization, if environment variables for configuration items are set, the values from the environment variables will be used as the configuration values in the generated config.toml.**

Correspondence of configuration items and environment variables: The names of the configuration items in the configuration file are converted to all uppercase and connected with underscores, with the prefix AXIOM_LEDGER_ (for genesis.toml it is AXIOM_LEDGER_GENESIS_; for consensus.toml it is AXIOM_LEDGER_CONSENSUS_).

For example, if the configuration item for the JSON-RPC port is port.jsonrpc, the corresponding environment variable is AXIOM_LEDGER_PORT_JSONRPC.

# Node Process Management

Nodes started with make cluster and make solo, as well as nodes deployed manually through make package, come with default node process management scripts. For specific usage, refer to the document ["node deploy Document"](./node-deploy.md).

**Note: When upgrading the binary, it is essential to use the update_binary.sh script in the deployment path, as the axiom-ledger file in the deployment path is not the binary but a proxy script (automatically setting the repo path and loading environment variables from .env.sh). The actual node binary is located in tools/bin/axiom-ledger within the deployment path.**

# Rapid Deployment of Multiple Nodes
AxiomLedger comes with a command line tool for multi-node configuration generation.

## Generate default 4 nodes config and keystore
1. Export some common config(support use env to set other config)
```shell
export AXIOM_LEDGER_GENESIS_CHAINID=1111
```
2. Generate
```shell
./axiom-ledger cluster generate-default --target ./nodes --force
```
Nodes configurations and private keys will be generated under target dir.

default info:
```golang
defaultAccountAddrs = []string{
"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
"0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
"0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC",
"0x90F79bf6EB2c4f870365E785982E1f101E93b906",
"0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65",
"0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc",
"0x976EA74026E726554dB657fA54763abd0C3a0aa9",
"0x14dC79964da2C08b23698B3D3cc7Ca32193d9955",
"0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f",
"0xa0Ee7A142d267C1f36714E4a8F75612F20a79720",
"0xBcd4042DE499D14e55001CcbB24a551F3b954096",
"0x71bE63f3384f5fb98995898A86B02Fb2426c5788",
"0xFABB0ac9d68B0B445fB7357272Ff202C5651694a",
"0x1CBd3b2770909D4e10f157cABC84C7264073C9Ec",
"0xdF3e18d64BC6A983f673Ab319CCaE4f1a57C7097",
"0xcd3B766CCDd6AE721141F452C550Ca635964ce71",
"0x2546BcD3c84621e976D8185a91A922aE77ECEc30",
"0xbDA5747bFD65F08deb54cb465eB87D40e51B197E",
"0xdD2FD4581271e230360230F9337D5c0430Bf44C0",
"0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199",
}

defaultCouncilMemberNames = []string{
"S2luZw==", // base64 encode King
"UmVk",     // base64 encode Red
"QXBwbGU=", // base64 encode Apple
"Q2F0",     // base64 encode Cat
}

defaultCouncilMemberKeys = []string{
"b6477143e17f889263044f6cf463dc37177ac4526c4c39a7a344198457024a2f",
"05c3708d30c2c72c4b36314a41f30073ab18ea226cf8c6b9f566720bfe2e8631",
"85a94dd51403590d4f149f9230b6f5de3a08e58899dcaf0f77768efb1825e854",
"72efcf4bb0e8a300d3e47e6a10f630bcd540de933f01ed5380897fc5e10dc95d",
}

defaultCouncilMemberAddrs = []string{
"0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013",
"0x79a1215469FaB6f9c63c1816b45183AD3624bE34",
"0x97c8B516D19edBf575D72a172Af7F418BE498C37",
"0xc0Ff2e0b3189132D815b8eb325bE17285AC898f8",
}

// nodes 

defaultConsensusKeys = []string{
"0x099383c2b41a282936fe9e656467b2ad6ecafd38753eefa080b5a699e3276372",
"0x5d21b741bd16e05c3a883b09613d36ad152f1586393121d247bdcfef908cce8f",
"0x42cc8e862b51a1c21a240bb2ae6f2dbad59668d86fe3c45b2e4710eebd2a63fd",
"0x6e327c2d5a284b89f9c312a02b2714a90b38e721256f9a157f03ec15c1a386a6",
}

defaultP2PKeys = []string{
"0xce374993d8867572a043e443355400ff4628662486d0d6ef9d76bc3c8b2aa8a8",
"0x43dd946ade57013fd4e7d0f11d84b94e2fda4336829f154ae345be94b0b63616",
"0x875e5ef34c34e49d35ff5a0f8a53003d8848fc6edd423582c00edc609a1e3239",
"0xf0aac0c25791d0bd1b96b2ec3c9c25539045cf6cc5cc9ad0f3cb64453d1f38c0",
}

defaultOperatorAddrs = []string{
"0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013",
"0x79a1215469FaB6f9c63c1816b45183AD3624bE34",
"0x97c8B516D19edBf575D72a172Af7F418BE498C37",
"0xc0Ff2e0b3189132D815b8eb325bE17285AC898f8",
}
```

## Quick generate nodes config and keystore by specifying the number of nodes
1. Export some common config(support use env to set other config)
```shell
export AXIOM_LEDGER_GENESIS_CHAINID=1111
```
2. Generate
```shell
./axiom-ledger cluster quick-generate --node-number 8 --target ./nodes --force
```
Nodes configurations and private keys will be generated under target dir.
Each node Operator key is generated in operator.key

## Generate nodes config and keystore by generate config file
1. Show generate config file template
```shell
./axiom-ledger cluster show-config-template
```
2. Edit generate config file (default is cmd/axiom-ledger/cluster_generate_config_temp.toml)
```toml
# if true, node port will be auto increased based on the `default_port`
enable_port_auto_increase = false
```
3. Export some common config(support use env to set other config)
```shell
export AXIOM_LEDGER_GENESIS_CHAINID=1111
```
4. Generate
```shell
./bin/axiom-ledger cluster generate-by-config --config-path ./generate_config.toml --target ./nodes --force
```
Nodes configurations and private keys will be generated under target dir.
Each node Operator key is generated in operator.key


# Debugging Techniques

## Consensus Debugging

The consensus module itself processes transactions and consensus messages in a serial manner (similar to a state machine). Currently, the only way to debug it is to compare the logs of multiple nodes.

**Note: During development and testing, make sure to set the log level of the consensus to debug (default is info), and synchronize the clocks of the machines during deployment.**

## Unrecoverable Node Panic and Restart

When a node experiences a panic and cannot recover after restart (due to consensus and block execution reasons), the node can be started in read-only mode (without starting the p2p and consensus modules) using the readonly parameter. Subsequently, the block and transaction data of the node can be queried through jsonrpc for debugging.

Enter the node's deployment directory and execute the following command to start the node in read-only mode:

```bash
cd ${deploy_path}
./axiom-ledger start --readonly
```