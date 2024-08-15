# gitforgefs

*Formerly gitlabfs*

`gitforgefs` allows you to mount and navigate git forges (Github, Gitlab, Gitea, etc.) as a [FUSE](https://github.com/libfuse/libfuse) filesystem with every groups, organization, and users represented as a folder and every repositories represented as a symlink pointing on a local clone of the project. This is helpful to automate the organization of your local clones.

To help illustrate, this is the output of `tree` in a filesystem exposing all the repositories of a github user.
```
$ tree
.
└── badjware
    ├── aws-cloud-gaming -> /home/marchambault/.local/share/gitforgefs/github.com/257091317
    ├── certbot -> /home/marchambault/.local/share/gitforgefs/github.com/122014287
    ├── certbot-dns-cpanel -> /home/marchambault/.local/share/gitforgefs/github.com/131224547
    ├── certbot-dns-ispconfig -> /home/marchambault/.local/share/gitforgefs/github.com/227005814
    ├── CommonLibVR -> /home/marchambault/.local/share/gitforgefs/github.com/832968971
    ├── community -> /home/marchambault/.local/share/gitforgefs/github.com/424689724
    ├── docker-postal -> /home/marchambault/.local/share/gitforgefs/github.com/132605640
    ├── dotfiles -> /home/marchambault/.local/share/gitforgefs/github.com/192993195
    ├── ecommerce-exporter -> /home/marchambault/.local/share/gitforgefs/github.com/562583906
    ├── FightClub5eXML -> /home/marchambault/.local/share/gitforgefs/github.com/246177579
    ├── gitforgefs -> /home/marchambault/.local/share/gitforgefs/github.com/324617595
    ├── kustomize-plugins -> /home/marchambault/.local/share/gitforgefs/github.com/263480122
    ├── librechat-mistral -> /home/marchambault/.local/share/gitforgefs/github.com/753193720
    ├── PapyrusExtenderSSE -> /home/marchambault/.local/share/gitforgefs/github.com/832969611
    ├── Parsec-Cloud-Preparation-Tool -> /home/marchambault/.local/share/gitforgefs/github.com/258052650
    ├── po3-Tweaks -> /home/marchambault/.local/share/gitforgefs/github.com/832969112
    ├── prometheus-ecs-discovery -> /home/marchambault/.local/share/gitforgefs/github.com/187891900
    ├── simplefuse -> /home/marchambault/.local/share/gitforgefs/github.com/111226611
    ├── tmux-continuum -> /home/marchambault/.local/share/gitforgefs/github.com/160746043
    ├── ttyd -> /home/marchambault/.local/share/gitforgefs/github.com/132514236
    ├── usb-libvirt-hotplug -> /home/marchambault/.local/share/gitforgefs/github.com/128696299
    └── vfio-win10 -> /home/marchambault/.local/share/gitforgefs/github.com/388475049

24 directories, 0 files
```

## Supported forges

Currently, the following forges are supported:

| Forge                           | Name in configuration | API token permissions, if using an API key             |
| ------------------------------- | --------------------- | ------------------------------------------------------ |
| [Gitlab](https://gitlab.com)    | `gitlab`              | `read_user`, `read_api`                                |
| [Github](https://github.com)    | `github`              | `repo`                                                 |
| [Gitea](https://gitea.com)      | `gitea`               | organization: `read`, repository: `read`, user: `read` |
| [Forgejo](https://forgejo.org/) | `gitea`               | organization: `read`, repository: `read`, user: `read` |

Merge requests to add support to other forges are welcome.

## Install

Install [go](https://golang.org/) and run
``` sh
go install github.com/badjware/gitforgefs@latest
```

The executable will be in `$GOPATH/bin/gitforgefs` or `~/go/bin/gitforgefs` by default. For convenience, add `~/go/bin` in your `$PATH` if not done already.

## Usage

Download the [example configuration file](./config.example.yaml) and edit the default configuration to suit your needs.

Then, you can run gitforgefs as follows:
``` sh
gitforgefs -config config.yaml /path/to/mountpoint
```

Stopping gitforgefs will unmount the filesystem. In the event the mountpoint is stuck in a bad state (eg: due to receiving a SIGKILL), you may need to manually cleanup using `umount`:
``` sh
sudo umount /path/to/mountpoint
```

### Running automatically on user login

See [./contrib/systemd](contrib/systemd) for instructions on how to configure a systemd service to automatically run gitforgefs on user login.

## Caching

### Filesystem cache

To reduce the number of calls to the APIs and improve the responsiveness of the filesystem, gitforgefs will cache the content of the forge in memory. If a group or project is renamed, created or deleted from the forge, these change will not appear in the filesystem immediately. To force gitforgefs to refresh its cache, use `touch .refresh` in the folder to signal gitforgefs to refresh this folder.

### Local repository cache

While the filesystem lives in memory, the git repositories that are cloned are saved on disk. By default, they are saved in `$XDG_DATA_HOME/gitforgefs` or `$HOME/.local/share/gitforgefs`, if `$XDG_DATA_HOME` is unset. `gitforgefs` symlink to the local clone of that repo. The local clone is unaffected by project rename or archive/unarchive in Gitlab and a given project will always point to the correct local folder.

## Future improvements
* Cache persists forever until a manual refresh is requested. Some way to automatically refresh after a timeout would be nice.

## Building from the repo

Simply use `make` to create the executable. The executable will be in `bin/`.

See `make help` for all available targets.