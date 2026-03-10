package models

import (
	"time"
)

// User represents a system user
type User struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Survey represents a survey
type Survey struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// SurveyQuestion represents a question in a survey
type SurveyQuestion struct {
	ID         int             `json:"id" db:"id"`
	SurveyID   int             `json:"survey_id" db:"survey_id"`
	Text       string          `json:"text" db:"text"`
	Type       string          `json:"type" db:"type"`
	IsRequired bool            `json:"is_required" db:"is_required"`
	Options    []string        `json:"options,omitempty" db:"options"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// SurveyResponse represents a user's response to a survey
type SurveyResponse struct {
	ID          int       `json:"id" db:"id"`
	SurveyID    int       `json:"survey_id" db:"survey_id"`
	UserID      int       `json:"user_id" db:"user_id"`
	SubmittedAt time.Time `json:"submitted_at" db:"submitted_at"`
}

// SurveyAnswer represents an answer to a specific question
type SurveyAnswer struct {
	ID          int       `json:"id" db:"id"`
	ResponseID  int       `json:"response_id" db:"response_id"`
	QuestionID  int       `json:"question_id" db:"question_id"`
	Value       string    `json:"value" db:"value"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Request/Response DTOs

// CreateSurveyRequest represents the request to create a survey
type CreateSurveyRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

// CreateQuestionRequest represents the request to create a question
type CreateQuestionRequest struct {
	Text       string   `json:"text" validate:"required,min=1,max=1000"`
	Type       string   `json:"type" validate:"required,oneof=single_choice text"`
	IsRequired bool     `json:"is_required"`
	Options    []string `json:"options,omitempty" validate:"required_if=Type single_choice"`
}

// SubmitResponseRequest represents the request to submit survey responses
type SubmitResponseRequest struct {
	Answers []AnswerInput `json:"answers" validate:"required,min=1"`
}

// AnswerInput represents a single answer input
type AnswerInput struct {
	QuestionID int    `json:"question_id" validate:"required"`
	Value      string `json:"value" validate:"required"`
}

// SurveyWithQuestions represents a survey with its questions
type SurveyWithQuestions struct {
	Survey
	Questions []SurveyQuestion `json:"questions"`
}

// SurveyResults represents aggregated results for a survey
type SurveyResults struct {
	Survey       Survey                    `json:"survey"`
	TotalResponses int                      `json:"total_responses"`
	QuestionResults []QuestionResult         `json:"question_results"`
}

// QuestionResult represents results for a specific question
type QuestionResult struct {
	Question SurveyQuestion `json:"question"`
	Answers  []AnswerCount  `json:"answers,omitempty"`
}

// AnswerCount represents count of answers for text questions
type AnswerCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}
