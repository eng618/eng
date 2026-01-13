/*
Package config is a helper package to facilitate getting and setting config
values using Viper.

This package provides functions to read and write configuration settings for
the eng CLI tool. It supports various configuration file formats and allows for
easy management of local configuration settings.

Key features:
- Load configuration from JSON, YAML, TOML, and other formats.
- Set and get configuration values programmatically.
- Support for environment variable overrides.
- Easy integration with the Viper library.

These are local configs related to the eng CLI, not deployment configs.
*/
package config
