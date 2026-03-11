# TUI (internal/tui)

Терминальный интерфейс на [Bubble Tea](https://github.com/charmbracelet/bubbletea) и [Lip Gloss](https://github.com/charmbracelet/lipgloss). Стили по спецификации 5.1 (MS-DOS/Norton Commander).

## Файлы пакета

| Файл | Назначение |
|------|------------|
| `model.go` | `Model`, `Screen` (enum), `ScreenStack`, `CurrentScreen()`, `PushScreen` / `PopScreen` / `ReplaceScreen`, общий `Update` / `View` с роутингом по текущему экрану |
| `styles.go` | Цвета, рамки, стили кнопок и текста (`FrameWithTitle`, `ContentBoxStyle`, `BodyStyle` и др.) |
| `screen_main.go` | Главное меню — список команд |
| `screen_token.go` | Ввод и сохранение API-токена |
| `screen_help.go` | Справка по клавишам |
| `screen_progress.go` | Экран прогресса (загрузка, установка) |
| `screen_update.go` | Экраны потока обновления: подтверждение, проверка, ошибки, успех |

## Навигация

Используется **стек экранов** (pushdown automaton):

- **PushScreen(m, s)** — открыть экран поверх текущего (есть «назад»).
- **PopScreen(m)** — закрыть текущий экран и вернуться к предыдущему.
- **ReplaceScreen(m, s)** — заменить текущий экран (переход вперёд без «назад» в этом шаге).

Текущий экран = `m.CurrentScreen()` (вершина `m.ScreenStack`). В `Update` и `View` по нему выбирается `update*Screen` / `view*Screen`.

## Добавление нового экрана

1. Добавить константу в enum `Screen` в `model.go`.
2. При необходимости — поле состояния в `Model` или переиспользовать существующее.
3. Реализовать `view*Screen(m Model) tea.View` и `update*Screen(m Model, msg tea.Msg) (tea.Model, tea.Cmd)` (новый файл `screen_*.go` или существующий).
4. В `model.go` в `Update` и `View` добавить ветку `case ScreenMyNew:` с вызовом новой функции.
5. Переходы только через `PushScreen` / `PopScreen` / `ReplaceScreen`.

Полное описание структуры проекта и навигации — в [.workflow/README.md](../../.workflow/README.md) (разделы «Структура проекта» и «Навигация TUI и добавление экранов»).
