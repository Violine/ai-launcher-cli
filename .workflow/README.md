# AI Launcher CLI — рабочая документация и план

Версия: 1.0 | Дата: 10.03.2026

## Назначение

Здесь ведётся вся рабочая документация по разработке AI Launcher CLI: план действий по ТЗ, текущая архитектура, сравнение подходов, история решений. Исходное техническое задание — [SPECIFICATION.md](./SPECIFICATION.md) (функциональные и нефункциональные требования, экраны UI, внешние интеграции, критерии приёмки).

---

## План действий (по спецификации)

План привязан к разделам 3 и 7 ТЗ. Отмечено выполненное; остальное — в работе или в бэклоге.

**Приоритеты реализации:**
1. **Автообновление ai-launcher-cli** (раздел 3.6) — в первую очередь.
2. **Обновление MCP для установленных AI-инструментов** — во вторую очередь: получение пакетов из registry, установка/обновление npm MCP, обновление MCP-конфигураций в конфигах AI-агентов (Claude, Cursor и др.).

### 3.1. Управление API-ключами (Config Manager)

- [x] FR-101: Запрос API-ключа при первом запуске
- [x] FR-102: Маскировка ввода ключа (*)
- [x] FR-103: Валидация формата ключа перед сохранением
- [x] FR-104: Ошибка и повторный ввод при невалидном ключе
- [x] FR-105: Ключ в ~/.config/ai-launcher/config.yaml
- [x] FR-106: Права доступа 0600 на файл конфигурации
- [x] FR-107: Сброс ключа через интерфейс (F7)

### 3.2. Получение списка моделей (API Client)

- [ ] FR-201: Запрос к API за доступными моделями после ввода ключа
- [ ] FR-202: Индикатор загрузки с анимацией (экран 5.2.2: каркас готов — ScreenProgress, прогресс-бар; использовать для загрузки моделей, обновления launcher, MCP)
- [ ] FR-203: Динамическое формирование списка инструментов из моделей
- [ ] FR-204: Кэширование списка моделей на 1 час
- [ ] FR-205: Принудительное обновление списка (R)

### 3.3. Список инструментов (TUI)

- [ ] FR-301: Таблица — #, Название, Модель, Статус, Избранное
- [ ] FR-302: Подсветка выбранной строки
- [ ] FR-303: Прокрутка ↑/↓, PgUp/PgDn, Home/End
- [ ] FR-304: Индикатор позиции (Item X of Y)
- [ ] FR-305: Enter — раскрытие строки с кнопками действий
- [ ] FR-306: Кнопки [▶ Run] [✎ Edit] [★ Fav] [◉ Toggle]
- [ ] FR-307: Навигация по кнопкам ←/→
- [ ] FR-308: Закрытие раскрытой строки (Esc / переход на другую строку)

### 3.4. Редактирование инструмента

- [ ] FR-401: Экран редактирования — Command, Model, ENV
- [ ] FR-402: Навигация Tab/Shift+Tab
- [ ] FR-403: Чекбоксы Enabled и Favorite
- [ ] FR-404: Кнопки [Save] и [Cancel]
- [ ] FR-405: Быстрое сохранение F2
- [ ] FR-406: Добавление произвольных ENV-переменных

### 3.5. Запуск инструмента (Process Executor)

- [ ] FR-501: Запуск инструмента как subprocess
- [ ] FR-502: Передача сконфигурированных ENV
- [ ] FR-503: Передача API-ключа в ANTHROPIC_API_KEY
- [ ] FR-504: stdin/stdout/stderr от launcher к инструменту
- [ ] FR-505: Возврат в launcher после завершения

### 3.6. Автообновление (Auto-Updater)

- [x] FR-601: Проверка обновлений при запуске (в фоне)
- [x] FR-602: Сравнение версий по semver (каркас)
- [ ] FR-603: Уведомление о новой версии
- [ ] FR-604: Скачивание и установка с подтверждением
- [ ] FR-605: Резервная копия перед обновлением
- [ ] FR-606: Откат при ошибке
- [ ] FR-607: Флаг --check-update

### 3.7. Обновление MCP для установленных AI-инструментов (приоритет 2)

- [x] Каркас: internal/gitlab, internal/mcp, плагин mcpupdate
- [ ] Получение списка MCP-пакетов из GitLab Package Registry (npm)
- [ ] Установка и обновление npm-пакетов MCP из registry
- [ ] Конфигурация registry URL и токена (в config или ENV)
- [ ] **Обновление MCP-конфигураций для AI-агентов:** актуализация конфигов установленных инструментов (Claude, Cursor, OpenCode и др.) после установки/обновления MCP-пакетов

### 3.8. Телеметрия (OpenTelemetry)

- [ ] FR-701: Метрики использования (какие инструменты, как часто)
- [ ] FR-702: Метрики производительности (время запуска, сессии)
- [ ] FR-703: Трейсинг ключевых операций
- [ ] FR-704: Отправка на OTEL Collector
- [ ] FR-705: Отключение телеметрии через конфиг/ENV
- [ ] FR-706: Не собирать персональные данные и содержимое запросов

### Дополнительно (остальные направления по ТЗ)

- [ ] Плагины: configgen, agentrun, autoupdate — реализация по ТЗ (configgen — генерация конфигов для Claude/OpenCode и др.)

### Этапы (раздел 7.1 ТЗ)

| Этап | Содержание | Срок | Статус |
|------|------------|------|--------|
| 1 | MVP: TUI + Config + базовый запуск | 2 нед | В плане |
| 2 | Интеграция с LLM Proxy API | 1 нед | В плане |
| 3 | OpenTelemetry телеметрия | 1 нед | В плане |
| 4 | Автообновление ai-launcher-cli | 1 нед | Каркас есть — **приоритет 1** |
| 4b | Обновление MCP для AI-инструментов и конфигов агентов | 1 нед | Каркас есть — **приоритет 2** |
| 5 | Тестирование и документация | 1 нед | — |
| 6 | Пилот на группе пользователей | 1 нед | — |

---

## Структура проекта

```
ai-launcher-cli/
├── .workflow/              # рабочая документация (здесь)
├── cmd/ai-launcher/        # точка входа (main.go — плагины, CLI)
├── internal/
│   ├── config/             # конфигурация (YAML, Load/Save)
│   ├── updater/            # самообновление (semver, backup, rollback)
│   ├── gitlab/             # клиент GitLab Package Registry и при необходимости Releases
│   ├── mcp/                # обновление MCP поверх internal/gitlab
│   ├── modules/            # плагины-команды
│   │   ├── configgen/      # генерация config для Claude/OpenCode
│   │   ├── mcpupdate/      # обновление MCP через GitLab
│   │   ├── agentrun/       # запуск AI-агентов (exec)
│   │   └── autoupdate/     # проверка обновлений (вызов internal/updater)
│   └── tui/                # TUI (Bubble Tea + Lip Gloss)
│       ├── model.go        # Model, ScreenStack, PushScreen/PopScreen/ReplaceScreen, Update/View
│       ├── styles.go       # стили (FrameWithTitle, кнопки, цвета по спецификации 5.1)
│       ├── screen_main.go  # главное меню (список команд)
│       ├── screen_token.go # ввод API-токена
│       ├── screen_help.go  # справка по клавишам
│       ├── screen_progress.go # экран прогресса (загрузка, установка)
│       └── screen_update.go   # экраны обновления (подтверждение, ошибки, успех)
├── pkg/plugin/             # интерфейс плагина (Name, Run)
├── api/                    # типы/клиенты протоколов MCP и Agent
├── configs/                # примеры YAML (config.example.yaml)
├── go.mod
└── go.sum
```

Подробнее про навигацию TUI и добавление новых экранов — см. [Навигация TUI и добавление экранов](#навигация-tui-и-добавление-экранов) ниже.

Отдельные пакеты `internal/executor`, `internal/telemetry`, `internal/api` (LLM Proxy) и при необходимости `internal/tools` (Tool Registry) появятся при реализации соответствующих FR. Экран загрузки с прогресс-баром (ТЗ 5.2.2, FR-202) — см. [.workflow/PROGRESS-SCREEN.md](.workflow/PROGRESS-SCREEN.md).

---

## Навигация TUI и добавление экранов

### Как устроена навигация

TUI использует **стек экранов** (pushdown automaton): текущий экран — вершина стека, переходы выполняются тремя операциями без мутации модели (в духе Elm/Bubble Tea).

| Операция | Назначение | Пример |
|----------|------------|--------|
| **PushScreen(m, s)** | Открыть экран поверх текущего; с него можно вернуться «назад» | F1 Help, F7 Token, пункт меню «Обновление» → подтверждение |
| **PopScreen(m)** | Закрыть текущий экран и вернуться к предыдущему | Esc в Help/Token, «Нет»/Esc в диалоге обновления, Cancel на экранах ошибок |
| **ReplaceScreen(m, s)** | Заменить текущий экран на другой (без добавления в стек); «назад» на предыдущий шаг потока нельзя | Подтверждение «Да» → Progress, Progress → Success/InstallError, Retry → Checking |

- **Модель:** `Model.ScreenStack []Screen` — слайс экранов, верхний = активный. Метод `CurrentScreen()` возвращает вершину стека (или `ScreenMain`, если стек пуст).
- **Роутинг:** в `Model.Update` и `Model.View` используется `m.CurrentScreen()`; по нему вызывается соответствующая функция `update*Screen` / `view*Screen`.
- **Глобальные сообщения** (размер окна, результат проверки обновлений, результат установки) обрабатываются в `Update` до делегирования по экрану; смена экрана по ним — через `ReplaceScreen` или `PushScreen` (например, фоновая ошибка проверки — `PushScreen(ScreenUpdateCheckError)`).

Подробнее: `internal/tui/model.go` (типы `Screen`, функции `PushScreen`, `PopScreen`, `ReplaceScreen`).

### Как добавить новый экран

1. **Добавить константу в enum `Screen`** в `internal/tui/model.go`:
   ```go
   const (
       // ...
       ScreenUpdateSuccess
       ScreenMyNew        // новый экран
   )
   ```

2. **При необходимости добавить состояние экрана** в `Model` (как `TokenScreenState`, `UpdateScreenState`) или переиспользовать существующие поля.

3. **Реализовать отрисовку и обработку сообщений.** Варианты:
   - добавить файл `internal/tui/screen_mynew.go` с функциями `viewMyNewScreen(m Model) tea.View` и `updateMyNewScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd)`;
   - либо разместить их в подходящем существующем файле (например, в `screen_update.go` для экранов того же потока).

4. **Подключить в роутинг** в `model.go`:
   - в `Update`: в `switch m.CurrentScreen()` добавить ветку `case ScreenMyNew: return updateMyNewScreen(m, msg)`;
   - в `View`: в `switch m.CurrentScreen()` добавить ветку `case ScreenMyNew: return viewMyNewScreen(m)`.

5. **Переходы на новый экран и обратно** — только через стек:
   - открыть поверх текущего (с возможностью «назад»): `m = PushScreen(m, ScreenMyNew)`; возврат: `m = PopScreen(m)`;
   - заменить текущий экран (без «назад» в рамках этого шага): `m = ReplaceScreen(m, ScreenMyNew)`.

6. **Глобальные сообщения:** если экран должен открываться по сообщению (аналогично `UpdateCheckErrorMsg`), в обработчике этого сообщения в `Model.Update` вызвать `PushScreen` или `ReplaceScreen`.

Стили рамок и текста — в `internal/tui/styles.go` (`FrameWithTitle`, `FrameWithTitleSubtitle`, `BodyStyle`, `ButtonStyle` и т.д.).

---

## Сборка артефактов для релиза

Артефакты для размещения в релизе (GitHub Releases или GitLab) собираются через Makefile с подстановкой версии в бинарь (для автообновления).

### Быстрый вариант

```bash
make release VERSION=1.0.0
```

В каталоге `release/` появятся четыре файла с именами, ожидаемыми при скачивании (см. ниже). Эти файлы нужно загрузить в раздел **Assets** при создании релиза по тегу `v1.0.0`.

### Что делает `make release`

1. **Сборка под все платформы** (`make build-all`) с передачей версии в бинарь:
   - `-ldflags "-X github.com/ai-launcher/cli/internal/modules/autoupdate.Version=$(VERSION)"`
   - Результат в `dist/`: `ai-launcher-darwin-arm64`, `ai-launcher-darwin-amd64`, `ai-launcher-linux-amd64`, `ai-launcher-windows-amd64.exe`.

2. **Копирование в `release/` с именами для GitHub** (с подчёркиваниями, без префикса `v` в имени файла):

| Платформа   | Имя файла в `release/`              |
|-------------|-------------------------------------|
| macOS arm64 | `ai-launcher_darwin_arm64`          |
| macOS Intel | `ai-launcher_darwin_amd64`          |
| Linux amd64 | `ai-launcher_linux_amd64`          |
| Windows amd64 | `ai-launcher_windows_amd64.exe`   |

Формат URL скачивания (GitHub):  
`https://github.com/{repo}/releases/download/{version}/ai-launcher_{GOOS}_{GOARCH}[.exe]`  
— поэтому имена артефактов должны быть с подчёркиваниями.

### Сборка без Makefile

Одна платформа с версией:

```bash
VERSION=1.0.0
go build -ldflags "-X github.com/ai-launcher/cli/internal/modules/autoupdate.Version=$VERSION" \
  -o dist/ai-launcher ./cmd/ai-launcher
```

Все платформы вручную (четыре команды, подставьте свой `VERSION`):

```bash
VERSION=1.0.0
LDFLAGS="-X github.com/ai-launcher/cli/internal/modules/autoupdate.Version=$VERSION"

GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/ai-launcher-darwin-arm64 ./cmd/ai-launcher
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/ai-launcher-darwin-amd64 ./cmd/ai-launcher
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/ai-launcher-linux-amd64 ./cmd/ai-launcher
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/ai-launcher-windows-amd64.exe ./cmd/ai-launcher
```

После этого переименовать файлы в имена с подчёркиваниями и загрузить в релиз (см. таблицу выше).

### Очистка

```bash
make clean
```

Удаляет каталоги `dist/` и `release/`.

Полный сценарий: создание тега, публикация релиза на GitHub и проверка автообновления — в [RELEASE-AND-TEST-UPDATE.md](./RELEASE-AND-TEST-UPDATE.md).

---

## Целевая архитектура

- **Плагинный слой:** контракт в `pkg/plugin` (Name, Run(ctx)). Реализации в `internal/modules/*`. Регистрация и запуск по имени команды в `cmd/ai-launcher`.
- **GitLab:** общий клиент в `internal/gitlab`, MCP — в `internal/mcp` поверх него.
- **Запуск агентов:** плагин `agentrun` вызывает логику subprocess/ENV (пакет `internal/executor` или внутри модуля).
- **TUI:** в `internal/tui` (Bubble Tea); режим «без аргументов → TUI» в main при реализации.
- **Конфиги:** примеры в `configs/`; протоколы при необходимости в `api/`.

---

## Сравнение подходов к архитектуре

Рассматривались два варианта: **структура по компонентам ТЗ (2.1.1)** — TUI-first, папки TUI/Config/Tools/Executor/Telemetry/Updater/API — и **текущая реализация** — CLI-first, плагины + общие сервисы (config, updater, gitlab, mcp).

| Аспект | По компонентам ТЗ | Текущая (plugin-first) |
|--------|-------------------|-------------------------|
| Точка входа | TUI: tea.NewProgram(tui.NewModel(cfg)) | CLI: os.Args[1] → плагин |
| Расширяемость | Новый экран/компонент в TUI | Новый плагин в modules/ + регистрация |
| Соответствие ТЗ | Прямое (диаграмма 2.1.1, экраны 5.2) | TUI — заглушка, часть пакетов отсутствует |
| Вне ТЗ | GitLab/MCP не заложены | gitlab + mcp уже в структуре |

**Плюсы подхода по ТЗ:** один сценарий (запуск → TUI), полный набор компонентов в коде, отдельный Tool Registry, готовность к экранам и стеку из ТЗ. **Минусы:** нет плагинного контракта для CLI-сценариев, GitLab/MCP добавлять отдельно, headless/CI сложнее.

**Плюсы текущей реализации:** расширяемость через плагины, CLI и автоматизация из коробки, уже есть gitlab/mcp, простой main. **Минусы:** отклонение от буквы ТЗ (TUI не главный интерфейс), нет пакетов executor/telemetry/api (LLM), реестр инструментов только в config.

**Рекомендация:** не переписывать всё под один подход. Сближение: в текущем проекте добавить пакеты по ТЗ (executor, telemetry, api для LLM, при необходимости tools), ввести режим «без аргументов → TUI»; при необходимости в варианте по ТЗ — добавить gitlab/mcp и плагинный контракт для CLI.

---

## Внешние зависимости (из ТЗ)

| Компонент | Назначение |
|-----------|------------|
| Update Server | GitLab Releases / Nexus |
| MCP | npm-пакеты в GitLab Registry |
| LLM Proxy API | GET /v1/models |
| OTEL Collector | OTLP/gRPC |

---

## История изменений

| Дата | Изменение |
|------|-----------|
| 10.03.2026 | Создан каркас проекта, .workflow, модули updater, mcp, executor |
| 10.03.2026 | Рефакторинг архитектуры: pkg/plugin, internal/modules, internal/gitlab, internal/mcp, configs/, api/, tui |
| 10.03.2026 | Вся документация и план перенесены в README; план действий приведён к разделам и FR спецификации; сравнение архитектур встроено в README |
| 10.03.2026 | Добавлены приоритеты: сначала автообновление launcher, затем обновление MCP для AI-инструментов; в план и спецификацию внесён пункт по обновлению MCP-конфигураций для AI-агентов |
| 10.03.2026 | FR-601: проверка обновлений при запуске в фоне (goroutine в main, CheckInBackground с callback для уведомления) |
| 10.03.2026 | TUI: экран ввода токена с Lip Gloss (FR-101–FR-104, FR-107), главный экран со стилями 5.1, передача config из main |
