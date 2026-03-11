# v1.2.0
* Fix a crash when `archived_project_handling` is set to "hidden" in gitlab
* Refactored caching, laying the groundwork for automatic cache invalidation.
* Default to symlink. Loopback mode is unstable.
* Add modification date to folder and symlink attributes
* Add stable inode based on the ids of the source
* Add more debugging options
* Improve loopback stability
* Add `-version` flag to print the version

# v1.1.0

* Now default to using a loopback instead of symlinks. This should improve compatibility.
  * Use `fs.use_symlinks: true` to revert to symlinks.
* Replaced taskq dependency with a simpler task queue implementation
* Fix an issue with context propagation not being done properly

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