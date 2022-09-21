package main

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	_ "embed"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	//go:embed sql/insert_user.sql
	insertUserSQL string
	//go:embed sql/insert_post.sql
	insertPostSQL string
	//go:embed sql/*.sql
	sqlQueriesFS embed.FS
)

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer db.Close()

	qs, err := NewQueryStore(sqlQueriesFS, "sql/")
	if err != nil {
		return err
	}

	// Run this once then comment it out and hardcode the userID
	userID, err := seed(db, "jon@calhoun.io")
	if err != nil {
		return fmt.Errorf("seed: %w", err)
	}

	rows, err := db.Query(qs.Q("user_posts"), userID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var email, title string
		var postID int
		err = rows.Scan(&email, &postID, &title)
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		fmt.Fprintf(stdout, "%s wrote %q (id:%d)\n", email, title, postID)
	}

	return nil
}

func seed(db *sql.DB, email string) (userID int, err error) {
	err = db.QueryRow(insertUserSQL, email).Scan(&userID)
	if err != nil {
		return -1, fmt.Errorf("inserting user: %w", err)
	}

	for i := 0; i < 10; i++ {
		title := fmt.Sprintf("Awesome Post #%d", i+1)
		markdown := `#WIP

		TODO: Write this post!`
		var postID int

		err = db.QueryRow(insertPostSQL, userID, title, markdown).Scan(&postID)
		if err != nil {
			return -1, fmt.Errorf("inserting post: %w", err)
		}
	}
	return userID, nil
}

func NewQueryStore(filesystem fs.FS, dir string) (QueryStore, error) {
	qs := make(QueryStore)
	pattern := dir + "*.sql"
	// I prefer to read all the queries ahead of time so I don't need to check for errors later.
	matches, err := fs.Glob(filesystem, pattern)
	if err != nil {
		return nil, fmt.Errorf("glob %s: %w", dir, err)
	}
	for _, match := range matches {
		bytes, err := fs.ReadFile(filesystem, match)
		if err != nil {
			return nil, fmt.Errorf("readfile %s: %w", match, err)
		}
		// remove the dir and extension from the name
		name := filepath.Base(match)
		name = name[:len(name)-4]
		qs[name] = string(bytes)
	}
	return qs, nil
}

type QueryStore map[string]string

func (qs QueryStore) Q(name string) string {
	return qs[name]
}
