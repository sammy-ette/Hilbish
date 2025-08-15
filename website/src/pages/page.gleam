import gleam/dict
import lustre/attribute
import lustre/element
import lustre/element/html
import lustre/ssg/djot
import post

pub fn page(p: post.Post) {
  html.div([attribute.class("mt-4 mx-auto md:w-1/2")], [
    html.h1([attribute.class("text-5xl mb-2")], [element.text(p.title)]),
    html.hr([attribute.class("my-4 text-pink-600")]),
    ..render_doc(p.contents)
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
