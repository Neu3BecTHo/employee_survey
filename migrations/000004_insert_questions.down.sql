-- Remove questions for surveys
DELETE FROM survey_questions WHERE survey_id IN (1, 2, 3);
