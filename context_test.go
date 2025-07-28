package disk

import (
	"testing"
	"time"
)

func TestContextManagement(t *testing.T) {
	t.Run("NewWithConfig creates client with custom timeout", func(t *testing.T) {
		config := &ClientConfig{
			DefaultTimeout: 45 * time.Second,
		}
		
		client, err := NewWithConfig(config, "test-token")
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
		
		if client.GetTimeout() != 45*time.Second {
			t.Errorf("Expected timeout to be 45s, got %v", client.GetTimeout())
		}
	})
	
	t.Run("SetTimeout updates client timeout", func(t *testing.T) {
		client, _ := New("test-token")
		client.SetTimeout(60 * time.Second)
		
		if client.GetTimeout() != 60*time.Second {
			t.Errorf("Expected timeout to be 60s, got %v", client.GetTimeout())
		}
	})
	
	t.Run("Context helper functions work correctly", func(t *testing.T) {
		// Test WithTimeout
		ctx, cancel := WithTimeout(5 * time.Second)
		defer cancel()
		
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("Expected context to have a deadline")
		}
		
		// Should be approximately 5 seconds from now
		expectedDeadline := time.Now().Add(5 * time.Second)
		if deadline.Before(expectedDeadline.Add(-100*time.Millisecond)) || 
		   deadline.After(expectedDeadline.Add(100*time.Millisecond)) {
			t.Error("Context deadline is not approximately 5 seconds from now")
		}
	})
	
	t.Run("Context cancellation is detected", func(t *testing.T) {
		ctx, cancel := WithCancel()
		cancel() // Cancel immediately
		
		client, _ := New("test-token")
		_, err := client.doRequest(ctx, GET, "", nil)
		
		if err == nil {
			t.Error("Expected error due to cancelled context")
		}
		
		if err.Error() != "request cancelled: context canceled" {
			t.Errorf("Expected cancellation error, got: %v", err)
		}
	})
}