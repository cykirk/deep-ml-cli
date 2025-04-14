package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cy/deep-ml-cli/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	client  *api.Client
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deep-ml",
	Short: "A CLI tool for solving problems from deep-ml.com",
	Long: `deep-ml is a command-line interface for interacting with deep-ml.com.
It allows you to browse problems, submit solutions, and track your progress
directly from your terminal.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.deep-ml/deep-ml.yaml)")

	// Initialize API client
	client = api.NewClient()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Create the config directory if it doesn't exist
		cfgDir := filepath.Join(home, ".deep-ml")
		if _, err := os.Stat(cfgDir); os.IsNotExist(err) {
			if err := os.MkdirAll(cfgDir, 0755); err != nil {
				fmt.Println("Error creating config directory:", err)
			}
		}
		
		// Set config path to ~/.deep-ml/deep-ml.yaml
		viper.AddConfigPath(cfgDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("deep-ml")
		
		// Create the problems directory if it doesn't exist
		problemsDir := filepath.Join(cfgDir, "problems")
		if _, err := os.Stat(problemsDir); os.IsNotExist(err) {
			if err := os.MkdirAll(problemsDir, 0755); err != nil {
				fmt.Println("Error creating problems directory:", err)
			}
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Set default values
	home, _ := os.UserHomeDir()
	viper.SetDefault("problems_dir", filepath.Join(home, ".deep-ml", "problems"))
	viper.SetDefault("include_problem_details", true)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// Load auth token info from config if exists
		token := viper.GetString("token")
		refreshToken := viper.GetString("refresh_token")
		expiresIn := viper.GetString("token_expires")
		
		if token != "" && refreshToken != "" {
			client.SetToken(token, refreshToken, expiresIn)
		}
	} else if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found, create it with default values
		if err := viper.WriteConfig(); err != nil {
			fmt.Println("Error creating config file:", err)
		}
	}
}