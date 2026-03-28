package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/mihari-bot/bot/internal/model"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func (b *Bot) prepareOpenAIClient(ctx context.Context, userConfig *model.Config) (*openai.Client, error) {
	// 检查是否完善了配置
	if err := b.checkBaseURL(ctx, userConfig.APIBaseURL); err != nil {
		return nil, fmt.Errorf("invalid baseurl: %w", err)
	}

	openaiClient := openai.NewClient(option.WithBaseURL(userConfig.APIBaseURL), option.WithAPIKey(userConfig.APIKey))
	return &openaiClient, nil
}

func (b *Bot) prepareEachRoundSystemMessage(past time.Time) openai.ChatCompletionMessageParamUnion {
	return openai.SystemMessage(fmt.Sprintf("当前时间:%s", past.Format(time.DateTime)))
}
