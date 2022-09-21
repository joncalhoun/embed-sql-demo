SELECT users.email,
  posts.id,
  posts.title
FROM posts
  JOIN users ON posts.user_id = users.id
WHERE users.id = $1;
