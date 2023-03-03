# NOMAD, a portable app installer (fork of [goappmation](https://github.com/josephspurrier/goappmation))

# Why ?
I needed a fast, reliable and simple way to manage portable apps.

# Other solutions ?
 * [portableapps](https://portableapps.com/) was to slow or buggy (do some fs stuff before/after run) 
 * [scoop](https://scoop.sh/) status was uncertain and I wanted to handle my apps without validation...

# How
 1. Download [latest release](https://github.com/jonathanMelly/portable-app-installer/releases/latest)
 2. To install / update / get status an app (*Filezilla* for instance), start a terminal and run
```bash 
nomad install filezilla
nomad update filezilla
nomad status filezilla
```

(available apps are listed [here](app-definitions) and you can add yours by adding any valid json file...)

# Essential options

| Flag                    | Description                                                         |
|-------------------------|---------------------------------------------------------------------|
| -configs &lt;folder&gt; | runs on all .json files in given folder                             |
| -force                  | force reinstall (removes existing folder)                           |
| -skip=false             | do not reuse already downloaded archive                             |
| -latest=false           | do not check for latest version (if url provided in config)         |
| -verbose                | verbose output useful for debug                                     |

# Other options
Please run
```bash 
nomad --help
```

# App definition structure
Please refer to [this file](installer/config.go)

# Next steps
Append defs to binary, get info from their if no custom is found...

