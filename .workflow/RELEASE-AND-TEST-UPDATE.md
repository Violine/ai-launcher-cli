# Релизы и проверка автообновления

**Целевой источник обновлений в проде:** GitLab (Releases / API).  
**Для тестов и отладки:** GitHub Releases — проще создать репозиторий, тег и релиз без корпоративного GitLab.

Ниже — как собрать бинарь с версией, выложить его на GitHub и проверить сценарий обновления.

---

## 1. Целевая схема

- **Продакшен:** приложение проверяет обновления в GitLab (API/Releases), скачивает бинарь с GitLab.
- **Тесты:** та же логика, но источник — GitHub (репозиторий вида `owner/repo`). Репозиторий на GitHub используется только для проверки: создаёте теговый релиз, вешаете бинарники — локальный бинарь «старой» версии видит новую и предлагает обновление.

В коде источник задаётся (например, repo `owner/repo` или base URL). Для GitHub сейчас зашит формат: `LatestRelease(repo)` через GitHub API и `DownloadURL(repo, version)` в виде  
`https://github.com/{repo}/releases/download/{version}/ai-launcher_{GOOS}_{GOARCH}[.exe]`.

---

## 2. Как задать версию тулзы

Версия берётся из переменной `autoupdate.Version`. По умолчанию в коде стоит `"0.0.0"`. Чтобы бинарь «считал» себя конкретной версией, её нужно прошить при сборке через `-ldflags`.

### 2.1 Одна сборка с версией (локально)

```bash
VERSION=1.0.0 go build -ldflags "-X github.com/ai-launcher/cli/internal/modules/autoupdate.Version=$(VERSION)" \
  -o dist/ai-launcher \
  ./cmd/ai-launcher
```

Проверка:

```bash
./dist/ai-launcher autoupdate
# или при наличии флага:
./dist/ai-launcher --check-update
```

В выводе или в логе при старте должна фигурировать версия `1.0.0`.

### 2.2 Сборка всех платформ с версией (для релиза)

Нужно, чтобы в каждой сборке подставлялась одна и та же версия (например, из тега `v1.0.0`). В Makefile добавляется переменная `VERSION` и `-ldflags` во все цели кросс-компиляции. Пример:

```makefile
LDFLAGS := -ldflags "-X github.com/ai-launcher/cli/internal/modules/autoupdate.Version=$(VERSION)"

build-darwin-arm64:
	@mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(OUT_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PKG)
# ... и так для остальных платформ
```

Сборка:

```bash
make build-all VERSION=1.0.0
```

В `dist/` появятся бинарники, «видящие» себя как 1.0.0.

---

## 3. Имена файлов для GitHub Releases

Функция `DownloadURL` в коде формирует URL так:

- `https://github.com/{repo}/releases/download/{version}/ai-launcher_{GOOS}_{GOARCH}`  
- для Windows добавляется `.exe`.

То есть имена артефактов на GitHub должны быть **с подчёркиваниями** и без префикса `v` в имени файла (тег может быть `v1.0.0`, в URL подставляется как есть):

| Платформа      | Имя файла в релизе          |
|----------------|-----------------------------|
| macOS arm64    | `ai-launcher_darwin_arm64`  |
| macOS Intel    | `ai-launcher_darwin_amd64`  |
| Linux amd64    | `ai-launcher_linux_amd64`   |
| Windows amd64  | `ai-launcher_windows_amd64.exe` |

Сейчас Makefile по умолчанию кладёт в `dist/` файлы с дефисами (`ai-launcher-darwin-arm64` и т.д.). Для выкладки на GitHub их нужно либо переименовать, либо завести отдельную цель `make release` (или `make build-release`), которая собирает с `VERSION` и копирует артефакты в каталог с именами через подчёркивание (например, `release/`), чтобы не путать с обычным `dist/`.

---

## 4. Пошаговый план: выложить бинарь на GitHub и проверить обновление

### Шаг 1: Репозиторий на GitHub

- Создайте репозиторий (например, `yourname/ai-launcher-cli` или форк).
- Код можно пушить туда только для тестов или не пушить вовсе — для проверки обновления достаточно создать **релиз по тегу** и приложить бинарники.

### Шаг 2: Сборка бинарников с версией

- В корне проекта задайте версию и соберите под нужные платформы (или только под свою):

```bash
export VERSION=1.0.0
make build-all VERSION=$VERSION
```

- Подготовьте имена под GitHub (из `dist/` в каталог, например `release/`):

```bash
mkdir -p release
cp dist/ai-launcher-darwin-arm64   release/ai-launcher_darwin_arm64
cp dist/ai-launcher-darwin-amd64   release/ai-launcher_darwin_amd64
cp dist/ai-launcher-linux-amd64    release/ai-launcher_linux_amd64
cp dist/ai-launcher-windows-amd64.exe release/ai-launcher_windows_amd64.exe
```

(Или добавьте в Makefile цель `release`, которая делает то же самое с учётом `VERSION`.)

### Шаг 3: Создать тег и релиз на GitHub

- В репозитории на GitHub: **Releases → Create a new release**.
- **Tag:** создайте тег, например `v1.0.0` (важно: тег в формате semver с `v`).
- **Release title:** например `v1.0.0`.
- В раздел **Assets** перетащите (или загрузите) все четыре файла из `release/` с именами как в таблице выше.
- Опубликуйте релиз.

После этого URL вида  
`https://github.com/yourname/ai-launcher-cli/releases/download/v1.0.0/ai-launcher_darwin_arm64`  
должен отдавать бинарь.

### Шаг 4: Указать репозиторий для проверки (свой GitHub)

Чтобы проверять обновление против **вашего** репозитория на GitHub, задайте его в настройках — не нужно менять код.

**Вариант A: конфиг**

В `~/.config/ai-launcher/config.yaml` добавьте поле (или создайте файл с ним):

```yaml
# Репозиторий для проверки обновлений (GitHub: owner/repo). Если не задан — используется значение по умолчанию из кода.
update_repo: "yourname/ai-launcher-cli"
```

При реализации автообновления код будет брать `update_repo` из конфига и передавать его в `LatestRelease(repo)` и `DownloadURL(repo, version)`. Для тестов подставьте сюда свой GitHub-репозиторий.

**Вариант B: переменная окружения (опционально)**

Можно дополнительно поддерживать `AI_LAUNCHER_UPDATE_REPO=yourname/ai-launcher-cli` (приоритет над конфигом или наоборот — на усмотрение). Удобно для одноразовых проверок без правки файла.

---

Раньше в шаге предлагалось временно заменить константу `DefaultRepo` в коде — этого не требуется, если в конфиге задан `update_repo`.

### Шаг 5: Собрать «старую» версию и проверить обновление

Идея: на машине крутится бинарь, который считает себя «старой» версией; на GitHub лежит релиз с более новой версией — приложение должно обнаружить обновление и (когда будет реализовано) предложить скачать и установить.

1. Соберите бинарь с версией **ниже** той, что на GitHub, например:

   ```bash
   make build VERSION=0.9.0
   # или для своей платформы:
   go build -ldflags "-X github.com/ai-launcher/cli/internal/modules/autoupdate.Version=0.9.0" \
     -o ./ai-launcher-old ./cmd/ai-launcher
   ```

2. Запустите:

   ```bash
   ./ai-launcher-old
   ```

   При старте в stderr должно появиться сообщение вроде «Update available: v1.0.0 (current: 0.9.0)».

3. Когда будет реализована команда установки (`ai-launcher autoupdate` с подтверждением):

   ```bash
   ./ai-launcher-old autoupdate
   ```

   Ожидается: проверка версии → обнаружение 1.0.0 → запрос подтверждения → скачивание с GitHub и замена бинарника.

### Шаг 6: Проверка «нет обновления»

- Соберите бинарь с той же версией, что и релиз на GitHub (например, `VERSION=1.0.0`), запустите — сообщения об обновлении быть не должно.
- Либо создайте релиз с тегом `v0.0.1`, а локально соберите с `VERSION=1.0.0` — приложение не должно предлагать «обновление» вниз.

---

## 5. Краткий чеклист для теста обновления на GitHub

1. Репозиторий на GitHub создан (или выбран существующий).
2. **В конфиге задан репозиторий:** в `~/.config/ai-launcher/config.yaml` указано `update_repo: "yourname/ai-launcher-cli"` (или через переменную окружения `AI_LAUNCHER_UPDATE_REPO` — при реализации).
3. `make build-all VERSION=1.0.0` (или аналог) — бинарники в `dist/`.
4. Файлы переименованы в имена с подчёркиваниями и загружены в релиз с тегом `v1.0.0`.
5. Собран локальный бинарь с `VERSION=0.9.0` (или другой меньшей версией).
6. Запуск этого бинарника — в stderr видно «Update available: v1.0.0».
7. (После реализации) `./ai-launcher-old autoupdate` → подтверждение → скачивание и замена на бинарь с GitHub.

---

## 6. Дальше: переход на GitLab

Когда будете подключать GitLab как основной источник:

- Реализовать получение последней версии и URL бинарника через GitLab API (Projects API / Releases).
- В конфиге или переменных окружения задавать base URL или ID проекта GitLab.
- Формат URL скачивания у GitLab другой (например, `/projects/:id/repository/archive` или ссылки на job artifacts) — его нужно заложить в аналог `DownloadURL` для GitLab.
- Имена артефактов при сборке для GitLab можно оставить такими же (darwin_arm64 и т.д.), чтобы один и тот же набор файлов подходил и для GitHub (тесты), и для GitLab (прод).

Этот документ можно использовать и как инструкцию для тестов на GitHub, и как базу для последующей документации по релизам в GitLab.
