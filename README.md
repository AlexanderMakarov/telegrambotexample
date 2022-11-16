# Example of the Telegram Bot

Based on https://iamondemand.com/blog/building-your-first-serverless-telegram-bot/
and https://gitlab.com/Athamaxy/telegram-bot-tutorial/-/blob/main/TutorialBot.go

Should cost nothing due to Google Cloud Provider Free plan.

# Setup

### Register new bot in Telegram

https://telegram.me/BotFather

It will provide TELEGRAM_APITOKEN and link to a bot in Telegram.

### Provide bot with TELEGRAM_APITOKEN and try locally

Create ".env" file in the root and set content as

```
TELEGRAM_APITOKEN=<token_received_in_step_before_without_triangle_brackets>
TELEGRAM_DEBUG=true
```

Start with `go run bot.go`. Process should start waiting requests and bBot in Telegram should start to work.
To stop press Ctrl + C.

### Load to GCP Function

TODO

### Inform Telegram to don't wait bot connection but reach it by WEB hook instead

Execute `curl --request POST --url https://api.telegram.org/bot<TELEGRAM_TOKEN>/setWebhook --header 'content-type: application/json' --data '{"url": "<LINK_YOU_GET_FROM_SERVERLESS_DEPLOY>"}'`