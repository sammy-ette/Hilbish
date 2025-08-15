---
title: Install
description: Steps on how to install Hilbish on all the OSes and distros supported.
layout: page
---

There are a small amount of ways to grab Hilbish. You can download the releases from GitHub, use your package manager, or build from source.

## Official Binaries

The easiest way to get Hilbish is to get a build directly from GitHub.
At any time, there are 2 versions of Hilbish available to install:
the latest stable release, and development builds from the master branch.\
You can download both at any time, but note that the development builds may have breaking changes.\
To download the latest stable release, [see here](https://github.com/sammy-ette/Hilbishsh/releases/latest)

For the latest development build, [click here](https://nightly.link/Rosettea/Hilbish/workflows/build/master)

## Package Repositories

Methods of installing Hilbish for your Linux distro.

### Fedora (COPR)

An official COPR is offered to install Hilbish easily on Fedora.
Enable the repo: `dnf copr enable sammyette/Hilbish`

And install Hilbish: `dnf install hilbish`

Or for the latest development build from master: `dnf install hilbish-git`

### Arch Linux (AUR)

Hilbish is on the AUR. Setup an AUR helper, and install.

Example with yay: `yay -S hilbish`

Or, from master branch: `yay -S hilbish-git`

### Alpine Linux

Hilbish is currentlty in the testing/edge repository for Alpine.
Follow the steps [here](https://wiki.alpinelinux.org/wiki/Enable_Community_Repository) (using testing repositories) and install: `apk add hilbish`

## Compiling From Source

To see steps on compiling Hilbish from source, [visit the GitHub repository](https://github.com/sammy-ette/Hilbishsh#build)
