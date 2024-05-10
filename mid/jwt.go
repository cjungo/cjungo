package mid

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func MakeJwtToken[T jwt.Claims](claims T) (string, error) {
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

func ParseJwtToken[T jwt.Claims](ctx echo.Context, claims T) (*jwt.Token, error) {
	key := os.Getenv("CJUNGO_JWT_KEY")
	if len(key) <= 0 {
		return nil, fmt.Errorf("解析 TOKEN 失败，没有配置 CJUNGO_JWT_KEY")
	}
	request := ctx.Request()
	auth := request.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, fmt.Errorf("不是有效的 JWT token")
	}
	tokenString := auth[7:]
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("加密算法不匹配 %v", token.Header["alg"])
		}
		k := []byte(key)
		return k, nil
	})
	if err != nil {
		return nil, fmt.Errorf("解析 Token 失败, %v, %s", err, tokenString)
	}
	return token, nil
}

func NewJwtAuthMiddleware[T jwt.Claims](provide func() T, onResult func(*jwt.Token, T) error) echo.MiddlewareFunc {
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
			claims := provide()
			token, err := ParseJwtToken(ctx, claims)
			if err != nil {
				return err
			}
			if err := onResult(token, claims); err != nil {
				return fmt.Errorf("自定义 Token 处理返回错误：%v", err)
			}

			return next(ctx)
		}
	}
}
