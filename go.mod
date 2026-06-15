module hilbish

go 1.25.0

require (
	github.com/arnodel/golua v0.2.0
	github.com/blackfireio/osinfo v1.0.5
	github.com/maxlandon/readline v1.0.14
	github.com/pborman/getopt v1.1.0
	github.com/sahilm/fuzzy v0.1.1
	golang.org/x/sys v0.46.0
	golang.org/x/term v0.22.0
	mvdan.cc/sh/v3 v3.8.0
)

require (
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/arnodel/strftime v0.1.6 // indirect
	github.com/evilsocket/islazy v1.11.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/olekukonko/ts v0.0.0-20171002115256-78ecb04241c0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/text v0.38.0 // indirect
)

replace mvdan.cc/sh/v3 => github.com/Rosettea/sh/v3 v3.4.0-0.dev.0.20240815163633-562273e09b73

replace github.com/maxlandon/readline => ./golibs/readline

replace layeh.com/gopher-luar => github.com/layeh/gopher-luar v1.0.10

replace github.com/arnodel/golua => github.com/Rosettea/golua v0.0.0-20241104031959-5551ea280f23
