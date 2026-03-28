# Mihari Bot

一只偏执地把对话当作剧情来写的私聊机器人。它通过 OneBot v11 的 WebSocket 接入消息流，用你提供的大模型与 TTS 服务，把文字与语音都做得更像“角色在说话”。

## 特性

- OneBot v11 WebSocket 接入：监听私聊消息并回复
- 多角色对话：启动时自动拉取远程角色卡（YAML），按角色注入用户名等变量
- 每人独立配置：BaseURL、Key、模型、角色、语音开关等存入本地数据库
- 语音模式：输出走 TTS，并用内置 HTTP 服务临时托管音频给 OneBot 客户端拉取
- 对话记忆（轮次）：按用户与角色维度持久化最近 N 轮消息

## 快速开始

### 1. 准备运行目录

程序启动后会在当前工作目录下使用 `./mihari` 作为数据根目录，至少需要放置一个 `config.json`：

```bash
mkdir -p mihari
$EDITOR mihari/config.json
```

### 2. 编写 `mihari/config.json`

```json
{
  "ws": {
    "url": "ws://127.0.0.1:3001",
    "accessToken": "your-onebot-access-token"
  },
  "db": {
    "provider": "sqlite",
    "dsn": "db.sqlite"
  },
  "chat": {
    "perCharDelay": { "min": "20ms", "max": "60ms" },
    "waitTime": { "min": "300ms", "max": "900ms" }
  },
  "voiceHttp": {
    "addr": "0.0.0.0:18080",
    "prefix": "http://127.0.0.1:18080"
  }
}
```

- `ws.url`：OneBot v11 的 WS 地址（机器人以客户端方式连接）
- `ws.accessToken`：对应 OneBot 端的 `access_token`
- `db.provider`：目前只支持 `sqlite`
- `db.dsn`：SQLite 文件路径（相对路径会自动落在 `./mihari` 目录下）
- `chat.*`：打字与分段发送的节奏控制
- `voiceHttp.*`：语音模式下用于临时托管音频的 HTTP 服务

`voiceHttp.prefix` 必须能被你的 OneBot 端访问到；如果机器人运行在内网容器里，请把这里配置成可被 OneBot 宿主机访问的地址。

### 3. 启动

```bash
go run ./cmd/mihari-bot
```

启动时会自动拉取角色卡仓库并加载本地数据库。

## 初次对话的“入场台词”

Mihari 的大模型与语音配置按用户存储，需要你在私聊里给它下达指令完成初始化。

### 必备指令

- `/set_base_url <url>`：设置大模型 API BaseURL
- `/set_key <key>`：设置大模型 API Key
- `/set_model <model>`：设置模型名
- `/roles`：查看可用角色列表（来自远程角色卡）
- `/set_role <role>`：选择角色

### 语音相关指令

- `/set_voice_base_url <url>`：设置 TTS BaseURL
- `/set_voice_key <key>`：设置 TTS Authorization/Key
- `/voice_roles`：从 TTS 服务端拉取可用音色列表
- `/set_voice_role <role>`：设置音色
- `/voice <on|off>`：开启或关闭语音模式

## 角色卡

启动时会将角色卡仓库拉取到 `./mihari/chardefs_remote`，并从其中的 `chardefs/*.yaml` 加载角色。

- 文件名（不含扩展名）即角色名
- 角色卡内容为 YAML；机器人会进行一次变量注入（如 `{{USERNAME}}`）并重新序列化

## 开发

```bash
go test ./...
```

```bash
go build ./cmd/mihari-bot
```

## License

见 [LICENSE](./LICENSE)。
