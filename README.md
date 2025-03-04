# Ovh Notifier

This is a simple go server that checks the availibility of some OVH dedicated servers by hitting their public REST API.
It is meant to run periodicly, every hour, and then sends a message to a discord channel via Webhook if there is stock somewhere. You can also just get the information from the logs.

## Configuration

You can configure the script with two env variables:
| ENV name  | Description |
| ------------- | ------------- |
| `DISCORD_WEBHOOK_URL`  | The webhook URl used to send a message via the discord API  |
| `DISCORD_MESSAGE_CONTENT` | The content that will be added to every discord message at the top, useful to add user/role pings to the discord message |

Both those env variables are not necessary for the script to run and otuput the results in logs. To check if they have been registered by the script, you can look at the logs when initializing the script as they will be printed out.

## Modifications

You can easily change the bahavior of the script to suits your needs:

- Each `checkAndCollectAvailability` call in the main function is made to check the availability of one ressource. You can make as many as you want. You need to pass it the OVH public API endpoint for the particular server, the ressource identifier to select which server configuration you want and finaly the server name use for the discord message.

- You can change how often the stock availability is checked at line 73 `ticker := time.NewTicker(1 * time.Hour)`. Checking every 4 hours would be `ticker := time.NewTicker(4 * time.Hour)` and checking every 30 minutes would be `ticker := time.NewTicker(30 * time.Minute)`. Remember that checking too often may rate limit your IP or even ban it and won't be that useful, checking every hour is fine, even checking every 6 hours woudl realisticly be enough for this kind of products!

- If you want to change where and how the script notifies you you can simply rewrite the `sendSummaryToDiscord` function. You can than use the allServersData array to create your own notification function.
