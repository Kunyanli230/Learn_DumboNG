core加入merkle tree模块

CLI用法
```
# Create a new wallet
go run ./009-MerkleTree createwallet

# List all wallet addresses
go run ./009-MerkleTree addresslists

# Create the genesis blockchain with a reward address
go run ./009-MerkleTree createblockchain -address <wallet address>

# Check the balance of an address
go run ./009-MerkleTree getbalance -address <wallet address>

# Send one transaction
go run ./009-MerkleTree send -from '["<from address>"]' -to '["<to address>"]' -amount '["10"]'

# Send multiple transactions in one command
go run ./009-MerkleTree send -from '["<from1>","<from2>"]' -to '["<to1>","<to2>"]' -amount '["10","20"]'

# Print all blocks in the local blockchain
go run ./009-MerkleTree printchain
```