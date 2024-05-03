package mid

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func NewJwtAuthMiddleware(onResult func(*jwt.Token) error) func(echo.HandlerFunc) echo.HandlerFunc {
	key := os.Getenv("CJUNGO_JWT_KEY")
	if len(key) <= 0 {
		key = "1234567890ABCDEF"
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			request := ctx.Request()
			auth := request.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				return fmt.Errorf("不是有效的 JWT token")
			}
			tokenString := auth[7:]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("加密算法不匹配 %v", token.Header["alg"])
				}
				k := []byte(key)
				return &k, nil
			})
			if err != nil {
				return err
			}
			if err := onResult(token); err != nil {
				return err
			}

			return next(ctx)
		}
	}
}
