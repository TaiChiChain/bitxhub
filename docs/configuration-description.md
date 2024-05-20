# genesis.toml - Genesis Configuration
```toml
# Chain ID
chainid = 1357
# Genesis block timestamp
timestamp = 1716185531
# Smart account system contract adminn address
smart_account_admin = '0xc0Ff2e0b3189132D815b8eb325bE17285AC898f8'
# Addresses of providers on the initial whitelist
whitelist_providers = []
# Native Coin `axc` Configuration
[axc]
# Total supply(e.g. 10000000axc; 1000gmol; 1000mol; 1000)
total_supply = '320000000axc'

# Incentive Configuration
[incentive]
  # Referral reward(not supported yet)
  [incentive.referral]
    avg_block_reward = '0'
    block_to_none = 0


# Governance council member list
[[council_members]]
# Committee member address
address = '0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013'
# Voting weight
weight = 1
# Name, base64 encoded
name = 'S2luZw=='

[[council_members]]
address = '0x79a1215469FaB6f9c63c1816b45183AD3624bE34'
weight = 1
name = 'UmVk'


# Initial token balance for accounts
balance = '1000000000000000000000000000'
# Addresses of providers on the initial whitelist
init_white_list_providers = [
    '0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013'
]
# List of initial accounts with a certain token balance
accounts = [
    '0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266'
]

# Epoch information configuration
[epoch_info]
version = 1
# Serial number of the initial epoch
epoch = 1
# Block interval for the epoch duration (epoch change occurs after this interval)
epoch_period = 10000
# Start block of the initial epoch(must be zero)
start_block = 0
# Consensus parameter configuration
  [epoch_info.consensus_params]
    # Proposer node election type:
    # wrf (weighted random function, randomly select a node from all validator nodes based on the consensus voting weight, with a term of checkpoint_period, and reselection after the term expires)
    # abnormal-rotation (original pbft logic, no rotation under normal circumstances, rotation occurs during viewchange)
    proposer_election_type = 'wrf'
    # Block interval for checkpoints (consensus waits for module execution and executes checkpoint logic after every checkpoint blocks to confirm blocks before the checkpoint)
    checkpoint_period = 1
    # Number of checkpoints within the high watermark (only processes consensus messages within the high watermark)
    high_watermark_checkpoint_period = 10
    # Minimum number of validators
    Min_validator_num = 4
    # Maximum number of validators (a batch of validator nodes will be elected if it exceeds the limit)
    max_validator_num = 4
    # Maximum number of transactions in a single block
    block_max_tx_num = 500
    # Enable timed generation of empty blocks (blocks are generated even without transactions)
    enable_timed_gen_empty_block = false
    # Reduced consensus voting weight when the primary node is inactive
    not_active_weight = 1
    # Interval for weight reduction recovery; reduced consensus voting weight will recover after epoch changes or after a certain number of checkpoint intervals
    abnormal_node_exclude_view = 10
    # Interval for nodes to propose blocks again as a percentage of the total number of validators, ensuring that nodes cannot propose blocks continuously (minimum 1, maximum total number of validators - 1)
    again_propose_interval_block_in_validators_num_percentage = 30
    # When non-proposing nodes discover locally packable transactions but receive a threshold number of null request heartbeats from primary nodes, trigger viewchange to change the primary node
    continuous_null_request_tolerance_number = 3
    # When non-proposing nodes discover locally packable transactions but receive a threshold number of null request heartbeats from primary nodes, replay these transactions
    rebroadcast_tolerance_number = 2

  # Token economic parameter
  [epoch_info.finance_params]
    # Gas limit for transactions
    gas_limit = 100000000
    # Minimum gas price(e.g. 10000000axc; 1000gmol; 1000mol; 1000)
    min_gas_price = '1000gmol'

  # Stake parameter
  [epoch_info.stake_params]
    # Whether to enable the stake module (currently cannot be disabled, this configuration is temporarily ineffective)
    stake_enable = true
    # The maximum ratio of tokens that can be added per epoch (maximum amount of stake added per epoch: max_add_stake_ratio/10000 * total ActiveStake of the previous epoch)
    max_add_stake_ratio = 1000
    # The maximum ratio of tokens that can be unlocked per epoch (maximum amount of stake unlocked per epoch: max_unlock_stake_ratio/10000 * total ActiveStake of the previous epoch)
    max_unlock_stake_ratio = 1000
    # The maximum number of unlocking records that can exist simultaneously for each lst token
    max_unlocking_record_num = 5
    # The lock-up period for unlocking tokens (in seconds, default is 3 days)
    unlock_period = 259200
    # The maximum ratio of validator nodes that can exit per epoch (maximum number of validator nodes exiting per epoch: max_pending_inactive_validator_ratio/10000 * total ActiveValidator count of the current epoch, rounded down)
    max_pending_inactive_validator_ratio = 1000
    # The minimum amount of stake required for delegated staking
    min_delegate_stake = '100axc'
    # The minimum amount of stake required for becoming a validator in the staking pool
    min_validator_stake = '10000000axc'
    # The maximum amount of stake allowed in the staking pool
    max_validator_stake = '50000000axc'

  # Other parameter configurations
  [epoch_info.misc_params]
    # Maximum size of transactions
    tx_max_size = 131072

# Genesis node list (DataSyncer nodes only synchronize blocks and do not participate in consensus)
[[nodes]]
  # Consensus publickey(hex encode)
  consensus_pubkey = '0xac9bb2675ab6b60b1c6d3ed60e95bdabb16517525458d8d50fa1065014184823556b0bd97922fab8c688788006e8b1030cd506d19101522e203769348ea10d21780e5c26a5c03c0cfcb8de23c7cf16d4d384140613bb953d446a26488fbaf6e0'
  # P2P publickey(hex encode)
  p2p_pubkey = '0xd4c0ac1567bcb2c855bb1692c09ab2a2e2c84c45376592674530ce95f1fda351'
  # Operator address 
  operator_address = '0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013'
  # Is DataSyncer flag(is true, stake_number must be zero)
  is_data_syncer = false
  # Initial funds for the node staking pool
  stake_number = '10000000axc'
  # Reward commission rate(0-10000)
  commission_rate = 0

  # Metadata
  [nodes.metadata]
    # Node name (unique, and cannot be empty)
    name = 'node1'
    # Node desc (can be empty)    
    desc = 'node1'
    # Node image url (can be empty)      
    image_url = ''
    # Node website url (can be empty) 
    website_url = ''

```

# config.toml - Basic Configuration
```toml
# Maximum number of handles the node process can open
ulimit = 65535

# Port Configuration (modify to ensure no port conflicts)
[port]
  # Listening port for jsonrpc
  jsonrpc = 8881
  # Listening port for websocket
  websocket = 9991
  # Listening port for p2p
  p2p = 4001
  # Listening port for golang pprof service
  pprof = 53121
  # Listening port for Prometheus metrics service
  monitor = 40011

# Node Configuration
[node]
  # Incentive tx address
  incentive_address = '0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013'

# Gas Price Oracle Configuration (Reference Go Ethereum)
[gas_price_oracle]
  blocks = 20
  percentile = 60
  max_header_history = 1024
  max_block_history = 1024
  default = '500gmol'
  max_price = '10000gmol'
  ignore_price = '2mol'

# JSONRPC Configuration
[jsonrpc]
  # Gas limit for executing eth_call and estimate_gas (prevents DoS attacks)
  gas_cap = 300000000
  # Timeout for executing eth_call and estimate_gas (prevents DoS attacks)
  evm_timeout = '5s'
  # Whether to reject transactions when consensus state is abnormal
  reject_txs_if_consensus_abnormal = false

  # Read request rate limiting configuration (uses token bucket algorithm, applies to all non-sendRawTransaction requests)
  [jsonrpc.read_limiter]
    # Interval for token replenishment
    interval = '50ns'
    # Number of tokens replenished each time
    quantum = 500
    # Token bucket capacity
    capacity = 10000
    # Enable rate limiting
    enable = false
    
  # Write request rate limiting configuration (uses token bucket algorithm, applies to sendRawTransaction requests)
  [jsonrpc.write_limiter]
    # Interval for token replenishment
    interval = '50ns'
    # Number of tokens replenished each time
    quantum = 500
    # Token bucket capacity
    capacity = 10000
    # Enable rate limiting
    enable = false

# P2P Configuration
[p2p]
  # Addresses of P2P bootstrap nodes; multiple nodes can connect indirectly through bootstrap nodes; address format: /ip4/127.0.0.1/tcp/4001/p2p/16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF
  bootstrap_node_addresses = []
  # Message encryption method: tls; noise
  security = 'tls'
  # Timeout for sending messages via p2p stream
  send_timeout = '5s'
  # Timeout for reading messages via p2p stream
  read_timeout = '5s'
  # Compression type (0: None; 1: Snappy; 2: Zstd)
  compression_option = 1
  # Enable metrics flag
  enable_metrics = true

  # Pipe protocol configuration (abstraction for asynchronous communication)
  [p2p.pipe]
    # Size of the message cache (if not consumed in time, subsequently received messages will be discarded)
    receive_msg_cache_size = 10240
    # Timeout for receiver to read messages in unicast
    unicast_read_timeout = '5s'
    # Number of retries for failed message sending in unicast
    unicast_send_retry_number = 5
    # Base interval for retrying failed message sending in unicast; the next retry time will double the previous one
    unicast_send_retry_base_time = '100ms'
    # Timeout for finding peer nodes via DHT
    find_peer_timeout = '10s'
    # Timeout for connecting to peer nodes
    connect_timeout = '1s'

    # Gossip broadcast protocol configuration
    [p2p.pipe.gossipsub]
      # Whether to disable custom message ID function (default is sufficient; disabling may lead to message loss)
      disable_custom_msg_id_fn = false
      # libp2p-pubsub-gossip configuration, buffer size for subscriptions
      sub_buffer_size = 10240
      # libp2p-pubsub-gossip configuration, buffer size for sending messages
      peer_outbound_buffer_size = 10240
      # libp2p-pubsub-gossip configuration, cache size for validators
      validate_buffer_size = 10240
      # libp2p-pubsub-gossip configuration, message ID expiration time (filters duplicate messages)
      seen_messages_ttl = '2m0s'
      # Whether to enable metrics data
      enable_metrics = true

# Block Sync Configuration
[sync]
  # Wait state response timeout
  wait_states_timeout = '30s'
  # Retry interval for block requester
  requester_retry_timeout = '30s'
  # If the number of timeouts for a node's block response exceeds this limit, the node will be removed and no longer requested for blocks
  timeout_count_limit = 10
  # Concurrency limit for block requests
  concurrency_limit = 1000

# Consensus Configuration, detailed configuration in consensus.toml
[consensus]
  # Consensus type: rbft (at least four nodes); solo (single node mode)
  type = 'rbft'
  # Database type for consensus: minifile
  storage_type = 'minifile'
  # Use bls consensus key as a signer key. If false, will use p2p key
  use_bls_key = false

# Storage
[storage]
  # Storage type: leveldb; pebble
  kv_type = 'pebble'
  # Cache size for pebble, in MB
  kv_cache_size = 128
  # Enable pebble sync option (real-time flushing is not enabled, data may be lost if the process is killed)
  sync = true

# Ledger Configuration
[ledger]
  # LRU cache size for ledger state
  chain_ledger_cache_size = 100
  # Cache size limit for state ledger account trie cache (in megabytes); larger values improve performance but increase memory usage
  state_ledger_account_trie_cache_megabytes_limit = 128
  # Cache size limit for state ledger storage trie cache (in megabytes); larger values improve performance but increase memory usage
  state_ledger_storage_trie_cache_megabytes_limit = 128
  # Cache size for account information in state ledger (number of accounts); caching account nonce, balance, code; larger values improve performance but increase memory usage
  state_ledger_account_cache_size = 1024
  # Enable prune
  enable_prune = true
  # If enable prue, state ledger reserved history block num
  state_ledger_reserved_history_block_num = 256

[snapshot]
  # Cache size limit for account snapshot (in megabytes); larger values improve performance but increase memory usage
  account_snapshot_cache_megabytes_limit = 128
  # Cache size limit for contract snapshot (in megabytes); larger values improve performance but increase memory usage
  contract_snapshot_cache_megabytes_limit = 128

# Executor Configuration
[executor]
  # Type: native (native mode); dev (development mode)
  type = 'native'
  # Whether to disable rollback functionality (when detecting that the height of at least quorum other nodes is higher than the local node, the node will roll back; if disabled, it will panic actively)
  disable_rollback = true

# Golang pprof Configuration
[pprof]
  # Whether to enable; when enabled, it will listen on the pprof port
  enbale = true
  ptype = 'http'
  mode = 'mem'
  duration = '30s'

# Metrics Service Configuration
[monitor]
  # Whether to enable; when enabled, it will listen on the monitor port
  enable = true
  # Whether to enable metrics for expensive operations (enabling will reduce performance)
  enable_expensive = false

# Log Level
[log]
  # Global log level
  level = 'info'
  # Log file name
  filename = 'axiom-ledger'
  # Whether to include log caller information
  report_caller = false
  # Whether to enable log compression
  enable_compress = false
  # Whether to use color in logs
  enable_color = true
  # Disable timestamp in logs
  disable_timestamp = false
  # Maximum retention days for logs
  max_age = 30
  # Maximum size for logs (in MB)
  max_size = 128
  # Log rotation interval
  rotation_time = '24h0m0s'

  # Log levels for different modules
  [log.module]
    p2p = 'info'
    consensus = 'debug'
    executor = 'info'
    governance = 'info'
    api = 'error'
    app = 'info'
    coreapi = 'info'
    storage = 'info'
    profile = 'info'
    finance = 'error'
    txpool = 'warn'
    access = 'info'
    blocksync = 'info'
    epoch = 'info'

# Access Control Configuration
[access]
  # Whether to enable whitelist
  enable_white_list = false
```

# consensus.toml - Basic Configuration
```toml
# Timed Block Generation Configuration
[timed_gen_block]
  # Block generation interval
  no_tx_batch_timeout = '2s'

# Flow Control Limit for P2P Transaction Broadcasting Recipients (Token Bucket)
[limit]
  # Enable or disable
  enable = false
  # Maximum number of tokens in the token bucket
  limit = 10000
  # Number of tokens restored per second
  burst = 10000

# Transaction Pool Configuration
[tx_pool]
  # Size of the transaction pool (stops accepting transactions after reaching the limit)
  pool_size = 50000
  # Interval for replaying transactions that have not been included in a block
  tolerance_time = '5m0s'
  # Time for removing transactions that have not been included in a block for a long time (after this duration, transactions will be deleted)
  tolerance_remove_time = '15m0s'
  # Time for cleaning empty accounts (has no txs) 
  clean_empty_account_time = '10m0s'
  # Maximum number of high-nonce transactions allowed for the same account
  tolerance_nonce_gap = 1000
  # Enable local persist (If enabled, no transactions will be lost on reboot)
  enable_locals_persist = true
  # Persist txs to local file interval
  rotate_tx_locals_interval = '1h0m0s'
  # TX min gas price
  price_limit = '1000gmol'
  # The higher gas price increase ratio required when the transaction is replaced
  price_bump = 10
  # Generate a batch type (fifo; price_priority)
  generate_batch_type = 'fifo'

# Transaction Cache Configuration (Responsible for Transaction Broadcasting)
[tx_cache]
  # Number of transactions broadcasted each time
  set_size = 50
  # Broadcasting interval
  set_timeout = '100ms'

# RBFT Configuration
[rbft]
  # Whether to enable metrics
  enable_metrics = true
  # Number of committed blocks cached
  committed_block_cache_number = 10

# Timeout Configuration
[rbft.timeout]
  null_request = '3s'
  request = '2s'
  resend_viewchange = '10s'
  clean_viewchange = '1m0s'
  new_view = '8s'
  sync_state = '1s'
  sync_state_restart = '10m0s'
  fetch_checkpoint = '5s'
  fetch_view = '1s'
  batch_timeout = '500ms'

# Solo Configuration
[solo]
  # Checkpoint interval
  checkpoint_period = 10
```