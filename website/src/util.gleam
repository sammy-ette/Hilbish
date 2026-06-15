import gleam/int
import gleam/string

import lustre/attribute
import lustre/element
import lustre/element/html

import conf
import glaml
import post
import theme

pub fn sort_weight(p1: #(String, post.Post), p2: #(String, post.Post)) {
  let p1_weight = case glaml.select_sugar({ p1.1 }.metadata, "weight") {
    Ok(glaml.NodeInt(w)) -> w
    _ -> 0
  }

  let p2_weight = case glaml.select_sugar({ p2.1 }.metadata, "weight") {
    Ok(glaml.NodeInt(w)) -> w
    _ -> 0
  }

  case p1_weight == p2_weight {
    True -> string.compare({ p1.1 }.name, { p2.1 }.name)
    False -> int.compare(p1_weight, p2_weight)
  }
}

/// An "external link" arrow icon, sized via the given Tailwind class.
pub fn external_link_icon(class: String) -> element.Element(a) {
  element.unsafe_raw_html(
    "",
    "tag",
    [],
    "<svg xmlns=\"http://www.w3.org/2000/svg\" fill=\"none\" viewBox=\"0 0 24 24\" stroke-width=\"1.5\" stroke=\"currentColor\" class=\""
      <> class
      <> "\">
  <path stroke-linecap=\"round\" stroke-linejoin=\"round\" d=\"M13.5 6H5.25A2.25 2.25 0 0 0 3 8.25v10.5A2.25 2.25 0 0 0 5.25 21h10.5A2.25 2.25 0 0 0 18 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25\" />
</svg>
",
  )
}

pub fn link(url: String, text: String, out: Bool) {
  html.a(
    [
      attribute.href(url),
      case out {
        False -> attribute.none()
        True -> attribute.target("_blank")
      },
    ],
    [
      html.span(
        [
          attribute.class("inline-flex hover:underline"),
          attribute.class(theme.accent_link),
        ],
        [
          case out {
            False -> element.none()
            True -> external_link_icon("size-6")
          },
          element.text(text),
        ],
      ),
    ],
  )
}

pub fn nav(show_sidebar_toggle: Bool) -> element.Element(a) {
  html.nav(
    [
      attribute.class(
        "sticky top-0 z-50 h-16 flex items-center bg-neutral-100/95 dark:bg-neutral-950/95 backdrop-blur-md border-b border-b-pink-300/30 dark:border-b-pink-500/20 px-4 sm:px-6",
      ),
    ],
    [
      html.div(
        [
          attribute.class(
            "w-full max-w-7xl mx-auto flex items-center justify-between gap-4",
          ),
        ],
        [
          html.div([attribute.class("flex items-center gap-3")], [
            case show_sidebar_toggle {
              False -> element.none()
              True ->
                html.label(
                  [
                    attribute.for("sidebar-toggle"),
                    attribute.class(
                      "sm:hidden -ml-1 cursor-pointer text-neutral-700 dark:text-neutral-200 hover:text-pink-600 dark:hover:text-pink-400 transition-colors",
                    ),
                    attribute.attribute("aria-label", "Toggle sidebar"),
                  ],
                  [hamburger_icon()],
                )
            },
            html.a(
              [
                attribute.href(conf.base_url_join("/")),
                attribute.class("flex items-center gap-2 group"),
              ],
              [
                html.img([
                  attribute.src(conf.base_url_join("/hilbish-flower.png")),
                  attribute.class("h-7 w-7"),
                ]),
                html.span(
                  [
                    attribute.class(
                      "text-xl font-bold text-neutral-900 dark:text-neutral-50 group-hover:text-pink-600 dark:group-hover:text-pink-300 transition-colors",
                    ),
                  ],
                  [element.text("Hilbish")],
                ),
              ],
            ),
          ]),
          html.div(
            [attribute.class("flex items-center gap-5 sm:gap-6 text-sm")],
            [
              link(conf.base_url_join("/docs"), "Docs", False),
              link(conf.base_url_join("/blog"), "Blog", False),
              html.a(
                [
                  attribute.class(
                    "px-3 py-1.5 rounded-md bg-pink-600 hover:bg-pink-700 text-white transition-colors font-medium",
                  ),
                  attribute.href(conf.base_url_join("/donate")),
                ],
                [element.text("Donate")],
              ),
            ],
          ),
        ],
      ),
    ],
  )
}

fn hamburger_icon() -> element.Element(a) {
  element.unsafe_raw_html(
    "",
    "tag",
    [],
    "<svg xmlns=\"http://www.w3.org/2000/svg\" height=\"24px\" viewBox=\"0 -960 960 960\" width=\"24px\" fill=\"currentColor\"><path d=\"M120-240v-80h240v80H120Zm0-200v-80h480v80H120Zm0-200v-80h720v80H120Z\"/></svg>",
  )
}

pub fn footer() -> element.Element(a) {
  html.footer(
    [
      attribute.class("px-6 py-12"),
      attribute.class(theme.border_t),
      attribute.class(theme.surface_bg),
    ],
    [
      html.div(
        [
          attribute.class(
            "max-w-5xl mx-auto flex flex-col gap-10 sm:flex-row sm:items-start sm:justify-between",
          ),
        ],
        [
          html.div([attribute.class("flex flex-col gap-3")], [
            html.a(
              [
                attribute.href(conf.base_url),
                attribute.class("flex items-center gap-3 group w-fit"),
              ],
              [
                html.img([
                  attribute.src(conf.base_url_join("/hilbish-flower.png")),
                  attribute.class("h-10 w-10"),
                ]),
                html.span(
                  [
                    attribute.class("text-2xl font-bold"),
                    attribute.class(theme.text_default),
                  ],
                  [element.text("Hilbish")],
                ),
              ],
            ),
            html.span([attribute.class(theme.text_muted)], [
              element.text("The Moon-powered shell!"),
            ]),
            html.span(
              [attribute.class("text-sm"), attribute.class(theme.text_muted)],
              [element.text("MIT License © sammyette 2026")],
            ),
          ]),
          html.div([attribute.class("flex flex-col gap-2.5")], [
            html.span(
              [
                attribute.class(
                  "text-xs font-semibold uppercase tracking-wider",
                ),
                attribute.class(theme.text_muted),
              ],
              [element.text("Links")],
            ),
            footer_link(conf.base_url_join("/docs"), "Docs", False),
            footer_link(conf.base_url_join("/blog"), "Blog", False),
            footer_link(conf.base_url_join("/donate"), "Donate", False),
            footer_link("https://github.com/sammy-ette/Hilbish", "GitHub", True),
          ]),
        ],
      ),
    ],
  )
}

fn footer_link(url: String, text: String, out: Bool) -> element.Element(a) {
  html.a(
    [
      attribute.href(url),
      case out {
        False -> attribute.none()
        True -> attribute.target("_blank")
      },
      attribute.class(
        "inline-flex items-center gap-1 text-sm transition-colors",
      ),
      attribute.class(theme.accent_link),
    ],
    [
      element.text(text),
      case out {
        False -> element.none()
        True -> external_link_icon("size-3.5")
      },
    ],
  )
}
