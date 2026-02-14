# feed-tgbot

Telegram bot for RSS/Atom feeds with manual `/news` requests and scheduled updates.

## Requirements

- Docker 24+ (or compatible)
- Docker Compose plugin (`docker compose`)
- Telegram bot token from BotFather

## Quick Start (Compose-first)

1. Create environment file:

```bash
cp .env.example .env
```

2. Edit `.env` and set your token:

```env
TG_BOT_TOKEN=123456:your_real_token
TZ=UTC
```

3. Build and start:

```bash
docker compose up -d --build
```

4. Check logs:

```bash
docker compose logs -f bot
```

5. Open Telegram and send `/start` to your bot.

## What Compose Does

- Builds the bot image locally from this repository.
- Runs container `feed-tgbot` with `restart: unless-stopped`.
- Persists SQLite DB (`bot.db`) in Docker named volume `bot_data`.

## Persistence and Data Safety

The bot stores subscriptions and state in SQLite (`bot.db`) inside the container workdir (`/app`).
Compose mounts named volume `bot_data` to keep data across container restarts/recreates.

### Backup database volume

```bash
docker run --rm \
  -v bot_data:/data \
  -v "$PWD":/backup \
  alpine:3.22 \
  sh -c 'cp /data/bot.db /backup/bot.db.backup'
```

### Restore database volume

```bash
docker run --rm \
  -v bot_data:/data \
  -v "$PWD":/backup \
  alpine:3.22 \
  sh -c 'cp /backup/bot.db.backup /data/bot.db'
```

## Upgrade / Redeploy

```bash
git pull
docker compose up -d --build
```

## Stop / Start / Remove

Stop:

```bash
docker compose stop
```

Start:

```bash
docker compose start
```

Remove containers (keep DB volume):

```bash
docker compose down
```

Remove containers and DB volume (data loss):

```bash
docker compose down -v
```

## Optional: Plain Docker Run (without Compose)

Build:

```bash
docker build -t feed-tgbot:local .
```

Run:

```bash
docker run -d \
  --name feed-tgbot \
  --restart unless-stopped \
  --env-file .env \
  -v bot_data:/app \
  feed-tgbot:local
```

Logs:

```bash
docker logs -f feed-tgbot
```

## Security Notes

- Never commit `.env` with a real `TG_BOT_TOKEN`.
- If a token is exposed, rotate it in BotFather immediately.
- This setup runs the bot process as non-root inside container.

## Troubleshooting

### Bot does not start

- Check container logs:

```bash
docker compose logs --tail=200 bot
```

- Common issue: missing/invalid `TG_BOT_TOKEN`.

### No news is fetched

- Confirm server has outbound internet access.
- Validate feed URLs are reachable from server network.

### Data disappeared

- Ensure you did not run `docker compose down -v`.
- Inspect volume:

```bash
docker volume ls
docker volume inspect bot_data
```
