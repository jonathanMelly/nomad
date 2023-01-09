# Portable app installer (fork of [goappmation](https://github.com/josephspurrier/goappmation))

# Why ?
I needed a fast, reliable and simple way to manage portable apps.

# Other solutions ?
 * [portableapps](https://portableapps.com/) was to slow or buggy (do some fs stuff before/after run) 
 * [scoop](https://scoop.sh/) status was uncertain and I wanted to handle my apps without validation...

# How
 1. Download [latest release](https://github.com/jonathanMelly/portable-app-installer/releases/latest)
 2. Run 
```bash 
portable-app-installer app.json 
```
(where *app.json* is a file with some info on the app like [these ones I personnaly use](app-definitions))

# Options

| Flag                                  | Description                                                         |
|---------------------------------------|---------------------------------------------------------------------|
| -configs &lt;folder&gt;               | runs on all .json files in given folder                             |
| -force                                | force reinstalls (removes existing folder)                          |
| -skip=false                           | do not reuse already downloaded archive                             |
| -envvar=&lt;envVarForShortcutPath&gt; | sets env var which points to shortcut (empty to use absolute paths) |

# App definition structure
Please refer to [this file](installer/config.go)

