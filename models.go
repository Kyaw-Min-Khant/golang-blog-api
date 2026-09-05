package main

import "time"

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	ID         int64     `json:"id"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Created_At time.Time `json:"created_at"`
	Updated_At time.Time `json:"updated_at"`
}

type Blog struct {
	ID         int64     `json:"id"`
	User_Id    int64     `json:"user_id"`
	Username   string    `json:"username"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Created_At time.Time `json:"created_at"`
	Updated_At time.Time `json:"updated_at"`
}

type createBlogRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}
