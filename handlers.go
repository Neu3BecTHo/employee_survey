package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"survey-app/internal"
)

// User handlers
func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := a.getAllUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
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
	rows, err := a.DB.Query(`
		SELECT s.id, s.title, s.description, s.status, s.created_at
		FROM surveys s
		LEFT JOIN survey_responses sr ON s.id = sr.survey_id AND sr.user_id = $1
		WHERE s.status = 'open' OR sr.user_id IS NOT NULL
		ORDER BY s.created_at DESC
	`, userID)
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

	surveyWithQuestions := models.SurveyWithQuestions{
		Survey:    *survey,
		Questions: questions,
	}

	respondWithJSON(w, http.StatusOK, surveyWithQuestions)
}

func (a *App) getSurveyByID(id int) (*models.Survey, error) {
	var s models.Survey
	err := a.DB.QueryRow("SELECT id, title, description, status, created_at FROM surveys WHERE id = $1", id).
		Scan(&s.ID, &s.Title, &s.Description, &s.Status, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (a *App) createSurvey(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSurveyRequest
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

	survey := models.Survey{
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

	var req models.CreateSurveyRequest
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
		"UPDATE surveys SET title = $1, description = $2 WHERE id = $3",
		req.Title, req.Description, id,
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

	var req models.CreateQuestionRequest
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

	err = a.DB.QueryRow(
		"INSERT INTO survey_questions (survey_id, text, type, is_required, options, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		question.SurveyID, question.Text, question.Type, question.IsRequired, question.Options, question.CreatedAt,
	).Scan(&question.ID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, question)
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

	var req models.SubmitResponseRequest
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

	response := models.SurveyResponse{
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

func (a *App) getSurveyResults(surveyID int) (*models.SurveyResults, error) {
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

	var questionResults []models.QuestionResult

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
				var answer models.AnswerCount
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
				answers = append(answers, models.AnswerCount{Value: value, Count: 1})
			}
		}

		questionResults = append(questionResults, models.QuestionResult{
			Question: question,
			Answers:  answers,
		})
	}

	return &models.SurveyResults{
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
		err := rows.Scan(&q.ID, &q.SurveyID, &q.Text, &q.Type, &q.IsRequired, &q.Options, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	return questions, nil
}

func (a *App) validateAnswers(questions []models.SurveyQuestion, answers []models.AnswerInput) error {
	questionMap := make(map[int]models.SurveyQuestion)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	answerMap := make(map[int]models.AnswerInput)
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
			for _, option := range q.Options {
				if a.Value == option {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid option '%s' for question %d", a.Value, a.QuestionID)
			}
		}
	}

	return nil
}
