CREATE TABLE posts (
  id SERIAL PRIMARY KEY,
  user_id INT,
  title TEXT,
  markdown TEXT
);
