package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"payment_bot/internal/usecase"
)

type stateType int

const (
	stateIdle        stateType = iota
	stateAwaitSalary           // bot asked user to enter salary
)

type Handler struct {
	calc      *usecase.SalaryCalculator
	users     usecase.UserRepository
	formatter *Formatter
	mu        sync.Mutex
	states    map[int64]stateType
}

func NewHandler(calc *usecase.SalaryCalculator, users usecase.UserRepository) *Handler {
	return &Handler{
		calc:      calc,
		users:     users,
		formatter: NewFormatter(),
		states:    make(map[int64]stateType),
	}
}

func (h *Handler) setState(userID int64, s stateType) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.states[userID] = s
}

func (h *Handler) getState(userID int64) stateType {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.states[userID]
}

func (h *Handler) HandleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	isPrivate := msg.Chat.IsPrivate()
	ctx := context.Background()

	switch msg.Command() {
	case "start":
		h.handleStart(ctx, bot, msg, isPrivate)
	case "calc":
		h.handleCalc(ctx, bot, msg, isPrivate)
	case "salary":
		if !isPrivate {
			return
		}
		h.handleSalaryCommand(ctx, bot, msg)
	case "help":
		h.handleHelp(bot, msg, isPrivate)
	case "":
		if isPrivate {
			h.handlePrivateText(ctx, bot, msg)
		}
	}
}

func (h *Handler) HandleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
	ctx := context.Background()

	bot.Request(tgbotapi.NewCallback(cb.ID, "")) //nolint:errcheck

	switch {
	case strings.HasPrefix(cb.Data, "update:"):
		salaryStr := strings.TrimPrefix(cb.Data, "update:")
		salary, err := strconv.ParseFloat(salaryStr, 64)
		if err != nil {
			return
		}
		if err := h.users.SetSalary(ctx, cb.From.ID, salary); err != nil {
			sendEditText(bot, cb.Message.Chat.ID, cb.Message.MessageID, "Ошибка сохранения. Попробуйте позже.")
			return
		}
		sendEditText(bot, cb.Message.Chat.ID, cb.Message.MessageID,
			fmt.Sprintf("Зарплата обновлена: %s", formatMoney(salary)))
		h.sendCalcPrivate(ctx, bot, cb.Message.Chat.ID, cb.From.ID)

	case cb.Data == "cancel":
		sendEditText(bot, cb.Message.Chat.ID, cb.Message.MessageID, "Зарплата не изменена.")
		h.sendCalcPrivate(ctx, bot, cb.Message.Chat.ID, cb.From.ID)
	}
}

func (h *Handler) handleStart(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message, isPrivate bool) {
	if !isPrivate {
		sendText(bot, msg.Chat.ID, "Привет! Используйте /calc для расчёта выплат за текущий месяц.")
		return
	}

	_, hasSalary, err := h.users.GetSalary(ctx, msg.From.ID)
	if err != nil {
		sendText(bot, msg.Chat.ID, "Ошибка. Попробуйте позже.")
		return
	}

	if !hasSalary {
		h.setState(msg.From.ID, stateAwaitSalary)
		sendText(bot, msg.Chat.ID,
			"Привет! Я помогу рассчитать аванс и остаток зарплаты.\n\nВведите вашу ежемесячную зарплату (например: 100000):")
	} else {
		sendText(bot, msg.Chat.ID,
			"Привет! Используйте /calc для расчёта выплат или /salary для управления зарплатой.")
	}
}

func (h *Handler) handleCalc(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message, isPrivate bool) {
	now := time.Now()
	year, month := now.Year(), now.Month()

	if arg := msg.CommandArguments(); arg != "" {
		var err error
		year, month, err = parsePeriod(arg, now)
		if err != nil {
			sendText(bot, msg.Chat.ID, "Некорректный месяц. Примеры: /calc 3, /calc март, /calc март 2025, /calc 03.2025")
			return
		}
	}

	if isPrivate {
		h.sendCalcPrivateFor(ctx, bot, msg.Chat.ID, msg.From.ID, year, month)
	} else {
		h.sendCalcGroupFor(ctx, bot, msg.Chat.ID, year, month)
	}
}

func (h *Handler) handleSalaryCommand(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	arg := msg.CommandArguments()

	if arg == "" {
		salary, hasSalary, err := h.users.GetSalary(ctx, msg.From.ID)
		if err != nil {
			sendText(bot, msg.Chat.ID, "Ошибка. Попробуйте позже.")
			return
		}
		if !hasSalary {
			h.setState(msg.From.ID, stateAwaitSalary)
			sendText(bot, msg.Chat.ID, "Зарплата не указана. Введите сумму:")
		} else {
			sendText(bot, msg.Chat.ID, fmt.Sprintf(
				"Ваша зарплата: %s\n\nДля изменения: /salary <сумма>", formatMoney(salary),
			))
		}
		return
	}

	salary, err := parseSalary(arg)
	if err != nil {
		sendText(bot, msg.Chat.ID, "Некорректная сумма. Пример: /salary 100000")
		return
	}

	if err := h.users.SetSalary(ctx, msg.From.ID, salary); err != nil {
		sendText(bot, msg.Chat.ID, "Ошибка сохранения. Попробуйте позже.")
		return
	}

	sendText(bot, msg.Chat.ID, fmt.Sprintf("Зарплата сохранена: %s", formatMoney(salary)))
	h.sendCalcPrivate(ctx, bot, msg.Chat.ID, msg.From.ID)
}

func (h *Handler) handleHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, isPrivate bool) {
	var text string
	if isPrivate {
		text = "Команды:\n" +
			"/calc — расчёт выплат за текущий месяц\n" +
			"/salary — показать текущую зарплату\n" +
			"/salary <сумма> — установить или изменить зарплату"
	} else {
		text = "Команды:\n" +
			"/calc — расчёт долей аванса и остатка за текущий месяц\n" +
			"/calc <месяц> [год] — расчёт за другой месяц (например: /calc 3, /calc март 2025, /calc 03.2025)"
	}
	sendText(bot, msg.Chat.ID, text)
}

func (h *Handler) handlePrivateText(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	salary, err := parseSalary(msg.Text)
	if err != nil {
		if h.getState(msg.From.ID) == stateAwaitSalary {
			sendText(bot, msg.Chat.ID, "Введите корректную сумму (например: 100000):")
		}
		return
	}

	if h.getState(msg.From.ID) == stateAwaitSalary {
		if err := h.users.SetSalary(ctx, msg.From.ID, salary); err != nil {
			sendText(bot, msg.Chat.ID, "Ошибка сохранения. Попробуйте позже.")
			return
		}
		h.setState(msg.From.ID, stateIdle)
		sendText(bot, msg.Chat.ID, fmt.Sprintf("Зарплата сохранена: %s", formatMoney(salary)))
		h.sendCalcPrivate(ctx, bot, msg.Chat.ID, msg.From.ID)
		return
	}

	// idle: check if salary already exists
	current, hasSalary, err := h.users.GetSalary(ctx, msg.From.ID)
	if err != nil {
		sendText(bot, msg.Chat.ID, "Ошибка. Попробуйте позже.")
		return
	}

	if !hasSalary {
		if err := h.users.SetSalary(ctx, msg.From.ID, salary); err != nil {
			sendText(bot, msg.Chat.ID, "Ошибка сохранения. Попробуйте позже.")
			return
		}
		sendText(bot, msg.Chat.ID, fmt.Sprintf("Зарплата сохранена: %s", formatMoney(salary)))
		h.sendCalcPrivate(ctx, bot, msg.Chat.ID, msg.From.ID)
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да, обновить", fmt.Sprintf("update:%.2f", salary)),
			tgbotapi.NewInlineKeyboardButtonData("Нет, оставить", "cancel"),
		),
	)
	reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf(
		"Обновить зарплату?\n%s → %s", formatMoney(current), formatMoney(salary),
	))
	reply.ReplyMarkup = keyboard
	bot.Send(reply) //nolint:errcheck
}

func (h *Handler) sendCalcPrivate(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64) {
	now := time.Now()
	h.sendCalcPrivateFor(ctx, bot, chatID, userID, now.Year(), now.Month())
}

func (h *Handler) sendCalcPrivateFor(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, year int, month time.Month) {
	salary, hasSalary, err := h.users.GetSalary(ctx, userID)
	if err != nil {
		sendText(bot, chatID, "Ошибка получения данных. Попробуйте позже.")
		return
	}
	if !hasSalary {
		h.setState(userID, stateAwaitSalary)
		sendText(bot, chatID, "Укажите вашу зарплату (например: 100000):")
		return
	}

	breakdown, err := h.calc.Calculate(ctx, salary, year, month)
	if err != nil {
		sendText(bot, chatID, "Не удалось получить производственный календарь. Попробуйте позже.")
		return
	}
	sendText(bot, chatID, h.formatter.FormatPrivate(breakdown))
}

func (h *Handler) sendCalcGroupFor(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, year int, month time.Month) {
	// salary=1 because only percentages matter for group output
	breakdown, err := h.calc.Calculate(ctx, 1, year, month)
	if err != nil {
		sendText(bot, chatID, "Не удалось получить производственный календарь. Попробуйте позже.")
		return
	}
	sendText(bot, chatID, h.formatter.FormatGroup(breakdown))
}

func sendText(bot *tgbotapi.BotAPI, chatID int64, text string) {
	bot.Send(tgbotapi.NewMessage(chatID, text)) //nolint:errcheck
}

func sendEditText(bot *tgbotapi.BotAPI, chatID int64, msgID int, text string) {
	bot.Send(tgbotapi.NewEditMessageText(chatID, msgID, text)) //nolint:errcheck
}

func parseSalary(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid salary: %q", s)
	}
	return v, nil
}

// parsePeriod разбирает аргумент /calc: "3", "март", "март 2025", "3 2025", "03.2025".
// Если год не указан, берётся год из now.
func parsePeriod(s string, now time.Time) (int, time.Month, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer(".", " ", "/", " ").Replace(s)
	parts := strings.Fields(s)
	if len(parts) == 0 || len(parts) > 2 {
		return 0, 0, fmt.Errorf("invalid period: %q", s)
	}

	month, ok := parseMonth(parts[0])
	if !ok {
		return 0, 0, fmt.Errorf("invalid month: %q", parts[0])
	}

	year := now.Year()
	if len(parts) == 2 {
		y, err := strconv.Atoi(parts[1])
		if err != nil || y < 2000 || y > 2100 {
			return 0, 0, fmt.Errorf("invalid year: %q", parts[1])
		}
		year = y
	}
	return year, month, nil
}

func parseMonth(s string) (time.Month, bool) {
	if n, err := strconv.Atoi(s); err == nil {
		if n >= 1 && n <= 12 {
			return time.Month(n), true
		}
		return 0, false
	}
	for m, name := range monthNames {
		// совпадает и с именительным ("март"), и с родительным ("марта") падежом
		if s == name || s == monthNamesGenitive[m] {
			return m, true
		}
	}
	return 0, false
}
