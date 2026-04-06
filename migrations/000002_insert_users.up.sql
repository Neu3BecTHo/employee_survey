-- Insert initial users only if they don't exist
INSERT INTO users (name, role) 
SELECT * FROM (VALUES 
    ('Иван Иванов', 'admin'),
    ('Петр Петров', 'employee'),
    ('Анна Сидорова', 'employee')
) AS v(name, role)
WHERE NOT EXISTS (SELECT 1 FROM users WHERE name = v.name);
