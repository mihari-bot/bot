// Package voice 提供 Voice API 的客户端封装.
package voice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client 是 Voice API 客户端.
type Client struct {
	http          *resty.Client
	authorization string
}

// Option 配置 Client 的初始化选项。
type Option func(*Client)

// WithAuthorization 设置 Authorization Header 的完整值。
// 例如: "Bearer <your_api_key>" 或直接 "<your_api_key>"。
func WithAuthorization(v string) Option {
	return func(c *Client) { c.authorization = strings.TrimSpace(v) }
}

// WithBearerToken 以 Bearer 形式设置 Authorization Header。
func WithBearerToken(token string) Option {
	return func(c *Client) { c.authorization = "Bearer " + strings.TrimSpace(token) }
}

// WithRestyClient 注入自定义的 resty.Client（用于自定义 Transport、代理、重试等）。
func WithRestyClient(rc *resty.Client) Option {
	return func(c *Client) { c.http = rc }
}

// New 创建 Voice API Client。
func New(baseURL string, opts ...Option) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("empty base url")
	}

	c := &Client{
		http: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(time.Minute * 10),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// GenerateRequest 对应 POST /generate 的请求体。
type GenerateRequest struct {
	Text string `json:"text"`
	Role string `json:"role"`
}

type apiError struct {
	method     string
	path       string
	statusCode int
	body       string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%s %s: status=%d body=%q", e.method, e.path, e.statusCode, e.body)
}

func (c *Client) applyAuth(r *resty.Request) {
	if c.authorization == "" {
		return
	}
	r.SetHeader("Authorization", c.authorization)
}

// Roles 请求 GET /roles 并返回角色列表。
func (c *Client) Roles(ctx context.Context) ([]string, error) {
	path := "/roles"

	var out []string
	req := c.http.R().SetContext(ctx).SetResult(&out)
	c.applyAuth(req)

	resp, err := req.Get(path)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, &apiError{method: "GET", path: path, statusCode: resp.StatusCode(), body: string(resp.Body())}
	}
	return out, nil
}

// Generate 请求 POST /generate 并返回音频二进制数据与 Content-Type。
func (c *Client) Generate(ctx context.Context, in GenerateRequest) (data []byte, contentType string, err error) {
	path := "/generate"

	req := c.http.R().
		SetContext(ctx).
		SetBody(in)
	c.applyAuth(req)

	resp, err := req.Post(path)
	if err != nil {
		return nil, "", err
	}
	if resp.IsError() {
		return nil, "", &apiError{method: "POST", path: path, statusCode: resp.StatusCode(), body: string(resp.Body())}
	}

	return resp.Body(), resp.Header().Get("Content-Type"), nil
}
