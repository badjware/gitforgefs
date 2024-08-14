# v1.0.0

* Added support for Github forge
* Added support for Gitea/Forgejo forge
* **BREAKING** Renamed project from `gitlabfs` to `gitforgefs`
* **BREAKING** Added mandatory configuration *fs.forge* (no default)
* **BREAKING** Changed Gitlab user configuration to use user names instead of user ids
* Handle archived repo as hidden files by default
* Improved support for old version of git
* Fixed various race conditions
* Fixed inode collision issue

## Migrating from gitlabfs

1. Run `mv ~/.local/share/gitlabfs ~/.local/share/gitforgefs` to move the cache to its new location
2. Install gitforgefs
3. Redo the configuration. gitlabfs configuration is not directly compatible with gitforgefs