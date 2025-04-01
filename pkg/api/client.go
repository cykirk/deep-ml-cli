package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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

	// Get user profile from deep-ml API using the token
	userProfile, err := c.GetUserProfile()
	if err != nil {
		// Non-critical error, we can still proceed with logged-in state
		authResp.User = models.User{
			Username: email, // Use email as username for now
		}
	} else {
		authResp.User = *userProfile
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
		ID               string              `json:"id"`
		Title            string              `json:"title"`
		Description      string              `json:"description"`      // base64 encoded
		Difficulty       string              `json:"difficulty"`
		Category         string              `json:"category"`
		Solution         string              `json:"solution"`         // base64 encoded
		TestCases        []map[string]string `json:"test_cases"`
		StarterCode      string              `json:"starter_code"`
		LearnSection     string              `json:"learn_section"`    // base64 encoded
		Example          map[string]string   `json:"example"`
		Likes            string              `json:"likes"`
		Dislikes         string              `json:"dislikes"`
		Video            string              `json:"video,omitempty"`
		Contributors     []map[string]string `json:"contributor,omitempty"`
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
		Likes:       encodedResp.Likes,
		Dislikes:    encodedResp.Dislikes,
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

// SubmitSolution submits a solution to a problem
func (c *Client) SubmitSolution(submission *models.Submission) (*models.SubmissionResult, error) {
	resp, err := c.post("/submissions", submission)
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
