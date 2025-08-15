> [!TIP]
> Check out [Hilbish: Midnight Edition](https://github.com/sammy-ette/Hilbish/tree/midnight-edition) if you want to use C Lua, LuaJIT or anything related!

<img src="./assets/hilbish-logo-and-text.png" width=512><br>
<blockquote>
🌓 The Moon-powered shell! A comfy and extensible shell for Lua fans! 🌺 ✨
</blockquote>

<img alt="GitHub commit activity" src="https://img.shields.io/github/commit-activity/m/sammy-ette/Hilbish?style=flat-square"><img alt="GitHub commits since latest release (by date)" src="https://img.shields.io/github/commits-since/sammy-ette/Hilbish/latest?style=flat-square"><img alt="GitHub contributors" src="https://img.shields.io/github/contributors/sammy-ette/Hilbish?style=flat-square"><br>
<a href="https://github.com/sammy-ette/Hilbish/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22"><img src="https://img.shields.io/github/issues/Hilbis/Hilbish/help%20wanted?style=flat-square&color=green" alt="help wanted"></a>
<a href="https://github.com/sammy-ette/Hilbish/blob/master/LICENSE"><img alt="GitHub license" src="https://img.shields.io/github/license/sammy-ette/Hilbish?style=flat-square"></a>
<a href="https://discord.gg/3PDdcQz"><img alt="Discord" src="https://img.shields.io/discord/732357621503229962?color=blue&style=flat-square"></a>
<br>

Hilbish is an extensible shell designed to be highly customizable.

It is configured in Lua, and provides a good range of features.
It aims to be easy to use for anyone, and powerful enough for
those who need more.

The motivation for choosing Lua was that its simpler and better to use
than old shell scripts. It's fine for basic interactive shell uses,
and supports [both Lua and Sh interactively](https://hilbish.sammyette.party/docs/features/runner-mode/).

That's the only place Hilbish can use traditional shell syntax though; 
everything else is Lua and aims to be infinitely configurable. 

If something isn't, open an issue!

# Screenshots
<div align="center">
<img src="gallery/tab.png">
</div>

# Getting Hilbish
**NOTE:** Hilbish is not guaranteed to work properly on Windows, starting
from the 2.0 version. It will still be able to compile, but functionality
may be lacking. If you want to contribute to make the situation better,
comment on the Windows discussion.

You can check the [install page](https://hilbish.sammyette.party/install/)
on the website for distributed binaries from GitHub or other package repositories.
Otherwise, continue reading for steps on compiling.

## Prerequisites
- [Go 1.22+](https://go.dev)
- [Task](https://taskfile.dev/installation/) (**Go on the hyperlink here to see Task's install method for your OS.**)

## Build
First, clone Hilbish. The recursive is required, as some Lua libraries
are submodules.  
```sh
git clone --recursive https://github.com/sammy-ette/Hilbish
cd Hilbish
go get -d ./...
```  

To build, run:
```
task
```  

Or, if you want a stable branch, run these commands:
```
git checkout $(git describe --tags `git rev-list --tags --max-count=1`)
task build
```  

After you did all that, run `sudo task install` to install Hilbish globally.

# Contributing
Any kind of contributions are welcome! Hilbish is very easy to contribute to.
Read [CONTRIBUTING.md](CONTRIBUTING.md) as a guideline to doing so.

**Thanks to everyone below who's contributed!**  
<a href="https://github.com/sammy-ette/Hilbish/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=sammy-ette/Hilbish" />
</a>

*Made with [contributors-img](https://contrib.rocks).*

# License
Hilbish is licensed under the [MIT license](LICENSE).  
[Images and assets](assets/) are licensed under CC-BY-SA 4.0
