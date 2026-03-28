// Package bot 是Mihari-Bot的实现
package bot

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mihari-bot/bot/internal/container"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Bot 是机器人实例
type Bot struct {
	logger              *zap.SugaredLogger
	db                  *gorm.DB
	callMp              callMp
	cmdMp               *orderedmap.OrderedMap[string, cmdDef]
	echoMp              *container.Map[int64, chan *gjson.Result]
	userLockMp          *container.Map[int64, sync.Mutex]
	chardefMp           *container.Map[string, string]
	userHelpMessageIDMp *container.Map[int64, int64]
	echoCounter         atomic.Int64
	voiceTmpMp          *container.Map[int64, voiceTempItem]
	voiceHTTP           *voiceHTTPServer
	profile             *gjson.Result
	runtimeConfig       struct {
		perCharDelayMin, perCharDelayMax time.Duration
		minWaitTime, maxWaitTime         time.Duration
	}
	baseDir string
}

// New 创建并返回一个Bot
func New(logger *zap.SugaredLogger, baseDir string) *Bot {
	return &Bot{
		logger:              logger,
		echoMp:              container.NewMap[int64, chan *gjson.Result](),
		cmdMp:               orderedmap.NewOrderedMap[string, cmdDef](),
		userLockMp:          container.NewMap[int64, sync.Mutex](),
		chardefMp:           container.NewMap[string, string](),
		userHelpMessageIDMp: container.NewMap[int64, int64](),
		voiceTmpMp:          container.NewMap[int64, voiceTempItem](),
		baseDir:             baseDir,
	}
}

// Init 负责初始化
func (b *Bot) Init(ctx context.Context) error {
	err := b.profileInit()
	if err != nil {
		return err
	}

	err = b.voiceHTTPInit(ctx)
	if err != nil {
		return err
	}

	err = b.dbInit()
	if err != nil {
		return err
	}

	err = b.chardefPullFromRemote(ctx)
	if err != nil {
		return err
	}

	b.callmapInit()
	b.cmdInit()

	chardefFolder, err := b.pthResolve("chardefs_remote/chardefs")
	if err != nil {
		return err
	}
	err = b.ChardefLoadFromFolder(chardefFolder)
	if err != nil {
		return err
	}

	return nil
}

// Start 将会启动配置的机器人, 并阻塞
func (b *Bot) Start(ctx context.Context) error {
	wsURL, accessToken, err := b.profileGetWsConfig()
	if err != nil {
		return err
	}

	return b.openWS(ctx, wsURL, accessToken)
}

type j = map[string]any
