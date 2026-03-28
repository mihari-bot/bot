package bot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mihari-bot/bot/internal/core"
	"github.com/mihari-bot/bot/internal/model"
	"github.com/mihari-bot/bot/internal/voice"
	"gopkg.in/yaml.v3"

	"github.com/openai/openai-go/v3"
	"go.uber.org/zap"
)

func (b *Bot) chat(ctx context.Context, cc *cctx) error {
	past := time.Now()
	userID := cc.getUserID()
	userName := cc.getUserName()
	userLogger := cc.getLogger()
	userLock, _ := b.userLockMp.GetOrSet(userID, sync.Mutex{})
	userLock.Lock()
	defer userLock.Unlock()

	if time.Since(past) > time.Second*5 {
		userLogger.Warn("Dropped")
		return nil
	}

	past = time.Now()

	userConfig, err := b.dbLoadUserConfig(ctx, userID)
	if err != nil {
		return err
	}

	if userConfig.Rounds <= 0 {
		return ErrInvalidRounds
	}

	openAiClient, err := b.prepareOpenAIClient(ctx, userConfig)
	if err != nil {
		return err
	}

	// 加载角色
	rolePrompt, ok := b.chardefMp.Get(userConfig.Role)
	if !ok {
		return ErrInvalidRole
	}
	// 注入一些变量
	rolePrompt = strings.ReplaceAll(rolePrompt, "{{USERNAME}}", userName)
	// 我们还需重新加载一遍yaml, 移除注释
	var rolePromptInMap j
	err = yaml.Unmarshal([]byte(rolePrompt), &rolePromptInMap)
	if err != nil {
		return err
	}
	// 注入Memory
	rolePromptInMap["memories"] = []string{} // TODO: Memory

	if userConfig.VoiceEnabled {
		ttsNotice := "注意：你的所有输出会被送入TTS生成语音。请不要使用emoji，不要用括号描述动作（如（挥手）），也不要输出除对白以外的内容。请尽可能在让用户满意的同时，保持内容的简短。"
		if v, ok := rolePromptInMap["speech_style"]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				rolePromptInMap["speech_style"] = strings.TrimSpace(s) + "\n" + ttsNotice
			} else {
				rolePromptInMap["speech_style"] = ttsNotice
			}
		} else {
			rolePromptInMap["speech_style"] = ttsNotice
		}

		if v, ok := rolePromptInMap["must_not_do"]; ok {
			switch vv := v.(type) {
			case []any:
				rolePromptInMap["must_not_do"] = append(vv, ttsNotice)
			case []string:
				rolePromptInMap["must_not_do"] = append(vv, ttsNotice)
			default:
				rolePromptInMap["must_not_do"] = []string{ttsNotice}
			}
		} else {
			rolePromptInMap["must_not_do"] = []string{ttsNotice}
		}
	}

	rolePromptInBytes, err := yaml.Marshal(rolePromptInMap)
	if err != nil {
		return err
	}
	rolePrompt = string(rolePromptInBytes)

	ccmpus, err := b.dbLoadCcmpu(ctx, userID, userConfig.Role, userConfig.Rounds)
	if err != nil {
		return err
	}

	if userConfig.VoiceEnabled {
		voiceClient, err := b.mustVoiceClient(ctx, userConfig)
		if err != nil {
			return err
		}

		return b.chatVoice(ctx, cc, userLogger, userConfig, openAiClient, rolePrompt, ccmpus, past, voiceClient)
	}

	return b.chatText(ctx, cc, userLogger, userConfig, openAiClient, rolePrompt, ccmpus, past)
}

func (b *Bot) chatText(ctx context.Context, cc *cctx, userLogger *zap.SugaredLogger, userConfig *model.Config, openAiClient *openai.Client, rolePrompt string, ccmpus []openai.ChatCompletionMessageParamUnion, past time.Time) error {
	message := cc.getMessage()
	textClient := core.NewTextClient(openAiClient, []openai.ChatCompletionToolUnionParam{}, rolePrompt, userConfig.APIModel, core.TextClientWithLogger(b.logger.Named("TextClient")), core.TextClientWithFlushMode(core.TextFlushModeEachLine))

	lastSendTime := time.Now()
	ch := textClient.Stream(ctx, []openai.ChatCompletionMessageParamUnion{
		b.prepareEachRoundSystemMessage(past),
		openai.UserMessage(message),
	}, ccmpus)

	cc.sendTyping(ctx)
	for out := range ch {
		cc.sendTyping(ctx)

		waitTime := b.mthClamp(
			b.mthCalcWaitTime(out, b.runtimeConfig.perCharDelayMin, b.runtimeConfig.perCharDelayMax),
			b.runtimeConfig.minWaitTime,
			b.runtimeConfig.maxWaitTime,
		)

		elapsed := time.Since(lastSendTime)
		if waitTime > elapsed {
			sleepDur := waitTime - elapsed
			time.Sleep(sleepDur)
		}

		_, err := cc.sendMessage(ctx, out)
		if err != nil {
			return err
		}

		lastSendTime = time.Now()

		userLogger.Debugw("已发送消息",
			"waitTime", waitTime,
			"out", out)
	}
	if err := textClient.GetErr(); err != nil {
		return err
	}

	return b.dbAppendCcmpu(ctx, cc.getUserID(), userConfig.Role, textClient.GetDiffed())
}

func (b *Bot) chatVoice(ctx context.Context, cc *cctx, userLogger *zap.SugaredLogger, userConfig *model.Config, openAiClient *openai.Client, rolePrompt string, ccmpus []openai.ChatCompletionMessageParamUnion, past time.Time, voiceClient *voice.Client) error {
	message := cc.getMessage()
	textClient := core.NewTextClient(openAiClient, []openai.ChatCompletionToolUnionParam{}, rolePrompt, userConfig.APIModel, core.TextClientWithLogger(b.logger.Named("TextClient")), core.TextClientWithFlushMode(core.TextFlushModeEachLine))

	lastSendTime := time.Now()
	ch := textClient.Stream(ctx, []openai.ChatCompletionMessageParamUnion{
		b.prepareEachRoundSystemMessage(past),
		openai.UserMessage(message),
	}, ccmpus)

	cc.sendTyping(ctx)
	for out := range ch {
		cc.sendTyping(ctx)

		waitTime := b.mthClamp(
			b.mthCalcWaitTime(out, b.runtimeConfig.perCharDelayMin, b.runtimeConfig.perCharDelayMax),
			b.runtimeConfig.minWaitTime,
			b.runtimeConfig.maxWaitTime,
		)

		elapsed := time.Since(lastSendTime)
		if waitTime > elapsed {
			sleepDur := waitTime - elapsed
			time.Sleep(sleepDur)
		}

		text := strings.TrimSpace(out)
		if text != "" {
			userLogger.Infow("voice_text",
				"text", text)
		}

		if text == "" {
			continue
		}
		if voiceClient == nil {
			return fmt.Errorf("voice client not ready")
		} else {
			voiceRole := strings.TrimSpace(userConfig.VoiceRole)
			if voiceRole == "" {
				return fmt.Errorf("voice role is not set")
			}

			audio, ct, err := voiceClient.Generate(ctx, voice.GenerateRequest{Text: text, Role: voiceRole})
			if err != nil {
				return err
			} else {
				userLogger.Infow("voice_generated",
					"bytes", len(audio),
					"contentType", ct)
				id := b.voiceTempPut(audio, ct)
				url := b.voiceTempURL(id)
				msg := j{
					"type": "record",
					"data": j{
						"file": url,
					},
				}
				cc.getLogger().Infow("send url", "url", url)
				_, err := cc.sendMessage(ctx, msg)
				if err != nil {
					return err
				}
			}
		}

		lastSendTime = time.Now()

		userLogger.Debugw("已发送消息",
			"waitTime", waitTime,
			"out", out)
	}
	if err := textClient.GetErr(); err != nil {
		return err
	}

	return b.dbAppendCcmpu(ctx, cc.getUserID(), userConfig.Role, textClient.GetDiffed())
}

func (b *Bot) mustVoiceClient(ctx context.Context, cfg *model.Config) (*voice.Client, error) {
	baseURL := strings.TrimSpace(cfg.VoiceBaseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("voice base url is empty")
	}
	if strings.TrimSpace(cfg.VoiceAuthorization) == "" {
		return nil, fmt.Errorf("voice authorization is empty")
	}
	if err := b.checkBaseURL(ctx, baseURL); err != nil {
		return nil, fmt.Errorf("invalid voice baseurl: %w", err)
	}
	return voice.New(baseURL, voice.WithAuthorization(cfg.VoiceAuthorization))
}
