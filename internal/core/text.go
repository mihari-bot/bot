package core

import (
	"context"
	"strings"

	"github.com/openai/openai-go/v3"
	"go.uber.org/zap"
)

// TextFlushMode 代表换行模式
type TextFlushMode int

// 换行模式的枚举
const (
	// TextFlushModeTwoLine 每两行刷新一次输出
	TextFlushModeTwoLine TextFlushMode = iota
	// TextFlushModeEachLine 每行刷新一次输出
	TextFlushModeEachLine
)

// ToolExecutor 用来执行某个工具
type ToolExecutor func(ctx context.Context, name, args string) (string, error)

// TextClient 封装了大模型调用并支持分段发送
type TextClient struct {
	openaiClient *openai.Client
	tools        []openai.ChatCompletionToolUnionParam
	toolExecutor ToolExecutor
	systemPrompt openai.ChatCompletionMessageParamUnion
	model        string
	logger       *zap.SugaredLogger
	flushMode    TextFlushMode
	err          error
	done         bool
	diffed       []openai.ChatCompletionMessageParamUnion
}

// TextOption 是 TextClient 的配置选项函数类型
type TextOption func(*TextClient)

// TextClientWithToolExecutor 设置 TextClient 的工具执行器
func TextClientWithToolExecutor(e ToolExecutor) TextOption {
	return func(c *TextClient) { c.toolExecutor = e }
}

// TextClientWithLogger 设置 TextClient 的日志记录器
func TextClientWithLogger(l *zap.SugaredLogger) TextOption {
	return func(c *TextClient) { c.logger = l }
}

// TextClientWithFlushMode 设置 TextClient 的换行刷新模式
func TextClientWithFlushMode(m TextFlushMode) TextOption {
	return func(c *TextClient) { c.flushMode = m }
}

// NewTextClient 创建一个新的 TextClient 实例
func NewTextClient(client *openai.Client, tools []openai.ChatCompletionToolUnionParam, sys, model string, opts ...TextOption) *TextClient {
	c := &TextClient{
		openaiClient: client,
		tools:        tools,
		systemPrompt: openai.SystemMessage(sys),
		model:        model,
		logger:       zap.NewNop().Sugar(),
		flushMode:    TextFlushModeTwoLine,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Stream 启动流式对话，返回一个字符串通道用于接收分段响应
// ctx: 上下文控制，支持取消和超时
// thisRound: 当前轮次的消息列表
// histories: 历史对话消息列表
func (c *TextClient) Stream(ctx context.Context, thisRound, histories []openai.ChatCompletionMessageParamUnion) <-chan string {
	out := make(chan string)
	if c.done {
		close(out)
		return out
	}
	c.done = true

	go func() {
		defer close(out)

		messages := append(append([]openai.ChatCompletionMessageParamUnion{c.systemPrompt}, histories...), thisRound...)
		baseLen := len(messages) - len(thisRound)

		suffix := "\n\n"
		if c.flushMode == TextFlushModeEachLine {
			suffix = "\n"
		}

		for {
			stream := c.openaiClient.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
				Messages: messages,
				Model:    openai.ChatModel(c.model),
				Tools:    c.tools,
			})

			var buf, full strings.Builder
			acc := &openai.ChatCompletionAccumulator{}
			var tcs []openai.FinishedChatCompletionToolCall

			for stream.Next() {
				chunk := stream.Current()
				acc.AddChunk(chunk)
				if len(chunk.Choices) == 0 {
					continue
				}
				if tc, ok := acc.JustFinishedToolCall(); ok {
					tcs = append(tcs, tc)
				}

				if delta := chunk.Choices[0].Delta.Content; delta != "" {
					full.WriteString(delta)
					buf.WriteString(delta)

					s := buf.String()
					for idx := strings.Index(s, suffix); idx != -1; idx = strings.Index(s, suffix) {
						select {
						case out <- s[:idx+len(suffix)]:
							s = s[idx+len(suffix):]
						case <-ctx.Done():
							return
						}
					}
					buf.Reset()
					buf.WriteString(s)
				}
			}

			if c.err = stream.Err(); c.err != nil {
				c.logger.Errorw("stream error",
					c.err)
				break
			}
			if buf.Len() > 0 {
				select {
				case out <- buf.String():
				case <-ctx.Done():
					return
				}
			}

			msg := openai.AssistantMessage(full.String())
			if len(tcs) > 0 {
				msg.OfAssistant.ToolCalls = make([]openai.ChatCompletionMessageToolCallUnionParam, len(tcs))
				for i, tc := range tcs {
					msg.OfAssistant.ToolCalls[i] = openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID:       tc.ID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{Name: tc.Name, Arguments: tc.Arguments},
						},
					}
				}
			}
			messages = append(messages, msg)

			if len(tcs) == 0 || c.toolExecutor == nil {
				break
			}

			for _, tc := range tcs {
				res, err := c.toolExecutor(ctx, tc.Name, tc.Arguments)
				if err != nil {
					res = "Error: " + err.Error()
					c.logger.Warnw("tool err",
						"tool", tc.Name,
						err)
				}
				messages = append(messages, openai.ToolMessage(res, tc.ID))
			}
		}

		c.diffed = messages[baseLen:]
	}()

	return out
}

// GetErr 获取 TextClient 最后一次操作产生的错误
func (c *TextClient) GetErr() error {
	return c.err
}

// GetDiffed 获取当前对话轮次新增的消息参数列表（不含历史记录）
func (c *TextClient) GetDiffed() []openai.ChatCompletionMessageParamUnion {
	return c.diffed
}
