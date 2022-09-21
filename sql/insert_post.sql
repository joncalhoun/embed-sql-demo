INSERT INTO posts (user_id, title, markdown)
VALUES ($1, $2, $3)
RETURNING id;
