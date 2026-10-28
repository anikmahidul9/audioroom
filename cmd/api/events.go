package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/anikmahidul9/voiceroom/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// 1. Create Event
func (app *application) createEventHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := "11111111-1111-1111-1111-111111111111" //r.Context().Value("userID").(string)
	hostID, _ := uuid.Parse(userIDStr)

	var input struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		CategoryID  *int       `json:"category_id"`
		Visibility  string     `json:"visibility"`
		ScheduledAt *time.Time `json:"scheduled_at"`
		StartNow    bool       `json:"start_now"`
	}

	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title == "" {
		app.badRequestResponse(w, r, fmt.Errorf("title is required"))
		return
	}
	if input.Visibility == "" {
		input.Visibility = "public"
	}

	status := "scheduled"
	if input.StartNow {
		status = "live"
	} else {
		if input.ScheduledAt == nil || !input.ScheduledAt.After(time.Now()) {
			app.badRequestResponse(w, r, store.ErrInvalidEventTime)
			return
		}
	}

	ev := &store.Event{
		HostID:      hostID,
		Title:       input.Title,
		Description: input.Description,
		CategoryID:  input.CategoryID,
		Visibility:  input.Visibility,
		Status:      status,
		ScheduledAt: input.ScheduledAt,
		RoomName:    "event-" + uuid.NewString(),
	}

	createdEvent, err := app.store.Events.Create(r.Context(), ev)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, createdEvent)
}

// 2. Feed (Live or Upcoming)
func (app *application) feedEventHandler(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	status := "live"
	if kind == "upcoming" {
		status = "scheduled"
	}

	events, err := app.store.Events.ListPublic(r.Context(), status, limit, offset)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, events)
}

// 3. Get Single Event
func (app *application) getEventHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid event id"))
		return
	}

	ev, err := app.store.Events.GetByID(r.Context(), id)
	if err != nil {
		app.notFoundResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, ev)
}

// 4. Start Event
func (app *application) startEventHandler(w http.ResponseWriter, r *http.Request) {
	app.changeEventStatus(w, r, "live")
}

// 5. End Event
func (app *application) endEventHandler(w http.ResponseWriter, r *http.Request) {
	app.changeEventStatus(w, r, "ended")
}

// Helper for Start/End
func (app *application) changeEventStatus(w http.ResponseWriter, r *http.Request, status string) {
	userIDStr := "11111111-1111-1111-1111-111111111111" //r.Context().Value("userID").(string)
	userID, _ := uuid.Parse(userIDStr)

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid event id"))
		return
	}

	ev, err := app.store.Events.GetByID(r.Context(), id)
	if err != nil {
		app.notFoundResponse(w, r, err)
		return
	}

	if ev.HostID != userID {
		app.forbiddenResponse(w, r)
		return
	}

	if err := app.store.Events.UpdateStatus(r.Context(), id, status); err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]string{"message": "event " + status})
}
