package router

import (
	"errors"
	"testing"

	"udp-hole-punch/pkg/models"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()

	if router == nil {
		t.Fatal("NewRouter() returned nil")
	}

	if router.methods == nil {
		t.Error("NewRouter() returned router with nil methods map")
	}

	if len(router.methods) != 0 {
		t.Errorf("NewRouter() returned router with non-empty methods map, got %d", len(router.methods))
	}
}

func TestRouter_AddRoute(t *testing.T) {
	router := NewRouter()

	testHandler := func(client *models.Client, payload string) error {
		return nil
	}

	result := router.AddRoute("test", testHandler)

	if result != router {
		t.Error("AddRoute() did not return self for chaining")
	}

	if len(router.methods) != 1 {
		t.Errorf("AddRoute() methods count = %d, want 1", len(router.methods))
	}

	if router.methods["test"] == nil {
		t.Error("AddRoute() did not add handler for 'test' event")
	}
}

func TestRouter_AddMultipleRoutes(t *testing.T) {
	router := NewRouter()

	handler1 := func(client *models.Client, payload string) error { return nil }
	handler2 := func(client *models.Client, payload string) error { return nil }
	handler3 := func(client *models.Client, payload string) error { return nil }

	router.AddRoute("event1", handler1).
		AddRoute("event2", handler2).
		AddRoute("event3", handler3)

	if len(router.methods) != 3 {
		t.Errorf("AddRoute() methods count = %d, want 3", len(router.methods))
	}

	events := []string{"event1", "event2", "event3"}
	for _, event := range events {
		if router.methods[event] == nil {
			t.Errorf("AddRoute() did not add handler for '%s' event", event)
		}
	}
}

func TestRouter_HandleEvent_ValidEvent(t *testing.T) {
	router := NewRouter()
	handlerCalled := false

	testHandler := func(client *models.Client, payload string) error {
		handlerCalled = true
		if payload != `{"test":"data"}` {
			t.Errorf("Handler received payload = %s, want %s", payload, `{"test":"data"}`)
		}
		return nil
	}

	router.AddRoute("test", testHandler)

	// Create a test client (you'll need to adjust this based on your Client implementation)
	client := models.NewClient()

	request := `{"event":"test","payload":"{\"test\":\"data\"}"}`
	router.HandleEvent(client, []byte(request))

	if !handlerCalled {
		t.Error("HandleEvent() did not call the registered handler")
	}
}

func TestRouter_HandleEvent_InvalidJSON(t *testing.T) {
	router := NewRouter()

	router.AddRoute("test", func(client *models.Client, payload string) error {
		t.Error("Handler should not be called for invalid JSON")
		return nil
	})

	client := models.NewClient()
	invalidJSON := []byte(`{invalid json}`)

	// Should not panic, just log error
	router.HandleEvent(client, invalidJSON)
}

func TestRouter_HandleEvent_UnregisteredEvent(t *testing.T) {
	router := NewRouter()

	router.AddRoute("registered", func(client *models.Client, payload string) error {
		t.Error("Handler should not be called for unregistered event")
		return nil
	})

	client := models.NewClient()
	request := `{"event":"unregistered","payload":""}`

	// Should not panic, just log that event not found
	router.HandleEvent(client, []byte(request))
}

func TestRouter_HandleEvent_HandlerError(t *testing.T) {
	router := NewRouter()
	testError := errors.New("test error")

	router.AddRoute("test", func(client *models.Client, payload string) error {
		return testError
	})

	client := models.NewClient()
	request := `{"event":"test","payload":""}`

	// Should not panic even if handler returns error
	router.HandleEvent(client, []byte(request))
}
