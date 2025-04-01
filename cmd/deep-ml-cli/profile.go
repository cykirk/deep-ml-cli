package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "View your profile",
	Long:  `View your deep-ml.com profile statistics.`,
	Run: func(cmd *cobra.Command, args []string) {
		user, err := client.GetUserProfile()
		if err != nil {
			fmt.Println("Error fetching profile:", err)
			return
		}

		fmt.Println("User Profile")
		fmt.Println("============")
		fmt.Printf("Username: %s\n", user.Username)
		if user.Email != "" {
			fmt.Printf("Email: %s\n", user.Email)
		}
		fmt.Printf("Problems Solved: %d / %d (%.1f%%)\n", 
			user.ProblemsSolved, 
			user.ProblemsTotal,
			float64(user.ProblemsSolved) / float64(user.ProblemsTotal) * 100)
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
}