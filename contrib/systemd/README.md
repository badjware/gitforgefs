This unit file allows you to automatically start gitlabfs as a systemd unit.

## Install
1. Install gitlabfs using `go get`
2. Copy **gitlabfs@.service** into **~/.config/systemd/user**. Create the folder if it does not exists.
3. Reload systemd: `systemctl --user daemon-reload`

## Usage
1. Create your gitlabfs config file in **~/.config/gitlabfs** eg: **~/.config/gitlabfs/gitlab.com.yaml**. Make sure the config file name ends with **.yaml** and a mountpoint is configured in the file.
2. Start your service with `systemctl --user start gitlabfs@<name of your config>.service`. eg: `systemctl --user start gitlabfs@gitlab.com.service`. Omit the **.yaml** in the name of the service.
3. Enable your service start on login with `systemctl --user enable gitlabfs@<name of your config>.service`. eg: `systemctl --user enable gitlabfs@gitlab.com.service`
