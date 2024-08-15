This unit file allows you to automatically start gitforgefs as a systemd unit.

## Setup
1. Install gitforgefs
2. Copy **gitforgefs@.service** into **$HOME/.config/systemd/user**. Create the folder if it does not exists.
``` sh
mkdir -p $HOME/.config/systemd/user
curl -o $HOME/.config/systemd/user/gitforgefs@.service https://raw.githubusercontent.com/badjware/gitforgefs/dev/contrib/systemd/gitforgefs%40.service
```
3. Reload systemd: `systemctl --user daemon-reload`

## Usage
1. Create your gitforgefs config file in **$HOME/.config/gitforgefs** eg: **$HOME/.config/gitforgefs/gitlab.com.yaml**. Make sure the config file name ends with **.yaml** and a mountpoint is configured in the file.
2. Start your service with `systemctl --user start gitforgefs@<name of your config>.service`. eg: `systemctl --user start gitforgefs@gitlab.com.service`. Omit the **.yaml** extension.
3. Enable your service to start on login with `systemctl --user enable gitforgefs@<name of your config>.service`. eg: `systemctl --user enable gitforgefs@gitlab.com.service`
