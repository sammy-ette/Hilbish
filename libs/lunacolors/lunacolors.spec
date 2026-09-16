Name: {{{ git_dir_name }}}
Version: {{{ git_dir_version }}}
Release: 1%{?dist}
Summary: ANSI colors library for Lua
License: MIT

Source: {{{ git_dir_pack }}}

Url: https://github.com/Rosettea/lunacolors
VCS: {{{ git_dir_vcs }}}
BuildArch: noarch

%description
Lunacolors is an elegant library for handling ANSI colors in Lua.  

%prep
{{{ git_dir_setup_macro }}}

%build

%install
install -Dm644 init.lua -t %{buildroot}/usr/share/hilbish/libs/lunacolors/

%files
%doc README.md
%license LICENSE
/usr/share/hilbish/libs/lunacolors/init.lua
