package delivery

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/sse"

	sselib "hrms/internal/pkg/sse"
)

type EventHandler struct {
	hub *sselib.Hub
}

func NewEventHandler(hub *sselib.Hub) *EventHandler {
	return &EventHandler{hub: hub}
}

func (h *EventHandler) Stream(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	return sse.New(sse.Config{
		Handler: func(c fiber.Ctx, stream *sse.Stream) error {
			punchCh, err := h.hub.Subscribe(stream.Context(), "punches")
			if err != nil {
				return err
			}
			notifCh, err := h.hub.Subscribe(stream.Context(), "notifications:"+userID)
			if err != nil {
				return err
			}

			for {
				select {
				case msg, ok := <-punchCh:
					if !ok {
						return nil
					}
					var data interface{}
					if json.Unmarshal([]byte(msg), &data) == nil {
						if err := stream.Event(sse.Event{Name: "punch", Data: data}); err != nil {
							return err
						}
					}
				case msg, ok := <-notifCh:
					if !ok {
						return nil
					}
					var data interface{}
					if json.Unmarshal([]byte(msg), &data) == nil {
						if err := stream.Event(sse.Event{Name: "notification", Data: data}); err != nil {
							return err
						}
					}
				case <-stream.Done():
					return stream.Err()
				}
			}
		},
	})(c)
}
