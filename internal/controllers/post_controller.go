package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PostController struct{}

type Post struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (pc *PostController) ListPosts(c *gin.Context) {
	posts := []Post{{ID: 1, Title: "Hello", Content: "First post"}}
	c.JSON(http.StatusOK, posts)
}

func (pc *PostController) CreatePost(c *gin.Context) {
	var req Post
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = 2
	c.JSON(http.StatusCreated, req)
}
