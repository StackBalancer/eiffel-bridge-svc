package gitlab

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"eiffel-bridge-svc/internal/eiffel"
)

func HandleWebhook(w http.ResponseWriter, r *http.Request, publish func(eiffel.Event) error) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	eventType := r.Header.Get("X-Gitlab-Event")
	log.Printf("Received GitLab event: %+v\n", payload)
	switch eventType {
	case "Push Hook":
		handlePush(payload, publish, w)
	case "Pipeline Hook":
		handlePipeline(payload, publish, w)
	default:
		http.Error(w, "unsupported event: "+eventType, http.StatusNotImplemented)
	}
}

func handlePush(payload map[string]any, publish func(eiffel.Event) error, w http.ResponseWriter) {
	user := payload["user_name"].(string)
	repo := payload["project"].(map[string]any)["git_http_url"].(string)
	ref := payload["ref"].(string) // "refs/heads/main"
	branch := ref[len("refs/heads/"):]
	commits := payload["commits"].([]any)

	for _, c := range commits {
		commit := c.(map[string]any)["id"].(string)
		event := eiffel.NewSourceChangeCreatedEvent(user, repo, branch, commit)
		if err := publish(event); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handlePipeline(payload map[string]any, publish func(eiffel.Event) error, w http.ResponseWriter) {
	obj := payload["object_attributes"].(map[string]any)
	status := obj["status"].(string)
	name := fmt.Sprintf("Pipeline #%v", obj["id"])

	// Triggered event
	triggered := eiffel.NewActivityTriggeredEvent(name)
	if err := publish(triggered); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Finished event
	if status == "success" || status == "failed" {
		finished := eiffel.NewActivityFinishedEvent(status)
		if err := publish(finished); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
