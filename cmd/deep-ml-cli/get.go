package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [problem-id]",
	Short: "Get a specific problem",
	Long:  `Get a specific problem by ID and display its details.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		problemID := args[0]
		
		problem, err := client.GetProblem(problemID)
		if err != nil {
			fmt.Println("Error fetching problem:", err)
			return
		}

		// Get display options
		showSolution, _ := cmd.Flags().GetBool("solution")
		showLearn, _ := cmd.Flags().GetBool("learn")
		showTests, _ := cmd.Flags().GetBool("tests")

		// Display problem details
		fmt.Printf("Problem %s: %s\n", problem.ID, problem.Title)
		fmt.Printf("Difficulty: %s\n", problem.Difficulty)
		fmt.Printf("Category: %s\n", problem.Category)
		
		if likes, err := strconv.Atoi(problem.Likes); err == nil {
			fmt.Printf("Likes: %d | ", likes)
		}
		if dislikes, err := strconv.Atoi(problem.Dislikes); err == nil {
			fmt.Printf("Dislikes: %d\n", dislikes)
		} else {
			fmt.Println()
		}
		
		if problem.Video != "" {
			fmt.Printf("Tutorial Video: %s\n", problem.Video)
		}
		fmt.Println()
		
		// Description
		fmt.Println("Description:")
		fmt.Println(problem.Description)
		fmt.Println()
		
		// Example
		if problem.Example != nil && len(problem.Example) > 0 {
			fmt.Println("Example:")
			if input, ok := problem.Example["input"]; ok {
				fmt.Printf("Input: %s\n", input)
			}
			if output, ok := problem.Example["output"]; ok {
				fmt.Printf("Output: %s\n", output)
			}
			if reasoning, ok := problem.Example["reasoning"]; ok {
				fmt.Printf("Reasoning: %s\n", reasoning)
			}
			fmt.Println()
		}

		// Show test cases if requested
		if showTests && len(problem.TestCases) > 0 {
			fmt.Println("Test Cases:")
			for i, tc := range problem.TestCases {
				fmt.Printf("Test Case %d:\n%s\n\n", i+1, tc)
			}
		}

		// Show starter code
		if problem.Starter != "" {
			fmt.Println("Starter Code:")
			fmt.Println(problem.Starter)
			fmt.Println()
		}

		// Show learning section if requested
		if showLearn && problem.LearnSection != "" {
			fmt.Println("Learning Material:")
			fmt.Println(problem.LearnSection)
			fmt.Println()
		}

		// Show solution if requested
		if showSolution && problem.Solution != "" {
			fmt.Println("Solution:")
			fmt.Println(problem.Solution)
			fmt.Println()
		}
		
		// Check if save flag is set
		save, _ := cmd.Flags().GetBool("save")
		if save {
			// Create a directory for the problem
			dir := filepath.Join(".", fmt.Sprintf("problem_%s", problem.ID))
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Println("Error creating directory:", err)
				return
			}
			
			// Save problem description
			descFile := filepath.Join(dir, "README.md")
			content := fmt.Sprintf("# Problem %s: %s\n\n", problem.ID, problem.Title)
			content += fmt.Sprintf("Difficulty: %s\n", problem.Difficulty)
			content += fmt.Sprintf("Category: %s\n\n", problem.Category)
			
			if problem.Video != "" {
				content += fmt.Sprintf("Tutorial Video: %s\n\n", problem.Video)
			}
			
			content += "## Description\n\n"
			content += problem.Description + "\n\n"
			
			if problem.Example != nil && len(problem.Example) > 0 {
				content += "## Example\n\n"
				if input, ok := problem.Example["input"]; ok {
					content += fmt.Sprintf("Input: %s\n", input)
				}
				if output, ok := problem.Example["output"]; ok {
					content += fmt.Sprintf("Output: %s\n", output)
				}
				if reasoning, ok := problem.Example["reasoning"]; ok {
					content += fmt.Sprintf("Reasoning: %s\n\n", reasoning)
				}
			}
			
			if len(problem.TestCases) > 0 {
				content += "## Test Cases\n\n"
				for i, tc := range problem.TestCases {
					content += fmt.Sprintf("### Test Case %d\n%s\n\n", i+1, tc)
				}
			}
			
			if problem.LearnSection != "" {
				content += "## Learning Material\n\n"
				content += problem.LearnSection + "\n\n"
			}

			if err := os.WriteFile(descFile, []byte(content), 0644); err != nil {
				fmt.Println("Error saving problem description:", err)
				return
			}
			
			// Save starter code if available
			if problem.Starter != "" {
				// Determine file extension based on likely language
				extension := ".py"  // Default to Python
				if strings.Contains(strings.ToLower(problem.Starter), "func ") {
					extension = ".go"
				} else if strings.Contains(strings.ToLower(problem.Starter), "function") && 
					      (strings.Contains(problem.Starter, "=>") || strings.Contains(problem.Starter, "return")) {
					extension = ".js"
				} else if strings.Contains(strings.ToLower(problem.Starter), "public class") {
					extension = ".java"
				}
				
				starterFile := filepath.Join(dir, "solution"+extension)
				if err := os.WriteFile(starterFile, []byte(problem.Starter), 0644); err != nil {
					fmt.Println("Error saving starter code:", err)
					return
				}
			}
			
			// Save solution if available
			if problem.Solution != "" {
				// Determine file extension similar to starter code
				extension := ".py"  // Default to Python
				if strings.Contains(strings.ToLower(problem.Solution), "func ") {
					extension = ".go"
				} else if strings.Contains(strings.ToLower(problem.Solution), "function") && 
					      (strings.Contains(problem.Solution, "=>") || strings.Contains(problem.Solution, "return")) {
					extension = ".js"
				} else if strings.Contains(strings.ToLower(problem.Solution), "public class") {
					extension = ".java"
				}
				
				solutionFile := filepath.Join(dir, "reference_solution"+extension)
				if err := os.WriteFile(solutionFile, []byte(problem.Solution), 0644); err != nil {
					fmt.Println("Error saving solution code:", err)
					return
				}
			}
			
			fmt.Printf("\nProblem saved to %s\n", dir)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("save", "s", false, "Save the problem locally")
	getCmd.Flags().BoolP("solution", "S", false, "Show the reference solution")
	getCmd.Flags().BoolP("learn", "l", false, "Show the learning material")
	getCmd.Flags().BoolP("tests", "t", false, "Show the test cases")
}