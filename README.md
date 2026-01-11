# Dota 2 AI Coach

AI-ассистент для Dota 2, который анализирует игровую ситуацию в реальном времени и даёт тактические советы через оверлей.

![Иллюстрация](dota.jpg)

## Возможности

- 🎮 Интеграция с Dota 2 через GSI (Game State Integration)
- 🤖 Поддержка AI провайдеров: Google Gemini и OpenRouter
- 💬 Автоматические советы каждые N секунд
- ❓ Возможность задавать вопросы прямо в игре
- 🖥️ Оверлей поверх игры
- ⌨️ Управление через горячие клавиши

## Требования

- Go 1.25.5+
- Windows 
- Dota 2
- API ключ от Gemini или OpenRouter

## Установка

```bash
git clone https://github.com/BrightGir/dota-ai-coach.git
cd dota-ai-coach
go mod download
```

## Настройка

### 1. API ключ

Создайте файл `.env` в корне проекта:

```env
API_KEY=your_api_key_here
```

**Где взять ключ:**
- Gemini: https://makersuite.google.com/app/apikey
- OpenRouter: https://openrouter.ai/keys

### 2. Конфигурация GSI в Dota 2

Создайте файл `gamestate_integration_aicoach.cfg` в папке с игрой:
```
C:\Program Files (x86)\Steam\steamapps\common\dota 2 beta\game\dota\cfg\gamestate_integration\
``` 
(или другая папка, куда у вас установлен Steam)

Содержимое файла:

```
"Dota 2 Integration Configuration"
{
    "uri"           "http://localhost:6000"
    "timeout"       "5.0"
    "buffer"        "0.1"
    "throttle"      "0.1"
    "heartbeat"     "30.0"
    "data"
    {
        "provider"      "1"
        "map"           "1"
        "player"        "1"
        "hero"          "1"
        "abilities"     "1"
        "items"         "1"
    }
}
```

### 3. Настройка приложения

Отредактируйте `config.json`:

```json
{
  "local_gsi_port": 6000,
  "provider": "openrouter",
  "model": "google/gemma-3-27b-it:free",
  "system_prompt": "Ты опытный тренер по Dota 2...",
  "request_interval_seconds": 15,
  "silence_duration_seconds": 15,
  "hotkey_turn_overlay": 120,
  "hotkey_focus_overlay": 121
}
```

**Параметры:**
- `local_gsi_port` — порт для GSI сервера
- `provider` — `gemini` или `openrouter`
- `model` — модель AI (например, `gemini-2.5-flash` или `google/gemma-3-27b-it:free`)
- `request_interval_seconds` — как часто запрашивать советы
- `silence_duration_seconds` — пауза после вопросов пользователя перед генерацией автоматического промпта
- `hotkey_turn_overlay` — клавиша вкл/выкл оверлея (F9 = 120)
- `hotkey_focus_overlay` — клавиша фокуса для ввода (F10 = 121)

> ⚠️ **Важно:** Качество и полезность советов напрямую зависит от выбранной модели. Более продвинутые модели дают значительно лучшие результаты, но стоят дороже. Бесплатные модели подходят для тестирования.

## Запуск

```bash
go run ./cmd/game-helper
```

Или скомпилировать:

```bash
go build -o dota-coach.exe ./cmd/game-helper
./dota-coach.exe
```

## Использование

1. Запустите приложение
2. Запустите Dota 2 и зайдите в игру
3. Оверлей появится автоматически

> 💡 **Примечание:** Dota 2 должна быть запущена в режиме **"В окне без рамки"** (Borderless Window) для корректной работы оверлея.

**Горячие клавиши: (по умолчанию)**
- `F9` — показать/скрыть оверлей
- `F10` — сфокусироваться на оверлее (для ввода текста)
- `Enter` — отправить вопрос AI

## Лицензия
Этот проект распространяется под лицензией [MIT License](LICENSE).