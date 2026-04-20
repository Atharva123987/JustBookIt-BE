package middleware

import "github.com/gin-gonic/gin"

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		deviceID := ctx.GetHeader("X-Device-Id")
		if deviceID == "" {
			ctx.AbortWithStatusJSON(400, gin.H{
				"error": "Missing X-Device-Id header",
			})
			return
		}
		ctx.Set("device_id", deviceID)

		ctx.Set("user_id", "68b19e59785ee9d9599cb55d") // set user ID for authentication purposes
		ctx.Next()                                     // continue to the next middleware/handler
	}
}

func AnonymousMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("lat", "19.215126175385713")
		ctx.Set("lng", "72.84057728596858")
	}
}
