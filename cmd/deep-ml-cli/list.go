package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available problems",
	Long:  `List problems available on deep-ml.com.`,
	Run: func(cmd *cobra.Command, args []string) {
		problems, err := client.GetProblems()
		if err != nil {
			fmt.Println("Error fetching problems:", err)
			return
		}

		if len(problems.Problems) == 0 {
			fmt.Println("No problems found.")
			return
		}

		// Get filter flags
		category, _ := cmd.Flags().GetString("category")
		difficulty, _ := cmd.Flags().GetString("difficulty")
		sortBy, _ := cmd.Flags().GetString("sort")
		showDaily, _ := cmd.Flags().GetBool("daily-only")

		// Show daily question if requested
		if showDaily {
			p := problems.DailyQuestion
			fmt.Println("🔥 Today's Daily Question:")
			fmt.Printf("ID: %s\n", p.ID)
			fmt.Printf("Title: %s\n", p.Title)
			fmt.Printf("Category: %s\n", p.Category)
			fmt.Printf("Difficulty: %s\n", p.Difficulty)
			return
		}

		// Filter problems
		var filteredProblems []struct {
			ID         string
			Title      string
			Difficulty string
			Category   string
			IdNum      int // For numerical sorting
		}

		for _, p := range problems.Problems {
			// Convert ID to integer for sorting
			idNum, _ := strconv.Atoi(p.ID)

			// Apply filters
			if category != "" && !strings.EqualFold(p.Category, category) {
				continue
			}
			if difficulty != "" && !strings.EqualFold(p.Difficulty, difficulty) {
				continue
			}

			filteredProblems = append(filteredProblems, struct {
				ID         string
				Title      string
				Difficulty string
				Category   string
				IdNum      int
			}{
				ID:         p.ID,
				Title:      p.Title,
				Difficulty: p.Difficulty,
				Category:   p.Category,
				IdNum:      idNum,
			})
		}

		// Sort problems
		switch sortBy {
		case "id":
			sort.Slice(filteredProblems, func(i, j int) bool {
				return filteredProblems[i].IdNum < filteredProblems[j].IdNum
			})
		case "title":
			sort.Slice(filteredProblems, func(i, j int) bool {
				return filteredProblems[i].Title < filteredProblems[j].Title
			})
		case "difficulty":
			difficultyRank := map[string]int{"easy": 1, "medium": 2, "hard": 3}
			sort.Slice(filteredProblems, func(i, j int) bool {
				return difficultyRank[strings.ToLower(filteredProblems[i].Difficulty)] < difficultyRank[strings.ToLower(filteredProblems[j].Difficulty)]
			})
		case "category":
			sort.Slice(filteredProblems, func(i, j int) bool {
				return filteredProblems[i].Category < filteredProblems[j].Category
			})
		default:
			// Default sort by ID
			sort.Slice(filteredProblems, func(i, j int) bool {
				return filteredProblems[i].IdNum < filteredProblems[j].IdNum
			})
		}

		fmt.Printf("Found %d problems:\n\n", len(filteredProblems))
		fmt.Printf("%-5s | %-40s | %-10s | %s\n", "ID", "TITLE", "DIFFICULTY", "CATEGORY")
		fmt.Println(strings.Repeat("-", 85))

		for _, p := range filteredProblems {
			title := p.Title
			if len(title) > 40 {
				title = title[:37] + "..."
			}
			
			category := p.Category
			if len(category) > 20 {
				category = category[:17] + "..."
			}
			
			fmt.Printf("%-5s | %-40s | %-10s | %s\n", 
				p.ID, 
				title, 
				p.Difficulty, 
				category)
		}

		// Show info about the daily question
		fmt.Printf("\n🔥 Today's Daily Question: #%s - %s (use --daily-only for details)\n", 
			problems.DailyQuestion.ID, problems.DailyQuestion.Title)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	
	listCmd.Flags().StringP("category", "c", "", "Filter problems by category (e.g., 'Machine Learning', 'NLP')")
	listCmd.Flags().StringP("difficulty", "d", "", "Filter problems by difficulty (easy, medium, hard)")
	listCmd.Flags().StringP("sort", "s", "id", "Sort problems by: id, title, difficulty, category")
	listCmd.Flags().BoolP("daily-only", "D", false, "Show only today's daily question")
}