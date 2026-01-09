package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("super-secret-key-change-in-production")

// UserClaims defines the JWT claims
type UserClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Org      string `json:"org"`
	jwt.RegisteredClaims
}

// GenerateToken creates a short-lived access token (15 minutes)
func GenerateToken(username, role, org string) (string, error) {
	claims := UserClaims{
		username,
		role,
		org,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    "ownership-registry",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// GenerateRefreshToken creates a long-lived refresh token (7 days)
func GenerateRefreshToken(username, role, org string) (string, error) {
	claims := UserClaims{
		username,
		role,
		org,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			Issuer:    "ownership-registry",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// Middleware validates the JWT token (Check Header or Cookie)
func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string

		// 1. Try Authorization Header
		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		// 2. Try Cookie if header is missing
		if tokenString == "" {
			tokenString = c.Cookies("access_token")
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}

		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired session"})
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid session claims"})
		}

		// Dictionary to store user info in context
		c.Locals("user", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("org", claims.Org)

		return c.Next()
	}
}
