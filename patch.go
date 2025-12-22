package main

import (
	"math"
	"time"

	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/time/slots"
	"github.com/ethereum/go-ethereum/core"
)

func SetForkTimes(genesisTime time.Time, elConfig *core.Genesis, cfg *params.BeaconChainConfig) {
	elConfig.Timestamp = uint64(genesisTime.Unix())

	// Shanghai fork time
	if cfg.CapellaForkEpoch != math.MaxUint64 {
		startSlot, err := slots.EpochStart(cfg.CapellaForkEpoch)
		if err == nil {
			startTime := slots.UnsafeStartTime(genesisTime, startSlot)
			newTime := uint64(startTime.Unix())
			elConfig.Config.ShanghaiTime = &newTime
		}
	}

	// Cancun fork time
	if cfg.DenebForkEpoch != math.MaxUint64 {
		startSlot, err := slots.EpochStart(cfg.DenebForkEpoch)
		if err == nil {
			startTime := slots.UnsafeStartTime(genesisTime, startSlot)
			newTime := uint64(startTime.Unix())
			elConfig.Config.CancunTime = &newTime
		}
	}

	// Prague fork time
	if cfg.ElectraForkEpoch != math.MaxUint64 {
		startSlot, err := slots.EpochStart(cfg.ElectraForkEpoch)
		if err == nil {
			startTime := slots.UnsafeStartTime(genesisTime, startSlot)
			newTime := uint64(startTime.Unix())
			elConfig.Config.PragueTime = &newTime
		}
	}

	// Osaka fork time
	if cfg.FuluForkEpoch != math.MaxUint64 {
		startSlot, err := slots.EpochStart(cfg.FuluForkEpoch)
		if err == nil {
			startTime := slots.UnsafeStartTime(genesisTime, startSlot)
			newTime := uint64(startTime.Unix())
			elConfig.Config.OsakaTime = &newTime
		}
	}
}
