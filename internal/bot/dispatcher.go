package bot

import "github.com/tidwall/gjson"

func (b *Bot) echoDispatch(echo int64, pk *gjson.Result) error {
	ch, has := b.echoMp.Get(echo)
	if !has {
		return ErrInvalidEcho
	}

	ch <- pk
	close(ch)

	return nil
}

func (b *Bot) echoCreate() (int64, func() *gjson.Result) {
	ch := make(chan *gjson.Result)
	echo := b.echoCounter.Add(1)
	b.echoMp.Set(echo, ch)

	return echo, func() *gjson.Result {
		res := <-ch
		b.echoMp.Delete(echo)
		return res
	}
}
