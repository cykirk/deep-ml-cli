package models

// Problem represents a problem from deep-ml.com
type Problem struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description,omitempty"`
	Difficulty   string            `json:"difficulty"` // easy, medium, hard
	Category     string            `json:"category"`   // NLP, Machine Learning, Deep Learning, etc.
	TestCases    []string          `json:"testCases,omitempty"`
	Starter      string            `json:"starter_code,omitempty"`  // Starter code for the problem
	LearnSection string            `json:"learn_section,omitempty"` // Educational content
	Example      map[string]string `json:"example,omitempty"`       // Example input/output
	Likes        string            `json:"-"`          // Not directly decoded, handled in custom logic
	Dislikes     string            `json:"-"`          // Not directly decoded, handled in custom logic
	Video        string            `json:"video,omitempty"`    // Video tutorial URL
	Solution     string            `json:"solution,omitempty"` // Solution code
}

// ProblemList represents a list of problems
type ProblemList struct {
	Problems      []Problem `json:"problems"`
	DailyQuestion Problem   `json:"daily_question"`
}

// Submission represents a problem submission
type Submission struct {
	UserCode   string `json:"user_code"`
	ProblemID  string `json:"problem_id"`
	Difficulty string `json:"difficulty"`
	UserID     string `json:"user_id"`
}

// SubmissionResponse represents the initial response from the submission API
type SubmissionResponse struct {
	TaskID string `json:"task_id"`
}

// SubmissionResult represents the result of a submission
type SubmissionResult struct {
	TestCases []struct {
		TestCase       string `json:"test_case"`
		ExpectedOutput string `json:"expected_output"`
		ActualOutput   string `json:"actual_output"`
		Passed         bool   `json:"passed"`
	} `json:""`
	Status    string `json:"status,omitempty"` // For backward compatibility
	ID        string `json:"id,omitempty"`
	Runtime   string `json:"runtime,omitempty"`
	Memory    string `json:"memory,omitempty"`
	Message   string `json:"message,omitempty"` // Error message if any
}

// User represents a user profile
type User struct {
	Username       string `json:"username"`
	Email          string `json:"email,omitempty"`
	UserID         string `json:"user_id,omitempty"` // ID for API calls
	ProblemsTotal  int    `json:"problemsTotal"`
	ProblemsSolved int    `json:"problemsSolved"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token        string `json:"idToken"`      // JWT token from Google Identity Toolkit
	RefreshToken string `json:"refreshToken"` // For refreshing the token
	ExpiresIn    string `json:"expiresIn"`    // Token expiry time in seconds
	LocalID      string `json:"localId"`      // User ID from Firebase Auth
	User         User   `json:"user"`
}

// GoogleSignInRequest represents a sign-in request to Google Identity Toolkit
type GoogleSignInRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

// GoogleRefreshTokenRequest represents a refresh token request
type GoogleRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	GrantType    string `json:"grant_type"`
}
