package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"bible-api-service/internal/bible"
)

// VersionsHandler handles requests for Bible versions.
type VersionsHandler struct {
	manager *bible.VersionManager
}

// NewVersionsHandler creates a new VersionsHandler.
func NewVersionsHandler(manager *bible.VersionManager) *VersionsHandler {
	return &VersionsHandler{manager: manager}
}

// ListVersions handles GET requests to list available Bible versions.
func (h *VersionsHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	versions := h.manager.GetAll()

	// Filter
	filtered := h.filterVersions(versions, r)

	// Sort
	h.sortVersions(filtered, r)

	// Pagination
	paginated, total, page, limit := h.paginateVersions(filtered, r)

	response := map[string]interface{}{
		"data":  paginated,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ServeHTTP implements http.Handler.
func (h *VersionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ListVersions(w, r)
}

func (h *VersionsHandler) filterVersions(versions []bible.Version, r *http.Request) []bible.Version {
	nameFilter := strings.ToLower(r.URL.Query().Get("name"))
	languageFilter := strings.ToLower(r.URL.Query().Get("language"))

	if nameFilter == "" && languageFilter == "" {
		// Return a copy to avoid modifying the original slice during sort
		dst := make([]bible.Version, len(versions))
		copy(dst, versions)
		return dst
	}

	var filtered []bible.Version
	for _, v := range versions {
		if nameFilter != "" && !strings.Contains(strings.ToLower(v.Name), nameFilter) {
			continue
		}
		if languageFilter != "" && !strings.Contains(strings.ToLower(v.Language), languageFilter) {
			continue
		}
		filtered = append(filtered, v)
	}
	return filtered
}

func (h *VersionsHandler) sortVersions(versions []bible.Version, r *http.Request) {
	// Default sort by code (value)
	sortField := r.URL.Query().Get("sort")
	if sortField == "" {
		sortField = "code"
	}

	sort.Slice(versions, func(i, j int) bool {
		switch sortField {
		case "name":
			return versions[i].Name < versions[j].Name
		case "language":
			return versions[i].Language < versions[j].Language
		default: // "code"
			return versions[i].Code < versions[j].Code
		}
	})
}

func (h *VersionsHandler) paginateVersions(versions []bible.Version, r *http.Request) ([]bible.Version, int, int, int) {
	total := len(versions)
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20 // Default limit
	}

	start := (page - 1) * limit
	if start >= total {
		return []bible.Version{}, total, page, limit
	}

	end := start + limit
	if end > total {
		end = total
	}

	return versions[start:end], total, page, limit
}
