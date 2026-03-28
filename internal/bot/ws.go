package bot

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/coder/websocket"
	"github.com/tidwall/gjson"
)

func (b *Bot) openWS(ctx context.Context, wsURL string, accessToken string) error {
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil
	}

	uq := u.Query()
	uq.Add("access_token", accessToken)
	u.RawQuery = uq.Encode()

	wsURL = u.String()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		return err
	}
	conn.SetReadLimit(-1)

	return b.serveWS(ctx, conn)
}

func (b *Bot) serveWS(ctx context.Context, conn *websocket.Conn) error {
	b.logger.Info("监听WS")

	for {
		// 读取消息
		pkTp, pkInBytes, err := conn.Read(ctx)
		if errors.Is(err, context.Canceled) {
			b.logger.Info("退出监听")
			break
		}
		if err != nil {
			return err
		}
		if pkTp != websocket.MessageText {
			b.logger.Warn("未知消息类型")
			continue
		}

		// 解析包
		pk := gjson.ParseBytes(pkInBytes)

		// 分发echo
		if echo := pk.Get("echo"); echo.Type == gjson.Number {
			go b.echoDispatch(echo.Int(), &pk)
			continue
		}

		cc := newCctx(b, conn, &pk)
		fn, ok := b.callmpLookup(&pk)
		if !ok {
			b.logger.Errorw("未实现这个packet的handler",
				"Data", string(pkInBytes))
		}
		if fn == nil {
			continue
		}

		go func() {
			err = fn(ctx, cc)
			if err != nil {
				b.logger.Errorw("Handler错误",
					err)
				time.Sleep(time.Second)
				cc.sendMessage(ctx, "呜哇！程序出了一点点意外错误呢... 💦\n"+err.Error())
			}
		}()
	}
	return nil
}
