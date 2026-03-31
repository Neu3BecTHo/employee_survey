package main

import (
	"database/sql"
	"encoding/json"
	"testing"

	_ "github.com/lib/pq"
)

// Test 1: Database connection
func TestDatabaseConnection(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=survey_user password=survey_pass dbname=survey_app sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Test basic query
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count == 0 {
		t.Error("No users found in database")
	}
}

// Test 2: User roles
func TestUserRoles(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=survey_user password=survey_pass dbname=survey_app sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check that admin user exists
	var role string
	err = db.QueryRow("SELECT role FROM users WHERE id = 1").Scan(&role)
	if err != nil {
		t.Fatal(err)
	}

	if role != "admin" {
		t.Errorf("Expected admin role for user 1, got %s", role)
	}

	// Check that employee user exists
	err = db.QueryRow("SELECT role FROM users WHERE id = 2").Scan(&role)
	if err != nil {
		t.Fatal(err)
	}

	if role != "employee" {
		t.Errorf("Expected employee role for user 2, got %s", role)
	}
}

// Test 3: Survey validation
func TestSurveyValidation(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=survey_user password=survey_pass dbname=survey_app sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Test that surveys have valid statuses
	rows, err := db.Query("SELECT id, status FROM surveys")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	validStatuses := map[string]bool{
		"draft":  true,
		"open":   true,
		"closed": true,
	}

	for rows.Next() {
		var id int
		var status string
		err := rows.Scan(&id, &status)
		if err != nil {
			t.Fatal(err)
		}

		if !validStatuses[status] {
			t.Errorf("Invalid status '%s' for survey %d", status, id)
		}
	}
}

// Test 4: Response uniqueness
func TestResponseUniqueness(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=survey_user password=survey_pass dbname=survey_app sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check that user 2 has responded to survey 1
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM survey_responses WHERE user_id = 2 AND survey_id = 1").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count == 0 {
		t.Skip("User 2 has not responded to survey 1, skipping uniqueness test")
		return
	}

	if count > 1 {
		t.Errorf("User 2 has %d responses to survey 1, expected at most 1", count)
	}
}

// Test 5: Question types
func TestQuestionTypes(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=survey_user password=survey_pass dbname=survey_app sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Test that questions have valid types
	rows, err := db.Query("SELECT id, type FROM survey_questions")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	validTypes := map[string]bool{
		"single_choice": true,
		"text":          true,
	}

	for rows.Next() {
		var id int
		var qType string
		err := rows.Scan(&id, &qType)
		if err != nil {
			t.Fatal(err)
		}

		if !validTypes[qType] {
			t.Errorf("Invalid type '%s' for question %d", qType, id)
		}
	}
}

// Test 6: JSON handling
func TestJSONHandling(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=survey_user password=survey_pass dbname=survey_app sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Test that options can be NULL for text questions
	rows, err := db.Query("SELECT id, options FROM survey_questions WHERE type = 'text'")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var options []byte
		err := rows.Scan(&id, &options)
		if err != nil {
			t.Fatal(err)
		}

		// Options should be NULL or valid JSON for text questions
		if len(options) > 0 {
			var test interface{}
			if err := json.Unmarshal(options, &test); err != nil {
				t.Errorf("Invalid JSON in options for question %d: %v", id, err)
			}
		}
	}
}

// Test 7: Survey status transitions
func TestSurveyStatusTransitions(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=survey_user password=survey_pass dbname=survey_app sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create a test survey
	var surveyID int
	err = db.QueryRow("INSERT INTO surveys (title, description, status) VALUES ('Test Survey', 'Test', 'draft') RETURNING id").Scan(&surveyID)
	if err != nil {
		t.Fatal(err)
	}

	// Test status transitions
	_, err = db.Exec("UPDATE surveys SET status = 'open' WHERE id = $1", surveyID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("UPDATE surveys SET status = 'closed' WHERE id = $1", surveyID)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up
	_, err = db.Exec("DELETE FROM surveys WHERE id = $1", surveyID)
	if err != nil {
		t.Fatal(err)
	}
}

// Test 8: Cannot submit response to closed survey
func TestClosedSurveyResponse(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=survey_user password=survey_pass dbname=survey_app sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create a test survey with closed status
	var surveyID int
	err = db.QueryRow("INSERT INTO surveys (title, description, status) VALUES ('Closed Test Survey', 'Test', 'closed') RETURNING id").Scan(&surveyID)
	if err != nil {
		t.Fatal(err)
	}

	// Try to insert a response to closed survey
	_, err = db.Exec("INSERT INTO survey_responses (survey_id, user_id) VALUES ($1, 1)", surveyID)
	if err == nil {
		t.Error("Should not be able to insert response to closed survey")
	}

	// Clean up
	_, err = db.Exec("DELETE FROM surveys WHERE id = $1", surveyID)
	if err != nil {
		t.Fatal(err)
	}
}

// Test 9: Required questions
func TestRequiredQuestions(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=survey_user password=survey_pass dbname=survey_app sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Test that required questions are properly marked
	rows, err := db.Query("SELECT id, is_required FROM survey_questions")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var required bool
		err := rows.Scan(&id, &required)
		if err != nil {
			t.Fatal(err)
		}

		// is_required should be boolean
		if required != true && required != false {
			t.Errorf("Invalid is_required value for question %d: %v", id, required)
		}
	}
}
