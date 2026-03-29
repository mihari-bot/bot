package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/glebarez/sqlite"
	"github.com/mihari-bot/bot/internal/model"

	"github.com/openai/openai-go/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func (b *Bot) dbLoadUserConfig(ctx context.Context, who int64) (*model.Config, error) {
	config, err := gorm.G[*model.Config](b.db).Where("user_id = ?", who).First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		config = &model.Config{UserID: who, Rounds: 16, VoiceEnabled: false}
		err = b.db.WithContext(ctx).Save(config).Error
		if err != nil {
			return nil, err
		}

		return config, nil
	}
	if err != nil {
		return nil, err
	}

	return config, err
}

func (b *Bot) dbLoadCcmpu(ctx context.Context, who int64, role string, roundsCount int) ([]openai.ChatCompletionMessageParamUnion, error) {
	res := []openai.ChatCompletionMessageParamUnion{}

	rounds, err := gorm.G[model.Round](b.db).Order("created_at DESC").Limit(roundsCount).Where("user_id = ? AND role = ?", who, role).Find(ctx)
	if err != nil {
		return nil, err
	}

	// Reverse
	slices.Reverse(rounds)

	for _, round := range rounds {
		tmp := []openai.ChatCompletionMessageParamUnion{}
		err = json.Unmarshal([]byte(round.DataInJSON), &tmp)
		if err != nil {
			return nil, err
		}

		res = append(res, tmp...)
	}

	return res, err
}

func (b *Bot) dbAppendCcmpu(ctx context.Context, who int64, role string, what []openai.ChatCompletionMessageParamUnion) error {
	inJSON, err := json.Marshal(what)
	if err != nil {
		return err
	}

	err = b.db.WithContext(ctx).Save(&model.Round{Role: role, UserID: who, DataInJSON: string(inJSON)}).Error
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) dbInit() error {
	provider, dsn, err := b.profileGetDbConfig()
	if err != nil {
		return err
	}

	var dialector gorm.Dialector
	switch provider {
	case "sqlite":
		dsn, err = b.pthResolve(dsn)
		if err != nil {
			return err
		}
		dialector = sqlite.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	default:
		return fmt.Errorf("unsupported database provider: \"%s\"", provider)
	}

	db, err := gorm.Open(dialector)
	if err != nil {
		return err
	}

	b.db = db

	return b.db.AutoMigrate(&model.Config{}, &model.Round{})
}
