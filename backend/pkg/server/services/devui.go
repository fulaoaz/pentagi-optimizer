package services

import (
	_ "embed"
	"github.com/gin-gonic/gin"
)

//go:embed templates/graphql-playground-cn.html
var playgroundCNHTML []byte

//go:embed templates/swagger-cn.html
var swaggerCNHTML []byte

// ServeGraphqlPlaygroundCN serves the Chinese-localized GraphQL Playground.
func (s *GraphqlService) ServeGraphqlPlaygroundCN(c *gin.Context) {
	w := c.Writer
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(playgroundCNHTML)
}

// ServeSwaggerCN serves the Chinese-localized Swagger UI shell.
func (s *GraphqlService) ServeSwaggerCN(c *gin.Context) {
	w := c.Writer
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(swaggerCNHTML)
}
