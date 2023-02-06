# !kaboom

> **Note**
>  You must have `ffmpeg` in your path and Opus libs already installed.

## Features

- 💣 Kick all users from voice channel
- 🎶 Playback sound (triggered custom emoji)

## Development

1. Copy `.env.sample`

```shell
> cp .env.sample .env.local
```

2. Edit `.env.local`

```shell
TOKEN=              # Bot token
GUILD_ID=           # Server ID
TEXT_CHANNEL_ID=    # Text channel ID for sending message
VOICE_CHANNEL_ID=   # Voice channel ID for play sounds
FOLDER=             # Directory path contains sound files
```

3. Run

```shell
# start bot with loading .env.local (default)
> go run main.go

# specified .env.* file (eg .env.production)
> GO_ENV=production go run main.go
```
