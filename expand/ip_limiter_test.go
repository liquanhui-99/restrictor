package expand

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/liquanhui-99/restrictor/single"
	"net/http"
	"time"
)

func ExampleIpLimiter_AllowIp() {
	limiter := single.NewSlideWindowLimiter(3*time.Millisecond, 10000)
	ipLimiter := NewIpLimiter(limiter, time.Minute, 10)
	defer ipLimiter.Close()
	r := gin.Default()
	r.Use(func(ctx *gin.Context) {
		ip := ctx.Request.URL.Host
		c, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		res, err := ipLimiter.AllowIp(c, ip)
		if err != nil {
			fmt.Println(err)
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		if !res {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	})

	r.GET("/profile", func(ctx *gin.Context) {
		ctx.Writer.WriteHeader(http.StatusOK)
		_, _ = ctx.Writer.Write([]byte("成功"))
	})

	if err := r.Run(":8082"); err != nil {
		panic(err)
	}
}
