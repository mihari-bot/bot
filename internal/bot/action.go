package bot

import (
	"context"
	"fmt"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/tidwall/gjson"
)

func (b *Bot) actionSend(ctx context.Context, conn *websocket.Conn, action string, params any) (*gjson.Result, error) {
	echo, waitDispatch := b.echoCreate()
	req := j{
		"action": action,
		"params": params,
		"echo":   echo,
	}

	err := wsjson.Write(ctx, conn, req)
	if err != nil {
		return nil, err
	}

	respInJSON := waitDispatch()
	respStatus := respInJSON.Get("status").String()
	respMessage := respInJSON.Get("message").String()
	respData := respInJSON.Get("data")
	if respStatus != "ok" {
		return nil, fmt.Errorf("bad request: \"%s\"", respMessage)
	}
	return &respData, nil
}

func (b *Bot) actionSendPrivate(ctx context.Context, conn *websocket.Conn, userid int64, message any) (int64, error) {
	result, err := b.actionSend(ctx, conn, "send_private_msg", j{
		"user_id": userid,
		"message": message,
	})
	if err != nil {
		return 0, fmt.Errorf("failed send private msg: %w", err)
	}

	return result.Get("message_id").Int(), nil
}
