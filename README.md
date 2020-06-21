Dolphin
--------

Connects Discord and Minecraft servers without mods or plugins.

[![Report](https://goreportcard.com/badge/github.com/EbonJaeger/dolphin)](https://goreportcard.com/report/github.com/EbonJaeger/dolphin) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
--------

## Building
Dolphin has a Makefile to make building and installing easier.  

To build the project, run `make`. To check the project and run tests, run `make check`.

## Installation
Create a Discord bot [here](https://discordapp.com/developers/applications/me). Next, add the bot to your Discord server using this link, replacing the Client ID with your bot's ID:
```
https://discordapp.com/oauth2/authorize?client_id=<CLIENT_ID>&scope=bot
```

In your Minecraft server.properties, set the following options and restart the server:
```
enable-rcon=true
rcon.password=<password>
rcon.port=<1-65535>
```

Place the downloaded or built binary where ever you want, and run it to generate the config. By default, the config is generated and looked for in `$HOME/.config/dolphin/dolphin.conf`. You can override this using the program's command flags.

### Using Discord Webhooks
Using a Discord webhook allows for much nicer messages to the Discord channel from Minecraft, such as using a different avatar for each Minecraft user and each message using their name. Setting it up is easy:

1) In Discord, go to your server settings, go to Webhooks, and create a new webhook for the channel you wish to use.

2) Copy the Webhook URL shown, and paste it in your Dolphin config, and enable using webhooks. Start Dolphin and that's it, you're done! :D

## Usage
```
./mcdolphin [OPTIONS]
```
Options:
```
-c, --config - The path to the configuration file to use
    --debug  - Print additional debug lines to stdout
-h, --help   - Print the help message
```

## License
Copyright © 2020 Evan Maddock (EbonJaeger)  
Makefile adapted from [usysconf](https://github.com/getsolus/usysconf), which is a Solus Project

Dolphin is available under the terms of the Apache-2.0 license
