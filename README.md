# NOMAD, a portable app installer 

# Credits
This initial structure of this app was based on [goappmation](https://github.com/josephspurrier/goappmation).
It has then a lot evolved...

# Why ?
I needed a fast, reliable and simple way to manage portable apps (mainly on windows).

## Just use Scoop ?
If you like scoop, save yourself some time and stop reading this ;-)
Otherwise, go ahead :-)

# Other solutions ?
 * [portableapps](https://portableapps.com/) was too slow or buggy (do some fs stuff before/after run) 
 * [scoop](https://scoop.sh/) status was uncertain (it seems back on track now)

## 3.3.2023 brief comparison with scoop

| feature                         | scoop | nomad | comments                                                          |
|---------------------------------|-------|-------|-------------------------------------------------------------------|
| checksums                       | yes   | no    | PR welcome                                                        |
| lots of apps                    | yes   | no    | Copy/paste a conf and adapt it for your needs                     |
| version pattern shortcut        | no    | yes   | I love it                                                         |
| shortcuts                       | yes   | yes   | Scoop uses shims... Nomad can use custom image index for shortcut |
| push for 100% portability       | no    | yes   | Example: putty conf not saved with scoop standard bucket          |
| uses github api (when possible) | no    | yes   | Consequence: Use less bandwidth but needs a PAT                   |
| single go binary                | no    | yes   | Scoop is a list of ps scripts                                     |

# Status
It is working (I’m using it at my work) but still need some improvements before being "prod nicely" :-)

# How
 1. Download [latest release](https://github.com/jonathanMelly/portable-app-installer/releases/latest)
2. Activate [developer mode](https://learn.microsoft.com/en-us/windows/apps/get-started/enable-your-device-for-development) (if needed) [this is for symlink]
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

## Other options
Please run
```bash 
nomad --help
```

# General config
## Github
To reduce network traffic, when possible, GitHub API is used to retrieve last release info.
As GitHub API limit traffic to guest requests, a PAT (GitHub token) is very useful.
If you have a PAT, please add it in your env (GITHUB_PAT) and put the following [file](example/nomad.toml) in 
the same directory as the binary (you may also write it directly into the file).

### Create a PAT
Please follow [this link](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token) to create a basic PAT.


# App definition structure
Please refer to [this file](installer/config.go).

# Next steps
Append defs to binary, get info from their if no custom is found...

