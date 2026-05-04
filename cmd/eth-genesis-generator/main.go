package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/islishude/eth-genesis-generator/internal/artifacts"
	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches CLI commands and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "generate":
		return runGenerate(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		_, _ = fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

// runInit writes a starter genesis.yaml without overwriting an existing file.
func runInit(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", "./devnet", "directory to write genesis.yaml into")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		_, _ = fmt.Fprintf(stderr, "create output directory: %v\n", err)
		return 1
	}

	configPath := filepath.Join(*out, "genesis.yaml")
	if _, err := os.Stat(configPath); err == nil {
		_, _ = fmt.Fprintf(stderr, "%s already exists\n", configPath)
		return 1
	} else if !os.IsNotExist(err) {
		_, _ = fmt.Fprintf(stderr, "stat %s: %v\n", configPath, err)
		return 1
	}

	if err := os.WriteFile(configPath, []byte(appconfig.ExampleYAML()), 0o644); err != nil {
		_, _ = fmt.Fprintf(stderr, "write %s: %v\n", configPath, err)
		return 1
	}

	_, _ = fmt.Fprintf(stdout, "wrote %s\n", configPath)
	return 0
}

// runGenerate loads a config file and writes all genesis artifacts.
func runGenerate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "path to genesis.yaml")
	out := fs.String("out", "./artifacts", "directory to write generated artifacts into")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *configPath == "" {
		_, _ = fmt.Fprintln(stderr, "missing required --config")
		return 2
	}

	cfg, err := appconfig.LoadFile(*configPath, time.Now())
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "load config: %v\n", err)
		return 1
	}

	manifest, err := artifacts.Generate(cfg, *out, time.Now())
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "generate artifacts: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintf(stdout, "wrote artifacts to %s\n", *out)
	_, _ = fmt.Fprintf(stdout, "execution genesis hash: %s\n", manifest.ExecutionGenesisHash)
	_, _ = fmt.Fprintf(stdout, "consensus state version: %s\n", manifest.StateVersion)
	_, _ = fmt.Fprintf(stdout, "validators: %d\n", manifest.ValidatorCount)
	return 0
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprint(w, `eth-genesis-generator generates local Ethereum PoS devnet genesis artifacts.

Usage:
  eth-genesis-generator init --out ./devnet
  eth-genesis-generator generate --config ./devnet/genesis.yaml --out ./artifacts
`)
}
