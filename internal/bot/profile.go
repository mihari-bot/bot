package bot

import (
	"fmt"
	"os"
	"time"

	"github.com/tidwall/gjson"
)

func (b *Bot) profileGetString(query string) (string, error) {
	obj := b.profile.Get(query)
	if !obj.Exists() {
		return "", fmt.Errorf("key not found: need key \"%s\"", query)
	}
	if obj.Type != gjson.String {
		return "", fmt.Errorf("type mismatched: need \"%s\" on \"%s\" but got \"%s\"", gjson.String, query, obj.Type)
	}

	return obj.String(), nil
}

func (b *Bot) profileGetStringOr(query string, def string) string {
	obj := b.profile.Get(query)
	if !obj.Exists() {
		return def
	}
	if obj.Type != gjson.String {
		return def
	}
	return obj.String()
}

func (b *Bot) profileGetInt(query string) (int64, error) {
	obj := b.profile.Get(query)
	if !obj.Exists() {
		return 0, fmt.Errorf("key not found: need key \"%s\"", query)
	}
	if obj.Type != gjson.Number {
		return 0, fmt.Errorf("type mismatched: need \"%s\" on \"%s\" but got \"%s\"", gjson.Number, query, obj.Type)
	}

	return obj.Int(), nil
}

func (b *Bot) profileGetBoolOr(query string, def bool) bool {
	obj := b.profile.Get(query)
	if !obj.Exists() {
		return def
	}
	if obj.Type != gjson.True && obj.Type != gjson.False {
		return def
	}
	return obj.Bool()
}

func (b *Bot) profileGetStringArrayOr(query string, def []string) []string {
	obj := b.profile.Get(query)
	if !obj.Exists() {
		return def
	}
	if !obj.IsArray() {
		return def
	}

	res := make([]string, 0, len(obj.Array()))
	obj.ForEach(func(_, v gjson.Result) bool {
		if v.Type == gjson.String {
			res = append(res, v.String())
		}
		return true
	})
	return res
}

func (b *Bot) profileInit() error {
	fp, err := b.pthResolve("config.json")
	if err != nil {
		return err
	}

	profileInBytes, err := os.ReadFile(fp)
	if err != nil {
		return err
	}

	res := gjson.ParseBytes(profileInBytes)
	b.profile = &res

	// Fill runtimeConfig
	perCharDelayMin, perCharDelayMax, err := b.profileGetPerCharDelayRange()
	if err != nil {
		return err
	}
	minWaitTime, maxWaitTime, err := b.profileGetWaitTimeRange()
	if err != nil {
		return err
	}
	b.runtimeConfig.perCharDelayMin, b.runtimeConfig.perCharDelayMax = perCharDelayMin, perCharDelayMax
	b.runtimeConfig.minWaitTime, b.runtimeConfig.maxWaitTime = minWaitTime, maxWaitTime

	return nil
}

func (b *Bot) profileGetWsConfig() ( /* wsURL */ string /* accessToken */, string, error) {
	wsURL, err := b.profileGetString("ws.url")
	if err != nil {
		return "", "", err
	}

	accessToken, err := b.profileGetString("ws.accessToken")
	if err != nil {
		return "", "", err
	}

	return wsURL, accessToken, nil
}

func (b *Bot) profileGetDbConfig() ( /* provider */ string /* dsn */, string, error) {
	provider, err := b.profileGetString("db.provider")
	if err != nil {
		return "", "", err
	}

	dsn, err := b.profileGetString("db.dsn")
	if err != nil {
		return "", "", err
	}
	return provider, dsn, nil
}

func (b *Bot) profileGetPerCharDelayRange() ( /* min */ time.Duration /* max */, time.Duration, error) {
	minDelayInString, err := b.profileGetString("chat.perCharDelay.min")
	if err != nil {
		return 0, 0, err
	}
	minDelay, err := time.ParseDuration(minDelayInString)
	if err != nil {
		return 0, 0, err
	}

	maxDelayInString, err := b.profileGetString("chat.perCharDelay.max")
	if err != nil {
		return 0, 0, err
	}
	maxDelay, err := time.ParseDuration(maxDelayInString)
	if err != nil {
		return 0, 0, err
	}

	return minDelay, maxDelay, err
}

func (b *Bot) profileGetWaitTimeRange() ( /* min */ time.Duration /* max */, time.Duration, error) {
	minWaitTimeInString, err := b.profileGetString("chat.waitTime.min")
	if err != nil {
		return 0, 0, err
	}
	minWaitTime, err := time.ParseDuration(minWaitTimeInString)
	if err != nil {
		return 0, 0, err
	}

	maxWaitTimeInString, err := b.profileGetString("chat.waitTime.max")
	if err != nil {
		return 0, 0, err
	}
	maxWaitTime, err := time.ParseDuration(maxWaitTimeInString)
	if err != nil {
		return 0, 0, err
	}

	return minWaitTime, maxWaitTime, err
}
