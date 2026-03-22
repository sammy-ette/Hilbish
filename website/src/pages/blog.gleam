import conf
import gleam/list
import lustre/attribute
import lustre/element
import lustre/element/html
import lustre/element/svg
import post
import util

pub fn page(blog_list: List(#(String, post.Post))) {
  html.div(
    [
      attribute.class(
        "p-8 h-full flex-1 mx-auto flex flex-col justify-center items-center",
      ),
    ],
    list.map(blog_list, fn(post) {
      html.a(
        [
          attribute.class("group"),
          attribute.href(conf.base_url_join({ post.1 }.slug)),
        ],
        [
          html.div([], [
            html.h2(
              [
                attribute.class(
                  "group-hover:underline dark:group-hover:text-pink-300 group-hover:text-pink-600",
                ),
              ],
              [element.text({ post.1 }.title)],
            ),
          ]),
        ],
      )
    }),
  )
}
