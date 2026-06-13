import gleam/dict.{type Dict}
import gleam/list
import gleam/option.{type Option}
import gleam/string
import jot

import lustre/attribute
import lustre/element
import lustre/element/html
import lustre/ssg/djot

import theme
import util

/// Renders djot content (frontmatter already stripped) into styled Lustre
/// elements. Shared by the docs and blog post page layouts.
pub fn render_doc(md: String) -> List(element.Element(a)) {
  djot.render(md, renderer())
}

fn renderer() -> djot.Renderer(element.Element(a)) {
  djot.Renderer(
    ..djot.default_renderer(),
    paragraph: paragraph,
    heading: heading,
    code: code,
    codeblock: codeblock,
    blockquote: blockquote,
    bullet_list: bullet_list,
    link: link,
    image: image,
    div: callout,
  )
}

fn to_attr(attrs: Dict(String, String)) -> List(attribute.Attribute(a)) {
  use acc, key, val <- dict.fold(attrs, [])
  [attribute.attribute(key, val), ..acc]
}

fn paragraph(
  attrs: Dict(String, String),
  content: List(element.Element(a)),
) -> element.Element(a) {
  html.p(
    [
      attribute.class(
        "my-5 leading-relaxed text-neutral-700 dark:text-neutral-200",
      ),
      ..to_attr(attrs)
    ],
    content,
  )
}

fn heading(
  attrs: Dict(String, String),
  level: Int,
  content: List(element.Element(a)),
) -> element.Element(a) {
  let size = case level {
    1 -> "text-5xl"
    2 -> "text-3xl"
    3 -> "text-2xl"
    _ -> "text-lg"
  }

  let margin = case level {
    1 -> "mt-10 mb-6"
    2 -> "mt-8 mb-4 pb-2 border-b border-pink-400/40"
    3 -> "mt-6 mb-3"
    _ -> "mt-5 mb-2"
  }

  let weight = case level {
    1 -> "font-bold"
    2 -> "font-semibold"
    3 -> "font-semibold"
    _ -> "font-semibold"
  }

  let class =
    margin
    <> " group "
    <> size
    <> " "
    <> weight
    <> " "
    <> theme.text_heading
    <> case level {
      1 -> " font-[Momo_Signature,cursive]"
      2 -> " font-[Momo_Signature,cursive]"
      _ -> ""
    }

  let content = case level > 1, dict.get(attrs, "id") {
    True, Ok(id) -> list.append(content, [anchor_link(id)])
    _, _ -> content |> list.prepend(element.text("❤ "))
  }

  let attr = to_attr(dict.insert(attrs, "class", class))

  case level {
    1 -> html.h1(attr, content)
    2 -> html.h2(attr, content)
    3 -> html.h3(attr, content)
    4 -> html.h4(attr, content)
    5 -> html.h5(attr, content)
    6 -> html.h6(attr, content)
    _ -> html.p(attr, content)
  }
}

fn anchor_link(id: String) -> element.Element(a) {
  html.a(
    [
      attribute.href("#" <> id),
      attribute.class(
        "ml-2 text-base font-normal "
        <> theme.text_muted
        <> " no-underline opacity-0 group-hover:opacity-100 hover:text-pink-400 transition-opacity",
      ),
      attribute.attribute("aria-label", "Link to this section"),
    ],
    [element.text("#")],
  )
}

fn code(content: String) -> element.Element(a) {
  html.code(
    [
      attribute.class(
        "px-2 py-1 rounded text-pink-300 bg-pink-500/20 font-mono text-[0.9em] font-medium",
      ),
    ],
    [element.text(content)],
  )
}

fn codeblock(
  _attrs: Dict(String, String),
  _lang: Option(String),
  code_text: String,
) -> element.Element(a) {
  html.pre(
    [
      attribute.class(
        "font-mono text-sm bg-neutral-900 dark:bg-neutral-900 border border-neutral-800 dark:border-pink-500/20 rounded-lg p-4 overflow-x-auto my-6 shadow-lg",
      ),
    ],
    [
      html.code([attribute.class("text-neutral-100")], [element.text(code_text)]),
    ],
  )
}

fn blockquote(
  _attrs: Dict(String, String),
  content: List(element.Element(a)),
) -> element.Element(a) {
  html.blockquote(
    [
      attribute.class(
        "border-l-4 border-pink-400 pl-5 py-3 my-5 italic text-neutral-300 dark:text-neutral-300 bg-pink-500/10 rounded-r-lg",
      ),
    ],
    content,
  )
}

fn bullet_list(
  layout: jot.ListLayout,
  _style: String,
  items: List(List(element.Element(a))),
) -> element.Element(a) {
  html.ul(
    [attribute.class("list-disc list-outside pl-6 my-5 space-y-2")],
    list.map(items, fn(item) {
      case layout {
        jot.Tight -> html.li([attribute.class(theme.text_default)], item)
        jot.Loose ->
          html.li([attribute.class(theme.text_default)], [
            html.p([attribute.class("my-1")], item),
          ])
      }
    }),
  )
}

fn link(
  destination: Option(String),
  attrs: Dict(String, String),
  content: List(element.Element(a)),
) -> element.Element(a) {
  let base_class =
    attribute.class(
      "text-pink-400 dark:text-pink-300 hover:text-pink-300 dark:hover:text-pink-200 font-semibold underline underline-offset-2 hover:underline-offset-4 transition-all hover:drop-shadow-[0_0_8px_rgba(236,72,153,0.5)]",
    )

  case destination {
    option.None -> html.span([base_class, ..to_attr(attrs)], content)
    option.Some(url) -> {
      let is_external =
        string.starts_with(url, "http://")
        || string.starts_with(url, "https://")

      case is_external {
        True ->
          html.a(
            [
              attribute.href(url),
              attribute.target("_blank"),
              base_class,
              ..to_attr(attrs)
            ],
            list.append(content, [
              util.external_link_icon(
                "size-4 inline ml-1 align-text-bottom opacity-70 group-hover:opacity-100",
              ),
            ]),
          )
        False ->
          html.a([attribute.href(url), base_class, ..to_attr(attrs)], content)
      }
    }
  }
}

fn image(
  destination: Option(String),
  attrs: Dict(String, String),
  alt: String,
) -> element.Element(a) {
  case destination {
    option.None -> html.span(to_attr(attrs), [element.text(alt)])
    option.Some(url) ->
      html.img([
        attribute.src(url),
        attribute.alt(alt),
        attribute.class(
          "rounded-lg max-w-full my-5 shadow-lg ring-1 ring-pink-500/20 hover:ring-pink-500/40 transition-all",
        ),
        ..to_attr(attrs)
      ])
  }
}

/// Renders `:::warning`, `:::note`, `:::tip`, `:::important` and `:::danger`
/// fenced divs as callout boxes. Any other div is passed through unstyled.
fn callout(
  attrs: Dict(String, String),
  content: List(element.Element(a)),
) -> element.Element(a) {
  case dict.get(attrs, "class") {
    Ok("warning") ->
      callout_box(
        "Warning",
        "border-yellow-500 bg-yellow-500/10 text-yellow-50",
        content,
      )
    Ok("note") ->
      callout_box("Note", "border-sky-400 bg-sky-500/10 text-sky-50", content)
    Ok("tip") ->
      callout_box(
        "Tip",
        "border-emerald-400 bg-emerald-500/10 text-emerald-50",
        content,
      )
    Ok("important") ->
      callout_box(
        "Important",
        "border-pink-400 bg-pink-500/15 text-pink-50",
        content,
      )
    Ok("danger") ->
      callout_box("Danger", "border-red-500 bg-red-500/10 text-red-50", content)
    _ -> html.div(to_attr(attrs), content)
  }
}

fn callout_box(
  label: String,
  classes: String,
  content: List(element.Element(a)),
) -> element.Element(a) {
  html.div(
    [
      attribute.class(
        "border-l-4 rounded-lg px-4 py-3 my-4 [&>p]:my-0 " <> classes,
      ),
    ],
    [
      html.p([attribute.class("font-semibold mb-2 text-sm")], [
        element.text(label),
      ]),
      ..content
    ],
  )
}
