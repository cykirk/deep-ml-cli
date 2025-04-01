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

## Usage

```
deep-ml is a command-line interface for interacting with deep-ml.com.
It allows you to browse problems, submit solutions, and track your progress
directly from your terminal.

Usage:
  deep-ml [command]

Available Commands:
  get         Get a specific problem
  help        Help about any command
  list        List available problems
  login       Login to deep-ml.com
  profile     View your profile
  submit      Submit a solution to a problem

Flags:
      --config string   config file (default is $HOME/.deep-ml.yaml)
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
Logged in successfully as your_username
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

### Submit Solution

```bash
$ deep-ml submit 1 ./solution.py
Submitting solution...
Submission ID: abcd1234
Status: Accepted
Runtime: 45ms
Memory: 14.2MB
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