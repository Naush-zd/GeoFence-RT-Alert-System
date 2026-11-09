package handlers

import (
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func (h Handler) AlertsWebSocket(c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	h.Hub.Add(conn)
	defer h.Hub.Remove(conn)
	defer conn.Close(websocket.StatusNormalClosure, "connection closed")

	for {
		if _, _, err := conn.Read(c.Request.Context()); err != nil {
			return
		}
	}
}
