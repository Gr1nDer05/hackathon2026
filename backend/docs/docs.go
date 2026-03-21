package docs

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed index.html openapi.yaml swagger-init.js
var embeddedFiles embed.FS

func RegisterRoutes(router *gin.Engine) {
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger/")
	})

	swagger := router.Group("/swagger")
	swagger.Use(swaggerHeaders())
	swagger.StaticFS("/", http.FS(mustAssetFS()))
}

func mustAssetFS() fs.FS {
	assets, err := fs.Sub(embeddedFiles, ".")
	if err != nil {
		panic(err)
	}

	return assets
}

func swaggerHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := c.Writer.Header()
		headers.Set(
			"Content-Security-Policy",
			"default-src 'self' https://cdn.jsdelivr.net; "+
				"connect-src 'self'; "+
				"img-src 'self' data: https://cdn.jsdelivr.net; "+
				"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
				"script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
				"font-src 'self' data: https://cdn.jsdelivr.net; "+
				"frame-ancestors 'none'; base-uri 'self'",
		)
		c.Next()
	}
}
