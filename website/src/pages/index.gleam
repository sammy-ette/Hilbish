import conf
import util

import lustre/attribute
import lustre/element
import lustre/element/html

pub fn page() -> element.Element(a) {
  html.main([attribute.class("flex flex-col")], [
    html.section(
      [
        attribute.class(
          "min-h-screen flex flex-col items-center justify-center px-6 py-20 text-center border-b border-b-pink-500/20",
        ),
      ],
      [
        html.div([attribute.class("max-w-3xl")], [
          html.h1(
            [
              attribute.class(
                "text-6xl sm:text-7xl font-bold font-[Momo_Signature] mb-6 text-neutral-900 dark:text-neutral-50",
              ),
            ],
            [element.text("Lua in your shell.")],
          ),
          html.p(
            [
              attribute.class(
                "text-xl text-neutral-600 dark:text-neutral-300 mb-8 leading-relaxed",
              ),
            ],
            [
              element.text(
                "Hilbish is a modern, extensible shell for everyone. Write your config and scripts in Lua, not bash.",
              ),
            ],
          ),
          html.div(
            [
              attribute.class(
                "flex flex-col sm:flex-row gap-4 justify-center mb-12",
              ),
            ],
            [
              html.a(
                [
                  attribute.href(conf.base_url_join("/docs/install")),
                  attribute.class(
                    "px-6 py-3 bg-pink-600 hover:bg-pink-700 text-white font-semibold rounded-lg transition-colors",
                  ),
                ],
                [element.text("Get Started")],
              ),
              html.a(
                [
                  attribute.href(conf.base_url_join("/docs")),
                  attribute.class(
                    "px-6 py-3 border border-pink-500/50 hover:border-pink-500 text-pink-300 hover:text-pink-200 font-semibold rounded-lg transition-colors",
                  ),
                ],
                [element.text("Read Docs")],
              ),
            ],
          ),
        ]),
      ],
    ),
    html.section(
      [
        attribute.class(
          "py-20 px-6 border-b border-b-pink-500/20 bg-white dark:bg-neutral-900",
        ),
      ],
      [
        html.div([attribute.class("max-w-5xl mx-auto")], [
          html.h2(
            [
              attribute.class(
                "text-4xl font-bold font-[Momo_Signature] mb-12 text-neutral-900 dark:text-neutral-50",
              ),
            ],
            [element.text("Why Lua?")],
          ),
          html.div([attribute.class("grid md:grid-cols-3 gap-8")], [
            why_card(
              "Truly Configurable",
              "Your shell, your way. Change prompts, keybinds, completions, all in Lua. The power is in your hands.",
            ),
            why_card(
              "Fast to Learn",
              "If you know Python or JavaScript, Lua will feel natural. Maybe you already know Lua too :)",
            ),
          ]),
        ]),
      ],
    ),
    html.section(
      [
        attribute.class(
          "py-20 px-6 border-b border-b-pink-500/20 bg-neutral-100 dark:bg-neutral-950",
        ),
      ],
      [
        html.div([attribute.class("max-w-5xl mx-auto")], [
          html.h2(
            [
              attribute.class(
                "text-4xl font-bold font-[Momo_Signature] mb-12 text-neutral-900 dark:text-neutral-50",
              ),
            ],
            [element.text("Built for Developers")],
          ),
          html.div([attribute.class("grid md:grid-cols-2 gap-12")], [
            feature(
              "Runner Mode",
              "Choose how to interpret commands. Lua-first (default), shell-first, pure Lua, or custom languages like Fennel.",
            ),
            feature(
              "Smart Completions",
              "Write contextual tab completions in Lua. Grid or list menus, descriptions, and flags, all fully customizable.",
            ),
            feature(
              "Customizable Line Editor",
              "Syntax highlighting, history search, vim mode are all tweakable in Lua. Full control over your input experience.",
            ),
            feature(
              "Interactive History Menu",
              "Visual history search. Browse and search your command history with an intuitive graphical interface.",
            ),
            feature(
              "Notification System",
              "In-shell messaging for background jobs and custom events. Display and manage shell notifications in Lua.",
            ),
            feature(
              "Full Lua Ecosystem",
              "Use any pure Lua library. LuaRocks packages, custom modules, whatever you need is all available.",
            ),
          ]),
        ]),
      ],
    ),
    html.section(
      [
        attribute.class(
          "py-20 px-6 border-b border-b-pink-500/20 bg-white dark:bg-neutral-900",
        ),
      ],
      [
        html.div([attribute.class("max-w-3xl mx-auto")], [
          html.h2(
            [
              attribute.class(
                "text-4xl font-bold font-[Momo_Signature] mb-8 text-neutral-900 dark:text-neutral-50",
              ),
            ],
            [element.text("Quick Example")],
          ),
          html.pre(
            [
              attribute.class(
                "bg-neutral-900 dark:bg-neutral-950 border border-pink-700/30 dark:border-pink-500/30 rounded-lg p-6 overflow-x-auto text-sm text-neutral-100",
              ),
            ],
            [
              html.code([], [
                element.text(
                  "local bait = require 'bait'\nlocal commander = require 'commander'\n\n-- Custom command\ncommander.register('mycommand', function()\n  print('Hello from Lua!')\nend)\n\n-- React to directory changes\nbait.catch('cd', function()\n  os.execute 'ls -la'\nend)\n\n-- Aliases\nhilbish.alias('ga', 'git add')\nhilbish.alias('gm', 'git commit -m')",
                ),
              ]),
            ],
          ),
        ]),
      ],
    ),
    html.section(
      [attribute.class("py-20 px-6 bg-neutral-100 dark:bg-neutral-950")],
      [
        html.div([attribute.class("max-w-3xl mx-auto text-center")], [
          html.h2(
            [
              attribute.class(
                "text-5xl font-bold font-[Momo_Signature] mb-6 text-neutral-900 dark:text-neutral-50",
              ),
            ],
            [element.text("Ready to dive in?")],
          ),
          html.p(
            [
              attribute.class(
                "text-lg text-neutral-600 dark:text-neutral-300 mb-8",
              ),
            ],
            [
              element.text("Install Hilbish and start scripting in Lua today."),
            ],
          ),
          html.div(
            [attribute.class("flex flex-col sm:flex-row gap-4 justify-center")],
            [
              html.a(
                [
                  attribute.href(conf.base_url_join("/docs/install")),
                  attribute.class(
                    "px-8 py-4 bg-pink-600 hover:bg-pink-700 text-white font-semibold rounded-lg transition-colors text-lg",
                  ),
                ],
                [element.text("Installation Guide")],
              ),
              html.a(
                [
                  attribute.href("https://github.com/sammy-ette/Hilbish"),
                  attribute.target("_blank"),
                  attribute.class(
                    "px-8 py-4 border border-pink-500/50 hover:border-pink-500 text-pink-300 hover:text-pink-200 font-semibold rounded-lg transition-colors text-lg flex items-center justify-center gap-2",
                  ),
                ],
                [
                  element.text("Star on GitHub"),
                  util.external_link_icon("h-5 w-5"),
                ],
              ),
            ],
          ),
        ]),
      ],
    ),
  ])
}

fn why_card(title: String, description: String) -> element.Element(a) {
  html.div(
    [
      attribute.class(
        "p-6 border border-pink-300/30 dark:border-pink-500/30 rounded-lg hover:border-pink-400/50 dark:hover:border-pink-500/50 transition-colors bg-neutral-50 dark:bg-neutral-800",
      ),
    ],
    [
      html.h3(
        [
          attribute.class(
            "text-xl font-semibold text-pink-700 dark:text-pink-300 mb-3",
          ),
        ],
        [
          element.text(title),
        ],
      ),
      html.p(
        [
          attribute.class(
            "text-neutral-600 dark:text-neutral-400 leading-relaxed",
          ),
        ],
        [
          element.text(description),
        ],
      ),
    ],
  )
}

fn feature(title: String, description: String) -> element.Element(a) {
  html.div([], [
    html.h3(
      [
        attribute.class(
          "text-2xl font-semibold text-neutral-900 dark:text-neutral-100 mb-3",
        ),
      ],
      [
        element.text(title),
      ],
    ),
    html.p(
      [
        attribute.class(
          "text-neutral-600 dark:text-neutral-400 leading-relaxed text-lg",
        ),
      ],
      [
        element.text(description),
      ],
    ),
  ])
}
