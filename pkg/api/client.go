package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cy/deep-ml-cli/pkg/models"
)

const (
	baseURL = "https://api.deep-ml.com"
	// Google Identity Toolkit API key and URL (replace with actual key when available)
	identityToolkitURL = "https://identitytoolkit.googleapis.com/v1"
	apiKey             = "AIzaSyCucvOHhInU5zY7YjHRatcgQ9hRgKh8Atc" // Replace with your Firebase API key
	secureTokenURL     = "https://securetoken.googleapis.com/v1"
)

// Client represents an API client for deep-ml.com
type Client struct {
	httpClient   *http.Client
	token        string
	refreshToken string
	expiresAt    time.Time
}

// NewClient creates a new API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the authentication token and related fields
func (c *Client) SetToken(token, refreshToken, expiresIn string) {
	c.token = token
	c.refreshToken = refreshToken

	// Parse expiration time
	expirySeconds, err := time.ParseDuration(expiresIn + "s")
	if err == nil {
		c.expiresAt = time.Now().Add(expirySeconds)
	}
}

// GetToken returns the current authentication token
func (c *Client) GetToken() string {
	return c.token
}

// GetRefreshToken returns the current refresh token
func (c *Client) GetRefreshToken() string {
	return c.refreshToken
}

// IsTokenExpired checks if the current token is expired
func (c *Client) IsTokenExpired() bool {
	// Add a small buffer (5 minutes) to ensure we refresh before actual expiry
	buffer := 5 * time.Minute
	return time.Now().Add(buffer).After(c.expiresAt)
}

// Login authenticates a user using Google Identity Toolkit and returns a token
func (c *Client) Login(email, password string) (*models.AuthResponse, error) {
	// Create sign-in request payload
	signInReq := models.GoogleSignInRequest{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}

	// URL for email/password authentication
	url := fmt.Sprintf("%s/accounts:signInWithPassword?key=%s", identityToolkitURL, apiKey)

	// Marshal the request body
	bodyBytes, err := json.Marshal(signInReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling login request: %w", err)
	}

	// Create and execute the request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing login request: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("login failed with status code %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("login failed: %s (code: %d)", errorResp.Error.Message, errorResp.Error.Code)
	}

	// Parse the successful response
	var authResp models.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("error parsing login response: %w", err)
	}

	// Store token information
	c.SetToken(authResp.Token, authResp.RefreshToken, authResp.ExpiresIn)
	
	// If LocalID is not provided directly in the sign-in response, get it from account info
	if authResp.LocalID == "" {
		// Get account info to obtain the user ID
		fmt.Println("Getting account information...")
		accountInfo, err := c.GetAccountInfo(authResp.Token)
		if err == nil && len(accountInfo.Users) > 0 {
			authResp.LocalID = accountInfo.Users[0].LocalID
		}
	}

	// Get user profile from deep-ml API using the token
	userProfile, err := c.GetUserProfile()
	if err != nil {
		// Non-critical error, we can still proceed with logged-in state
		authResp.User = models.User{
			Username: email, // Use email as username for now
			UserID:   authResp.LocalID, // Set user ID from Firebase auth
		}
	} else {
		authResp.User = *userProfile
		// If the user profile doesn't have a UserID, set it from Firebase
		if authResp.User.UserID == "" {
			authResp.User.UserID = authResp.LocalID
		}
	}

	return &authResp, nil
}

// GetProblems returns a list of problems
func (c *Client) GetProblems() (*models.ProblemList, error) {
	resp, err := c.get("/list-problems")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var problems models.ProblemList
	if err := json.NewDecoder(resp.Body).Decode(&problems); err != nil {
		return nil, err
	}

	return &problems, nil
}

// GetProblem returns a specific problem by ID
func (c *Client) GetProblem(id string) (*models.Problem, error) {
	resp, err := c.get(fmt.Sprintf("/fetch-problem?problem_id=%s", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response into a special format first
	var encodedResp struct {
		ID               string                 `json:"id"`
		Title            string                 `json:"title"`
		Description      string                 `json:"description"`      // base64 encoded
		Difficulty       string                 `json:"difficulty"`
		Category         string                 `json:"category"`
		Solution         string                 `json:"solution"`         // base64 encoded
		TestCases        []map[string]string    `json:"test_cases"`
		StarterCode      string                 `json:"starter_code"`
		LearnSection     string                 `json:"learn_section"`    // base64 encoded
		Example          map[string]string      `json:"example"`
		Likes            json.RawMessage        `json:"likes"`           // Can be either string or number
		Dislikes         json.RawMessage        `json:"dislikes"`        // Can be either string or number
		Video            string                 `json:"video,omitempty"`
		Contributors     []map[string]string    `json:"contributor,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&encodedResp); err != nil {
		return nil, fmt.Errorf("error decoding problem response: %w", err)
	}

	// Decode base64 encoded fields
	description, err := base64Decode(encodedResp.Description)
	if err != nil {
		return nil, fmt.Errorf("error decoding description: %w", err)
	}

	solution, err := base64Decode(encodedResp.Solution)
	if err != nil {
		return nil, fmt.Errorf("error decoding solution: %w", err)
	}

	learnSection, err := base64Decode(encodedResp.LearnSection)
	if err != nil {
		return nil, fmt.Errorf("error decoding learn section: %w", err)
	}

	// Extract test cases
	testCases := make([]string, 0, len(encodedResp.TestCases))
	for _, tc := range encodedResp.TestCases {
		testCases = append(testCases, fmt.Sprintf("Test: %s\nExpected: %s", tc["test"], tc["expected_output"]))
	}

	// Handle likes and dislikes which can be either string or number
	likes := "0"
	dislikes := "0"
	
	// Try to parse likes
	if len(encodedResp.Likes) > 0 {
		// Check if it's a JSON string (has quotes)
		if encodedResp.Likes[0] == '"' {
			// It's a string, remove the quotes
			var likesStr string
			if err := json.Unmarshal(encodedResp.Likes, &likesStr); err == nil {
				likes = likesStr
			}
		} else {
			// It's a number
			var likesNum int
			if err := json.Unmarshal(encodedResp.Likes, &likesNum); err == nil {
				likes = fmt.Sprintf("%d", likesNum)
			}
		}
	}
	
	// Try to parse dislikes
	if len(encodedResp.Dislikes) > 0 {
		// Check if it's a JSON string (has quotes)
		if encodedResp.Dislikes[0] == '"' {
			// It's a string, remove the quotes
			var dislikesStr string
			if err := json.Unmarshal(encodedResp.Dislikes, &dislikesStr); err == nil {
				dislikes = dislikesStr
			}
		} else {
			// It's a number
			var dislikesNum int
			if err := json.Unmarshal(encodedResp.Dislikes, &dislikesNum); err == nil {
				dislikes = fmt.Sprintf("%d", dislikesNum)
			}
		}
	}

	// Build the final problem object
	problem := &models.Problem{
		ID:          encodedResp.ID,
		Title:       encodedResp.Title,
		Description: description,
		Difficulty:  encodedResp.Difficulty,
		Category:    encodedResp.Category,
		TestCases:   testCases,
		Starter:     encodedResp.StarterCode,
		LearnSection: learnSection,
		Example:     encodedResp.Example,
		Likes:       likes,
		Dislikes:    dislikes,
		Video:       encodedResp.Video,
		Solution:    solution,
	}

	return problem, nil
}

// base64Decode decodes a base64 encoded string
func base64Decode(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// SubmitSolution submits a solution to a problem and returns a task ID
func (c *Client) SubmitSolution(submission *models.Submission) (*models.SubmissionResponse, error) {
	// Use the execute-code endpoint with user_id as a query parameter
	url := fmt.Sprintf("/execute-code?user_id=%s", submission.UserID)
	
	resp, err := c.post(url, submission)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body into a buffer
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse the response as a task ID
	var taskResponse models.SubmissionResponse
	if err := json.Unmarshal(bodyBytes, &taskResponse); err != nil {
		// If we couldn't parse as a task response, try parsing as an error object
		var errorObj struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		
		if err := json.Unmarshal(bodyBytes, &errorObj); err == nil {
			errorMessage := errorObj.Error
			if errorMessage == "" {
				errorMessage = errorObj.Message
			}
			
			if errorMessage != "" {
				return nil, fmt.Errorf("submission failed: %s", errorMessage)
			}
		}

		// If we get here, the response format is unexpected
		return nil, fmt.Errorf("unexpected submission response format: %s", string(bodyBytes))
	}

	return &taskResponse, nil
}

// GetSubmissionResult fetches the test results for a completed submission
// This can be called multiple times until a result is available or max attempts is reached
// The function tries multiple potential endpoints because the deep-ml.com platform
// may use Firestore or another real-time database for delivering results
func (c *Client) GetSubmissionResult(taskID string, waitSeconds int, maxAttempts int) (*models.SubmissionResult, error) {
	// Attempt to fetch submission results from multiple possible endpoints
	// First try the direct submission result endpoint
	primaryUrl := fmt.Sprintf("/submission-result/%s", taskID)
	// Fallback to possible Firestore-like endpoint
	fallbackUrl := fmt.Sprintf("/execute-result/%s", taskID)
	// Start with primary URL
	url := primaryUrl
	// We'll alternate between URLs on subsequent attempts
	useAlternateUrl := false
	
	// Try to fetch the results, with retries
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Make the request
		resp, err := c.get(url)
		if err != nil {
			lastErr = err
			// Alternate between URLs on retry
			useAlternateUrl = !useAlternateUrl
			if useAlternateUrl {
				url = fallbackUrl
			} else {
				url = primaryUrl
			}
			// Wait and retry
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}
		
		// Read the response
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("error reading response body: %w", err)
			// Alternate between URLs on retry
			useAlternateUrl = !useAlternateUrl
			if useAlternateUrl {
				url = fallbackUrl
			} else {
				url = primaryUrl
			}
			// Wait and retry
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}
		
		// First, try parsing the response as an array of test cases directly
		// This matches the format in the submit-response file
		var testCases []struct {
			TestCase       string `json:"test_case"`
			ExpectedOutput string `json:"expected_output"`
			ActualOutput   string `json:"actual_output"`
			Passed         bool   `json:"passed"`
		}
		
		if err := json.Unmarshal(bodyBytes, &testCases); err == nil && len(testCases) > 0 {
			// Successfully parsed as test cases
			result := &models.SubmissionResult{
				TestCases: make([]struct {
					TestCase       string `json:"test_case"`
					ExpectedOutput string `json:"expected_output"`
					ActualOutput   string `json:"actual_output"`
					Passed         bool   `json:"passed"`
				}, len(testCases)),
			}
			
			// Check if all tests passed
			allPassed := true
			for i, tc := range testCases {
				result.TestCases[i] = tc
				if !tc.Passed {
					allPassed = false
				}
			}
			
			if allPassed {
				result.Status = "accepted"
			} else {
				result.Status = "failed"
			}
			
			return result, nil
		}
		
		// Try parsing the response as a full SubmissionResult object
		var fullResult models.SubmissionResult
		if err := json.Unmarshal(bodyBytes, &fullResult); err == nil && (len(fullResult.TestCases) > 0 || fullResult.Status != "") {
			return &fullResult, nil
		}
		
		// Parse response status to check if it's still processing
		var statusResp struct {
			Status string `json:"status"`
		}
		
		if err := json.Unmarshal(bodyBytes, &statusResp); err == nil {
			if statusResp.Status == "processing" {
				// Still processing, wait and retry
				time.Sleep(time.Duration(waitSeconds) * time.Second)
				continue
			}
		}
		
		// Try parsing as an error object
		var errorObj struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		
		if err := json.Unmarshal(bodyBytes, &errorObj); err == nil {
			errorMessage := errorObj.Error
			if errorMessage == "" {
				errorMessage = errorObj.Message
			}
			
			if errorMessage != "" {
				// Create a result with the error message
				result := &models.SubmissionResult{
					Status:  "error",
					Message: errorMessage,
				}
				return result, nil
			}
		}
		
		// If the response is empty, we're probably still waiting for results
		if len(bodyBytes) == 0 || string(bodyBytes) == "null" || string(bodyBytes) == "{}" {
			fmt.Println("No results available yet, waiting...")
			// Alternate between URLs on retry
			useAlternateUrl = !useAlternateUrl
			if useAlternateUrl {
				url = fallbackUrl
			} else {
				url = primaryUrl
			}
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}
		
		// Unexpected response format, wait and retry
		lastErr = fmt.Errorf("unexpected response format: %s", string(bodyBytes))
		// Alternate between URLs on retry
		useAlternateUrl = !useAlternateUrl
		if useAlternateUrl {
			url = fallbackUrl
		} else {
			url = primaryUrl
		}
		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}
	
	// If we get here, we've exceeded max attempts
	if lastErr != nil {
		return nil, fmt.Errorf("maximum attempts exceeded: %w", lastErr)
	}
	
	return nil, fmt.Errorf("maximum attempts exceeded, no results available")
}

// RunLocalTests executes test cases locally for a given problem and user code.
// It returns a SubmissionResult with test case results without sending to the server.
func (c *Client) RunLocalTests(problemID, userCode string) (*models.SubmissionResult, error) {
	// Fetch the problem to get test cases
	problem, err := c.GetProblem(problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem: %w", err)
	}

	// If there are no test cases, return an error
	if len(problem.TestCases) == 0 {
		return nil, fmt.Errorf("no test cases found for problem %s", problemID)
	}

	// Parse the test cases from the problem
	result := &models.SubmissionResult{
		TestCases: make([]struct {
			TestCase       string `json:"test_case"`
			ExpectedOutput string `json:"expected_output"`
			ActualOutput   string `json:"actual_output"`
			Passed         bool   `json:"passed"`
		}, 0),
		Status: "processing",
	}

	// Execute each test case
	for _, tcRaw := range problem.TestCases {
		// Parse the test case text into TestCase and ExpectedOutput
		var testCase, expectedOutput string
		
		// Format: "Test: <testcase>\nExpected: <expected_output>"
		lines := strings.Split(tcRaw, "\n")
		if len(lines) >= 2 {
			testCase = strings.TrimPrefix(lines[0], "Test: ")
			expectedOutput = strings.TrimPrefix(lines[1], "Expected: ")
		} else {
			// If format is unexpected, use the raw text
			testCase = tcRaw
			expectedOutput = ""
		}

		// Execute the code with the test case
		// The test case will typically have code that calls the user's function and prints the result
		actualOutput, err := executeCode(userCode, testCase)
		
		// Create a test case result
		tc := struct {
			TestCase       string `json:"test_case"`
			ExpectedOutput string `json:"expected_output"`
			ActualOutput   string `json:"actual_output"`
			Passed         bool   `json:"passed"`
		}{
			TestCase:       testCase,
			ExpectedOutput: expectedOutput,
			ActualOutput:   actualOutput,
		}

		// Normalize line endings and whitespace for comparison
		normalizedActual := strings.TrimSpace(actualOutput)
		normalizedExpected := strings.TrimSpace(expectedOutput)
		
		// Check if test passed (handle case where there was an error during execution)
		if err == nil {
			// Compare the normalized outputs
			tc.Passed = normalizedActual == normalizedExpected
		} else {
			tc.Passed = false
			tc.ActualOutput = fmt.Sprintf("Error: %v", err)
		}

		result.TestCases = append(result.TestCases, tc)
	}

	// Calculate the final result status
	allPassed := true
	for _, tc := range result.TestCases {
		if !tc.Passed {
			allPassed = false
			break
		}
	}

	if allPassed {
		result.Status = "accepted"
	} else {
		result.Status = "failed"
	}

	return result, nil
}

// executeCode runs a Python script with the given test input and returns the output.
// This version combines the user's solution code with the test case to execute it properly.
func executeCode(userCode, testInput string) (string, error) {
	// Create a temporary file that combines the user code and test input
	tmpFile, err := os.CreateTemp("", "deepml-*.py")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	
	// Combine user code with the test case
	// We add the user's code, then the test case which will usually call functions defined in the user code
	combinedCode := userCode + "\n\n# Test case\n" + testInput
	
	if _, err := tmpFile.WriteString(combinedCode); err != nil {
		return "", fmt.Errorf("failed to write code to temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	
	// Run the command with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "python", tmpFile.Name())
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("execution timed out after 10 seconds")
	}
	
	if err != nil {
		// If there's stderr output, return that as the error message
		if stderr.Len() > 0 {
			return "", fmt.Errorf("execution error: %s", stderr.String())
		}
		return "", fmt.Errorf("execution error: %w", err)
	}
	
	// If there's stdout, return that as the output
	if stdout.Len() > 0 {
		return stdout.String(), nil
	}
	
	// If there's no stdout but there is stderr (but no error), return stderr as output
	if stderr.Len() > 0 {
		return stderr.String(), nil
	}
	
	// No output
	return "", nil
}

// GetSubmission retrieves a specific submission by ID
func (c *Client) GetSubmission(id string) (*models.SubmissionResult, error) {
	resp, err := c.get(fmt.Sprintf("/submissions/%s", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.SubmissionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserProfile gets the current user's profile
func (c *Client) GetUserProfile() (*models.User, error) {
	resp, err := c.get("/user/profile")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAccountInfo gets the user account information from Firebase
func (c *Client) GetAccountInfo(idToken string) (*struct {
	Users []struct {
		LocalID string `json:"localId"`
		Email   string `json:"email"`
	} `json:"users"`
}, error) {
	// URL for getting account info
	url := fmt.Sprintf("%s/accounts:lookup?key=%s", identityToolkitURL, apiKey)

	// Create request payload
	reqBody := struct {
		IDToken string `json:"idToken"`
	}{
		IDToken: idToken,
	}

	// Marshal the request body
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling account info request: %w", err)
	}

	// Create and execute the request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating account info request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing account info request: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("account info request failed with status code %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("account info request failed: %s (code: %d)", errorResp.Error.Message, errorResp.Error.Code)
	}

	// Parse the successful response
	var accountInfo struct {
		Users []struct {
			LocalID string `json:"localId"`
			Email   string `json:"email"`
		} `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accountInfo); err != nil {
		return nil, fmt.Errorf("error parsing account info response: %w", err)
	}

	return &accountInfo, nil
}

// RefreshToken refreshes an expired token using the refresh token
func (c *Client) RefreshToken() error {
	if c.refreshToken == "" {
		return errors.New("no refresh token available")
	}

	// Create refresh token request payload
	refreshReq := models.GoogleRefreshTokenRequest{
		RefreshToken: c.refreshToken,
		GrantType:    "refresh_token",
	}

	// URL for token refresh
	url := fmt.Sprintf("%s/token?key=%s", secureTokenURL, apiKey)

	// Marshal the request body
	bodyBytes, err := json.Marshal(refreshReq)
	if err != nil {
		return fmt.Errorf("error marshaling refresh request: %w", err)
	}

	// Create and execute the request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("error creating refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing refresh request: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return fmt.Errorf("token refresh failed with status code %d", resp.StatusCode)
		}
		return fmt.Errorf("token refresh failed: %s (code: %d)", errorResp.Error.Message, errorResp.Error.Code)
	}

	// Parse the successful response
	var refreshResp struct {
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    string `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		return fmt.Errorf("error parsing refresh response: %w", err)
	}

	// Update token information
	c.SetToken(refreshResp.IDToken, refreshResp.RefreshToken, refreshResp.ExpiresIn)
	return nil
}

// request helpers
func (c *Client) get(path string) (*http.Response, error) {
	// Check if token needs to be refreshed
	if c.IsTokenExpired() && c.refreshToken != "" {
		if err := c.RefreshToken(); err != nil {
			// Log but don't fail - we'll try with the existing token
			fmt.Fprintf(os.Stderr, "Warning: Failed to refresh token: %v\n", err)
		}
	}
	return c.request(http.MethodGet, path, nil)
}

func (c *Client) post(path string, body interface{}) (*http.Response, error) {
	// Check if token needs to be refreshed
	if c.IsTokenExpired() && c.refreshToken != "" {
		if err := c.RefreshToken(); err != nil {
			// Log but don't fail - we'll try with the existing token
			fmt.Fprintf(os.Stderr, "Warning: Failed to refresh token: %v\n", err)
		}
	}
	return c.request(http.MethodPost, path, body)
}

func (c *Client) request(method, path string, body interface{}) (*http.Response, error) {
	url := baseURL + path

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errorResp struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		resp.Body.Close()
		return nil, errors.New(errorResp.Message)
	}

	return resp, nil
}