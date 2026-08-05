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

func (app *Service) GetMyNotifications(ctx context.Context) ([]models.Notification, error) {
	res := []models.Notification{}
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return res, apierrors.ErrInternalServer
	}

	res, err := app.db.GetMyNotifications(ctx, claims.UserID)
	return res, err
}

func (app *Service) SetReadStatusOfNotification(ctx context.Context) error {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	err := app.db.SetReadStatusOfNotification(ctx, claims.UserID)
	return err
}

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

	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrJSONEncodeFailed, err)
	}

	app.Hub.Send(notification.UserID, string(notificationJSON))

	return nil
}

type Hub struct {
	hub map[int]*Client
	mu  sync.RWMutex
}

func newHub() *Hub {
	return &Hub{
		hub: make(map[int]*Client),
	}
}

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

func (h *Hub) Send(userID int, msg string) {
	h.mu.RLock()
	client, ok := h.hub[userID]
	if !ok {
		log.Println("Not user in hub.")
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

type Client struct {
	UserID int
	Conn   *websocket.Conn
	Send   chan string
	Hub    *Hub
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		err := c.Conn.WriteJSON(msg)
		if err != nil {
			log.Println(err)
			break
		}
	}

}

func (c *Client) ReadPump() {
	defer func() {
		c.Conn.Close()
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
