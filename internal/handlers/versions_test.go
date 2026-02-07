package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"bible-api-service/internal/bible"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersionsHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if err := os.MkdirAll("tmp", 0755); err != nil {
			t.Fatal(err)
		}
		tmpDir, err := os.MkdirTemp("tmp", "TestNewVersionsHandler*")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			os.RemoveAll(tmpDir)
		})
		configPath := filepath.Join(tmpDir, "versions.yaml")
		content := `
- name: English Standard Version
  code: ESV
  language: English
- name: King James Version
  code: KJV
  language: English
`
		err = os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		vm, err := bible.NewVersionManager(configPath)
		require.NoError(t, err)

		h := NewVersionsHandler(vm)
		versions := h.manager.GetAll()
		assert.Len(t, versions, 2)
		assert.Equal(t, "ESV", versions[0].Code)
	})
}

// Helper to create handler with pre-populated versions
func createTestVersionsHandler(t *testing.T, versions []bible.Version) *VersionsHandler {
	return nil
}

// Copied from previous attempt but modified to use file creation
func TestVersionsHandler_ListVersions(t *testing.T) {
	content := `
- name: English Standard Version
  code: ESV
  language: English
- name: King James Version
  code: KJV
  language: English
- name: Reina-Valera 1960
  code: RVR1960
  language: Spanish
- name: La Bible du Semeur
  code: BDS
  language: French
`
	if err := os.MkdirAll("tmp", 0755); err != nil {
		t.Fatal(err)
	}
	tmpDir, err := os.MkdirTemp("tmp", "TestVersionsHandler_ListVersions*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	configPath := filepath.Join(tmpDir, "versions.yaml")
	err = os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	vm, err := bible.NewVersionManager(configPath)
	require.NoError(t, err)

	h := NewVersionsHandler(vm)

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
		assert.Equal(t, "ESV", data[0].(map[string]interface{})["code"])
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
		assert.Equal(t, "RVR1960", data[0].(map[string]interface{})["code"])
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
		assert.Equal(t, "BDS", data[0].(map[string]interface{})["code"])
		assert.Equal(t, "ESV", data[1].(map[string]interface{})["code"])
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
