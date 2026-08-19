package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// GetMyNotifications returns the notifications of the authenticated user.
func (app *Service) GetMyNotifications(ctx context.Context) ([]models.Notification, error) {
	res := []models.Notification{}
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return res, apierrors.ErrInternalServer
	}

	res, err := app.db.GetMyNotifications(ctx, claims.UserID)
	return res, err
}

// SetReadStatusOfNotification marks all notifications of the authenticated user as read.
func (app *Service) SetReadStatusOfNotification(ctx context.Context) error {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	err := app.db.SetReadStatusOfNotification(ctx, claims.UserID)
	return err
}

// InvateIntern sends an internship invitation to an intern and notifies them.
func (app *Service) InvateIntern(ctx context.Context, internshipID int, invateIntern models.InvateIntern) error {
	err := app.validate.Struct(invateIntern)
	if err != nil {
		return apierrors.Wrap(err, apierrors.ErrBadRequest)
	}

	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	companyID, err := app.db.GetCompanyIDByUserID(ctx, claims.UserID)
	if err != nil {
		return err
	}

	notification, err := app.db.NewInviteNotification(ctx, invateIntern.UserID, internshipID, companyID, invateIntern.Message)
	if err != nil {
		return err
	}

	app.sendRecsysAction(ctx, invateIntern.UserID, internshipID, constants.RecsysActionInvate)

	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrJSONEncodeFailed, err)
	}

	app.Hub.Send(notification.UserID, string(notificationJSON))

	return nil
}

// Hub manages the websocket connections of online users.
type Hub struct {
	hub map[int]*Client
	mu  sync.RWMutex
}

func newHub() *Hub {
	return &Hub{
		hub: make(map[int]*Client),
	}
}

// Register registers the given client for the given user.
func (h *Hub) Register(userID int, client *Client) {
	h.mu.Lock()
	h.hub[userID] = client
	h.mu.Unlock()
}

func (h *Hub) unregister(userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, ok := h.hub[userID]
	if !ok {
		log.Println("Not user in hub.")
		return
	}

	close(client.Send)
	delete(h.hub, userID)
}

// GetOnlineCount returns the number of currently connected users.
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.hub)
}

// KickUser disconnects the user with the given ID and sends the given reason.
func (h *Hub) KickUser(userID int, reason string) {
	h.mu.RLock()
	client, ok := h.hub[userID]
	if !ok {
		log.Println("Not user in hub.")
		h.mu.RUnlock()
		return
	}
	h.mu.RUnlock()

	select {
	case client.Send <- reason:
	default:
	}

	h.unregister(userID)
}

// Send sends the given message to the user with the given ID.
func (h *Hub) Send(userID int, msg string) {
	h.mu.RLock()
	client, ok := h.hub[userID]
	if !ok {
		log.Println("Not user in hub.")
		h.mu.RUnlock()
		return
	}
	h.mu.RUnlock()
	select {
	case client.Send <- msg:
	default:
		h.mu.Lock()
		close(client.Send)
		delete(h.hub, userID)
		h.mu.Unlock()
	}

}

// Client represents a websocket connection of a user.
type Client struct {
	UserID int
	Conn   *websocket.Conn
	Send   chan string
	Hub    *Hub
}

// WritePump writes messages from the Send channel to the websocket connection.
func (c *Client) WritePump() {
	defer func() {
		if cerr := c.Conn.Close(); cerr != nil {
			log.Printf("ws: close connection: %v", cerr)
		}
	}()

	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println(err)
			break
		}
	}

}

// ReadPump reads messages from the websocket connection and responds to pings.
func (c *Client) ReadPump() {
	defer func() {
		if cerr := c.Conn.Close(); cerr != nil {
			log.Printf("ws: close connection: %v", cerr)
		}
		c.Hub.unregister(c.UserID)
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		if string(msg) == "ping" {
			select {
			case c.Send <- "pong":
			default:
				log.Println("Send channel closed.")
			}

		}
	}

}
