package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ১. ইভেন্টে জয়েন করা (FR-JOIN-001, FR-JOIN-002)
func (app *application) joinEventHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr, _ := r.Context().Value("userID").(string)
	userID, _ := uuid.Parse(userIDStr)

	eventIDStr := chi.URLParam(r, "id")
	eventID, _ := uuid.Parse(eventIDStr)

	// ইভেন্টটি লাইভ কিনা চেক করা (FR-ROOM-009)
	ev, err := app.store.Events.GetByID(r.Context(), eventID)
	if err != nil {
		app.notFoundResponse(w, r, err)
		return
	}
	if ev.Status != "live" {
		app.badRequestResponse(w, r, fmt.Errorf("can only join live events"))
		return
	}

	// ডিফল্টভাবে লিসেনার হিসেবে জয়েন করবে
	err = app.store.Members.Join(r.Context(), eventID, userID, "listener")
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]string{"message": "joined as listener"})
}

// ২. ইভেন্ট থেকে লিভ করা (FR-JOIN-003, FR-JOIN-004)
func (app *application) leaveEventHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr, _ := r.Context().Value("userID").(string)
	userID, _ := uuid.Parse(userIDStr)

	eventIDStr := chi.URLParam(r, "id")
	eventID, _ := uuid.Parse(eventIDStr)

	err := app.store.Members.Leave(r.Context(), eventID, userID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]string{"message": "left the event"})
}

// ৩. মেম্বার লিস্ট দেখা
func (app *application) listMembersHandler(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "id")
	eventID, _ := uuid.Parse(eventIDStr)

	members, err := app.store.Members.ListActive(r.Context(), eventID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, members)
}

// ৪. স্পিকার বানানো বা রোল পরিবর্তন করা (Host/Co-host এর জন্য)
func (app *application) updateMemberRoleHandler(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "id")
	eventID, _ := uuid.Parse(eventIDStr)

	targetUserIDStr := chi.URLParam(r, "userID")
	targetUserID, _ := uuid.Parse(targetUserIDStr)

	var input struct {
		Role string `json:"role"` // "speaker" or "listener"
	}
	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// (ভবিষ্যতে এখানে চেক করতে হবে রিকোয়েস্ট পাঠানো ইউজার Host বা Co-host কিনা)

	err := app.store.Members.UpdateRole(r.Context(), eventID, targetUserID, input.Role)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]string{"message": "role updated to " + input.Role})
}