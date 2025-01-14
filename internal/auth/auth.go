package auth

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matthewgaim/loudmouth_api/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

func Signup(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errors.RespondWithError(w, http.StatusMethodNotAllowed, "only POST method is allowed")
			return
		}

		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			errors.RespondWithError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validateInput(user); err != nil {
			errors.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", user.Email).Scan(&exists)
		if err != nil {
			errors.RespondWithError(w, http.StatusInternalServerError, "database error")
			return
		}
		if exists {
			errors.RespondWithError(w, http.StatusConflict, "email already registered")
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			errors.RespondWithError(w, http.StatusInternalServerError, "error processing password")
			return
		}

		err = db.QueryRow(`
			INSERT INTO users (display_name, email, password)
			VALUES ($1, $2, $3)
			RETURNING id
		`, user.DisplayName, user.Email, string(hashedPassword)).Scan(&user.ID)

		if err != nil {
			errors.RespondWithError(w, http.StatusInternalServerError, "error creating user")
			return
		}

		token, err := generateToken(user)
		if err != nil {
			errors.RespondWithError(w, http.StatusInternalServerError, "error generating token")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": token,
		})
	}
}

func Signin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errors.RespondWithError(w, http.StatusMethodNotAllowed, "only POST method is allowed")
			return
		}

		var creds SigninCreds

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			errors.RespondWithError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		var user User
		err := db.QueryRow(`
			SELECT id, display_name, email, password 
			FROM users 
			WHERE email = $1
		`, creds.Email).Scan(&user.ID, &user.DisplayName, &user.Email, &user.Password)

		if err == sql.ErrNoRows {
			errors.RespondWithError(w, http.StatusUnauthorized, "invalid email or password")
			return
		} else if err != nil {
			errors.RespondWithError(w, http.StatusInternalServerError, "database error")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			if subtle.ConstantTimeCompare([]byte(user.Password), []byte(creds.Password)) != 1 {
				errors.RespondWithError(w, http.StatusUnauthorized, "invalid email or password")
				return
			}
		}

		token, err := generateToken(user)
		if err != nil {
			fmt.Println(err.Error())
			errors.RespondWithError(w, http.StatusInternalServerError, "error generating token")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": token,
		})
	}
}

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

var jwtKey = []byte(os.Getenv("JWT_KEY"))

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func validateInput(user User) error {
	if user.DisplayName == "" || len(user.DisplayName) > 255 {
		return fmt.Errorf("invalid display name length")
	}
	if !emailRegex.MatchString(user.Email) || len(user.Email) > 255 {
		return fmt.Errorf("invalid email format")
	}
	if len(user.Password) < 8 || len(user.Password) > 255 {
		return fmt.Errorf("password must be between 8 and 255 characters")
	}
	return nil
}

func generateToken(user User) (string, error) {
	expirationTime := time.Now().Add(720 * time.Hour) // 1 month
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (*Claims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errors.RespondWithError(w, http.StatusUnauthorized, "no authorization header")
			return
		}

		claims, err := VerifyToken(authHeader)
		if err != nil {
			errors.RespondWithError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
