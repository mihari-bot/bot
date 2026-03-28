package bot

import (
	"context"

	"github.com/tidwall/gjson"
)

type cmpDef = map[string]string
type cmpFn = func(ctx context.Context, cc *cctx) error
type cmpItem = struct {
	def cmpDef
	fn  cmpFn
}
type callMp = []cmpItem

// CallmpRegister 注册一个消息处理器。
func (b *Bot) CallmpRegister(def cmpDef, fn cmpFn) {
	b.callMp = append(b.callMp, cmpItem{def, fn})
}

func (b *Bot) callmpLookup(pk *gjson.Result) (cmpFn, bool) {
	for _, i := range b.callMp {
		finded := true
		for k, v := range i.def {
			if j := pk.Get(k); !(j.Exists() && j.String() == v) {
				finded = false
				break
			}
		}

		if finded {
			return i.fn, true
		}
	}
	return nil, false
}

func (b *Bot) callmapInit() {
	b.CallmpRegister(cmpDef{
		"post_type":    "message",
		"message_type": "private",
		"sub_type":     "friend",
	}, b.handlerPrivateChat)
	b.CallmpRegister(cmpDef{
		"post_type": "meta_event",
	}, nil)
	b.CallmpRegister(cmpDef{
		"post_type":   "notice",
		"notice_type": "notify",
		"sub_type":    "input_status",
	}, nil)
}
