package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockClient(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	t.Run("CreateAPIKey", func(t *testing.T) {
		key, err := client.CreateAPIKey(ctx, "test-client", 100)
		assert.NoError(t, err)
		assert.NotEmpty(t, key.Key)
		assert.Equal(t, "test-client", key.ClientName)
		assert.Equal(t, 100, key.DailyLimit)
		assert.True(t, key.Active)
	})

	t.Run("GetAPIKey", func(t *testing.T) {
		created, _ := client.CreateAPIKey(ctx, "getter", 50)

		got, err := client.GetAPIKey(ctx, created.Key)
		assert.NoError(t, err)
		assert.Equal(t, created.Key, got.Key)

		// Not found
		none, err := client.GetAPIKey(ctx, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, none)
	})

	t.Run("IncrementUsage", func(t *testing.T) {
		key, _ := client.CreateAPIKey(ctx, "limiter", 2)

		// 1st request
		count, exceeded, err := client.IncrementUsage(ctx, key.Key)
		assert.NoError(t, err)
		assert.False(t, exceeded)
		assert.Equal(t, 1, count)

		// 2nd request
		count, exceeded, err = client.IncrementUsage(ctx, key.Key)
		assert.NoError(t, err)
		assert.False(t, exceeded)
		assert.Equal(t, 2, count)

		// 3rd request (Limit is 2, so this should fail? Or match limit?)
		// Logic: if apiKey.RequestCount >= apiKey.DailyLimit { exceeded = true }
		// Current count is 2. Limit is 2.
		// Next call checks `if apiKey.RequestCount >= apiKey.DailyLimit`. 2 >= 2 is True.
		// Returns count (2), true.

		count, exceeded, err = client.IncrementUsage(ctx, key.Key)
		assert.NoError(t, err)
		assert.True(t, exceeded)
		assert.Equal(t, 2, count)
	})

	t.Run("IncrementUsage_DateReset", func(t *testing.T) {
		key, _ := client.CreateAPIKey(ctx, "resetter", 10)

		// Manually tamper with the key in the map to simulate yesterday
		client.mu.Lock()
		k := client.Keys[key.Key]
		k.LastUsageDate = "2000-01-01"
		k.RequestCount = 5
		client.Keys[key.Key] = k
		client.mu.Unlock()

		// Increment should reset
		count, exceeded, err := client.IncrementUsage(ctx, key.Key)
		assert.NoError(t, err)
		assert.False(t, exceeded)
		assert.Equal(t, 1, count)

		// Verify date is today
		updated, _ := client.GetAPIKey(ctx, key.Key)
		assert.Equal(t, time.Now().Format("2006-01-02"), updated.LastUsageDate)
	})

	t.Run("Close", func(t *testing.T) {
		err := client.Close()
		assert.NoError(t, err)
	})
}
