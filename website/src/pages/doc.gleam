import gleam/dict
import gleam/list
import gleam/order
import gleam/string

import glaml
import lustre/attribute
import lustre/element
import lustre/element/html

import conf
import post
import render
import theme
import util

pub fn page(
  p: post.Post,
  this_slug: String,
  doc_pages_list,
) -> element.Element(a) {
  html.div(
    [
      attribute.class("flex flex-1 min-h-0 overflow-hidden"),
      attribute.class(theme.page_bg),
    ],
    [
      html.input([
        attribute.type_("checkbox"),
        attribute.id("sidebar-toggle"),
        attribute.class("peer hidden"),
      ]),
      // Dimmed backdrop on mobile; tapping it closes the sidebar.
      html.label(
        [
          attribute.for("sidebar-toggle"),
          attribute.class(
            "fixed inset-0 top-16 sm:hidden bg-black/50 opacity-0 pointer-events-none peer-checked:opacity-100 peer-checked:pointer-events-auto transition-opacity z-30",
          ),
        ],
        [],
      ),
      html.div(
        [
          attribute.class("w-64 flex-shrink-0 flex flex-col overflow-hidden"),
          attribute.class(
            "bg-neutral-100 dark:bg-neutral-950 border-r border-r-pink-300/30 dark:border-r-pink-500/20",
          ),
          attribute.class(
            "fixed left-0 top-16 bottom-0 sm:static sm:top-0 z-40 sm:z-auto",
          ),
          attribute.class(
            "transition-transform duration-300 -translate-x-full sm:translate-x-0 peer-checked:translate-x-0",
          ),
        ],
        [
          html.ul(
            [
              attribute.class(
                "flex-1 overflow-y-auto scrollbar-pink px-3 py-4 text-sm flex flex-col gap-1",
              ),
            ],
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
                    // If it is a string, it's just saying to be grouped
                    // in the menu.
                    // So use the title instead, because titles are unique?
                    { post.1 }.title
                  }
                  Ok(_) -> panic as "wrong type fool"
                  Error(_) -> {
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
                  html.li([attribute.class("pt-2 first:pt-0")], [
                    html.a(
                      [
                        attribute.href(conf.base_url_join(
                          { parent_post.1 }.slug,
                        )),
                        attribute.class(
                          "block px-3 py-2 rounded-md font-medium transition-colors",
                        ),
                        attribute.class(
                          case this_slug == { parent_post.1 }.slug {
                            False ->
                              "text-neutral-600 dark:text-neutral-300 hover:text-pink-600 dark:hover:text-pink-300 hover:bg-pink-500/10"
                            True ->
                              "text-pink-600 dark:text-pink-300 bg-pink-500/15"
                          },
                        ),
                      ],
                      [element.text({ parent_post.1 }.title)],
                    ),
                  ]),
                  case list.length(group.1) {
                    1 -> element.none()
                    _ ->
                      html.ul(
                        [
                          attribute.class(
                            "ml-3 mt-1 border-l border-l-pink-300/40 dark:border-l-pink-500/30 pl-3 space-y-0.5",
                          ),
                        ],
                        list.sort(group.1, util.sort_weight)
                          |> list.filter(fn(p1) {
                            { p1.1 }.title != { parent_post.1 }.title
                          })
                          |> list.map(fn(post: #(String, post.Post)) {
                            html.li([], [
                              html.a(
                                [
                                  attribute.href(conf.base_url_join(post.0)),
                                  attribute.class(
                                    "block px-3 py-1.5 rounded-md text-sm transition-colors",
                                  ),
                                  attribute.class(
                                    case this_slug == { post.1 }.slug {
                                      False ->
                                        "text-neutral-500 dark:text-neutral-400 hover:text-pink-600 dark:hover:text-pink-300 hover:bg-pink-500/10"
                                      True ->
                                        "text-pink-600 dark:text-pink-300 bg-pink-500/15"
                                    },
                                  ),
                                ],
                                [element.text({ post.1 }.title)],
                              ),
                            ])
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
            "flex-1 min-w-0 overflow-y-auto scrollbar-pink flex flex-col",
          ),
        ],
        [
          html.div(
            [
              attribute.class(
                "w-full flex-1 max-w-4xl mx-auto px-6 sm:px-8 py-8 sm:py-12",
              ),
            ],
            [
              html.h1(
                [
                  attribute.class("font-bold text-5xl mb-3"),
                  attribute.class(theme.text_heading),
                  attribute.class(
                    "bg-gradient-to-r from-pink-500 via-pink-400 to-pink-500 bg-clip-text text-transparent",
                  ),
                ],
                [element.text(p.title)],
              ),
              html.p(
                [
                  attribute.class("text-lg mb-8"),
                  attribute.class(theme.text_muted),
                ],
                [element.text(p.description)],
              ),
              html.div(
                [attribute.class("max-w-none")],
                render.render_doc(p.contents),
              ),
            ],
          ),
          html.div([attribute.class("mt-16")], [util.footer()]),
        ],
      ),
    ],
  )
}
