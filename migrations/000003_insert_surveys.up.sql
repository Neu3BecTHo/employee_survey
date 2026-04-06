-- Insert test surveys only if they don't exist
INSERT INTO surveys (title, description, status) 
SELECT * FROM (VALUES 
    ('Обратная связь по проекту', 'Как вы оцениваете коммуникацию в команде?', 'open'),
    ('Оценка рабочих условий', 'Насколько вам комфортно работать в офисе?', 'draft'),
    ('Техническое оснащение', 'Достаточно ли вам оборудования для работы?', 'closed')
) AS v(title, description, status)
WHERE NOT EXISTS (SELECT 1 FROM surveys WHERE title = v.title);
