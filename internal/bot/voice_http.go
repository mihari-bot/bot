package bot

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mihari-bot/bot/internal/container"
	"go.uber.org/zap"
)

type voiceTempItem struct {
	data        []byte
	contentType string
	expireAt    time.Time
}

type voiceHTTPServer struct {
	logger       *zap.SugaredLogger
	store        *container.Map[int64, voiceTempItem]
	addr         string
	publicPrefix string
	app          *fiber.App
	started      atomic.Bool
}

func newVoiceHTTPServer(logger *zap.SugaredLogger, store *container.Map[int64, voiceTempItem], addr string, publicPrefix string) *voiceHTTPServer {
	return &voiceHTTPServer{
		logger:       logger,
		store:        store,
		addr:         addr,
		publicPrefix: publicPrefix,
	}
}

func (s *voiceHTTPServer) urlForID(id int64) string {
	return strings.TrimRight(s.publicPrefix, "/") + "/voice/" + strconv.FormatInt(id, 10)
}

func (s *voiceHTTPServer) Start(ctx context.Context) error {
	if s.started.Swap(true) {
		return nil
	}

	s.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	s.app.Get("/voice/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		item, ok := s.store.Get(id)
		if !ok {
			return c.SendStatus(fiber.StatusNotFound)
		}

		if time.Now().After(item.expireAt) {
			s.store.Delete(id)
			return c.SendStatus(fiber.StatusNotFound)
		}

		s.store.Delete(id)
		if item.contentType != "" {
			c.Set(fiber.HeaderContentType, item.contentType)
		}
		return c.Send(item.data)
	})

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	go func() {
		err := s.app.Listener(ln)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			s.logger.Errorw("voice http server stopped",
				"err", err)
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.app.ShutdownWithContext(shutdownCtx)
	}()

	go s.gcLoop(ctx)

	return nil
}

func (s *voiceHTTPServer) gcLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			for _, id := range s.store.Keys() {
				item, ok := s.store.Get(id)
				if !ok {
					continue
				}
				if now.After(item.expireAt) {
					s.store.Delete(id)
				}
			}
		}
	}
}

func (b *Bot) voiceHTTPInit(ctx context.Context) error {
	if b.voiceHTTP != nil {
		return nil
	}

	addr := b.profileGetStringOr("voiceHttp.addr", "0.0.0.0:18080")
	publicPrefix := b.profileGetStringOr("voiceHttp.prefix", "http://127.0.0.1:18080")
	if strings.TrimSpace(publicPrefix) == "" {
		publicPrefix = "http://127.0.0.1:18080"
	}

	b.voiceHTTP = newVoiceHTTPServer(b.logger.Named("VoiceHTTP"), b.voiceTmpMp, addr, publicPrefix)
	return b.voiceHTTP.Start(ctx)
}

func (b *Bot) voiceTempPut(audio []byte, contentType string) int64 {
	id := b.echoCounter.Add(1)
	audioCopy := make([]byte, len(audio))
	copy(audioCopy, audio)

	b.voiceTmpMp.Set(id, voiceTempItem{
		data:        audioCopy,
		contentType: contentType,
		expireAt:    time.Now().Add(5 * time.Minute),
	})

	return id
}

func (b *Bot) voiceTempURL(id int64) string {
	return b.voiceHTTP.urlForID(id)
}
