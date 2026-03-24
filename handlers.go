package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	. "survey-app/internal"
	models "survey-app/internal"

	"github.com/gorilla/mux"
)

// User handlers
func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := a.getAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (a *App) getAllUsers() ([]models.User, error) {
	rows, err := a.DB.Query("SELECT id, name, role, created_at FROM users ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Name, &u.Role, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (a *App) isUserAdmin(userID int) (bool, error) {
	var role string
	err := a.DB.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil {
		return false, err
	}
	return role == "admin", nil
}

// Survey handlers
func (a *App) getSurveys(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-Id")
	var surveys []models.Survey
	var err error

	if userIDStr != "" {
		userID, _ := strconv.Atoi(userIDStr)
		surveys, err = a.getSurveysForUser(userID)
	} else {
		surveys, err = a.getAllSurveys()
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, surveys)
}

func (a *App) getAllSurveys() ([]models.Survey, error) {
	rows, err := a.DB.Query("SELECT id, title, description, status, created_at FROM surveys ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var surveys []models.Survey
	for rows.Next() {
		var s models.Survey
		err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Status, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		surveys = append(surveys, s)
	}

	return surveys, nil
}

func (a *App) getSurveysForUser(userID int) ([]models.Survey, error) {
	// Check if user is admin
	isAdmin, err := a.isUserAdmin(userID)
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	if isAdmin {
		// Admin can see all surveys
		rows, err = a.DB.Query(`
			SELECT s.id, s.title, s.description, s.status, s.created_at
			FROM surveys s
			ORDER BY s.created_at DESC
		`)
	} else {
		// Employee can only see open surveys or surveys they've responded to
		rows, err = a.DB.Query(`
			SELECT s.id, s.title, s.description, s.status, s.created_at
			FROM surveys s
			LEFT JOIN survey_responses sr ON s.id = sr.survey_id AND sr.user_id = $1
			WHERE s.status = 'open' OR sr.user_id IS NOT NULL
			ORDER BY s.created_at DESC
		`, userID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var surveys []models.Survey
	for rows.Next() {
		var s models.Survey
		err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Status, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		surveys = append(surveys, s)
	}

	return surveys, nil
}

func (a *App) getSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid survey ID")
		return
	}

	survey, err := a.getSurveyByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Survey not found")
		return
	}

	questions, err := a.getQuestionsBySurveyID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Ensure questions is never nil
	if questions == nil {
		questions = []models.SurveyQuestion{}
	}

	surveyWithQuestions := SurveyWithQuestions{
		Survey:    *survey,
		Questions: questions,
	}

	respondWithJSON(w, http.StatusOK, surveyWithQuestions)
}

func (a *App) getSurveyByID(id int) (*Survey, error) {
	var s Survey
	err := a.DB.QueryRow("SELECT id, title, description, status, created_at FROM surveys WHERE id = $1", id).
		Scan(&s.ID, &s.Title, &s.Description, &s.Status, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (a *App) createSurvey(w http.ResponseWriter, r *http.Request) {
	var req CreateSurveyRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate required fields
	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	survey := Survey{
		Title:       req.Title,
		Description: req.Description,
		Status:      "draft",
		CreatedAt:   time.Now(),
	}

	err := a.DB.QueryRow(
		"INSERT INTO surveys (title, description, status, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		survey.Title, survey.Description, survey.Status, survey.CreatedAt,
	).Scan(&survey.ID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, survey)
}

func (a *App) updateSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid survey ID")
		return
	}

	var req CreateSurveyRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if req.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	_, err = a.DB.Exec(
		"UPDATE surveys SET title = $1, description = $2, status = $3 WHERE id = $4",
		req.Title, req.Description, req.Status, id,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	survey, err := a.getSurveyByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, survey)
}

func (a *App) createQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	surveyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid survey ID")
		return
	}

	var req CreateQuestionRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate
	if req.Text == "" {
		respondWithError(w, http.StatusBadRequest, "Question text is required")
		return
	}

	if req.Type != "single_choice" && req.Type != "text" {
		respondWithError(w, http.StatusBadRequest, "Invalid question type")
		return
	}

	if req.Type == "single_choice" && len(req.Options) == 0 {
		respondWithError(w, http.StatusBadRequest, "Options are required for single choice questions")
		return
	}

	// Check if survey exists and is in draft status
	var status string
	err = a.DB.QueryRow("SELECT status FROM surveys WHERE id = $1", surveyID).Scan(&status)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Survey not found")
		return
	}

	if status != "draft" {
		respondWithError(w, http.StatusBadRequest, "Can only add questions to draft surveys")
		return
	}

	question := models.SurveyQuestion{
		SurveyID:   surveyID,
		Text:       req.Text,
		Type:       req.Type,
		IsRequired: req.IsRequired,
		Options:    req.Options,
		CreatedAt:  time.Now(),
	}

	// Convert options to JSON for JSONB column
	var optionsJSON []byte
	if len(question.Options) > 0 {
		optionsJSON, err = json.Marshal(question.Options)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	err = a.DB.QueryRow(
		"INSERT INTO survey_questions (survey_id, text, type, is_required, options, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		question.SurveyID, question.Text, question.Type, question.IsRequired, optionsJSON, question.CreatedAt,
	).Scan(&question.ID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, question)
}

func (a *App) updateQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	surveyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid survey ID")
		return
	}

	questionID, err := strconv.Atoi(vars["questionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}

	var req models.SurveyQuestion
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate request
	if req.Text == "" {
		respondWithError(w, http.StatusBadRequest, "Question text is required")
		return
	}

	if req.Type != "text" && req.Type != "single_choice" {
		respondWithError(w, http.StatusBadRequest, "Invalid question type")
		return
	}

	if req.Type == "single_choice" && len(req.Options) == 0 {
		respondWithError(w, http.StatusBadRequest, "Options are required for single choice questions")
		return
	}

	// Check if survey is in draft status
	var status string
	err = a.DB.QueryRow("SELECT status FROM surveys WHERE id = $1", surveyID).Scan(&status)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if status != "draft" {
		respondWithError(w, http.StatusBadRequest, "Can only update questions in draft surveys")
		return
	}

	// Update question
	optionsJSON, err := json.Marshal(req.Options)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = a.DB.Exec(
		"UPDATE survey_questions SET text = $1, type = $2, is_required = $3, options = $4 WHERE id = $5 AND survey_id = $6",
		req.Text, req.Type, req.IsRequired, optionsJSON, questionID, surveyID,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get updated question
	updatedQuestion, err := a.getQuestionByID(questionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, updatedQuestion)
}

func (a *App) deleteQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	surveyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid survey ID")
		return
	}

	questionID, err := strconv.Atoi(vars["questionId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}

	// Check if survey is in draft status
	var status string
	err = a.DB.QueryRow("SELECT status FROM surveys WHERE id = $1", surveyID).Scan(&status)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if status != "draft" {
		respondWithError(w, http.StatusBadRequest, "Can only delete questions from draft surveys")
		return
	}

	// Delete question
	_, err = a.DB.Exec("DELETE FROM survey_questions WHERE id = $1 AND survey_id = $2", questionID, surveyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Question deleted successfully"})
}

func (a *App) getQuestionByID(questionID int) (*models.SurveyQuestion, error) {
	var q models.SurveyQuestion
	var optionsJSON []byte

	err := a.DB.QueryRow(
		"SELECT id, survey_id, text, type, is_required, options, created_at FROM survey_questions WHERE id = $1",
		questionID,
	).Scan(&q.ID, &q.SurveyID, &q.Text, &q.Type, &q.IsRequired, &optionsJSON, &q.CreatedAt)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(optionsJSON, &q.Options); err != nil {
		q.Options = []string{}
	}

	return &q, nil
}

func (a *App) openSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid survey ID")
		return
	}

	_, err = a.DB.Exec("UPDATE surveys SET status = 'open' WHERE id = $1 AND status = 'draft'", id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	survey, err := a.getSurveyByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, survey)
}

func (a *App) closeSurvey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid survey ID")
		return
	}

	_, err = a.DB.Exec("UPDATE surveys SET status = 'closed' WHERE id = $1 AND status = 'open'", id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	survey, err := a.getSurveyByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, survey)
}

func (a *App) submitResponse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	surveyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid survey ID")
		return
	}

	userIDStr := r.Header.Get("X-User-Id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req SubmitResponseRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Validate survey is open
	var status string
	err = a.DB.QueryRow("SELECT status FROM surveys WHERE id = $1", surveyID).Scan(&status)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Survey not found")
		return
	}

	if status != "open" {
		respondWithError(w, http.StatusBadRequest, "Survey is not open")
		return
	}

	// Check if user already responded
	var existingResponseID int
	err = a.DB.QueryRow("SELECT id FROM survey_responses WHERE survey_id = $1 AND user_id = $2", surveyID, userID).Scan(&existingResponseID)
	if err == nil {
		respondWithError(w, http.StatusConflict, "User has already responded to this survey")
		return
	}

	// Get questions for validation
	questions, err := a.getQuestionsBySurveyID(surveyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Validate answers
	if err := a.validateAnswers(questions, req.Answers); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Begin transaction
	tx, err := a.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	// Insert response
	var responseID int
	err = tx.QueryRow(
		"INSERT INTO survey_responses (survey_id, user_id, submitted_at) VALUES ($1, $2, $3) RETURNING id",
		surveyID, userID, time.Now(),
	).Scan(&responseID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Insert answers
	for _, answer := range req.Answers {
		_, err = tx.Exec(
			"INSERT INTO survey_answers (response_id, question_id, value) VALUES ($1, $2, $3)",
			responseID, answer.QuestionID, answer.Value,
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := SurveyResponse{
		ID:          responseID,
		SurveyID:    surveyID,
		UserID:      userID,
		SubmittedAt: time.Now(),
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (a *App) getResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	surveyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid survey ID")
		return
	}

	results, err := a.getSurveyResults(surveyID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, results)
}

func (a *App) getSurveyResults(surveyID int) (*SurveyResults, error) {
	// Get survey
	survey, err := a.getSurveyByID(surveyID)
	if err != nil {
		return nil, err
	}

	// Get total responses
	var totalResponses int
	err = a.DB.QueryRow("SELECT COUNT(*) FROM survey_responses WHERE survey_id = $1", surveyID).Scan(&totalResponses)
	if err != nil {
		return nil, err
	}

	// Get questions
	questions, err := a.getQuestionsBySurveyID(surveyID)
	if err != nil {
		return nil, err
	}

	var questionResults []QuestionResult

	for _, question := range questions {
		var answers []models.AnswerCount

		if question.Type == "single_choice" {
			// For single choice, count occurrences of each option
			rows, err := a.DB.Query(`
				SELECT value, COUNT(*) as count
				FROM survey_answers
				WHERE question_id = $1
				GROUP BY value
				ORDER BY count DESC
			`, question.ID)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			for rows.Next() {
				var answer AnswerCount
				err := rows.Scan(&answer.Value, &answer.Count)
				if err != nil {
					return nil, err
				}
				answers = append(answers, answer)
			}
		} else {
			// For text questions, get all answers
			rows, err := a.DB.Query(`
				SELECT value
				FROM survey_answers
				WHERE question_id = $1
				ORDER BY created_at DESC
			`, question.ID)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			for rows.Next() {
				var value string
				err := rows.Scan(&value)
				if err != nil {
					return nil, err
				}
				answers = append(answers, AnswerCount{Value: value, Count: 1})
			}
		}

		questionResults = append(questionResults, QuestionResult{
			Question: question,
			Answers:  answers,
		})
	}

	return &SurveyResults{
		Survey:          *survey,
		TotalResponses:  totalResponses,
		QuestionResults: questionResults,
	}, nil
}

func (a *App) getQuestionsBySurveyID(surveyID int) ([]models.SurveyQuestion, error) {
	rows, err := a.DB.Query("SELECT id, survey_id, text, type, is_required, options, created_at FROM survey_questions WHERE survey_id = $1 ORDER BY id", surveyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.SurveyQuestion
	for rows.Next() {
		var q models.SurveyQuestion
		var optionsJSON []byte

		err := rows.Scan(&q.ID, &q.SurveyID, &q.Text, &q.Type, &q.IsRequired, &optionsJSON, &q.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSONB to []string if not null
		if optionsJSON != nil {
			if err := json.Unmarshal(optionsJSON, &q.Options); err != nil {
				return nil, err
			}
		} else {
			q.Options = []string{}
		}

		questions = append(questions, q)
	}

	return questions, nil
}

func (a *App) getMyResponses(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-Id")
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, "User ID required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	responses, err := a.getUserResponses(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, responses)
}

func (a *App) getResponseDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	responseID, err := strconv.Atoi(vars["responseId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid response ID")
		return
	}

	userIDStr := r.Header.Get("X-User-Id")
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, "User ID required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get response details
	response, err := a.getResponseByID(responseID, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (a *App) getResponseByID(responseID, userID int) (*models.ResponseDetail, error) {
	// Get response info
	var response models.ResponseDetail
	err := a.DB.QueryRow(`
		SELECT sr.id, sr.survey_id, sr.user_id, sr.submitted_at, s.title, s.description
		FROM survey_responses sr
		JOIN surveys s ON sr.survey_id = s.id
		WHERE sr.id = $1 AND sr.user_id = $2
	`, responseID, userID).Scan(&response.ID, &response.SurveyID, &response.UserID, &response.SubmittedAt, &response.Survey.Title, &response.Survey.Description)

	if err != nil {
		return nil, err
	}

	// Get questions
	questions, err := a.getQuestionsBySurveyID(response.SurveyID)
	if err != nil {
		return nil, err
	}

	// Get answers for this response
	rows, err := a.DB.Query(`
		SELECT question_id, value
		FROM survey_answers
		WHERE response_id = $1
	`, responseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []models.SurveyAnswer
	for rows.Next() {
		var answer models.SurveyAnswer
		err := rows.Scan(&answer.QuestionID, &answer.Value)
		if err != nil {
			return nil, err
		}
		answers = append(answers, answer)
	}

	response.Questions = questions
	response.Answers = answers

	return &response, nil
}

func (a *App) getUserResponses(userID int) ([]models.SurveyResponse, error) {
	rows, err := a.DB.Query(`
		SELECT sr.id, sr.survey_id, sr.user_id, sr.submitted_at, s.title, s.description
		FROM survey_responses sr
		JOIN surveys s ON sr.survey_id = s.id
		WHERE sr.user_id = $1
		ORDER BY sr.submitted_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []models.SurveyResponse
	for rows.Next() {
		var response models.SurveyResponse
		err := rows.Scan(&response.ID, &response.SurveyID, &response.UserID, &response.SubmittedAt, &response.SurveyTitle, &response.SurveyDescription)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}

	return responses, nil
}

func (a *App) validateAnswers(questions []models.SurveyQuestion, answers []AnswerInput) error {
	questionMap := make(map[int]models.SurveyQuestion)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	answerMap := make(map[int]AnswerInput)
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}

	// Check required questions are answered
	for _, q := range questions {
		if q.IsRequired {
			if _, exists := answerMap[q.ID]; !exists {
				return fmt.Errorf("required question %d is not answered", q.ID)
			}
		}
	}

	// Check answers are for valid questions and have valid values
	for _, a := range answers {
		q, exists := questionMap[a.QuestionID]
		if !exists {
			return fmt.Errorf("answer for invalid question %d", a.QuestionID)
		}

		if a.Value == "" {
			return fmt.Errorf("empty answer for question %d", a.QuestionID)
		}

		if q.Type == "single_choice" {
			valid := false
			if len(q.Options) > 0 {
				for _, option := range q.Options {
					if a.Value == option {
						valid = true
						break
					}
				}
			}
			if !valid {
				return fmt.Errorf("invalid option '%s' for question %d", a.Value, a.QuestionID)
			}
		}
	}

	return nil
}
