-- Insert questions only if they don't exist
INSERT INTO survey_questions (survey_id, text, type, is_required, options, created_at) 
SELECT v.survey_id, v.text, v.type, v.is_required, v.options::jsonb, CURRENT_TIMESTAMP 
FROM (VALUES 
    (1, 'Как вы оцениваете коммуникацию в команде?', 'single_choice', true, '["Отлично", "Хорошо", "Удовлетворительно", "Плохо"]'),
    (1, 'Что можно улучшить в процессе коммуникации?', 'text', false, NULL),
    (2, 'Насколько вам комфортно работать в офисе?', 'single_choice', true, '["Очень комфортно", "Комфортно", "Некомфортно"]'),
    (3, 'Достаточно ли вам оборудования для работы?', 'text', false, NULL)
) AS v(survey_id, text, type, is_required, options)
WHERE NOT EXISTS (SELECT 1 FROM survey_questions WHERE text = v.text);
