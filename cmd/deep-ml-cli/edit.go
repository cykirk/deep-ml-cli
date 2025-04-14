package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit [problem-id]",
	Short: "Edit a problem solution",
	Long:  `Open a problem solution in your default editor. If the file doesn't exist, it will be created from the problem's starter code.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		problemID := args[0]
		
		// Get the problems directory from config
		problemsDir := viper.GetString("problems_dir")
		includeProblemDetails := viper.GetBool("include_problem_details")
		
		// Ensure the problems directory exists
		if err := os.MkdirAll(problemsDir, 0755); err != nil {
			fmt.Printf("Error creating problems directory: %s\n", err)
			return
		}
		
		// Determine the file path
		filePath := filepath.Join(problemsDir, fmt.Sprintf("%s.py", problemID))
		
		// Check if the file already exists
		fileExists := true
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fileExists = false
			
			// Get the problem from the API to get the starter code
			fmt.Printf("Fetching problem %s from deep-ml.com...\n", problemID)
			problem, err := client.GetProblem(problemID)
			if err != nil {
				fmt.Printf("Error fetching problem: %s\n", err)
				return
			}
			
			// Create the file with starter code
			var fileContent strings.Builder
			
			// Include problem details as a comment if configured
			if includeProblemDetails {
				fileContent.WriteString(fmt.Sprintf("'''\n# Problem %s: %s\n", problem.ID, problem.Title))
				fileContent.WriteString(fmt.Sprintf("Difficulty: %s\n", problem.Difficulty))
				fileContent.WriteString(fmt.Sprintf("Category: %s\n\n", problem.Category))
				fileContent.WriteString(fmt.Sprintf("%s\n\n", problem.Description))
				
				if len(problem.TestCases) > 0 {
					fileContent.WriteString("Test Cases:\n")
					for _, tc := range problem.TestCases {
						fileContent.WriteString(fmt.Sprintf("- %s\n", tc))
					}
					fileContent.WriteString("\n")
				}
				
				if problem.Example != nil && len(problem.Example) > 0 {
					fileContent.WriteString("Example:\n")
					for k, v := range problem.Example {
						fileContent.WriteString(fmt.Sprintf("%s: %s\n", k, v))
					}
					fileContent.WriteString("\n")
				}
				
				fileContent.WriteString("'''\n\n")
			}
			
			// Add the starter code
			if problem.Starter != "" {
				fileContent.WriteString(problem.Starter)
			} else {
				// If no starter code, create a basic template
				fileContent.WriteString(fmt.Sprintf("# Solution for problem %s\n\n", problemID))
			}
			
			// Write the file
			if err := os.WriteFile(filePath, []byte(fileContent.String()), 0644); err != nil {
				fmt.Printf("Error creating solution file: %s\n", err)
				return
			}
			
			fmt.Printf("Created solution file: %s\n", filePath)
		}
		
		// Open the file in the user's editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			// Default editors by platform
			switch runtime.GOOS {
			case "windows":
				editor = "notepad"
			case "darwin": // macOS
				editor = "open -t" // Opens in default text editor
			default: // Linux and others
				// Try common editors
				for _, e := range []string{"vim", "nano", "code", "gedit"} {
					if _, err := exec.LookPath(e); err == nil {
						editor = e
						break
					}
				}
				// If no editor found, default to vim
				if editor == "" {
					editor = "vim"
				}
			}
		}
		
		parts := strings.Fields(editor)
		var execCmd *exec.Cmd
		if len(parts) > 1 {
			execCmd = exec.Command(parts[0], append(parts[1:], filePath)...)
		} else {
			execCmd = exec.Command(editor, filePath)
		}
		
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		
		fmt.Printf("Opening %s with %s...\n", filePath, editor)
		if err := execCmd.Run(); err != nil {
			fmt.Printf("Error opening editor: %s\n", err)
			return
		}
		
		if !fileExists {
			fmt.Printf("\nTip: When you're ready, submit your solution with:\n")
			fmt.Printf("deep-ml submit %s\n", problemID)
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}