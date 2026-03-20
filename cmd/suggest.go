package cmd

import (
	"fmt"
	"strings"

	"github.com/osir/cli/internal/api/models"
	"github.com/osir/cli/internal/output"
	"github.com/spf13/cobra"
)

func addSuggestCommands(parent *cobra.Command) {
	suggestCmd := &cobra.Command{
		Use:   "suggest",
		Short: "Domain name suggestion tools",
		Long:  "Generate domain name suggestions using various strategies: AI generation, word spinning, prefixes, suffixes, bulk suggestions, and keyword availability checks.",
	}

	suggestGenerateCmd := &cobra.Command{
		Use:   "generate <name>",
		Short: "Generate domain name suggestions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			tlds, _ := cmd.Flags().GetString("tlds")
			lang, _ := cmd.Flags().GetString("lang")
			numbers, _ := cmd.Flags().GetBool("numbers")
			max, _ := cmd.Flags().GetInt("max")

			result, err := app.Client.GenerateSuggestions(ctx, args[0], tlds, lang, numbers, max)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to generate suggestions: %s", err))
				return err
			}

			printSuggestions(app.Output, result)
			return nil
		},
	}

	suggestSpinCmd := &cobra.Command{
		Use:   "spin <name>",
		Short: "Generate suggestions by spinning/replacing words",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			position, _ := cmd.Flags().GetInt("position")
			similarity, _ := cmd.Flags().GetFloat64("similarity")
			tlds, _ := cmd.Flags().GetString("tlds")
			lang, _ := cmd.Flags().GetString("lang")
			max, _ := cmd.Flags().GetInt("max")

			result, err := app.Client.SpinWords(ctx, args[0], position, similarity, tlds, lang, max)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to spin words: %s", err))
				return err
			}

			printSuggestions(app.Output, result)
			return nil
		},
	}

	suggestPrefixCmd := &cobra.Command{
		Use:   "prefix <name>",
		Short: "Generate suggestions by adding prefixes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			vocabulary, _ := cmd.Flags().GetString("vocabulary")
			tlds, _ := cmd.Flags().GetString("tlds")
			lang, _ := cmd.Flags().GetString("lang")
			max, _ := cmd.Flags().GetInt("max")

			result, err := app.Client.AddPrefix(ctx, args[0], vocabulary, tlds, lang, max)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to add prefix: %s", err))
				return err
			}

			printSuggestions(app.Output, result)
			return nil
		},
	}

	suggestSuffixCmd := &cobra.Command{
		Use:   "suffix <name>",
		Short: "Generate suggestions by adding suffixes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			vocabulary, _ := cmd.Flags().GetString("vocabulary")
			tlds, _ := cmd.Flags().GetString("tlds")
			lang, _ := cmd.Flags().GetString("lang")
			max, _ := cmd.Flags().GetInt("max")

			result, err := app.Client.AddSuffix(ctx, args[0], vocabulary, tlds, lang, max)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			printSuggestions(app.Output, result)
			return nil
		},
	}

	suggestBulkCmd := &cobra.Command{
		Use:   "bulk <keywords...>",
		Short: "Generate bulk domain suggestions for multiple keywords",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			tlds, _ := cmd.Flags().GetString("tlds")
			lang, _ := cmd.Flags().GetString("lang")
			numbers, _ := cmd.Flags().GetBool("numbers")
			max, _ := cmd.Flags().GetInt("max")

			req := models.BulkSuggestRequest{
				Names:             args,
				TLDs:              tlds,
				Lang:              lang,
				UseNumbers:        numbers,
				MaxResultsPerName: max,
			}

			result, err := app.Client.BulkSuggest(ctx, req)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to get bulk suggestions: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				if len(result.Errors) > 0 {
					for _, e := range result.Errors {
						app.Output.PrintError(fmt.Sprintf("Error for '%s': %s", e.OriginalName, e.ErrorMessage))
					}
					app.Output.Println("")
				}

				if len(result.Suggestions) == 0 {
					app.Output.Println("No suggestions found")
					return nil
				}

				for _, group := range result.Suggestions {
					app.Output.Println(fmt.Sprintf("── %s ──", group.OriginalName))
					if len(group.Suggestions) == 0 {
						app.Output.Println("  No suggestions")
					} else {
						headers := []string{"DOMAIN", "AVAILABLE"}
						rows := make([][]string, len(group.Suggestions))
						for i, s := range group.Suggestions {
							rows[i] = []string{s.Name, formatAvailability(s.Availability)}
						}
						app.Output.PrintTable(headers, rows)
					}
					app.Output.Println("")
				}
			}

			return nil
		},
	}

	suggestKeywordCmd := &cobra.Command{
		Use:   "keyword <keyword>",
		Short: "Check keyword availability across TLDs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			registries, _ := cmd.Flags().GetString("registries")
			tlds, _ := cmd.Flags().GetString("tlds")

			result, err := app.Client.CheckKeywordAvailability(ctx, args[0], registries, tlds)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to check keyword availability: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintKeyValue("Keyword", result.Keyword)
				app.Output.PrintKeyValue("Total Domains", fmt.Sprintf("%d", result.TotalDomains))
				app.Output.PrintKeyValue("Available", fmt.Sprintf("%d", result.AvailableDomains))
				app.Output.PrintKeyValue("Unavailable", fmt.Sprintf("%d", result.UnavailableDomains))
				if result.ProcessingTimeMs > 0 {
					app.Output.PrintKeyValue("Processing Time", fmt.Sprintf("%dms", result.ProcessingTimeMs))
				}

				if len(result.Results) > 0 {
					app.Output.Println("")
					headers := []string{"DOMAIN", "TLD", "AVAILABLE", "PRICE", "PREMIUM"}
					rows := make([][]string, len(result.Results))
					for i, r := range result.Results {
						price := ""
						if r.TotalPrice > 0 {
							price = fmt.Sprintf("%.2f", r.TotalPrice)
						} else if r.Price > 0 {
							price = fmt.Sprintf("%.2f", r.Price)
						}
						premium := ""
						if r.Premium {
							premium = "Yes"
						}
						rows[i] = []string{r.Domain, r.TLD, formatAvailability(r.Availability), price, premium}
					}
					app.Output.PrintTable(headers, rows)
				}
			}

			return nil
		},
	}

	suggestKeywordSummaryCmd := &cobra.Command{
		Use:   "keyword-summary <keyword>",
		Short: "Check keyword availability summary (faster)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			registries, _ := cmd.Flags().GetString("registries")
			tlds, _ := cmd.Flags().GetString("tlds")

			result, err := app.Client.CheckKeywordSummary(ctx, args[0], registries, tlds)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to check keyword summary: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintKeyValue("Keyword", result.Keyword)
				app.Output.PrintKeyValue("Total Domains", fmt.Sprintf("%d", result.TotalDomains))
				app.Output.PrintKeyValue("Available", fmt.Sprintf("%d", result.AvailableDomains))
				app.Output.PrintKeyValue("Unavailable", fmt.Sprintf("%d", result.UnavailableDomains))
				if result.ProcessingTimeMs > 0 {
					app.Output.PrintKeyValue("Processing Time", fmt.Sprintf("%dms", result.ProcessingTimeMs))
				}
			}

			return nil
		},
	}

	// generate flags
	suggestGenerateCmd.Flags().String("tlds", "", "Comma-separated TLDs (e.g. com,net,org)")
	suggestGenerateCmd.Flags().String("lang", "", "Language code (e.g. eng)")
	suggestGenerateCmd.Flags().Bool("numbers", false, "Include numbers in suggestions")
	suggestGenerateCmd.Flags().Int("max", 20, "Maximum number of results")

	// spin flags
	suggestSpinCmd.Flags().Int("position", 0, "Word position to spin (0-based index)")
	suggestSpinCmd.Flags().Float64("similarity", 0.7, "Word similarity threshold (0.0-1.0)")
	suggestSpinCmd.Flags().String("tlds", "", "Comma-separated TLDs (e.g. com,net,org)")
	suggestSpinCmd.Flags().String("lang", "", "Language code (e.g. eng)")
	suggestSpinCmd.Flags().Int("max", 20, "Maximum number of results")

	// prefix flags
	suggestPrefixCmd.Flags().String("vocabulary", "@prefixes", "Vocabulary source for prefixes")
	suggestPrefixCmd.Flags().String("tlds", "", "Comma-separated TLDs (e.g. com,net,org)")
	suggestPrefixCmd.Flags().String("lang", "", "Language code (e.g. eng)")
	suggestPrefixCmd.Flags().Int("max", 20, "Maximum number of results")

	// suffix flags
	suggestSuffixCmd.Flags().String("vocabulary", "@suffixes", "Vocabulary source for suffixes")
	suggestSuffixCmd.Flags().String("tlds", "", "Comma-separated TLDs (e.g. com,net,org)")
	suggestSuffixCmd.Flags().String("lang", "", "Language code (e.g. eng)")
	suggestSuffixCmd.Flags().Int("max", 20, "Maximum number of results")

	// bulk flags
	suggestBulkCmd.Flags().String("tlds", "", "Comma-separated TLDs (e.g. com,net,org)")
	suggestBulkCmd.Flags().String("lang", "", "Language code (e.g. eng)")
	suggestBulkCmd.Flags().Bool("numbers", false, "Include numbers in suggestions")
	suggestBulkCmd.Flags().Int("max", 20, "Maximum results per name")

	// keyword flags
	suggestKeywordCmd.Flags().String("registries", "", "Comma-separated registries (e.g. verisign,pir)")
	suggestKeywordCmd.Flags().String("tlds", "", "Comma-separated TLDs (e.g. com,net,org)")

	// keyword-summary flags
	suggestKeywordSummaryCmd.Flags().String("registries", "", "Comma-separated registries (e.g. verisign,pir)")
	suggestKeywordSummaryCmd.Flags().String("tlds", "", "Comma-separated TLDs (e.g. com,net,org)")

	suggestCmd.AddCommand(
		suggestGenerateCmd,
		suggestSpinCmd,
		suggestPrefixCmd,
		suggestSuffixCmd,
		suggestBulkCmd,
		suggestKeywordCmd,
		suggestKeywordSummaryCmd,
	)
	parent.AddCommand(suggestCmd)
}

func formatAvailability(availability string) string {
	switch strings.ToLower(availability) {
	case "available":
		return "Yes"
	case "unavailable":
		return "No"
	default:
		return availability
	}
}

func printSuggestions(out *output.Formatter, result *models.SuggestResponse) {
	if out.IsJSON() {
		out.PrintResult(result)
		return
	}

	if len(result.Results) == 0 {
		out.Println("No suggestions found")
		return
	}

	headers := []string{"DOMAIN", "AVAILABLE"}
	rows := make([][]string, len(result.Results))
	for i, s := range result.Results {
		rows[i] = []string{s.Name, formatAvailability(s.Availability)}
	}
	out.PrintTable(headers, rows)
}
