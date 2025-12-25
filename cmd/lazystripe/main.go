package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.ibm.com/soub4i/lazystripe/internal/config"
	"github.ibm.com/soub4i/lazystripe/internal/ui"
	"github.ibm.com/soub4i/lazystripe/internal/version"

	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v84"
)

func main() {
	var rootCmd = &cobra.Command{Use: "lazystripe"}

	var initCmd = &cobra.Command{
		Use:   "init <apikey>",
		Short: "Initialize config with API key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cp, _ := os.UserHomeDir()
			if err := os.MkdirAll(filepath.Join(cp, ".lazystripe"), 0755); err != nil {
				log.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(cp, ".lazystripe", "config"), []byte(args[0]), 0600); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Config created at ~/.lazystripe/config")
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Lazystripe: %s\n", version.String())
			fmt.Printf("Stripe: %s\n", stripe.APIVersion)
		},
	}

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the TUI",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			if err := ui.Run(cfg.APIKey); err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.AddCommand(initCmd, versionCmd, runCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
