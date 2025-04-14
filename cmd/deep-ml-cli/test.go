package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test [problem-id] [solution-file]",
	Short: "Test a solution locally",
	Long:  `Test a solution against a problem's test cases locally without submitting to the server. If no solution file is provided, it will look for a file in the problems directory with the same name as the problem ID.`,
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
		
		// Run the local tests
		fmt.Println("Running tests locally...")
		result, err := client.RunLocalTests(problemID, string(code))
		if err != nil {
			fmt.Printf("Error running local tests: %s\n", err)
			return
		}
		
		// Display result
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
			
			// Provide a hint for submitting if all tests passed
			if passedCount == len(result.TestCases) {
				fmt.Println("\nAll tests passed! You can submit your solution with:")
				fmt.Printf("deep-ml submit %s\n", problemID)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}