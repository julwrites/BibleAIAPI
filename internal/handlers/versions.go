package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"bible-api-service/internal/biblegateway"

	"gopkg.in/yaml.v2"
)

// VersionsHandler handles requests for Bible versions.
type VersionsHandler struct {
	versions []biblegateway.Version
}

// NewVersionsHandler creates a new VersionsHandler by loading versions from the config file.
func NewVersionsHandler(configPath string) (*VersionsHandler, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions config: %w", err)
	}

	var versions []biblegateway.Version
	if err := yaml.Unmarshal(data, &versions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal versions config: %w", err)
	}

	return &VersionsHandler{versions: versions}, nil
}

// ListVersions handles GET requests to list available Bible versions.
func (h *VersionsHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Filter
	filtered := h.filterVersions(r)

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

func (h *VersionsHandler) filterVersions(r *http.Request) []biblegateway.Version {
	nameFilter := strings.ToLower(r.URL.Query().Get("name"))
	languageFilter := strings.ToLower(r.URL.Query().Get("language"))

	if nameFilter == "" && languageFilter == "" {
		// Return a copy to avoid modifying the original slice during sort
		dst := make([]biblegateway.Version, len(h.versions))
		copy(dst, h.versions)
		return dst
	}

	var filtered []biblegateway.Version
	for _, v := range h.versions {
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

func (h *VersionsHandler) sortVersions(versions []biblegateway.Version, r *http.Request) {
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
			return versions[i].Value < versions[j].Value
		}
	})
}

func (h *VersionsHandler) paginateVersions(versions []biblegateway.Version, r *http.Request) ([]biblegateway.Version, int, int, int) {
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
		return []biblegateway.Version{}, total, page, limit
	}

	end := start + limit
	if end > total {
		end = total
	}

	return versions[start:end], total, page, limit
}
