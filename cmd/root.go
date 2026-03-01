package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/the20100/fal-cli/internal/api"
	"github.com/the20100/fal-cli/internal/config"
)

var (
	// Persistent flags
	jsonFlag   bool
	prettyFlag bool

	// Global API client, set in PersistentPreRunE
	client *api.Client

	// Global config, set in PersistentPreRunE
	cfg *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "fal",
	Short: "fal.ai CLI — run generative AI models via fal.ai",
	Long: `fal is a CLI tool for the fal.ai API.

It outputs JSON when piped (for agent use) and human-readable tables in a terminal.

Token resolution order:
  1. FAL_KEY env var (or aliases: FAL_API_KEY, FAL_API, API_KEY_FAL, ...)
  2. Own config  (~/.config/fal/config.json  via: fal auth set-key)

Examples:
  fal auth set-key
  fal models list --category text-to-image
  fal models pricing fal-ai/nano-banana-pro
  fal run fal-ai/nano-banana-pro --input '{"prompt":"a cat"}'
  fal queue submit fal-ai/flux/dev --input '{"prompt":"a cat"}'
  fal queue status fal-ai/flux/dev <request_id>
  fal generate "a cat wearing a hat"
  fal edit "make it night time" --image https://example.com/photo.jpg`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Force JSON output")
	rootCmd.PersistentFlags().BoolVar(&prettyFlag, "pretty", false, "Force pretty-printed JSON output (implies --json)")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if isAuthCommand(cmd) || cmd.Name() == "info" {
			return nil
		}

		key, err := resolveAPIKey()
		if err != nil {
			return err
		}

		client = api.NewClient(key)
		return nil
	}

	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show tool info: config path, key status, and environment",
	Run: func(cmd *cobra.Command, args []string) {
		printInfo()
	},
}

func printInfo() {
	fmt.Println("fal — fal.ai CLI")
	fmt.Println()

	exe, _ := os.Executable()
	fmt.Printf("  binary:  %s\n", exe)
	fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	fmt.Println("  config paths by OS:")
	fmt.Println("    macOS:    ~/Library/Application Support/fal/config.json")
	fmt.Println("    Linux:    ~/.config/fal/config.json")
	fmt.Println("    Windows:  %AppData%\\fal\\config.json")
	fmt.Printf("  config:   %s\n", config.Path())
	fmt.Println()

	keySource := "(not set)"
	if t := resolveEnv(
		"FAL_KEY", "FAL_API_KEY", "FAL_API", "API_KEY_FAL", "API_FAL", "FAL_PK", "FAL_PUBLIC",
		"FAL_API_SECRET", "FAL_SECRET_KEY", "FAL_API_SECRET_KEY", "FAL_SECRET", "SECRET_FAL", "API_SECRET_FAL", "SK_FAL", "FAL_SK",
	); t != "" {
		keySource = "FAL_KEY env var (or alias)"
	} else if c, err := config.Load(); err == nil && c.APIKey != "" {
		keySource = "config file"
	}
	fmt.Printf("  key source: %s\n", keySource)
	fmt.Println()
	fmt.Println("  env vars:")
	fmt.Printf("    FAL_KEY = %s  (also accepts aliases: FAL_API_KEY, FAL_API, ...)\n", maskOrEmpty(os.Getenv("FAL_KEY")))
	fmt.Println()
	fmt.Println("  key resolution order:")
	fmt.Println("    1. FAL_KEY env var (or aliases)")
	fmt.Println("    2. config file  (fal auth set-key)")
}

func maskOrEmpty(v string) string {
	if v == "" {
		return "(not set)"
	}
	if len(v) <= 8 {
		return "***"
	}
	return v[:4] + "..." + v[len(v)-4:]
}

// resolveEnv returns the value of the first non-empty environment variable from the given names.
func resolveEnv(names ...string) string {
	for _, name := range names {
		if v := os.Getenv(name); v != "" {
			return v
		}
	}
	return ""
}

// resolveAPIKey returns the best available API key.
func resolveAPIKey() (string, error) {
	// 1. Env var aliases (key and secret variants)
	if k := resolveEnv(
		"FAL_KEY", "FAL_API_KEY", "FAL_API", "API_KEY_FAL", "API_FAL", "FAL_PK", "FAL_PUBLIC",
		"FAL_API_SECRET", "FAL_SECRET_KEY", "FAL_API_SECRET_KEY", "FAL_SECRET", "SECRET_FAL", "API_SECRET_FAL", "SK_FAL", "FAL_SK",
	); k != "" {
		return k, nil
	}

	// 2. Config file
	var err error
	cfg, err = config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.APIKey != "" {
		return cfg.APIKey, nil
	}

	return "", fmt.Errorf("not authenticated — run: fal auth set-key\nor set FAL_KEY env var")
}

// isAuthCommand returns true if cmd is a child of the "auth" command.
func isAuthCommand(cmd *cobra.Command) bool {
	if cmd.Name() == "auth" {
		return true
	}
	p := cmd.Parent()
	for p != nil {
		if p.Name() == "auth" {
			return true
		}
		p = p.Parent()
	}
	return false
}
