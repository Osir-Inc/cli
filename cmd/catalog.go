package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/osir/cli/internal/output"
	"github.com/spf13/cobra"
)

func addCatalogCommands(parent *cobra.Command) {
	catalogCmd := &cobra.Command{
		Use:   "catalog",
		Short: "Product catalog browsing",
		Long: `Browse the OSIR product catalog.

Available catalogs:
  domains [ext]  - Domain extensions and pricing (e.g. catalog domains com)
  servers        - Dedicated server configurations and pricing`,
	}

	catalogDomainsCmd := &cobra.Command{
		Use:   "domains [extension]",
		Short: "List domain extensions and pricing",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			extension := ""
			if len(args) > 0 {
				extension = args[0]
			}

			result, err := app.Client.ListDomainExtensions(ctx, extension)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to list domain extensions: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printCatalogDomains(app.Output, result)
			return nil
		},
	}

	catalogServersCmd := &cobra.Command{
		Use:   "servers",
		Short: "List dedicated server configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.ListDedicatedServers(ctx)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to list dedicated servers: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printCatalogServers(app.Output, result)
			return nil
		},
	}

	catalogCmd.AddCommand(catalogDomainsCmd, catalogServersCmd)
	parent.AddCommand(catalogCmd)
}

func printCatalogDomains(out *output.Formatter, raw json.RawMessage) {
	var resp struct {
		Extensions []struct {
			Extension         string  `json:"extension"`
			ExtensionType     string  `json:"extensionType"`
			Registrar         string  `json:"registrar"`
			RegistrationPrice float64 `json:"registrationPrice"`
			RenewalPrice      float64 `json:"renewalPrice"`
			TransferPrice     float64 `json:"transferPrice"`
		} `json:"extensions"`
		TotalExtensions int `json:"totalExtensions"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		out.PrintResult(raw)
		return
	}

	out.Println(fmt.Sprintf("Total extensions: %d", resp.TotalExtensions))
	out.Println("")

	headers := []string{"EXTENSION", "TYPE", "REGISTER", "RENEW", "TRANSFER", "REGISTRAR"}
	rows := make([][]string, len(resp.Extensions))
	for i, ext := range resp.Extensions {
		rows[i] = []string{
			ext.Extension,
			ext.ExtensionType,
			fmt.Sprintf("$%.2f", ext.RegistrationPrice/100),
			fmt.Sprintf("$%.2f", ext.RenewalPrice/100),
			fmt.Sprintf("$%.2f", ext.TransferPrice/100),
			ext.Registrar,
		}
	}
	out.PrintTable(headers, rows)
}

func printCatalogServers(out *output.Formatter, raw json.RawMessage) {
	var resp struct {
		TotalPackages int `json:"totalPackages"`
		Packages      []struct {
			Name           string  `json:"name"`
			CPUModel       string  `json:"cpuModel"`
			CPUCores       int     `json:"cpuCores"`
			MemoryGB       int     `json:"memoryGb"`
			MemoryType     string  `json:"memoryType"`
			PrimaryStorage int     `json:"primaryStorageGb"`
			PrimaryType    string  `json:"primaryStorageType"`
			PrimaryCount   int     `json:"primaryStorageCount"`
			RAID           string  `json:"raidConfiguration"`
			BandwidthGbps  float64 `json:"bandwidthGbps"`
			TrafficTB      int     `json:"trafficTb"`
			IPv4           int     `json:"ipv4Addresses"`
			PriceMonthly   int     `json:"priceMonthly"`
			SetupFee       int     `json:"setupFee"`
			StockAvailable int     `json:"stockAvailable"`
			InStock        bool    `json:"inStock"`
			Category       struct {
				Name string `json:"name"`
			} `json:"category"`
			Datacenter struct {
				DisplayName string `json:"displayName"`
			} `json:"datacenter"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		out.PrintResult(raw)
		return
	}

	out.Println(fmt.Sprintf("Total packages: %d", resp.TotalPackages))
	out.Println("")

	headers := []string{"NAME", "CPU", "RAM", "STORAGE", "BW", "MONTHLY", "SETUP", "STOCK", "LOCATION"}
	rows := make([][]string, len(resp.Packages))
	for i, p := range resp.Packages {
		storage := fmt.Sprintf("%dx%dGB %s", p.PrimaryCount, p.PrimaryStorage, p.PrimaryType)
		if p.RAID != "" {
			storage += " " + p.RAID
		}
		stock := strconv.Itoa(p.StockAvailable)
		if !p.InStock {
			stock = "OUT"
		}
		rows[i] = []string{
			p.Name,
			fmt.Sprintf("%s (%dC)", p.CPUModel, p.CPUCores),
			fmt.Sprintf("%dGB %s", p.MemoryGB, p.MemoryType),
			storage,
			fmt.Sprintf("%.0fGbps/%dTB", p.BandwidthGbps, p.TrafficTB),
			fmt.Sprintf("$%.2f", float64(p.PriceMonthly)/100),
			fmt.Sprintf("$%.2f", float64(p.SetupFee)/100),
			stock,
			p.Datacenter.DisplayName,
		}
	}
	out.PrintTable(headers, rows)
}
