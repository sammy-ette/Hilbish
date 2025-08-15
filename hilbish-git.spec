%global _missing_build_ids_terminate_build 0
%global debug_package %{nil}

Name: hilbish-git
Version: {{{ git_tag_version }}}.{{{ git_short_hash }}}
Release: 1%{?dist}
Summary: The flower shell. A comfy and nice little shell for Lua fans!
License: MIT

Source: {{{ git_dir_pack }}}
BuildRequires: git golang go-task
Requires: inspect succulent lunacolors

Url: https://github.com/sammy-ette/Hilbish
VCS: {{{ git_dir_vcs }}}

%description
Hilbish is a extensible shell (framework). It was made to be very customizable
via the Lua programming language. It aims to be easy to use for the casual
people but powerful for those who want to tinker more with their shell,
the thing used to interface with most of the system.

The motivation for choosing Lua was that its simpler and better to use
than old shell script. It's fine for basic interactive shell uses,
but that's the only place Hilbish has shell script; everything else is Lua
and aims to be infinitely configurable. If something isn't, open an issue!

%prep
{{{ git_dir_setup_macro }}}
sed -i '\|/etc/shells|d' Taskfile.yaml

%build
go-task

%install
go-task install PREFIX=%{buildroot}/usr BINDIR=%{buildroot}/%{_bindir}

%post
if [ "$1" = 1 ]; then
	if [ ! -f %{_sysconfdir}/shells ] ; then
		echo "%{_bindir}/hilbish" > %{_sysconfdir}/shells
		echo "/bin/hilbish" >> %{_sysconfdir}/shells
	else
		grep -q "^%{_bindir}/hilbish$" %{_sysconfdir}/shells || echo "%{_bindir}/hilbish" >> %{_sysconfdir}/shells
		grep -q "^/bin/hilbish$" %{_sysconfdir}/shells || echo "/bin/hilbish" >> %{_sysconfdir}/shells
	fi
fi
 
%postun
if [ "$1" = 0 ] && [ -f %{_sysconfdir}/shells ] ; then
	sed -i '\!^%{_bindir}/hilbish$!d' %{_sysconfdir}/shells
	sed -i '\!^/bin/hilbish$!d' %{_sysconfdir}/shells
fi

%files
%doc README.md
%license LICENSE
%{_bindir}/hilbish
%{_datadir}/hilbish
