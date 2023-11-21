# Go telegram notification

## Description
This service allows you to send mass messages by your bots in telegram. Also supports multiple bots to use for a single message.
By default, there's a limit of 25 messages per second. <br>
You can change it in the ``internal/config/settings.go`` file.

---
## Instalation

1. Create new ``.env`` file (you can use the ``.env.example`` template)
```shell
cp .env.example .env
```
2. Install docker (https://www.docker.com/products/docker-desktop/)
3. Build docker 
```shell
sudo docker compose build
```
4. Run docker
```shell
sudo docker compose up
```

---
## Usage
### General
First of all, you need to create your bots. This service can handle multiple bots. <br>
To add your bot you need to obtain your bot token at https://t.me/BotFather <br>
After you obtained a token you need to create new row in the database. <br>
Connect to the database in any way you want and insert your bot settings:
```sql
insert into telegram_bots(bot_token, bot_host, enabled)
values ('<YOUR TOKEN HERE>', '<BOT HOST KEY>', True);
```
Explanation:<br>
``bot_token`` - is the token you obtained from the BotFather <br>
``bot_host``  - is the filter key for this bot. Every message settings will contain host key to determine which bot should process that message.

### Sending messages
Invoke an ``AddNotificationsToQueue`` method of the gRPC with this structure:
```json
{
    "message_settings": [
      {"telegram_user_id": 0, "telegram_bot_host": "string"},
      {"telegram_user_id": 0, "telegram_bot_host": "string"}
    ],
    "text": "string"
}
```
``message_settings`` - is the list of the users to receive message, provided in ``text`` field. <br>
``telegram_bot_host`` - is the key of the bot, that would send message to user, provided in ``telegram_user_id`` field. <br>
This method will add all of the ``message_settings`` objects to the global message queue. Message processor will try to process 25 messages per second by default.
