package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cy/deep-ml-cli/pkg/models"
	"github.com/spf13/cobra"
)

// submitCmd represents the submit command
var submitCmd = &cobra.Command{
	Use:   "submit [problem-id] [solution-file]",
	Short: "Submit a solution to a problem",
	Long:  `Submit a solution to a specific problem.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		problemID := args[0]
		solutionFile := args[1]
		
		// Determine language from file extension
		ext := strings.ToLower(filepath.Ext(solutionFile))
		var language string
		switch ext {
		case ".py":
			language = "python"
		case ".go":
			language = "go"
		case ".js":
			language = "javascript"
		case ".java":
			language = "java"
		case ".cpp", ".cc":
			language = "cpp"
		case ".c":
			language = "c"
		default:
			fmt.Printf("Unsupported file extension: %s\n", ext)
			return
		}
		
		// Read solution file
		code, err := os.ReadFile(solutionFile)
		if err != nil {
			fmt.Printf("Error reading solution file: %s\n", err)
			return
		}
		
		// Create submission
		submission := &models.Submission{
			ProblemID: problemID,
			Code:      string(code),
			Language:  language,
		}
		
		// Submit solution
		fmt.Println("Submitting solution...")
		result, err := client.SubmitSolution(submission)
		if err != nil {
			fmt.Printf("Error submitting solution: %s\n", err)
			return
		}
		
		// Display result
		fmt.Printf("Submission ID: %s\n", result.ID)
		fmt.Printf("Status: %s\n", result.Status)
		
		if result.Runtime != "" {
			fmt.Printf("Runtime: %s\n", result.Runtime)
		}
		
		if result.Memory != "" {
			fmt.Printf("Memory: %s\n", result.Memory)
		}
		
		if result.Message != "" {
			fmt.Printf("\nMessage: %s\n", result.Message)
		}
		
		// Show test case results if available
		if len(result.TestCases) > 0 {
			fmt.Println("\nTest Cases:")
			for i, tc := range result.TestCases {
				fmt.Printf("\n[Test Case %d]\n", i+1)
				fmt.Printf("Input: %s\n", tc.Input)
				fmt.Printf("Expected: %s\n", tc.Output)
				fmt.Printf("Result: %s\n", tc.Result)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
}