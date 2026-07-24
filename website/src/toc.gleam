import gleam/dict
import gleam/list
import jot
import lustre/attribute
import lustre/element
import lustre/element/html
import theme

/// A single entry in the table of contents.
pub type TocEntry {
  TocEntry(level: Int, text: String, id: String)
}

/// Recursively extracts plain text from a list of jot Inline nodes.
fn inline_to_text(inlines: List(jot.Inline)) -> String {
  list.fold(inlines, "", fn(acc, inline) {
    case inline {
      jot.Text(t) -> acc <> t
      jot.Code(t) -> acc <> t
      jot.Strong(content) -> acc <> inline_to_text(content)
      jot.Emphasis(content) -> acc <> inline_to_text(content)
      jot.Link(_, content, _) -> acc <> inline_to_text(content)
      jot.Span(_, content) -> acc <> inline_to_text(content)
      _ -> acc
    }
  })
}

/// Parses a Djot string and extracts all h2–h3 headings as TocEntry values.
/// h4+ (individual function blocks) are excluded — too numerous on long API pages.
pub fn extract(djot: String) -> List(TocEntry) {
  let doc = jot.parse(djot)
  list.filter_map(doc.content, fn(container) {
    case container {
      jot.Heading(attrs, level, content) if level >= 2 && level <= 3 ->
        case dict.get(attrs, "id") {
          Ok(id) -> Ok(TocEntry(level, inline_to_text(content), id))
          Error(_) -> Error(Nil)
        }
      _ -> Error(Nil)
    }
  })
}

/// Renders an "On this page" table of contents box from TocEntry values.
/// Returns element.none() when there are fewer than 3 entries.
pub fn render(entries: List(TocEntry)) -> element.Element(a) {
  case list.length(entries) < 3 {
    True -> element.none()
    False ->
      html.nav(
        [
          attribute.class(
            "mt-6 mb-8 rounded-lg border border-neutral-800 bg-neutral-900/40 px-5 py-4 text-sm",
          ),
          attribute.attribute("aria-label", "On this page"),
        ],
        [
          html.p(
            [
              attribute.class(
                "text-xs font-semibold uppercase tracking-wider mb-3 "
                <> theme.text_muted,
              ),
            ],
            [element.text("On this page")],
          ),
          html.ul(
            [attribute.class("space-y-1.5")],
            list.map(entries, fn(entry) {
              let indent = case entry.level {
                2 -> ""
                _ -> "ml-4 "
              }
              html.li(
                [],
                [
                  html.a(
                    [
                      attribute.href("#" <> entry.id),
                      attribute.class(
                        indent
                        <> "block truncate text-neutral-400 hover:text-pink-300 transition-colors",
                      ),
                    ],
                    [element.text(entry.text)],
                  ),
                ],
              )
            }),
          ),
        ],
      )
  }
}

/// Convenience function: parse and render in one call.
pub fn toc(djot: String) -> element.Element(a) {
  djot |> extract |> render
}
