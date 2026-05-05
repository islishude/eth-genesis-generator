package consensus

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethpandaops/eth-beacon-genesis/beaconchain"
	"github.com/ethpandaops/eth-beacon-genesis/beaconconfig"
	"github.com/ethpandaops/eth-beacon-genesis/validators"
	eth2http "github.com/ethpandaops/go-eth2-client/http"
	"github.com/ethpandaops/go-eth2-client/spec"
)

// GenesisResult contains the beacon genesis state and metadata derived from it.
type GenesisResult struct {
	// State is kept for callers that need to inspect the in-memory beacon state.
	State *spec.VersionedBeaconState
	// SSZ is the serialized consensus/genesis.ssz payload.
	SSZ []byte
	// JSON is the optional serialized consensus/genesis.json payload.
	JSON []byte
	// Validators is the number of validators present in the final state.
	Validators int
	// StateVersion is the active fork version name reported by the generated state.
	StateVersion string
	// ETH1Hash is the execution genesis block hash embedded in eth1_data.
	ETH1Hash common.Hash
}

// BuildGenesis creates a beacon genesis state from execution genesis and CL input files.
func BuildGenesis(elGenesis *core.Genesis, clConfigPath string, mnemonicsPath string, outputJSON bool) (*GenesisResult, error) {
	clConfig, err := beaconconfig.LoadConfig(clConfigPath)
	if err != nil {
		return nil, fmt.Errorf("load consensus config: %w", err)
	}

	genesisValidators, err := validators.GenerateValidatorsByMnemonic(mnemonicsPath)
	if err != nil {
		return nil, fmt.Errorf("generate validators: %w", err)
	}

	builder := beaconchain.NewGenesisBuilder(elGenesis, clConfig)
	if builder == nil {
		return nil, fmt.Errorf("no beacon genesis builder for configured fork")
	}

	builder.AddValidators(genesisValidators)
	state, err := builder.BuildState()
	if err != nil {
		return nil, fmt.Errorf("build beacon state: %w", err)
	}

	sszBytes, err := builder.Serialize(state, eth2http.ContentTypeSSZ)
	if err != nil {
		return nil, fmt.Errorf("serialize beacon state as ssz: %w", err)
	}

	var jsonBytes []byte
	if outputJSON {
		jsonBytes, err = builder.Serialize(state, eth2http.ContentTypeJSON)
		if err != nil {
			return nil, fmt.Errorf("serialize beacon state as json: %w", err)
		}
	}

	// Keep the EL and CL artifacts tied to the same genesis block before writing output.
	eth1Hash, err := ETH1BlockHash(state)
	if err != nil {
		return nil, err
	}

	expected := elGenesis.ToBlock().Hash()
	if !bytes.Equal(eth1Hash.Bytes(), expected.Bytes()) {
		return nil, fmt.Errorf("consensus eth1_data.block_hash %s does not match execution genesis hash %s", eth1Hash, expected)
	}

	stateValidators, err := state.Validators()
	if err != nil {
		return nil, err
	}

	return &GenesisResult{
		State:        state,
		SSZ:          sszBytes,
		JSON:         jsonBytes,
		Validators:   len(stateValidators),
		StateVersion: state.Version.String(),
		ETH1Hash:     eth1Hash,
	}, nil
}

// ETH1BlockHash extracts eth1_data.block_hash across supported beacon state versions.
func ETH1BlockHash(state *spec.VersionedBeaconState) (common.Hash, error) {
	switch state.Version {
	case spec.DataVersionFulu:
		return common.BytesToHash(state.Fulu.ETH1Data.BlockHash), nil
	case spec.DataVersionElectra:
		return common.BytesToHash(state.Electra.ETH1Data.BlockHash), nil
	case spec.DataVersionDeneb:
		return common.BytesToHash(state.Deneb.ETH1Data.BlockHash), nil
	case spec.DataVersionCapella:
		return common.BytesToHash(state.Capella.ETH1Data.BlockHash), nil
	case spec.DataVersionBellatrix:
		return common.BytesToHash(state.Bellatrix.ETH1Data.BlockHash), nil
	case spec.DataVersionAltair:
		return common.BytesToHash(state.Altair.ETH1Data.BlockHash), nil
	case spec.DataVersionPhase0:
		return common.BytesToHash(state.Phase0.ETH1Data.BlockHash), nil
	default:
		return common.Hash{}, fmt.Errorf("unsupported beacon state version %s", state.Version)
	}
}
