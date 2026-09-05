package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultBlogsLimit = 10
	maxBlogsLimit     = 100
)

type api struct {
	db *sql.DB
}

func (a *api) healthCheck(c *gin.Context) {
	respondSuccess(c, http.StatusOK, "Server is running ....")
}

func (a *api) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		slog.Error("hash password failed", "error", err)
		respondError(c, http.StatusInternalServerError, "Could not process password")
		return
	}

	userId, err := createUser(a.db, req.Username, hash)
	if err != nil {
		slog.Error("create user failed", "error", err)
		respondError(c, http.StatusInternalServerError, "Could not create user")
		return
	}
	respondSuccess(c, http.StatusCreated, gin.H{"user_id": userId})
}

func (a *api) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := findUserByName(a.db, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(c, http.StatusUnauthorized, "Invalid username or password")
			return
		}
		slog.Error("find user failed", "error", err)
		respondError(c, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		respondError(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	token, err := generateToken(user.ID, user.Username)
	if err != nil {
		slog.Error("generate token failed", "error", err)
		respondError(c, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"token": token, "username": user.Username, "user_id": user.ID})
}

func (a *api) me(c *gin.Context) {
	respondSuccess(c, http.StatusOK, gin.H{
		"user_id":  c.GetInt64("userID"),
		"username": c.GetString("username"),
	})
}

func (a *api) createBlog(c *gin.Context) {
	var req createBlogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	userId := c.GetInt64("userID")
	username := c.GetString("username")

	blog, err := createBlogByUserId(a.db, userId, username, req.Title, req.Content)
	if err != nil {
		slog.Error("create blog failed", "error", err)
		respondError(c, http.StatusInternalServerError, "Could not create blog")
		return
	}
	respondSuccess(c, http.StatusCreated, blog)
}

func (a *api) getBlogs(c *gin.Context) {
	limit, offset := parsePagination(c, defaultBlogsLimit, 0)

	blogs, err := getAllBlogs(a.db, limit, offset)
	if err != nil {
		slog.Error("list blogs failed", "error", err)
		respondError(c, http.StatusInternalServerError, "Could not fetch blogs")
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{"blogs": blogs, "limit": limit, "page": offset})
}

func (a *api) getBlogsWithUserID(c *gin.Context) {
	userId := c.GetInt64("userID")
	limit, offset := parsePagination(c, defaultBlogsLimit, 1)

	blogs, err := getBlogsByUserId(a.db, int(userId), limit, offset)
	if err != nil {
		slog.Error("list user blogs failed", "error", err)
		respondError(c, http.StatusInternalServerError, "Could not fetch blogs")
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{"blogs": blogs, "limit": limit, "page": offset})
}

func (a *api) updateBlog(c *gin.Context) {
	blogId, err := strconv.ParseInt(c.Param("blog_id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "Blog id is required")
		return
	}

	var req createBlogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	userId := c.GetInt64("userID")
	blog, err := updateBlogByUserId(a.db, blogId, userId, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, ErrBlogNotFound) {
			respondError(c, http.StatusNotFound, "Blog not found")
			return
		}
		slog.Error("update blog failed", "error", err)
		respondError(c, http.StatusInternalServerError, "Could not update blog")
		return
	}
	respondSuccess(c, http.StatusOK, blog)
}

func (a *api) deleteBlog(c *gin.Context) {
	blogId, err := strconv.ParseInt(c.Param("blog_id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "Blog id is required")
		return
	}

	userId := c.GetInt64("userID")
	if err := deleteBlogByUserId(a.db, blogId, userId); err != nil {
		if errors.Is(err, ErrBlogNotFound) {
			respondError(c, http.StatusNotFound, "Blog not found")
			return
		}
		slog.Error("delete blog failed", "error", err)
		respondError(c, http.StatusInternalServerError, "Could not delete blog")
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{"message": "Blog deleted"})
}

func parsePagination(c *gin.Context, defaultLimit, defaultOffset int) (limit, offset int) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxBlogsLimit {
		limit = maxBlogsLimit
	}

	offset, err = strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultOffset)))
	if err != nil || offset < 0 {
		offset = defaultOffset
	}
	return limit, offset
}
