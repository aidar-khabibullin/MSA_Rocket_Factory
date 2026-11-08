package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/config"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

type client struct {
	bot                      *bot.Bot
	chatID                   int64
	userRegistrationCallback func(ctx context.Context, username string, chatID int64) error
}

type noopClient struct {
	chatID                   int64
	userRegistrationCallback func(ctx context.Context, username string, chatID int64) error
	// Добавляем поле для хранения последнего сообщения /start
	lastStartMessage string
}

func (c *client) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if c.userRegistrationCallback != nil {
		username := update.Message.From.Username
		if username == "" {
			username = fmt.Sprintf("user_%d", update.Message.From.ID)
		}

		err := c.userRegistrationCallback(ctx, username, update.Message.Chat.ID)
		if err != nil {
			logger.Error(ctx, "Ошибка при регистрации пользователя из Telegram",
				zap.Error(err),
				zap.String("username", username),
				zap.Int64("chat_id", update.Message.Chat.ID))
		}
	}

	message := "🚀 Добро пожаловать в Rocket Factory!\n\n" +
		"Этот бот уведомляет о важных событиях в системе сборки ракет.\n" +
		"Вы будете получать уведомления о:\n" +
		"• Оплаченных заказах\n" +
		"• Собранных кораблях\n" +
		"• Других важных событиях\n\n" +
		"Бот работает автоматически и не требует дополнительных команд."

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
	if err != nil {
		logger.Error(ctx, "Failed to send start command response",
			zap.Error(err),
			zap.Int64("chat_id", update.Message.Chat.ID))
	}
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Игнорируем все сообщения, кроме команды /start
}

func NewClient(ctx context.Context, cfg config.TelegramConfig) (TelegramClient, error) {
	if cfg.SkipAPICheck() {
		logger.Info(ctx, "Telegram API check skipped - using noop client for development")
		return &noopClient{}, nil
	}

	token := cfg.BotToken()
	if token == "" {
		return nil, fmt.Errorf("telegram bot token is empty")
	}

	token = strings.TrimSpace(token)

	if !strings.Contains(token, ":") {
		return nil, fmt.Errorf("invalid telegram bot token format: must contain ':'")
	}

	b, err := bot.New(token, bot.WithDefaultHandler(defaultHandler))
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	reportBot := &client{
		bot: b,
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, reportBot.startHandler)

	me, err := b.GetMe(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to telegram API: %w", err)
	}

	logger.Info(ctx, "Telegram bot connected successfully",
		zap.String("bot_username", me.Username),
		zap.Int64("bot_id", me.ID))

	return reportBot, nil
}

func (c *client) Start(ctx context.Context) error {
	go func() {
		c.bot.Start(ctx)
	}()

	return nil
}

func (c *client) SendMessage(ctx context.Context, chatID int64, message string) error {
	targetChatID := chatID
	if targetChatID == 0 {
		targetChatID = c.chatID
	}

	_, err := c.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: targetChatID,
		Text:   message,
	})
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}

func (c *client) Close(ctx context.Context) error {
	if c.bot != nil {
		_, err := c.bot.Close(ctx)
		if err != nil {
			logger.Error(ctx, "Failed to close telegram bot", zap.Error(err))
		}
	}
	return nil
}

func (c *client) SetUserRegistrationCallback(callback func(ctx context.Context, username string, chatID int64) error) {
	c.userRegistrationCallback = callback
}

// HandleStartCommand обрабатывает команду /start (для совместимости с интерфейсом)
func (c *client) HandleStartCommand(ctx context.Context, username string, chatID int64) error {
	// В реальном клиенте команда /start обрабатывается через startHandler
	// Этот метод добавлен для совместимости с интерфейсом
	return nil
}

// noopClient методы

func (n *noopClient) SetUserRegistrationCallback(callback func(ctx context.Context, username string, chatID int64) error) {
	n.userRegistrationCallback = callback
}

func (n *noopClient) Start(ctx context.Context) error {
	logger.Info(ctx, "Development mode: Telegram bot started (simulated)")

	// В режиме разработки запускаем бесконечный цикл для имитации работы бота
	go func() {
		<-ctx.Done()
		logger.Info(ctx, "Development mode: Telegram bot stopped")
	}()

	return nil
}

func (n *noopClient) SendMessage(ctx context.Context, chatID int64, message string) error {
	targetChatID := chatID
	if targetChatID == 0 {
		targetChatID = n.chatID
	}

	logger.Info(ctx, "Development mode: Telegram message simulated",
		zap.Int64("chat_id", targetChatID),
		zap.String("message", message))
	return nil
}

func (n *noopClient) Close(ctx context.Context) error {
	logger.Info(ctx, "Development mode: Telegram client close skipped")
	return nil
}

// HandleStartCommand обрабатывает команду /start в режиме разработки
func (n *noopClient) HandleStartCommand(ctx context.Context, username string, chatID int64) error {
	logger.Info(ctx, "Development mode: Обработка команды /start",
		zap.String("username", username),
		zap.Int64("chat_id", chatID))

	// Если есть callback для регистрации, вызываем его
	if n.userRegistrationCallback != nil {
		err := n.userRegistrationCallback(ctx, username, chatID)
		if err != nil {
			logger.Error(ctx, "Development mode: Ошибка при регистрации пользователя",
				zap.Error(err),
				zap.String("username", username),
				zap.Int64("chat_id", chatID))
			return err
		}
		logger.Info(ctx, "Development mode: Пользователь успешно зарегистрирован",
			zap.String("username", username),
			zap.Int64("chat_id", chatID))
	} else {
		logger.Warn(ctx, "Development mode: Callback для регистрации не установлен")
	}

	// Сохраняем сообщение для отладки
	n.lastStartMessage = fmt.Sprintf("🚀 Добро пожаловать в Rocket Factory! (Development mode)\n\nПользователь: %s\nChat ID: %d", username, chatID)

	return nil
}
