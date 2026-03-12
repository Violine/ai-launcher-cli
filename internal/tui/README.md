# TUI (internal/tui)

Терминальный интерфейс на [Bubble Tea](https://github.com/charmbracelet/bubbletea) и [Lip Gloss](https://github.com/charmbracelet/lipgloss). Стили по спецификации 5.1 (MS-DOS/Norton Commander).

## Архитектура

- **RootModel** — корневая `tea.Model`: хранит стек экранов (`[]ScreenModel`), общее состояние (`*SharedState`) и фабрику экранов. Обрабатывает глобальные сообщения (размер окна, результат проверки/установки обновлений) и навигационные сообщения; остальное делегирует активному экрану.
- **SharedState** — общие данные: Config, Commands, CommandNames, Width, Height, версии, RunCommandIndex, тексты ошибок обновления; метод `ContentWidth()`.
- **ScreenModel** — интерфейс экрана: `tea.Model` + `ID() Screen`. Каждый экран — отдельный тип со своим состоянием и логикой Init/Update/View.
- **Навигация** — через сообщения `PushScreenMsg`, `PopScreenMsg`, `ReplaceScreenMsg` и команды `PushScreenCmd(s)`, `PopScreenCmd()`, `ReplaceScreenCmd(s)`. Экраны возвращают эти команды из `Update`; корень обрабатывает сообщения и обновляет стек.

## Файлы пакета

| Файл | Назначение |
|------|------------|
| `model.go` | `Screen` (enum), навигационные сообщения и Cmd, `SharedState`, `ScreenModel`, `RootModel`, `NewModel`, фабрика экранов `newScreenFactory` |
| `styles.go` | Цвета, рамки, стили кнопок и текста (`FrameWithTitle`, `ContentBoxStyle`, `BodyStyle` и др.) |
| `screen_main.go` | `MainModel` — главное меню, список команд |
| `screen_token.go` | `TokenModel` — ввод и сохранение API-токена |
| `screen_help.go` | `HelpModel` — справка по клавишам |
| `screen_progress.go` | `ProgressModel` — экран прогресса (и вариант для проверки обновлений) |
| `screen_update.go` | `UpdateConfirmModel`, `UpdateCheckErrorModel`, `UpdateInstallErrorModel`, `UpdateSuccessModel`, а также `runCheckUpdateCmd` и `runInstallCmd` |

## Навигация

Стек экранов (pushdown automaton), управляемый сообщениями:

- **PushScreenCmd(s)** — открыть экран поверх текущего (есть «назад»). Экраны возвращают из `Update`: `return m, PushScreenCmd(ScreenHelp)`.
- **PopScreenCmd()** — закрыть текущий экран и вернуться к предыдущему. Пример: `return m, PopScreenCmd()`.
- **ReplaceScreenCmd(s)** — заменить текущий экран (переход вперёд без «назад» в этом шаге). Пример: `return m, ReplaceScreenCmd(ScreenProgress)`.

Корневая модель обрабатывает эти сообщения в `Update` и изменяет стек; переключение по экранам делается без `switch` по `Screen` — активный экран берётся как `current()` (вершина стека).

## Добавление нового экрана

1. Добавить константу в enum **Screen** в `model.go`.
2. Создать тип, реализующий **ScreenModel** (поле `*SharedState` при необходимости, свои поля состояния), и методы:
   - `ID() Screen`
   - `Init() tea.Cmd`
   - `View() tea.View`
   - `Update(msg tea.Msg) (tea.Model, tea.Cmd)`
3. Зарегистрировать экран в **newScreenFactory** в `model.go`: в `switch s` добавить `case ScreenMyNew: return NewMyNewModel(root.Shared)` (и при необходимости конструктор `NewMyNewModel`).
4. Переходы с других экранов на новый — через возврат команд навигации: `return m, PushScreenCmd(ScreenMyNew)` и т.д.
5. Если экран должен открываться по глобальному сообщению (как ошибка проверки обновлений), в `RootModel.Update` обработать это сообщение и вызвать `m.pushScreen(ScreenMyNew)` (или заменить вершину стека).

Стили — в `internal/tui/styles.go` (`FrameWithTitle`, `FrameWithTitleSubtitle`, `BodyStyle`, `ButtonStyle` и т.д.).

Полное описание структуры проекта — в [.workflow/README.md](../../.workflow/README.md) (разделы «Структура проекта» и «Навигация TUI и добавление экранов»).
