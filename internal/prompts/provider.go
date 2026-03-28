package prompts

import "github.com/mihari-bot/bot/internal/container"

type Provider interface {
	Load(mp *container.Map[string, string]) error
}
