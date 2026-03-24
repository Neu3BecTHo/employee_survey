-- Remove test surveys
DELETE FROM surveys WHERE title IN ('Обратная связь по проекту', 'Оценка рабочих условий', 'Техническое оснащение');
