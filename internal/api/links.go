package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tesserabox/bentotask/internal/model"
)

// createLinkRequest is the JSON body for POST /tasks/:id/links.
type createLinkRequest struct {
	TargetID string `json:"target_id"`
	Type     string `json:"type"`
}

func (s *Server) handleCreateLink(w http.ResponseWriter, r *http.Request) {
	sourceID := chi.URLParam(r, "id")

	var req createLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondBadRequest(w, "invalid JSON: "+err.Error())
		return
	}

	if req.TargetID == "" {
		respondValidationError(w, "target_id is required")
		return
	}

	lt := model.LinkType(req.Type)
	if lt == "" {
		lt = model.LinkRelatedTo
	}
	if lt != model.LinkDependsOn && lt != model.LinkBlocks && lt != model.LinkRelatedTo {
		respondValidationError(w, "invalid link type: "+req.Type)
		return
	}

	s.mu.Lock()
	source, target, err := s.app.LinkTasks(sourceID, req.TargetID, lt)
	s.mu.Unlock()
	if err != nil {
		if isNotFound(err) {
			respondNotFound(w, err.Error())
			return
		}
		if isConflict(err) {
			respondConflict(w, err.Error())
			return
		}
		respondInternalError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"source_id": source.ID,
		"target_id": target.ID,
		"link_type": string(lt),
	})
}

func (s *Server) handleGetLinks(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.RLock()
	links, err := s.app.GetTaskLinks(id)
	s.mu.RUnlock()
	if err != nil {
		if isNotFound(err) {
			respondNotFound(w, err.Error())
			return
		}
		respondInternalError(w, err)
		return
	}

	items := make([]map[string]string, len(links))
	for i, l := range links {
		items[i] = map[string]string{
			"type":       string(l.Type),
			"direction":  l.Direction,
			"task_id":    l.TaskID,
			"task_title": l.TaskTitle,
		}
	}

	respondJSON(w, http.StatusOK, collectionResponse{Items: items, Count: len(items)})
}

func (s *Server) handleDeleteLink(w http.ResponseWriter, r *http.Request) {
	sourceID := chi.URLParam(r, "id")
	targetID := chi.URLParam(r, "targetId")

	linkType := r.URL.Query().Get("type")
	lt := model.LinkType(linkType)
	if lt == "" {
		lt = model.LinkRelatedTo
	}

	s.mu.Lock()
	_, _, err := s.app.UnlinkTasks(sourceID, targetID, lt)
	s.mu.Unlock()
	if err != nil {
		if isNotFound(err) {
			respondNotFound(w, err.Error())
			return
		}
		respondInternalError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"removed": true})
}
