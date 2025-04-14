# Deep ML CLI

A command-line interface for interacting with [deep-ml.com](https://deep-ml.com), allowing you to browse problems, submit solutions, and track your progress directly from your terminal.

## Installation

```bash
go install github.com/cy/deep-ml-cli/cmd/deep-ml-cli@latest
```

Or build from source:

```bash
git clone https://github.com/cy/deep-ml-cli.git
cd deep-ml-cli
go build -o deep-ml ./cmd/deep-ml-cli
```

## Configuration

Before using the CLI, you need to set your Firebase API key in the source code. Open `pkg/api/client.go` and replace the `apiKey` constant with your actual Firebase API key:

```go
const (
    // ...
    apiKey = "YOUR_FIREBASE_API_KEY" // Replace with your Firebase API key
    // ...
)
```

The CLI uses a YAML configuration file located at `~/.deep-ml/deep-ml.yaml`. Here are the available configuration options:

```yaml
# Directory where problem solutions are stored
problems_dir: ~/.deep-ml/problems

# Whether to include problem details as comments in solution files
include_problem_details: true

# User ID for API requests (saved during login)
user_id: your_deep_ml_user_id

# Authentication tokens (automatically managed by the CLI)
token: your_authentication_token
refresh_token: your_refresh_token
token_expires: expiration_time
```

You can modify this file manually or use the `--config` flag to specify a different configuration file.

## Usage

```
deep-ml is a command-line interface for interacting with deep-ml.com.
It allows you to browse problems, submit solutions, and track your progress
directly from your terminal.

Usage:
  deep-ml [command]

Available Commands:
  edit        Edit a problem solution
  get         Get a specific problem
  help        Help about any command
  list        List available problems
  login       Login to deep-ml.com
  profile     View your profile
  submit      Submit a solution to a problem
  test        Test a solution locally without submitting

Flags:
      --config string   config file (default is $HOME/.deep-ml/deep-ml.yaml)
  -h, --help            help for deep-ml

Use "deep-ml [command] --help" for more information about a command.
```

## Examples

### Login

```bash
$ deep-ml login
Email: your.email@example.com
Password: 
Logging in with Google Identity Toolkit...
User ID saved automatically.
Logged in successfully as your_username
```

Your user ID is automatically retrieved from Firebase authentication and saved in the configuration file. It will be used automatically for submissions.

If needed, you can override the automatically detected user ID with the `--user-id` flag:

```bash
$ deep-ml login --user-id eJUGOmlFP9ctLBR45ty8EiV7g2o1
```

### List Problems

```bash
$ deep-ml list
Found 115 problems:

ID    | TITLE                                   | DIFFICULTY | CATEGORY
-------------------------------------------------------------------------------------
1     | Matrix-Vector Dot Product               | easy       | Linear Algebra
2     | Transpose of a Matrix                   | easy       | Linear Algebra
3     | Reshape Matrix                          | easy       | Linear Algebra
...

🔥 Today's Daily Question: #66 - Implement Orthogonal Projection of a Vector onto a Line (use --daily-only for details)
```

You can filter and sort the problem list:

```bash
# Show only ML problems
$ deep-ml list --category "Machine Learning"

# Show only easy problems
$ deep-ml list --difficulty easy

# Sort by difficulty
$ deep-ml list --sort difficulty

# See the daily challenge
$ deep-ml list --daily-only
```

### Get Problem Details

```bash
$ deep-ml get 1
Problem 1: Matrix-Vector Dot Product
Difficulty: easy
Category: Linear Algebra
Likes: 0 | Dislikes: 0
Tutorial Video: https://youtu.be/DNoLs5tTGAw?si=vpkPobZMA8YY10WY

Description:
Write a Python function that computes the dot product of a matrix and a vector. The function should return a list representing the resulting vector if the operation is valid, or -1 if the matrix and vector dimensions are incompatible. A matrix (a list of lists) can be dotted with a vector (a list) only if the number of columns in the matrix equals the length of the vector. For example, an n x m matrix requires a vector of length m.

Example:
Input: a = [[1, 2], [2, 4]], b = [1, 2]
Output: [5, 10]
Reasoning: Row 1: (1 * 1) + (2 * 2) = 1 + 4 = 5; Row 2: (1 * 2) + (2 * 4) = 2 + 8 = 10

Starter Code:
def matrix_dot_vector(a: list[list[int|float]], b: list[int|float]) -> list[int|float]:
	# Return a list where each element is the dot product of a row of 'a' with 'b'.
	# If the number of columns in 'a' does not match the length of 'b', return -1.
	pass

# View test cases, learning material, and solution with flags
$ deep-ml get 1 --tests
$ deep-ml get 1 --learn
$ deep-ml get 1 --solution

# Save problem locally with all materials
$ deep-ml get 1 --save
Problem saved to ./problem_1
```

### Edit a Problem Solution

```bash
$ deep-ml edit 1
Fetching problem 1 from deep-ml.com...
Created solution file: /home/user/.deep-ml/problems/1.py
Opening /home/user/.deep-ml/problems/1.py with vim...

# Your editor opens with the problem details and starter code
```

### Submit Solution

```bash
# Test a solution locally before submitting
$ deep-ml test 1
Running tests locally...
Status: accepted
Test Cases: 3/3 tests passed

# Submit a specific file (user ID is loaded from config)
$ deep-ml submit 1 ./solution.py
Running tests locally before submitting...
Status: accepted
Test Cases: 3/3 tests passed

Submitting solution...
Task ID: a96f84e0-1a06-4298-b1b6-5d59864a2e7c
Waiting for execution results (polling every 3 seconds, max 10 attempts)...
Status: accepted
Test Cases: 3/3 tests passed

# Or submit using the default file location (after editing with 'deep-ml edit')
$ deep-ml submit 1
Running tests locally before submitting...
Status: accepted
Test Cases: 3/3 tests passed

Submitting solution...
Task ID: 5b92c3d7-8a15-4f01-a3e4-2c1690bef47d
Waiting for execution results (polling every 3 seconds, max 10 attempts)...
Status: accepted
Test Cases: 3/3 tests passed

# Submit without running tests locally first
$ deep-ml submit 1 --run-local-first=false
Submitting solution...
Task ID: 5b92c3d7-8a15-4f01-a3e4-2c1690bef47d
Waiting for execution results (polling every 3 seconds, max 10 attempts)...
Status: accepted
Test Cases: 3/3 tests passed

# Run tests locally and submit without waiting for server results
$ deep-ml submit 1
Running tests locally before submitting...
Status: accepted
Test Cases: 3/3 tests passed

Submitting solution...
Task ID: 5b92c3d7-8a15-4f01-a3e4-2c1690bef47d
Submission sent successfully. Skipping remote execution result polling.
You can check the result later by visiting:
https://www.deep-ml.com/problems/1

# Explicitly skip remote result polling
$ deep-ml submit 1 --skip-polling
Running tests locally before submitting...
Status: accepted
Test Cases: 3/3 tests passed

Submitting solution...
Task ID: 5b92c3d7-8a15-4f01-a3e4-2c1690bef47d
Submission sent successfully. Skipping remote execution result polling.
You can check the result later by visiting:
https://www.deep-ml.com/problems/1

# Test only without submitting
$ deep-ml test 1
Running tests locally...
Status: accepted
Test Cases: 3/3 tests passed

# Alternative way to test only without submitting
$ deep-ml submit 1 --test-only
Running tests locally...
Status: accepted
Test Cases: 3/3 tests passed

# You can customize polling behavior
$ deep-ml submit 1 --wait-seconds 5 --max-attempts 20

# You can specify a user ID if needed
$ deep-ml submit 1 --user-id eJUGOmlFP9ctLBR45ty8EiV7g2o1
```

### View Profile

```bash
$ deep-ml profile
User Profile
============
Username: your_username
Email: your_email@example.com
Problems Solved: 15 / 42 (35.7%)
```

## Authentication

This CLI uses Google Identity Toolkit (Firebase Authentication) for authentication. Your login token is securely stored in your home directory in the `.deep-ml.yaml` file. The CLI will automatically refresh your token when it expires.