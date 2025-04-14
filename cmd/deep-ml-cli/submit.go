package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cy/deep-ml-cli/pkg/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// submitCmd represents the submit command
var submitCmd = &cobra.Command{
	Use:   "submit [problem-id] [solution-file]",
	Short: "Submit a solution to a problem",
	Long:  `Submit a solution to a specific problem. If no solution file is provided, it will look for a file in the problems directory with the same name as the problem ID.`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		problemID := args[0]
		var solutionFile string
		
		// Check if a solution file was provided
		if len(args) == 2 {
			solutionFile = args[1]
		} else {
			// Use the default file location based on problem ID
			problemsDir := viper.GetString("problems_dir")
			solutionFile = filepath.Join(problemsDir, fmt.Sprintf("%s.py", problemID))
			
			// Check if the file exists
			if _, err := os.Stat(solutionFile); os.IsNotExist(err) {
				fmt.Printf("Solution file not found: %s\n", solutionFile)
				fmt.Println("You can create it with: deep-ml edit", problemID)
				return
			}
		}
		
		// Read solution file
		code, err := os.ReadFile(solutionFile)
		if err != nil {
			fmt.Printf("Error reading solution file: %s\n", err)
			return
		}

		// Check if test-only mode is enabled
		testOnly, _ := cmd.Flags().GetBool("test-only")
		
		// If test-only mode is enabled, run tests locally without submitting
		if testOnly {
			fmt.Println("Running tests locally...")
			result, err := client.RunLocalTests(problemID, string(code))
			if err != nil {
				fmt.Printf("Error running local tests: %s\n", err)
				return
			}
			
			// Display result
			displayTestResults(result, problemID)
			return
		}
		
		// Otherwise, normal submission flow
		// Get user ID from flag, config, or environment
		userID, err := cmd.Flags().GetString("user-id")
		if err != nil || userID == "" {
			// Try to get from config
			userID = viper.GetString("user_id")
			if userID == "" {
				// Try to get from environment
				userID = os.Getenv("DEEP_ML_USER_ID")
				if userID == "" {
					fmt.Println("Error: user-id is required. Options:")
					fmt.Println("  1. Login with 'deep-ml login' to save it")
					fmt.Println("  2. Set it with --user-id flag")
					fmt.Println("  3. Set DEEP_ML_USER_ID environment variable")
					return
				}
			}
		}
		
		// Get difficulty from flag or default to "easy"
		difficulty, err := cmd.Flags().GetString("difficulty")
		if err != nil || difficulty == "" {
			difficulty = "easy"
		}
		
		// Create submission
		submission := &models.Submission{
			UserCode:   string(code),
			ProblemID:  problemID,
			Difficulty: difficulty,
			UserID:     userID,
		}

		// Get timeout and polling parameters
		waitSeconds, err := cmd.Flags().GetInt("wait-seconds")
		if err != nil || waitSeconds <= 0 {
			waitSeconds = 3 // Default to 3 seconds between polls
		}
		
		maxAttempts, err := cmd.Flags().GetInt("max-attempts")
		if err != nil || maxAttempts <= 0 {
			maxAttempts = 10 // Default to 10 attempts (30 seconds total with default wait)
		}
		
		// Run local tests first
		runLocalFirst, _ := cmd.Flags().GetBool("run-local-first")
		if runLocalFirst {
			fmt.Println("Running tests locally before submitting...")
			localResult, err := client.RunLocalTests(problemID, string(code))
			if err != nil {
				fmt.Printf("Warning: Could not run local tests: %s\n", err)
				fmt.Println("Continuing with remote submission...")
			} else {
				displayTestResults(localResult, problemID)
				
				// If local tests failed, ask for confirmation before submitting remotely
				if localResult.Status != "accepted" {
					fmt.Println("\nLocal tests did not pass. Submit anyway? (y/N): ")
					var response string
					fmt.Scanln(&response)
					if response != "y" && response != "Y" {
						fmt.Println("Submission canceled.")
						return
					}
				}
			}
		}
		
		// Submit solution
		fmt.Println("Submitting solution...")
		taskResponse, err := client.SubmitSolution(submission)
		if err != nil {
			fmt.Printf("Error submitting solution: %s\n", err)
			return
		}
		
		// Show task ID
		fmt.Printf("Task ID: %s\n", taskResponse.TaskID)
		
		// Skip polling for results if we've already run tests locally
		skipPolling, _ := cmd.Flags().GetBool("skip-polling")
		if runLocalFirst || skipPolling {
			fmt.Println("Submission sent successfully. Skipping remote execution result polling.")
			fmt.Println("You can check the result later by visiting:")
			fmt.Printf("https://www.deep-ml.com/problems/%s\n", problemID)
			return
		}
		
		// Poll for results if not skipping
		fmt.Printf("Waiting for execution results (polling every %d seconds, max %d attempts)...\n", waitSeconds, maxAttempts)
		result, err := client.GetSubmissionResult(taskResponse.TaskID, waitSeconds, maxAttempts)
		if err != nil {
			fmt.Printf("Error getting submission result: %s\n", err)
			fmt.Println("\nYou can check the result later by visiting:")
			fmt.Printf("https://www.deep-ml.com/problems/%s\n", problemID)
			return
		}
		
		// Display result
		displayTestResults(result, problemID)
	},
}

// displayTestResults shows test results in a formatted way
func displayTestResults(result *models.SubmissionResult, problemID string) {
	fmt.Printf("Status: %s\n", result.Status)
	
	if result.Runtime != "" {
		fmt.Printf("Runtime: %s\n", result.Runtime)
	}
	
	if result.Memory != "" {
		fmt.Printf("Memory: %s\n", result.Memory)
	}
	
	if result.Message != "" {
		fmt.Printf("\nMessage: %s\n", result.Message)
		// If we only got an error message, no test cases, just return
		if result.Status == "error" && len(result.TestCases) == 0 {
			return
		}
	}
	
	// Show test case results if available
	if len(result.TestCases) > 0 {
		fmt.Println("\nTest Cases:")
		passedCount := 0
		for i, tc := range result.TestCases {
			fmt.Printf("\n[Test Case %d]\n", i+1)
			fmt.Printf("Test: %s\n", tc.TestCase)
			fmt.Printf("Expected: %s\n", tc.ExpectedOutput)
			fmt.Printf("Actual: %s\n", tc.ActualOutput)
			fmt.Printf("Passed: %t\n", tc.Passed)
			
			if tc.Passed {
				passedCount++
			}
		}
		
		fmt.Printf("\nSummary: %d/%d tests passed\n", passedCount, len(result.TestCases))
	}
}

func init() {
	rootCmd.AddCommand(submitCmd)
	
	// Add flags for submission
	submitCmd.Flags().String("user-id", "", "User ID for submission (optional if saved during login)")
	submitCmd.Flags().String("difficulty", "easy", "Problem difficulty (easy, medium, hard)")
	submitCmd.Flags().Int("wait-seconds", 3, "Seconds to wait between result polling attempts")
	submitCmd.Flags().Int("max-attempts", 10, "Maximum number of polling attempts before giving up")
	
	// Add flags for local testing
	submitCmd.Flags().Bool("test-only", false, "Run tests locally without submitting to the server")
	submitCmd.Flags().Bool("run-local-first", true, "Run tests locally before submitting to the server")
	submitCmd.Flags().Bool("skip-polling", false, "Skip waiting for remote execution results (still submits the solution)")
}