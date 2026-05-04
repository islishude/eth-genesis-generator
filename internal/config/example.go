package config

// ExampleYAML returns the starter genesis.yaml written by the init command.
func ExampleYAML() string {
	return `network:
  name: devnet
  chain_id: 32382
  # Leave empty or set 0 to use current time + 60 seconds.
  genesis_time: 0

execution:
  gas_limit: 36000000
  extra_data: "0x"
  base_fee_per_gas: "1000000000"
  prefund:
    "0x1000000000000000000000000000000000000001": "1000000000000000000000000000"

consensus:
  fork: fulu
  preset_base: mainnet
  validator_count: 64
  validator_balance_gwei: 32000000000
  # Leave empty to generate a new 24-word mnemonic.
  mnemonic: ""
  withdrawal_address: "0x1000000000000000000000000000000000000001"
  withdrawal_prefix: "0x02"
  deposit_contract_address: "0x4242424242424242424242424242424242424242"
  output_json: true
  # Optional 4-byte fork version overrides.
  # fork_versions:
  #   genesis: "0x00007e7e"
  #   altair: "0x01007e7e"
  #   bellatrix: "0x02007e7e"
  #   capella: "0x03007e7e"
  #   deneb: "0x04007e7e"
  #   electra: "0x05007e7e"
  #   fulu: "0x06007e7e"
`
}
