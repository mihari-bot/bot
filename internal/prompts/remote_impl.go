package prompts

import (
	"github.com/mihari-bot/bot/internal/container"

	"go.uber.org/zap"
)

type RemoteImpl struct {
	storeToWhere string
	logger       *zap.Logger
	url          string
}

func (r *RemoteImpl) Load(mp *container.Map[string, string]) error {
	r.logger.Info("开始从远程拉取")
	return nil
}

var _ Provider = &RemoteImpl{}
