# Contributing to Hilbish
Thanks for your interest in contributing to Hilbish! No matter if it's
a new feature, documentation, or bug report, we appreciate any kind
of contributions.

This file is a document to state the steps/rules to take when making
a contribution. Be sure to read through it.

## Bug Reports // Feature Requests
Use GitHub Issues to report any bugs or to request any features
that may be useful to *anyone* else.

Check [currently open issues](https://github.com/sammy-ette/Hilbish/issues)
and [closed ones](https://github.com/sammy-ette/Hilbish/issues?q=is%3Aissue+is%3Aclosed)
to make sure someone else hasn't already made the issue.

For bug reports, be sure to include:
- Hilbish Version (`hilbish -v`)
- Ways to reproduce

## Code
For any code contributions (Lua and/or Go), you should follow these rules:  
- Tab size 4 indentation
- 80 line column limit, unless it breaks code or anything like that
- In Lua prefer no braces `()` if the function takes 1 string argument
- Use camelCase

### Making the Pull Request
1. Ensure that any new install or build dependencies are documented in
the README.md and pull request.

2. Mention any and all changes, feature additons, removals, etc. This includes
useful file locations and breaking changes. Document them in the [changelog](CHANGELOG.md)
in the pull request.

3. We use [Semver](http://semver.org/) for versioning and
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
for commit messages. Please document any breaking changes and be sure to
write proper commits, or your pull request will not be considered.

4. Finally, make the pull request.

## Finding Issues to Contribute to
You can check out the [help wanted](https://github.com/sammy-ette/Hilbish/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22+)
labels to figure out what we need your help working on.  

The [up for grabs](https://github.com/sammy-ette/Hilbish/issues?q=is%3Aissue+is%3Aopen+label%3A%22up+for+grabs%22+)
labeled issues are low hanging fruit that should be easy for anyone. You can
use this to get started on contributing!
