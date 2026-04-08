package api

import (
	"net/http"
	"strconv"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
)

func (s *Server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	availTime := 60
	if v := q.Get("time"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			availTime = parsed
		}
	}

	energy := model.Energy(q.Get("energy"))
	if energy == "" {
		energy = model.EnergyMedium
	}

	context := q.Get("context")

	count := 5
	if v := q.Get("count"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			count = parsed
		}
	}

	opts := app.SuggestOptions{
		AvailableTime: availTime,
		Energy:        energy,
		Context:       context,
	}

	s.mu.RLock()
	suggestions, err := s.app.Suggest(opts, count)
	s.mu.RUnlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}

	items := SuggestionsToJSON(suggestions)
	respondJSON(w, http.StatusOK, collectionResponse{Items: items, Count: len(items)})
}

func (s *Server) handlePlanToday(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	availTime := 480
	if v := q.Get("time"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			availTime = parsed
		}
	}

	energy := model.Energy(q.Get("energy"))
	if energy == "" {
		energy = model.EnergyMedium
	}

	context := q.Get("context")

	opts := app.SuggestOptions{
		AvailableTime: availTime,
		Energy:        energy,
		Context:       context,
	}

	s.mu.RLock()
	result, err := s.app.PlanDay(opts)
	s.mu.RUnlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, PlanToJSON(result, availTime))
}

func (s *Server) handleRebuildIndex(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	count, err := s.app.RebuildIndex()
	s.mu.Unlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"indexed": count})
}

func (s *Server) handleMetaTags(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	tags, err := s.app.CompleteTags()
	s.mu.RUnlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}
	if tags == nil {
		tags = []string{}
	}

	respondJSON(w, http.StatusOK, collectionResponse{Items: tags, Count: len(tags)})
}

func (s *Server) handleMetaBoxes(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	boxes, err := s.app.CompleteBoxes()
	s.mu.RUnlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}
	if boxes == nil {
		boxes = []string{}
	}

	respondJSON(w, http.StatusOK, collectionResponse{Items: boxes, Count: len(boxes)})
}

func (s *Server) handleMetaContexts(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	contexts, err := s.app.CompleteContexts()
	s.mu.RUnlock()
	if err != nil {
		respondInternalError(w, err)
		return
	}
	if contexts == nil {
		contexts = []string{}
	}

	respondJSON(w, http.StatusOK, collectionResponse{Items: contexts, Count: len(contexts)})
}
