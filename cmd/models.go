package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vincentmaurin/fal-cli/internal/output"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Browse and search the fal.ai model catalog",
}

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List models from the fal.ai catalog",
	Long: `List models from the fal.ai catalog with optional filters.

Examples:
  fal models list
  fal models list --category text-to-image
  fal models list --search "flux"
  fal models list --category image-to-video --limit 10`,
	RunE: runModelsList,
}

var modelsPricingCmd = &cobra.Command{
	Use:   "pricing <model-id> [model-id...]",
	Short: "Show pricing for one or more models",
	Long: `Show pricing for one or more fal.ai model endpoints.

Examples:
  fal models pricing fal-ai/nano-banana-pro
  fal models pricing fal-ai/flux/dev fal-ai/flux/schnell
  fal models pricing fal-ai/nano-banana-pro fal-ai/nano-banana-pro/edit`,
	Args: cobra.MinimumNArgs(1),
	RunE: runModelsPricing,
}

var (
	modelsSearchFlag   string
	modelsCategoryFlag string
	modelsLimitFlag    int
)

func init() {
	modelsListCmd.Flags().StringVar(&modelsSearchFlag, "search", "", "Free-text search query")
	modelsListCmd.Flags().StringVar(&modelsCategoryFlag, "category", "", "Filter by category (e.g. text-to-image, image-to-video)")
	modelsListCmd.Flags().IntVar(&modelsLimitFlag, "limit", 20, "Max number of models to return")

	modelsCmd.AddCommand(modelsListCmd, modelsPricingCmd)
	rootCmd.AddCommand(modelsCmd)
}

func runModelsList(cmd *cobra.Command, args []string) error {
	resp, err := client.ListModels(modelsSearchFlag, modelsCategoryFlag, "", modelsLimitFlag)
	if err != nil {
		return err
	}

	if output.IsJSON(cmd) {
		return output.PrintJSON(resp.Models, output.IsPretty(cmd))
	}

	if len(resp.Models) == 0 {
		fmt.Println("No models found.")
		return nil
	}

	headers := []string{"ENDPOINT ID", "NAME", "CATEGORY", "STATUS"}
	rows := make([][]string, len(resp.Models))
	for i, m := range resp.Models {
		rows[i] = []string{
			m.EndpointID,
			output.Truncate(m.Metadata.DisplayName, 35),
			m.Metadata.Category,
			m.Metadata.Status,
		}
	}
	output.PrintTable(headers, rows)

	if resp.HasMore {
		fmt.Printf("\n(%d shown, more available — increase --limit to see more)\n", len(resp.Models))
	}
	return nil
}

func runModelsPricing(cmd *cobra.Command, args []string) error {
	resp, err := client.GetModelPricing(args)
	if err != nil {
		return err
	}

	if output.IsJSON(cmd) {
		return output.PrintJSON(resp.Prices, output.IsPretty(cmd))
	}

	if len(resp.Prices) == 0 {
		fmt.Println("No pricing info found for the given model(s).")
		return nil
	}

	headers := []string{"ENDPOINT ID", "PRICE", "UNIT", "CURRENCY"}
	rows := make([][]string, len(resp.Prices))
	for i, p := range resp.Prices {
		rows[i] = []string{
			p.EndpointID,
			fmt.Sprintf("%.4f", p.UnitPrice),
			p.Unit,
			strings.ToUpper(p.Currency),
		}
	}
	output.PrintTable(headers, rows)
	return nil
}
