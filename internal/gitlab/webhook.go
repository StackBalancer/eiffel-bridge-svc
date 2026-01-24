package gitlab

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"eiffel-bridge-svc/internal/eiffel"
)

func HandleWebhook(w http.ResponseWriter, r *http.Request, publish func(eiffel.Event) error) {
	var payload map[string]any
	
	// GitLab can send webhooks as form data or JSON
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// Parse form data - GitLab sends JSON in 'payload' field
		if err := r.ParseForm(); err != nil {
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}
		payloadStr := r.FormValue("payload")
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			http.Error(w, "invalid JSON in payload field", http.StatusBadRequest)
			return
		}
	} else {
		// Parse as JSON body
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
	}

	eventType := r.Header.Get("X-Gitlab-Event")
	log.Printf("Received GitLab event: %s", eventType)
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
	// Safe type assertions with error handling
	user, ok := payload["user_name"].(string)
	if !ok {
		http.Error(w, "missing or invalid user_name", http.StatusBadRequest)
		return
	}
	
	project, ok := payload["project"].(map[string]any)
	if !ok {
		http.Error(w, "missing or invalid project", http.StatusBadRequest)
		return
	}
	
	repo, ok := project["git_http_url"].(string)
	if !ok {
		http.Error(w, "missing or invalid git_http_url", http.StatusBadRequest)
		return
	}
	
	ref, ok := payload["ref"].(string)
	if !ok || !strings.HasPrefix(ref, "refs/heads/") {
		http.Error(w, "missing or invalid ref", http.StatusBadRequest)
		return
	}
	
	branch := ref[len("refs/heads/"):]
	commits, ok := payload["commits"].([]any)
	if !ok {
		http.Error(w, "missing or invalid commits", http.StatusBadRequest)
		return
	}

	for _, c := range commits {
		commitMap, ok := c.(map[string]any)
		if !ok {
			continue // Skip invalid commit entries
		}
		
		commit, ok := commitMap["id"].(string)
		if !ok {
			continue // Skip commits without valid ID
		}
		
		event := eiffel.NewSourceChangeCreatedEvent(user, repo, branch, commit)
		if err := publish(event); err != nil {
			log.Printf("Failed to publish event for commit %s: %v", commit, err)
			http.Error(w, "failed to publish event", http.StatusInternalServerError)
			return
		}
		log.Printf("Published SourceChangeCreated event for commit %s", commit)
	}

	w.WriteHeader(http.StatusOK)
}

func handlePipeline(payload map[string]any, publish func(eiffel.Event) error, w http.ResponseWriter) {
	obj, ok := payload["object_attributes"].(map[string]any)
	if !ok {
		http.Error(w, "missing or invalid object_attributes", http.StatusBadRequest)
		return
	}
	
	status, ok := obj["status"].(string)
	if !ok {
		http.Error(w, "missing or invalid status", http.StatusBadRequest)
		return
	}
	
	pipelineID := obj["id"]
	name := fmt.Sprintf("Pipeline #%v", pipelineID)

	// Triggered event
	triggered := eiffel.NewActivityTriggeredEvent(name)
	if err := publish(triggered); err != nil {
		log.Printf("Failed to publish ActivityTriggered event: %v", err)
		http.Error(w, "failed to publish triggered event", http.StatusInternalServerError)
		return
	}
	log.Printf("Published ActivityTriggered event for %s", name)

	// Finished event
	if status == "success" || status == "failed" {
		finished := eiffel.NewActivityFinishedEvent(status)
		if err := publish(finished); err != nil {
			log.Printf("Failed to publish ActivityFinished event: %v", err)
			http.Error(w, "failed to publish finished event", http.StatusInternalServerError)
			return
		}
		log.Printf("Published ActivityFinished event with outcome: %s", status)
	}

	w.WriteHeader(http.StatusOK)
}
