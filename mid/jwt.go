package mid

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func MakeJwtToken(claims jwt.Claims) (string, error) {
	key := os.Getenv("CJUNGO_JWT_KEY")
	if len(key) <= 0 {
		return "", fmt.Errorf("生成 TOKEN 失败，没有配置 CJUNGO_JWT_KEY")
	}

	k := []byte(key)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(k)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Bearer %s", tokenString), nil
}

func NewJwtAuthMiddleware(onResult func(*jwt.Token) error) echo.MiddlewareFunc {
	key := os.Getenv("CJUNGO_JWT_KEY")
	if len(key) <= 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				return fmt.Errorf("解析 TOKEN 失败，没有配置 CJUNGO_JWT_KEY")
			}
		}
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
				return k, nil
			})
			if err != nil {
				return fmt.Errorf("解析 Token 失败, %v, %s", err, tokenString)
			}
			if err := onResult(token); err != nil {
				return fmt.Errorf("自定义 Token 处理返回错误：%v", err)
			}

			return next(ctx)
		}
	}
}
