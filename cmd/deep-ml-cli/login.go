package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to deep-ml.com",
	Long:  `Login to your deep-ml.com account using Google Identity to access problems and submit solutions.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		// Prompt for email
		fmt.Print("Email: ")
		email, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading email:", err)
			return
		}
		email = strings.TrimSpace(email)

		// Prompt for password (without echoing)
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("\nError reading password:", err)
			return
		}
		password := string(passwordBytes)
		fmt.Println() // Add a newline after password input

		// Attempt login
		fmt.Println("Logging in with Google Identity Toolkit...")
		authResp, err := client.Login(email, password)
		if err != nil {
			fmt.Println("Login failed:", err)
			return
		}

		// Save token information to config
		viper.Set("token", authResp.Token)
		viper.Set("refresh_token", authResp.RefreshToken)
		viper.Set("token_expires", authResp.ExpiresIn)
		
		// Save user ID (either from authResp.LocalID or user profile)
		userID := authResp.LocalID
		if userID == "" && authResp.User.UserID != "" {
			userID = authResp.User.UserID
		}
		
		// If we have a user ID, save it
		if userID != "" {
			viper.Set("user_id", userID)
			fmt.Println("User ID saved automatically.")
		} else {
			// Still allow manual override if desired
			userIDOverride, err := cmd.Flags().GetString("user-id")
			if err == nil && userIDOverride != "" {
				viper.Set("user_id", userIDOverride)
				fmt.Println("User ID set from flag.")
			}
		}
		
		if err := viper.WriteConfig(); err != nil {
			fmt.Println("Failed to save configuration:", err)
			return
		}

		fmt.Printf("Logged in successfully as %s\n", authResp.User.Username)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("user-id", "", "Override the automatically detected user ID")
}