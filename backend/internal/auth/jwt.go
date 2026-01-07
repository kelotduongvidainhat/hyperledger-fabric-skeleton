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

// GenerateToken creates a signed JWT
func GenerateToken(username, role, org string) (string, error) {
	claims := UserClaims{
		username,
		role,
		org,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "ownership-registry",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// Middleware validates the JWT token
func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Token from Header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing Authorization Header"})
		}

		// Basic "Bearer <token>" check
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Token Format"})
		}
		tokenString := authHeader[7:]

		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Token"})
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Claims"})
		}

		// Dictionary to store user info in context
		c.Locals("user", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("org", claims.Org)

		return c.Next()
	}
}
