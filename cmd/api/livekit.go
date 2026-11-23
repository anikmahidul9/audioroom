package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
)

func (app *application) joinRoomHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(r.Context().Value("userID").(string))
	eventID, _ := uuid.Parse(chi.URLParam(r, "id"))


	ev, err := app.store.Events.GetByID(r.Context(), eventID)
	if err != nil {
		app.notFoundResponse(w, r, err)
		return
	}
	if ev.Status != "live" {
		app.badRequestResponse(w, r, fmt.Errorf("event is not live yet"))
		return
	}


	if ev.Visibility == "private" {
		hasAccess, _ := app.store.Events.CheckEventAccess(r.Context(), eventID, userID)
		if !hasAccess {
			app.forbiddenResponse(w, r)
			return
		}
	}

	role := "listener"
	if ev.HostID == userID {
		role = "host"
	} else {
		existingRole, _, err := app.store.Members.GetMemberRole(r.Context(), eventID, userID)
		if err == nil && existingRole != "" {
			role = existingRole 
		}
	}

	err = app.store.Members.Join(r.Context(), eventID, userID, role)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")
	at := auth.NewAccessToken(apiKey, apiSecret)
	grant := &auth.VideoGrant{RoomJoin: true, Room: ev.RoomName}

	if role == "speaker" || role == "host" || role == "co_host" {
		grant.SetCanPublish(true) 
	} else {
		grant.SetCanPublish(false) 
	}

	at.AddGrant(grant).SetIdentity(userID.String()).SetValidFor(3 * time.Hour)
	token, err := at.ToJWT()
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]string{
		"token":      token,
		"server_url": os.Getenv("LIVEKIT_URL"),
		"role":       role,
	})
}