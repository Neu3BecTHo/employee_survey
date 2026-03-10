package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"survey-app/internal"
	"github.com/gorilla/mux"
)

var a App

func TestMain(m *testing.M) {
	// Setup test database connection
	a = App{}
	a.Initialize(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"),
		5432,
	)

	code := m.Run()

	// Teardown
	a.DB.Close()

	os.Exit(code)
}

func TestGetUsers(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var users []models.User
	json.Unmarshal(response.Body.Bytes(), &users)

	// Should have at least the test users
	if len(users) < 2 {
		t.Errorf("Expected at least 2 users, got %d", len(users))
	}
}

func TestCreateSurvey(t *testing.T) {
	payload := map[string]interface{}{
		"title":       "Test Survey",
		"description": "Test Description",
	}

	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/surveys", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "1") // Admin user

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var survey models.Survey
	json.Unmarshal(response.Body.Bytes(), &survey)

	if survey.Title != "Test Survey" {
		t.Errorf("Expected survey title to be 'Test Survey', got '%s'", survey.Title)
	}
}

func TestCannotSubmitResponseToClosedSurvey(t *testing.T) {
	// First create a survey and close it
	payload := map[string]interface{}{
		"title": "Closed Survey Test",
	}

	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/surveys", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "1")

	response := executeRequest(req)
	var survey models.Survey
	json.Unmarshal(response.Body.Bytes(), &survey)

	// Close the survey
	req2, _ := http.NewRequest("POST", "/surveys/"+string(rune(survey.ID))+"/close", nil)
	req2.Header.Set("X-User-Id", "1")
	executeRequest(req2)

	// Try to submit response to closed survey
	responsePayload := map[string]interface{}{
		"answers": []map[string]interface{}{
			{"question_id": 1, "value": "Test"},
		},
	}

	jsonResponsePayload, _ := json.Marshal(responsePayload)
	req3, _ := http.NewRequest("POST", "/surveys/"+string(rune(survey.ID))+"/responses", bytes.NewBuffer(jsonResponsePayload))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-User-Id", "2") // Employee user

	response3 := executeRequest(req3)
	checkResponseCode(t, http.StatusBadRequest, response3.Code)
}

func TestCannotSubmitDuplicateResponse(t *testing.T) {
	// Create an open survey
	payload := map[string]interface{}{
		"title": "Duplicate Response Test",
	}

	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/surveys", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "1")

	response := executeRequest(req)
	var survey models.Survey
	json.Unmarshal(response.Body.Bytes(), &survey)

	// Open the survey
	req2, _ := http.NewRequest("POST", "/surveys/"+string(rune(survey.ID))+"/open", nil)
	req2.Header.Set("X-User-Id", "1")
	executeRequest(req2)

	// Submit first response
	responsePayload := map[string]interface{}{
		"answers": []map[string]interface{}{
			{"question_id": 1, "value": "Test"},
		},
	}

	jsonResponsePayload, _ := json.Marshal(responsePayload)
	req3, _ := http.NewRequest("POST", "/surveys/"+string(rune(survey.ID))+"/responses", bytes.NewBuffer(jsonResponsePayload))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-User-Id", "2")

	response3 := executeRequest(req3)
	checkResponseCode(t, http.StatusCreated, response3.Code)

	// Try to submit second response from same user
	req4, _ := http.NewRequest("POST", "/surveys/"+string(rune(survey.ID))+"/responses", bytes.NewBuffer(jsonResponsePayload))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("X-User-Id", "2")

	response4 := executeRequest(req4)
	checkResponseCode(t, http.StatusConflict, response4.Code)
}

func TestValidateRequiredQuestions(t *testing.T) {
	// This test would require setting up survey with required questions
	// For now, just ensure the validation logic exists
	questions := []models.SurveyQuestion{
		{ID: 1, IsRequired: true, Type: "text"},
		{ID: 2, IsRequired: false, Type: "text"},
	}

	answers := []models.AnswerInput{
		{QuestionID: 2, Value: "Optional answer"},
		// Missing required question 1
	}

	if validateAnswers(questions, answers) {
		t.Error("Expected validation to fail due to missing required question")
	}

	// Add required answer
	answers = append(answers, models.AnswerInput{QuestionID: 1, Value: "Required answer"})

	if !validateAnswers(questions, answers) {
		t.Error("Expected validation to pass with all required questions answered")
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d, got %d", expected, actual)
	}
}
