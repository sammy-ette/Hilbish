import gleam/int
import gleam/option
import gleam/string

import lustre/attribute
import lustre/element
import lustre/element/html

import conf
import glaml
import post

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
          attribute.class(
            "inline-flex text-light dark:text-pink-300 dark:hover:text-pink-200 text-pink-600 hover:text-pink-300 hover:underline",
          ),
        ],
        [
          case out {
            False -> element.none()
            True ->
              element.unsafe_raw_html(
                "",
                "tag",
                [],
                "<svg xmlns=\"http://www.w3.org/2000/svg\" fill=\"none\" viewBox=\"0 0 24 24\" stroke-width=\"1.5\" stroke=\"currentColor\" class=\"size-6\">
  <path stroke-linecap=\"round\" stroke-linejoin=\"round\" d=\"M13.5 6H5.25A2.25 2.25 0 0 0 3 8.25v10.5A2.25 2.25 0 0 0 5.25 21h10.5A2.25 2.25 0 0 0 18 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25\" />
</svg>
",
              )
          },
          element.text(text),
        ],
      ),
    ],
  )
}

pub fn nav() -> element.Element(a) {
  html.nav(
    [
      attribute.class(
        "z-100 sticky top-0 bg-stone-100/80 dark:bg-neutral-950/80 flex justify-around items-center w-full border-b border-b-zinc-300 backdrop-blur-md h-12",
      ),
    ],
    [
      html.div([attribute.class("flex my-auto px-2")], [
        html.div([], [
          html.a(
            [
              attribute.href(conf.base_url_join("/")),
              attribute.class("flex items-center gap-1"),
            ],
            [
              html.img([
                attribute.src(conf.base_url_join("/hilbish-flower.png")),
                attribute.class("h-8"),
              ]),
              html.span([attribute.class("self-center text-3xl font-medium")], [
                element.text("Hilbish"),
              ]),
            ],
          ),
        ]),
      ]),
      html.div([attribute.class("flex gap-3")], [
        link(conf.base_url_join("/docs"), "Docs", False),
        link(conf.base_url_join("/blog"), "Blog", False),
        html.a(
          [
            attribute.class("bg-pink-500/50 hover:bg-pink-400  px-2 rounded-sm"),
            attribute.href(conf.base_url_join("/donate")),
          ],
          [element.text("Donate")],
        ),
      ]),
    ],
  )
}

pub fn footer() -> element.Element(a) {
  html.footer(
    [
      attribute.class(
        "z-100 py-4 px-6 flex flex-row justify-around border-t border-t-zinc-300",
      ),
    ],
    [
      html.div([attribute.class("flex flex-col")], [
        html.a(
          [
            attribute.href(conf.base_url),
            attribute.class("flex items-center gap-1"),
          ],
          [
            html.img([
              attribute.src(conf.base_url_join("/hilbish-flower.png")),
              attribute.class("h-24"),
            ]),
            html.span([attribute.class("self-center text-6xl")], [
              element.text("Hilbish"),
            ]),
          ],
        ),
        html.span([attribute.class("text-xl")], [
          element.text("The Moon-powered shell!"),
        ]),
        html.span([attribute.class("text-light text-neutral-500")], [
          element.text("MIT License, copyright sammyette 2025"),
        ]),
      ]),
      html.div([attribute.class("flex flex-col")], [
        link("https://github.com/sammy-ette/Hilbish", "GitHub", True),
      ]),
    ],
  )
}
