package main

import (
	"database/sql"
	"errors"
	"time"
)

var ErrBlogNotFound = errors.New("blog not found")

func createUser(db *sql.DB, username, hash string) (int, error) {
	res, err := db.Exec(`INSERT INTO users (username,password) VALUES (?,?)`, username, hash)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func findUserByName(db *sql.DB, username string) (User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, password, created_at FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password, &user.Created_At)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func createBlogByUserId(db *sql.DB, user_id int64, username, title, content string) (Blog, error) {
	res, err := db.Exec(`INSERT INTO blogs (user_id,title,content) VALUES (?,?,?)`, user_id, title, content)
	if err != nil {
		return Blog{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Blog{}, err
	}
	now := time.Now()
	return Blog{ID: id, User_Id: user_id, Username: username, Title: title, Content: content, Created_At: now, Updated_At: now}, nil
}

func getAllBlogs(db *sql.DB, limit, offset int) ([]Blog, error) {
	rows, err := db.Query(`SELECT blogs.id, blogs.user_id, COALESCE(users.username, ''), blogs.title, blogs.content, blogs.created_at
		FROM blogs LEFT JOIN users ON blogs.user_id = users.id
		WHERE blogs.user_id != 0
		ORDER BY blogs.id
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []Blog
	for rows.Next() {
		var blog Blog
		if err := rows.Scan(&blog.ID, &blog.User_Id, &blog.Username, &blog.Title, &blog.Content, &blog.Created_At); err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return blogs, nil
}

func getBlogsByUserId(db *sql.DB, user_id, limit, offset int) ([]Blog, error) {
	rows, err := db.Query(`SELECT blogs.id, blogs.user_id, COALESCE(users.username, ''), blogs.title, blogs.content, blogs.created_at
		FROM blogs LEFT JOIN users ON blogs.user_id = users.id
		WHERE blogs.user_id = ?
		ORDER BY blogs.id
		LIMIT ? OFFSET ?`, user_id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []Blog
	for rows.Next() {
		var blog Blog
		if err := rows.Scan(&blog.ID, &blog.User_Id, &blog.Username, &blog.Title, &blog.Content, &blog.Created_At); err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return blogs, nil
}

func updateBlogByUserId(db *sql.DB, blogId, userId int64, title, content string) (Blog, error) {
	now := time.Now()
	res, err := db.Exec(`UPDATE blogs SET title = ?, content = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		title, content, now, blogId, userId)
	if err != nil {
		return Blog{}, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return Blog{}, err
	}
	if rowsAffected == 0 {
		return Blog{}, ErrBlogNotFound
	}

	return Blog{
		ID:         blogId,
		User_Id:    userId,
		Title:      title,
		Content:    content,
		Updated_At: now,
	}, nil
}

func deleteBlogByUserId(db *sql.DB, blogId, userId int64) error {
	res, err := db.Exec(`DELETE FROM blogs WHERE id = ? AND user_id = ?`, blogId, userId)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrBlogNotFound
	}
	return nil
}
