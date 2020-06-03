# RVM Bundler Cloud Native Buildpack

The RVM Bundler Cloud Native Buildpack installs Bundler in an OCI image. It requrires the [RVM CNB](https://github.com/avarteqgmbh/rvm-cnb).

## Functionality

1. The RVM Bundler CNB installs Bundler into its own layer. The version of Bundler to be installed can be configured in [buildpack.toml](buildpack.toml) or in `buildpack.yml`.
1. It also executes `bundle install` to install the Gemfile's gems into its own layer.

## Dependencies

This CNB requires the [RVM CNB](https://github.com/avarteqgmbh/rvm-cnb) as a dependency in the build and launch layers.

## TODO

1. Add configuration options for bundler.
