package main

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

// LiveKit Server Client তৈরি করার হেল্পার
func getRoomClient() *lksdk.RoomServiceClient {
	url := os.Getenv("LIVEKIT_URL")
	// SDK-এর জন্য URL কে https:// এ কনভার্ট করতে হয়
	if len(url) > 4 && url[:4] == "wss:" {
		url = "https:" + url[4:]
	}
	return lksdk.NewRoomServiceClient(url, os.Getenv("LIVEKIT_API_KEY"), os.Getenv("LIVEKIT_API_SECRET"))
}

// ১. নির্দিষ্ট একজনকে মিউট করা
func (app *application) forceMuteHandler(w http.ResponseWriter, r *http.Request) {
	hostID, _ := uuid.Parse(r.Context().Value("userID").(string))
	eventID, _ := uuid.Parse(chi.URLParam(r, "id"))
	targetID := chi.URLParam(r, "userID")

	ev, err := app.store.Events.GetByID(r.Context(), eventID)
	if err != nil || ev.HostID != hostID {
		app.forbiddenResponse(w, r)
		return
	}

	// ডাটাবেসে রোল লিসেনার করা
	app.store.Members.UpdateRole(r.Context(), eventID, uuid.MustParse(targetID), "listener")
	
	// LiveKit-এ পারমিশন আপডেট করা
	getRoomClient().UpdateParticipant(context.Background(), &livekit.UpdateParticipantRequest{
		Room:     ev.RoomName,
		Identity: targetID,
		Permission: &livekit.ParticipantPermission{
			CanPublish:   false,
			CanSubscribe: true,
		},
	})

	app.jsonResponse(w, http.StatusOK, map[string]string{"message": "user muted"})
}

// ২. সবাইকে মিউট করা
func (app *application) muteAllSpeakersHandler(w http.ResponseWriter, r *http.Request) {
	hostID, _ := uuid.Parse(r.Context().Value("userID").(string))
	eventID, _ := uuid.Parse(chi.URLParam(r, "id"))

	ev, err := app.store.Events.GetByID(r.Context(), eventID)
	if err != nil || ev.HostID != hostID {
		app.forbiddenResponse(w, r)
		return
	}

	// ডাটাবেসে সব স্পিকারকে লিসেনার বানানো
	app.store.Members.MuteAllSpeakers(r.Context(), eventID)

	// LiveKit-এ রুমের সবাইকে মিউট করা
	res, _ := getRoomClient().ListParticipants(context.Background(), &livekit.ListParticipantsRequest{Room: ev.RoomName})
	if res != nil {
		for _, p := range res.Participants {
			getRoomClient().UpdateParticipant(context.Background(), &livekit.UpdateParticipantRequest{
				Room:     ev.RoomName,
				Identity: p.Identity,
				Permission: &livekit.ParticipantPermission{
					CanPublish:   false,
					CanSubscribe: true,
				},
			})
		}
	}
	app.jsonResponse(w, http.StatusOK, map[string]string{"message": "all speakers muted"})
}

// ৩. কাউকে রুম থেকে কিক করা
func (app *application) kickParticipantHandler(w http.ResponseWriter, r *http.Request) {
	hostID, _ := uuid.Parse(r.Context().Value("userID").(string))
	eventID, _ := uuid.Parse(chi.URLParam(r, "id"))
	targetID := chi.URLParam(r, "userID")

	ev, err := app.store.Events.GetByID(r.Context(), eventID)
	if err != nil || ev.HostID != hostID {
		app.forbiddenResponse(w, r)
		return
	}

	// LiveKit থেকে রিমুভ করা
	getRoomClient().RemoveParticipant(context.Background(), &livekit.RoomParticipantIdentity{
		Room:     ev.RoomName,
		Identity: targetID,
	})
	
	// ডাটাবেসে লিভ করানো
	app.store.Members.Leave(r.Context(), eventID, uuid.MustParse(targetID))

	app.jsonResponse(w, http.StatusOK, map[string]string{"message": "user kicked"})
}

func (app *application) unmuteAllSpeakersHandler(w http.ResponseWriter, r *http.Request) {
	hostID, _ := uuid.Parse(r.Context().Value("userID").(string))
	eventID, _ := uuid.Parse(chi.URLParam(r, "id"))

	// ১. চেক: রিকোয়েস্ট পাঠানো লোক হোস্ট কিনা
	ev, err := app.store.Events.GetByID(r.Context(), eventID)
	if err != nil || ev.HostID != hostID {
		app.forbiddenResponse(w, r)
		return
	}

	// ২. ডাটাবেসে সব লিসেনারকে স্পিকার বানানো
	err = app.store.Members.UnmuteAllSpeakers(r.Context(), eventID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	// ৩. LiveKit-এ সবার পারমিশন আপডেট করা (CanPublish: true)
	roomClient := getRoomClient()
	res, err := roomClient.ListParticipants(context.Background(), &livekit.ListParticipantsRequest{
		Room: ev.RoomName,
	})
	
	if err == nil {
		for _, p := range res.Participants {
			// (ঐচ্ছিক: চাইলে এখানে চেক করতে পারেন সে হোস্ট কিনা, হোস্টকে স্কিপ করতে পারেন)
			roomClient.UpdateParticipant(context.Background(), &livekit.UpdateParticipantRequest{
				Room:     ev.RoomName,
				Identity: p.Identity,
				Permission: &livekit.ParticipantPermission{
					CanPublish:   true, // মাইক অন করার অনুমতি
					CanSubscribe: true, // শোনার অনুমতি
				},
			})
		}
	}

	app.jsonResponse(w, http.StatusOK, map[string]string{
		"message": "all listeners promoted to speakers",
	})
}