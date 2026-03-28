package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chhongzh/shlex"
)

type cmdUsage struct{ format, mean string }
type cmdUsages []cmdUsage
type cmdFn func(ctx context.Context, cc *cctx, args []string) error
type cmdDef struct {
	name   string
	short  string
	usages []cmdUsage
	fn     cmdFn
}

func (b *Bot) cmdHelp(ctx context.Context, cc *cctx, _ []string) error {
	if replyID, hint := b.userHelpMessageIDMp.Get(cc.getUserID()); hint {
		cc.getLogger().Debug("Hint help cache.")
		_, err := cc.sendMessage(ctx, fmt.Sprintf("[CQ:reply,id=%d].", replyID))
		if err != nil {
			return err
		}
		return nil
	}

	msgs := []string{}

	for _, cmd := range b.cmdMp.AllFromFront() {
		for _, usage := range cmd.usages {
			line := ""
			if usage.format != "" {
				line = fmt.Sprintf("\"/%s %s\" %s", cmd.name, usage.format, usage.mean)
			} else {
				line = fmt.Sprintf("\"/%s\" %s", cmd.name, usage.mean)
			}

			msgs = append(msgs, line)
		}
	}

	helpMessageID, err := cc.sendMessage(ctx, strings.Join(msgs, "\n"))
	if err != nil {
		return err
	}
	b.userHelpMessageIDMp.Set(cc.getUserID(), helpMessageID)

	return nil
}
func (b *Bot) cmdSetBaseURL(ctx context.Context, cc *cctx, args []string) error {
	if len(args) != 1 {
		return ErrInvalidArgs
	}

	baseURL := args[0]

	config, err := b.dbLoadUserConfig(ctx, cc.getUserID())
	if err != nil {
		return err
	}
	config.APIBaseURL = baseURL
	err = b.db.WithContext(ctx).Save(config).Error
	if err != nil {
		return err
	}

	_, err = cc.sendMessage(ctx, "好哒！已经更新成功啦~")
	if err != nil {
		return err
	}
	return nil
}
func (b *Bot) cmdSetKey(ctx context.Context, cc *cctx, args []string) error {
	if len(args) != 1 {
		return ErrInvalidArgs
	}

	config, err := b.dbLoadUserConfig(ctx, cc.getUserID())
	if err != nil {
		return err
	}
	config.APIKey = args[0]
	err = b.db.WithContext(ctx).Save(config).Error
	if err != nil {
		return err
	}

	_, err = cc.sendMessage(ctx, "好哒！已经更新成功啦~")
	if err != nil {
		return err
	}
	return nil
}
func (b *Bot) cmdSetModel(ctx context.Context, cc *cctx, args []string) error {
	if len(args) != 1 {
		return ErrInvalidArgs
	}

	config, err := b.dbLoadUserConfig(ctx, cc.getUserID())
	if err != nil {
		return err
	}
	config.APIModel = args[0]
	err = b.db.WithContext(ctx).Save(config).Error
	if err != nil {
		return err
	}

	_, err = cc.sendMessage(ctx, "好哒！已经更新成功啦~")
	if err != nil {
		return err
	}
	return nil
}
func (b *Bot) cmdSetRole(ctx context.Context, cc *cctx, args []string) error {
	if len(args) != 1 {
		return ErrInvalidArgs
	}

	role := args[0]
	if !b.chardefMp.Exists(role) {
		_, err := cc.sendMessage(ctx, "呜哇，找不到这个角色呢... 检查一下名字有没有写错呀？")
		if err != nil {
			return err
		}
		return nil
	}

	config, err := b.dbLoadUserConfig(ctx, cc.getUserID())
	if err != nil {
		return err
	}
	config.Role = role
	err = b.db.WithContext(ctx).Save(config).Error
	if err != nil {
		return err
	}

	_, err = cc.sendMessage(ctx, "好哒！已经更新成功啦~")
	if err != nil {
		return err
	}
	return nil
}
func (b *Bot) cmdRoles(ctx context.Context, cc *cctx, args []string) error {
	if len(args) != 0 {
		return ErrInvalidArgs
	}

	roles := b.chardefMp.Keys()
	buf := []string{}

	buf = append(buf, "可用角色列表如下")
	buf = append(buf, "")
	for _, role := range roles {
		buf = append(buf, role)
	}

	_, err := cc.sendMessage(ctx, strings.Join(buf, "\n"))
	if err != nil {
		return err
	}
	return nil
}
func (b *Bot) cmdSetVoiceBaseURL(ctx context.Context, cc *cctx, args []string) error {
	if len(args) != 1 {
		return ErrInvalidArgs
	}

	baseURL := args[0]

	config, err := b.dbLoadUserConfig(ctx, cc.getUserID())
	if err != nil {
		return err
	}
	config.VoiceBaseURL = baseURL
	err = b.db.WithContext(ctx).Save(config).Error
	if err != nil {
		return err
	}

	_, err = cc.sendMessage(ctx, "好哒！已经更新成功啦~")
	if err != nil {
		return err
	}
	return nil
}
func (b *Bot) cmdSetVoiceKey(ctx context.Context, cc *cctx, args []string) error {
	if len(args) != 1 {
		return ErrInvalidArgs
	}

	config, err := b.dbLoadUserConfig(ctx, cc.getUserID())
	if err != nil {
		return err
	}
	config.VoiceAuthorization = args[0]
	err = b.db.WithContext(ctx).Save(config).Error
	if err != nil {
		return err
	}

	_, err = cc.sendMessage(ctx, "好哒！已经更新成功啦~")
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) cmdVoiceRoles(ctx context.Context, cc *cctx, args []string) error {
	if len(args) != 0 {
		return ErrInvalidArgs
	}

	config, err := b.dbLoadUserConfig(ctx, cc.getUserID())
	if err != nil {
		return err
	}

	vc, err := b.mustVoiceClient(ctx, config)
	if err != nil {
		return err
	}
	roles, err := vc.Roles(ctx)
	if err != nil {
		return err
	}
	buf := []string{}
	buf = append(buf, "Voice列表")
	buf = append(buf, "")
	for _, r := range roles {
		buf = append(buf, r)
	}
	_, err = cc.sendMessage(ctx, strings.Join(buf, "\n"))
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) cmdSetVoiceRole(ctx context.Context, cc *cctx, args []string) error {
	if len(args) != 1 {
		return ErrInvalidArgs
	}

	role := strings.TrimSpace(args[0])
	if role == "" {
		return ErrInvalidArgs
	}

	config, err := b.dbLoadUserConfig(ctx, cc.getUserID())
	if err != nil {
		return err
	}

	vc, err := b.mustVoiceClient(ctx, config)
	if err != nil {
		return err
	}
	roles, err := vc.Roles(ctx)
	if err != nil {
		return err
	}
	found := false
	for _, r := range roles {
		if r == role {
			found = true
			break
		}
	}
	if !found {
		_, err2 := cc.sendMessage(ctx, "呜哇，找不到这个Voice角色呢… 可以先用 /voice_roles 看看有哪些~")
		if err2 != nil {
			return err2
		}
		return nil
	}

	config.VoiceRole = role
	err = b.db.WithContext(ctx).Save(config).Error
	if err != nil {
		return err
	}

	_, err = cc.sendMessage(ctx, "好哒！已经更新成功啦~")
	if err != nil {
		return err
	}
	return nil
}
func (b *Bot) cmdVoice(ctx context.Context, cc *cctx, args []string) error {
	if len(args) != 1 {
		return ErrInvalidArgs
	}

	v := strings.ToLower(strings.TrimSpace(args[0]))
	var enabled bool
	switch v {
	case "on", "1", "true", "enable", "enabled":
		enabled = true
	case "off", "0", "false", "disable", "disabled":
		enabled = false
	default:
		return ErrInvalidArgs
	}

	config, err := b.dbLoadUserConfig(ctx, cc.getUserID())
	if err != nil {
		return err
	}

	if enabled {
		if strings.TrimSpace(config.VoiceRole) == "" {
			_, err2 := cc.sendMessage(ctx, "要先设置Voice角色才可以开启语音模式哟~（可以先用 /voice_roles 看列表）")
			if err2 != nil {
				return err2
			}
			return nil
		}

		vc, err := b.mustVoiceClient(ctx, config)
		if err != nil {
			return err
		}
		roles, err := vc.Roles(ctx)
		if err != nil {
			return err
		}
		found := false
		for _, r := range roles {
			if r == config.VoiceRole {
				found = true
				break
			}
		}
		if !found {
			_, err2 := cc.sendMessage(ctx, "你设置的Voice角色在服务端找不到呢… 可以先用 /voice_roles 重新选一个~")
			if err2 != nil {
				return err2
			}
			return nil
		}
	}

	config.VoiceEnabled = enabled
	err = b.db.WithContext(ctx).Save(config).Error
	if err != nil {
		return err
	}

	if enabled {
		_, err = cc.sendMessage(ctx, "好耶！语音模式已开启~")
	} else {
		_, err = cc.sendMessage(ctx, "好哒，语音模式已关闭~")
	}
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) cmdInit() {
	b.cmdRegister(
		"help",
		"显示帮助菜单",
		cmdUsages{
			{"", "列出所有的帮助"},
		},
		b.cmdHelp,
	)
	b.cmdRegister(
		"set_base_url",
		"设置Api调用的BaseURL",
		cmdUsages{
			{"<url>", "BaseURL由您的大模型提供商提供"},
		},
		b.cmdSetBaseURL,
	)
	b.cmdRegister(
		"set_key",
		"设置Api调用的Key",
		cmdUsages{
			{"<key>", "Key也由您的大模型提供商提供"},
		},
		b.cmdSetKey,
	)
	b.cmdRegister(
		"set_model",
		"设置Api调用所使用的模型",
		cmdUsages{
			{"<model>", "可用模型由您的大模型提供商提供"},
		},
		b.cmdSetModel,
	)
	b.cmdRegister(
		"roles",
		"获取可用角色列表",
		cmdUsages{
			{"", "列出所有可用角色列表"},
		},
		b.cmdRoles,
	)
	b.cmdRegister(
		"set_role",
		"设置角色",
		cmdUsages{
			{"<role>", "角色名"},
		},
		b.cmdSetRole,
	)
	b.cmdRegister(
		"set_voice_base_url",
		"设置Voice Api调用的BaseURL",
		cmdUsages{
			{"<url>", "Voice BaseURL由您的TTS提供商提供"},
		},
		b.cmdSetVoiceBaseURL,
	)
	b.cmdRegister(
		"set_voice_key",
		"设置Voice Api调用的Key/Authorization",
		cmdUsages{
			{"<key>", "你的key"},
		},
		b.cmdSetVoiceKey,
	)
	b.cmdRegister(
		"voice_roles",
		"获取Voice服务端的角色列表",
		cmdUsages{
			{"", "从Voice API获取可用角色列表"},
		},
		b.cmdVoiceRoles,
	)
	b.cmdRegister(
		"set_voice_role",
		"设置Voice角色",
		cmdUsages{
			{"<role>", "role来自 /voice_roles 返回的列表"},
		},
		b.cmdSetVoiceRole,
	)
	b.cmdRegister(
		"voice",
		"开启或关闭语音模式",
		cmdUsages{
			{"<on|off>", "on=开启, off=关闭"},
		},
		b.cmdVoice,
	)
}

func (b *Bot) cmdRegister(name string, short string, usage cmdUsages, fn cmdFn) {
	b.cmdMp.Set(name, cmdDef{name, short, usage, fn})
}

func (b *Bot) cmd(ctx context.Context, cc *cctx) error {
	// Random delay
	delay := b.mthRandN(time.Millisecond*500, time.Second*3)
	time.Sleep(delay)

	cmds, err := shlex.Split(cc.getMessage()[1:])
	if err != nil {
		return err
	}
	if len(cmds) == 0 {
		_, err = cc.sendMessage(ctx, "呜哇，这个指令我不认识呢... 是不是打错字了？")
		if err != nil {
			return err
		}
		return nil
	}

	name, args := cmds[0], cmds[1:]
	def, ok := b.cmdMp.Get(name)
	if !ok {
		_, err = cc.sendMessage(ctx, "呜哇，没听懂你在说什么呢，可以再试一次吗？")
		if err != nil {
			return err
		}
		return nil
	}

	err = def.fn(ctx, cc, args)
	if errors.Is(err, ErrInvalidArgs) {
		_, err = cc.sendMessage(ctx, "呜哇，没听懂你在说什么呢，可以再试一次吗？")
		if err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}
