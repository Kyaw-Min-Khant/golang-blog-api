package main

import "github.com/gin-gonic/gin"

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"status": "error", "message": message})
}

func respondSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"status": "success", "data": data})
}
