package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/the20100/fal-cli/internal/config"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage fal.ai authentication",
}

var authSetKeyCmd = &cobra.Command{
	Use:   "set-key <api-key>",
	Short: "Save a fal.ai API key to the config file",
	Long: `Save a fal.ai API key to the local config file.

Get your API key from: https://fal.ai/dashboard/keys

The key is stored at:
  macOS:   ~/Library/Application Support/fal/config.json
  Linux:   ~/.config/fal/config.json
  Windows: %AppData%\fal\config.json

You can also set the FAL_KEY env var instead of using this command.`,
	Args:    cobra.ExactArgs(1),
	RunE:    runAuthSetKey,
	Example: "  fal auth set-key your_api_key_here",
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE:  runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the saved API key from the config file",
	RunE:  runAuthLogout,
}

func init() {
	authCmd.AddCommand(authSetKeyCmd, authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuthSetKey(cmd *cobra.Command, args []string) error {
	key := args[0]
	if len(key) < 8 {
		return fmt.Errorf("API key looks too short — check your key at https://fal.ai/dashboard/keys")
	}

	c := &config.Config{APIKey: key}
	if err := config.Save(c); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("API key saved to %s\n", config.Path())
	fmt.Printf("Key: %s\n", maskOrEmpty(key))
	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	c, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Printf("Config: %s\n", config.Path())
	fmt.Println()

	if envKey := os.Getenv("FAL_KEY"); envKey != "" {
		fmt.Println("Key source: FAL_KEY env var (takes priority over config)")
		fmt.Printf("Key:        %s\n", maskOrEmpty(envKey))
	} else if c.APIKey != "" {
		fmt.Println("Key source: config file")
		fmt.Printf("Key:        %s\n", maskOrEmpty(c.APIKey))
	} else {
		fmt.Println("Status: not authenticated")
		fmt.Println()
		fmt.Println("Run: fal auth set-key <your-api-key>")
		fmt.Println("Or:  export FAL_KEY=<your-api-key>")
	}
	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	if err := config.Clear(); err != nil {
		return fmt.Errorf("removing config: %w", err)
	}
	fmt.Println("API key removed from config.")
	fmt.Println("Set FAL_KEY env var if you still need access.")
	return nil
}
