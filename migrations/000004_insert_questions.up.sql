-- Insert questions for surveys
INSERT INTO survey_questions (survey_id, text, type, is_required, options, created_at) VALUES 
    (1, 'Как вы оцениваете коммуникацию в команде?', 'single_choice', true, NOW()),
    (1, 'Что можно улучшить в процессе коммуникации?', 'text', false, NULL, NOW()),
    (2, 'Насколько вам комфортно работать в офисе?', 'single_choice', true, NOW()),
    (3, 'Достаточно ли вам оборудования для работы?', 'text', false, NULL, NOW());
