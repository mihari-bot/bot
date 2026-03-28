package bot

import (
	"context"
	"fmt"

	"github.com/coder/websocket"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type cctx struct {
	bot  *Bot
	conn *websocket.Conn
	pk   *gjson.Result
}

func newCctx(bot *Bot, conn *websocket.Conn, pk *gjson.Result) *cctx {
	return &cctx{bot: bot, conn: conn, pk: pk}
}

func (cc *cctx) getMessage() string {
	return cc.pk.Get("raw_message").String()
}

func (cc *cctx) getUserID() int64 {
	return cc.pk.Get("sender.user_id").Int()
}

func (cc *cctx) getUserName() string {
	return cc.pk.Get("sender.nickname").String()
}

func (cc *cctx) getLogger() *zap.SugaredLogger {
	return cc.bot.logger.Named(fmt.Sprintf("Cctx-%d-\"%s\"", cc.getUserID(), cc.getUserName()))
}

func (cc *cctx) sendMessage(ctx context.Context, message any) (int64, error) {
	return cc.bot.actionSendPrivate(ctx, cc.conn, cc.getUserID(), message)
}

func (cc *cctx) sendTyping(ctx context.Context) error {
	_, err := cc.bot.actionSend(ctx, cc.conn, "set_input_status", j{
		"user_id":    cc.getUserID(),
		"event_type": 1,
	})
	return err
}
