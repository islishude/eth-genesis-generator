package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/islishude/eth-genesis-generator/internal/artifacts"
	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
	"github.com/urfave/cli/v2"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the CLI and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	app := newApp(stdout, stderr)
	if err := app.Run(append([]string{app.Name}, args...)); err != nil {
		if msg := err.Error(); msg != "" {
			_, _ = fmt.Fprintln(stderr, msg)
		}
		if exitErr, ok := err.(interface{ ExitCode() int }); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}

func newApp(stdout, stderr io.Writer) *cli.App {
	return &cli.App{
		Name:                 "eth-genesis-generator",
		Usage:                "Generate local Ethereum PoS devnet genesis artifacts",
		Writer:               stdout,
		ErrWriter:            stderr,
		HideHelpCommand:      true,
		EnableBashCompletion: false,
		Action: func(ctx *cli.Context) error {
			if err := cli.ShowAppHelp(ctx); err != nil {
				return err
			}
			return cli.Exit("", 2)
		},
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Write a configurable genesis.yaml template",
				Flags: initFlags(),
				Action: func(ctx *cli.Context) error {
					return runInit(ctx, stdout)
				},
			},
			{
				Name:  "generate",
				Usage: "Generate all genesis artifacts from genesis.yaml",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Usage:    "path to genesis.yaml",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "out",
						Usage: "directory to write generated artifacts into",
						Value: "./artifacts",
					},
				},
				Action: func(ctx *cli.Context) error {
					return runGenerate(ctx, stdout)
				},
			},
		},
	}
}

func initFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "out",
			Usage: "directory to write genesis.yaml into",
			Value: "./devnet",
		},
		&cli.StringFlag{
			Name:  "network-name",
			Usage: "network name to write into genesis.yaml",
			Value: appconfig.DefaultNetworkName,
		},
		&cli.Uint64Flag{
			Name:  "chain-id",
			Usage: "execution and deposit chain ID",
			Value: appconfig.DefaultChainID,
		},
		&cli.Uint64Flag{
			Name:  "genesis-time",
			Usage: "Unix genesis timestamp; 0 means generate uses current time + 60 seconds",
			Value: 0,
		},
		&cli.Uint64Flag{
			Name:  "gas-limit",
			Usage: "execution genesis gas limit",
			Value: appconfig.DefaultGasLimit,
		},
		&cli.StringFlag{
			Name:  "base-fee-per-gas",
			Usage: "execution genesis base fee per gas as a decimal string",
			Value: appconfig.DefaultBaseFeePerGas,
		},
		&cli.StringFlag{
			Name:  "execution-contracts",
			Usage: fmt.Sprintf("execution contract predeploy profiles; comma-separated supported profiles: %s", appconfig.SupportedExecutionContractProfiles()),
		},
		&cli.StringSliceFlag{
			Name:  "prefund",
			Usage: "prefunded execution account as ADDRESS=AMOUNT; repeatable",
		},
		&cli.StringFlag{
			Name:  "fork",
			Usage: fmt.Sprintf("consensus fork activated at genesis; supported: %s", joinForks()),
			Value: appconfig.DefaultFork,
		},
		&cli.StringFlag{
			Name:  "preset-base",
			Usage: "consensus preset base",
			Value: appconfig.DefaultPresetBase,
		},
		&cli.Uint64Flag{
			Name:  "validator-count",
			Usage: "number of pre-filled validators",
			Value: appconfig.DefaultValidatorCount,
		},
		&cli.Uint64Flag{
			Name:  "validator-balance-gwei",
			Usage: "initial validator balance in gwei",
			Value: appconfig.DefaultValidatorBalanceGwei,
		},
		&cli.StringFlag{
			Name:  "mnemonic",
			Usage: "validator mnemonic; empty means generate a new 24-word mnemonic during generate",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "withdrawal-address",
			Usage: "validator withdrawal Ethereum address",
			Value: appconfig.DefaultWithdrawalAddress,
		},
		&cli.StringFlag{
			Name:  "withdrawal-prefix",
			Usage: "one-byte 0x-prefixed withdrawal credential prefix",
			Value: appconfig.DefaultWithdrawalPrefix,
		},
		&cli.StringFlag{
			Name:  "deposit-contract-address",
			Usage: "deposit contract Ethereum address",
			Value: appconfig.DefaultDepositContractAddress,
		},
		&cli.BoolFlag{
			Name:  "output-json",
			Usage: "write consensus/genesis.json during generate",
			Value: true,
		},
		&cli.BoolFlag{
			Name:  "force",
			Usage: "overwrite an existing genesis.yaml",
			Value: false,
		},
	}
}

// runInit writes a starter genesis.yaml without generating artifacts.
func runInit(ctx *cli.Context, stdout io.Writer) error {
	cfg := appconfig.DefaultInitConfig()
	cfg.Network.Name = ctx.String("network-name")
	cfg.Network.ChainID = ctx.Uint64("chain-id")
	cfg.Network.GenesisTime = ctx.Uint64("genesis-time")
	cfg.Execution.GasLimit = ctx.Uint64("gas-limit")
	cfg.Execution.BaseFeePerGas = ctx.String("base-fee-per-gas")
	cfg.Consensus.Fork = ctx.String("fork")
	cfg.Consensus.PresetBase = ctx.String("preset-base")
	cfg.Consensus.ValidatorCount = ctx.Uint64("validator-count")
	cfg.Consensus.ValidatorBalanceGwei = ctx.Uint64("validator-balance-gwei")
	cfg.Consensus.Mnemonic = ctx.String("mnemonic")
	cfg.Consensus.WithdrawalAddress = ctx.String("withdrawal-address")
	cfg.Consensus.WithdrawalPrefix = ctx.String("withdrawal-prefix")
	cfg.Consensus.DepositContractAddress = ctx.String("deposit-contract-address")
	cfg.Consensus.OutputJSON = new(ctx.Bool("output-json"))

	if ctx.IsSet("execution-contracts") {
		contracts, err := appconfig.ParseExecutionContractProfiles(ctx.String("execution-contracts"))
		if err != nil {
			return err
		}
		cfg.Execution.Contracts = contracts
	}
	if ctx.IsSet("prefund") {
		prefund, err := appconfig.ParsePrefundEntries(ctx.StringSlice("prefund"))
		if err != nil {
			return err
		}
		cfg.Execution.Prefund = prefund
	}
	if err := cfg.ValidateInit(); err != nil {
		return err
	}

	out := ctx.String("out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	configPath := filepath.Join(out, "genesis.yaml")
	if !ctx.Bool("force") {
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("%s already exists; use --force to overwrite", configPath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s: %w", configPath, err)
		}
	}

	if err := os.WriteFile(configPath, []byte(appconfig.RenderInitYAML(cfg)), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", configPath, err)
	}

	_, _ = fmt.Fprintf(stdout, "wrote %s\n", configPath)
	return nil
}

// runGenerate loads a config file and writes all genesis artifacts.
func runGenerate(ctx *cli.Context, stdout io.Writer) error {
	out := ctx.String("out")
	cfg, err := appconfig.LoadFile(ctx.String("config"), time.Now())
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	manifest, err := artifacts.Generate(cfg, out, time.Now())
	if err != nil {
		return fmt.Errorf("generate artifacts: %w", err)
	}

	_, _ = fmt.Fprintf(stdout, "wrote artifacts to %s\n", out)
	_, _ = fmt.Fprintf(stdout, "execution genesis hash: %s\n", manifest.ExecutionGenesisHash)
	_, _ = fmt.Fprintf(stdout, "consensus state version: %s\n", manifest.StateVersion)
	_, _ = fmt.Fprintf(stdout, "validators: %d\n", manifest.ValidatorCount)
	_, _ = fmt.Fprintf(stdout, "validator keystores: %d\n", manifest.ValidatorKeystoreCount)
	_, _ = fmt.Fprintf(stdout, "validator keystore password: %s\n", filepath.Join(out, "validators", "keystores", "password.txt"))
	return nil
}

func joinForks() string {
	forks := appconfig.SupportedForks()
	if len(forks) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(forks[0])
	for _, fork := range forks[1:] {
		out.WriteString(", " + fork)
	}
	return out.String()
}
