package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (app *application) joinEventHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(r.Context().Value("userID").(string))
	evID, _ := uuid.Parse(chi.URLParam(r, "id"))

	ev, err := app.store.Events.GetByID(r.Context(), evID)
	if err != nil {
		app.notFoundResponse(w, r, err)
		return
	}
	if ev.Status != "live" {
		app.badRequestResponse(w, r, fmt.Errorf("event is not live yet"))
		return
	}

	role := "listener"

	if ev.HostID == userID {
		role = "host"
	} else {
		existingRole, _, err := app.store.Members.GetMemberRole(r.Context(), evID, userID)
		if err == nil && existingRole != "" {
			role = existingRole
		}
	}

	err = app.store.Members.Join(r.Context(), evID, userID, role)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]string{"message": "joined as " + role})
}

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

func (app *application) updateMemberRoleHandler(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "id")
	eventID, _ := uuid.Parse(eventIDStr)

	targetUserIDStr := chi.URLParam(r, "userID")
	targetUserID, _ := uuid.Parse(targetUserIDStr)

	var input struct {
		Role string `json:"role"`
	}
	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err := app.store.Members.UpdateRole(r.Context(), eventID, targetUserID, input.Role)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]string{"message": "role updated to " + input.Role})
}
