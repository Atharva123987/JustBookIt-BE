package middleware

import "github.com/gin-gonic/gin"

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("user_id", "68b19e59785ee9d9599cb55d") // set user ID for authentication purposes
		ctx.Next()                                     // continue to the next middleware/handler
	}
}
