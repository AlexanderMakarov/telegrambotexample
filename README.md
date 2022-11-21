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

Try bot in "poll Telegram" mode with `go run bot-local-polling.go` from repo root.
After 'Start listening for updates. Press Enter key to stop' bot in Telegram application should start to work - try to write phrases or '/help' to it.
To stop press Enter.

To check that function works in "listen requests" mode go to "telegrambot" folder and run `go run cmd/bot-local-listen.go` from here.
After 'Serving function: "HandleTelegramWebHook"' message try to check it with `curl -X POST http://localhost:8080/ -d '{"some": true}'` - app should print something like "Received unsupported update: {0 <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil>}".

### Load to GCP Function

- [Install `gcloud` CLI](https://cloud.google.com/sdk/docs/install). It should support "Generation v2" functions (`--gen2` argument in command `gcloud functions deploy` which is listed below).
- Login in `gcloud` with `gcloud auth login` command.
- Create new project with `gcloud projects create {PROJECT_NAME}`. Note that [there are a lot of limitations to project name](https://cloud.google.com/resource-manager/docs/creating-managing-projects#creating_a_project).
- Switch `gcloud` to use this project with `gcloud config set project {PROJECT_NAME}`.
- Get project unique number with `gcloud projects describe {PROJECT_NAME}` - it would be in "projectNumber" field. BTW only after this command `gcloud projects list` will start to show your new project. New project may be located somewhere on https://console.cloud.google.com/home.
- Get your Go version with `go version`. Find closest or more suitable on https://cloud.google.com/functions/docs/concepts/go-runtime#gcloud to use as an argument to the next command.
- Create a vendor directory using the contents of go.mod file with `go mod vendor`. It will created "vendor" directory with dependencies.
- TODO .env.yaml
- Create secret to store Telegram Bot API with `gcloud secrets create telegram-bot-secret --replication-policy="automatic"` (type "y" when/if it asks for enabling API). Upload secret with `gcloud secrets versions add telegram-bot-secret --data-file=".env"`. It may be verified on https://console.cloud.google.com/security/secret-manager/secret/telegram-bot-secret/.
- Run <pre>`gcloud functions deploy telegram-bot \`
    `--gen2 \`
    `--runtime=go116 \`
    `--region={REGION} \`
    `--source=. \`
    `--entry-point=HandleTelegramWebHook \`
    `--trigger-http \`
    `--allow-unauthenticated \`
    `--env-vars-file .env.yaml`</pre>
    - It will ask you _"API [cloudfunctions.googleapis.com/artifactregistry.googleapis.com/run.googleapis.com press] not enabled "_ - just press'y' for all and wait.
    - If it failed with _"...failed as the billing account is not available"_ then need to open https://console.cloud.google.com/home/dashboard?project={PROJECT_NAME} and enable billing in it - https://cloud.google.com/billing/docs/how-to/modify-project (it will ask you for billing data to charge in case of breaking free limits).
    - If it failed with _"Cloud Build API has not been used in project ... before or it is disabled. Enable it by visiting ..."_ then need to enable it and then wait few minutes until it start work.
    - If it failed with _"... Permission 'storage.objects.get' denied on resource ..."_ then it is probably because of issue above and you need to wait more.

vendor/**.md
vendor/**.txt
vendor/**.yaml
vendor/**Makefile
vendor/**.sh
vendor/**.tmpl
vendor/**.gitignore
vendor/**.lock
vendor/**LICENSE


### Inform Telegram to don't wait bot connection but reach it by WEB hook instead

Execute `curl --request POST --url https://api.telegram.org/bot<TELEGRAM_TOKEN>/setWebhook --header 'content-type: application/json' --data '{"url": "<LINK_YOU_GET_FROM_SERVERLESS_DEPLOY>"}'`