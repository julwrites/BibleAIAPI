package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"bible-api-service/internal/biblegateway"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersionsHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "versions.yaml")
		content := `
- name: English Standard Version
  value: ESV
  language: English
- name: King James Version
  value: KJV
  language: English
`
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		h, err := NewVersionsHandler(configPath)
		require.NoError(t, err)
		assert.Len(t, h.versions, 2)
		assert.Equal(t, "ESV", h.versions[0].Value)
	})

	t.Run("FileNotFound", func(t *testing.T) {
		_, err := NewVersionsHandler("nonexistent.yaml")
		assert.Error(t, err)
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(configPath, []byte("invalid yaml content"), 0644)
		require.NoError(t, err)

		_, err = NewVersionsHandler(configPath)
		assert.Error(t, err)
	})
}

func TestVersionsHandler_ListVersions(t *testing.T) {
	v := []biblegateway.Version{
		{Name: "English Standard Version", Value: "ESV", Language: "English"},
		{Name: "King James Version", Value: "KJV", Language: "English"},
		{Name: "Reina-Valera 1960", Value: "RVR1960", Language: "Spanish"},
		{Name: "La Bible du Semeur", Value: "BDS", Language: "French"},
	}
	// Need to access internal field, so we use reflection or just assume the struct is available if in same package
	// Since we are in handlers package (white-box testing), we can access unexported fields if we create the struct directly.
	// However, VersionsHandler struct definition has `versions` as exported? No, it is `versions`.
	// Let's check `internal/handlers/versions.go` again.
	// type VersionsHandler struct { versions []biblegateway.Version }
	// It is unexported.
	// But `versions_test.go` is package `handlers`, so it can access unexported fields.

	h := &VersionsHandler{versions: v}

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/versions", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("NoFilters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/versions?limit=10", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, float64(4), resp["total"])
		data := resp["data"].([]interface{})
		assert.Len(t, data, 4)
	})

	t.Run("FilterByName", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/versions?name=Standard", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, float64(1), resp["total"])
		data := resp["data"].([]interface{})
		assert.Len(t, data, 1)
		assert.Equal(t, "ESV", data[0].(map[string]interface{})["value"])
	})

	t.Run("FilterByLanguage", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/versions?language=spanish", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, float64(1), resp["total"])
		data := resp["data"].([]interface{})
		assert.Len(t, data, 1)
		assert.Equal(t, "RVR1960", data[0].(map[string]interface{})["value"])
	})

	t.Run("SortByName", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/versions?sort=name", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].([]interface{})
		assert.Equal(t, "English Standard Version", data[0].(map[string]interface{})["name"]) // E
		assert.Equal(t, "King James Version", data[1].(map[string]interface{})["name"])       // K
		assert.Equal(t, "La Bible du Semeur", data[2].(map[string]interface{})["name"])       // L
		assert.Equal(t, "Reina-Valera 1960", data[3].(map[string]interface{})["name"])        // R
	})

	t.Run("SortByCodeDefault", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/versions", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].([]interface{})
		// BDS, ESV, KJV, RVR1960
		assert.Equal(t, "BDS", data[0].(map[string]interface{})["value"])
		assert.Equal(t, "ESV", data[1].(map[string]interface{})["value"])
	})

	t.Run("Pagination", func(t *testing.T) {
		// Page 1, Limit 2
		req := httptest.NewRequest(http.MethodGet, "/versions?page=1&limit=2", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].([]interface{})
		assert.Len(t, data, 2)

		// Page 2, Limit 2
		req2 := httptest.NewRequest(http.MethodGet, "/versions?page=2&limit=2", nil)
		w2 := httptest.NewRecorder()
		h.ServeHTTP(w2, req2)

		var resp2 map[string]interface{}
		json.Unmarshal(w2.Body.Bytes(), &resp2)
		data2 := resp2["data"].([]interface{})
		assert.Len(t, data2, 2)
		assert.NotEqual(t, data[0], data2[0])

		// Page 3, Limit 2 (Empty)
		req3 := httptest.NewRequest(http.MethodGet, "/versions?page=3&limit=2", nil)
		w3 := httptest.NewRecorder()
		h.ServeHTTP(w3, req3)

		var resp3 map[string]interface{}
		json.Unmarshal(w3.Body.Bytes(), &resp3)
		data3 := resp3["data"].([]interface{})
		assert.Len(t, data3, 0)
	})

	t.Run("PaginationOutOfBounds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/versions?page=100&limit=10", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].([]interface{})
		assert.Len(t, data, 0)
	})

	t.Run("InvalidPaginationParams", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/versions?page=invalid&limit=-5", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		// Should verify defaults (page 1, limit 20)
		assert.Equal(t, float64(1), resp["page"])
		assert.Equal(t, float64(20), resp["limit"])
	})
}
