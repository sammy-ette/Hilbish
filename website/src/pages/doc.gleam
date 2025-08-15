import gleam/dict
import gleam/io
import gleam/list
import gleam/option
import gleam/order
import gleam/result
import gleam/string

import glaml
import lustre/attribute
import lustre/element
import lustre/element/html
import lustre/ssg/djot

import conf
import jot
import post
import util

pub fn page(
  p: post.Post,
  this_slug: String,
  doc_pages_list,
) -> element.Element(a) {
  html.div([attribute.class("flex-1 flex flex-col overflow-hidden")], [
    html.div(
      [
        attribute.class(
          "sm:hidden h-10 flex py-2 px-4 border-b border-b-zinc-300 w-full gap-2 backdrop-blur-sm bg-zinc-300/50 dark:bg-zinc-800/50",
        ),
      ],
      [
        html.label(
          [attribute.for("sidebar-toggle"), attribute.class("cursor-pointer")],
          [
            element.unsafe_raw_html(
              "",
              "tag",
              [],
              "<svg xmlns=\"http://www.w3.org/2000/svg\" height=\"24px\" viewBox=\"0 -960 960 960\" width=\"24px\" class=\"fill-white\"><path d=\"M120-240v-80h240v80H120Zm0-200v-80h480v80H120Zm0-200v-80h720v80H120Z\"/></svg>",
            ),
          ],
        ),
        html.span([attribute.class("font-bold")], [element.text(p.title)]),
      ],
    ),
    html.div([attribute.class("flex-1 sm:flex grid overflow-hidden")], [
      html.input([
        attribute.type_("checkbox"),
        attribute.id("sidebar-toggle"),
        attribute.class("peer hidden"),
      ]),
      html.div(
        [
          attribute.class(
            "overflow-y-scroll p-4 sm:border-r sm:border-r-zinc-300 col-start-1 row-start-1 bg-neutral-100 dark:bg-neutral-950 basis-2/10 transition-transform duration-300 -translate-x-full peer-checked:translate-x-0 sm:translate-x-0 z-30",
          ),
        ],
        [
          html.ul(
            [attribute.class("text-lg flex flex-col gap-2")],
            list.flatten(
              list.group(doc_pages_list, fn(post: #(String, post.Post)) {
                case glaml.select_sugar({ post.1 }.metadata, "menu") {
                  Ok(glaml.NodeMap(menu)) -> {
                    let assert Ok(menu_first) = list.first(menu)
                    let assert Ok(glaml.NodeStr(parent)) =
                      glaml.select_sugar(menu_first.1, "parent")
                    parent
                  }
                  Ok(glaml.NodeStr(_)) -> {
                    // If it is a sring, it's just saying to be grouped
                    // in the menu.
                    // So use the title instead, because titles are unique?
                    { post.1 }.title
                  }
                  Ok(_) -> panic as "wrong type fool"
                  Error(e) -> {
                    echo { post.1 }.slug
                    let assert Ok(title) =
                      { post.1 }.slug |> string.split("/") |> list.last
                    title |> string.capitalise
                  }
                }
              })
              |> dict.to_list()
              |> list.sort(fn(group1, group2) {
                let assert Ok(group_1_parent_post) =
                  list.filter(doc_pages_list, fn(p) {
                    { p.1 }.title == group1.0
                  })
                  |> list.first()
                let assert Ok(group_2_parent_post) =
                  list.filter(doc_pages_list, fn(p) {
                    { p.1 }.title == group2.0
                  })
                  |> list.first()

                let sort_weight_reverse = order.reverse(util.sort_weight)
                sort_weight_reverse(group_1_parent_post, group_2_parent_post)
              })
              |> list.map(fn(group: #(String, List(#(String, post.Post)))) {
                let assert Ok(parent_post) =
                  list.filter(doc_pages_list, fn(p: #(String, post.Post)) {
                    { p.1 }.title == group.0
                  })
                  |> list.first()
                [
                  html.li(
                    [
                      attribute.class(
                        "font-bold"
                        <> case this_slug == { parent_post.1 }.slug {
                          False -> {
                            ""
                          }
                          True -> " text-pink-400"
                        },
                      ),
                    ],
                    [
                      html.a(
                        [
                          attribute.href(conf.base_url_join(
                            { parent_post.1 }.slug,
                          )),
                        ],
                        [
                          element.text(
                            case this_slug == { parent_post.1 }.slug {
                              False -> ""
                              True -> " -> "
                            }
                            <> { parent_post.1 }.title,
                          ),
                        ],
                      ),
                    ],
                  ),
                  case list.length(group.1) {
                    1 -> element.none()
                    _ ->
                      html.ul(
                        [attribute.class("pl-4")],
                        list.sort(group.1, util.sort_weight)
                          |> list.filter(fn(p1) {
                            { p1.1 }.title != { parent_post.1 }.title
                          })
                          |> list.map(fn(post: #(String, post.Post)) {
                            html.li(
                              [
                                attribute.class(
                                  "mb-2"
                                  <> case this_slug == { post.1 }.slug {
                                    False -> {
                                      ""
                                    }
                                    True -> " text-pink-400"
                                  },
                                ),
                              ],
                              [
                                html.a(
                                  [attribute.href(conf.base_url_join(post.0))],
                                  [
                                    element.text(
                                      case this_slug == { post.1 }.slug {
                                        False -> ""
                                        True -> " -> "
                                      }
                                      <> { post.1 }.title,
                                    ),
                                  ],
                                ),
                              ],
                            )
                          }),
                      )
                  },
                ]
              }),
            ),
          ),
        ],
      ),
      html.main(
        [
          attribute.class(
            "flex-1 flex justify-center basis-7/7 col-start-1 row-start-1 transition-all duration-300 peer-checked:filter peer-checked:blur-sm peer-checked:bg-black/30",
          ),
        ],
        [
          html.div([attribute.class("flex-1 flex flex-col overflow-y-auto")], [
            // todo: add date of publishing
            //html.time([], [])
            //html.small([], [element.text({{p.contents |> string.split(" ") |> list.length} / 200} |> int.to_string <> " min read")]),
            //element.unsafe_raw_html("namespace", "Tag", [], render_doc(p.contents))
            html.div([attribute.class("flex-1 w-3/4 self-center p-8")], [
              html.h1([attribute.class("my-3 font-bold text-4xl")], [
                element.text(p.title),
              ]),
              html.i([], [element.text(p.description)]),
              ..render_doc(p.contents)
            ]),
            util.footer(),
          ]),
        ],
      ),
    ]),
  ])
}

fn render_doc(md: String) {
  let renderer =
    djot.Renderer(
      ..djot.default_renderer(),
      heading: fn(attrs, level, content) {
        let size = case level {
          1 -> "text-4xl"
          2 -> "text-3xl"
          3 -> "text-2xl"
          _ -> "text-xl"
        }

        let margin = case level {
          1 -> "my-4"
          2 -> "my-2"
          _ -> "my-1"
        }

        let attr =
          dict.insert(
            attrs,
            "class",
            margin
              <> " text-neutral-800 dark:text-neutral-300 font-bold "
              <> size,
          )

        case level {
          1 -> html.h1(to_attr(attr), content)
          2 -> html.h2(to_attr(attr), content)
          3 -> html.h3(to_attr(attr), content)
          4 -> html.h4(to_attr(attr), content)
          5 -> html.h5(to_attr(attr), content)
          6 -> html.h6(to_attr(attr), content)
          _ -> html.p(to_attr(attr), content)
        }
      },
      code: fn(content) {
        html.code([attribute.class("text-violet-600 dark:text-violet-400")], [
          element.text(content),
        ])
      },
    )
  djot.render(md, renderer)
}

fn to_attr(attrs) {
  use attrs, key, val <- dict.fold(attrs, [])
  [attribute.attribute(key, val), ..attrs]
}
