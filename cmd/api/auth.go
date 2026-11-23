package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/anikmahidul9/voiceroom/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type registerInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ১. Register Handler
func (app *application) registerHandler(w http.ResponseWriter, r *http.Request) {
	var input registerInput
	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// পাসওয়ার্ড হ্যাশ
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	user := &store.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
	}

	if err := app.store.Users.Create(r.Context(), user); err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("could not create user"))
		return
	}

	app.jsonResponse(w, http.StatusCreated, map[string]string{"message": "user registered"})
}

// ২. Login Handler
func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input loginInput
	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), input.Email)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid credentials"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid credentials"))
		return
	}

	// JWT টোকেন তৈরি
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	})
	
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	app.jsonResponse(w, http.StatusOK, map[string]string{"token": tokenString})
}