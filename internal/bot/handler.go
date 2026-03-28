package bot

import (
	"context"
	"strings"
)

func (b *Bot) handlerPrivateChat(ctx context.Context, cc *cctx) error {
	userName, message, userID := cc.getUserName(), cc.getMessage(), cc.getUserID()

	b.logger.Infow("收到消息",
		"name", userName,
		"message", message,
		"userID", userID)

	if strings.HasPrefix(message, "/") {
		return b.cmd(ctx, cc)
	}

	return b.chat(ctx, cc)
}
