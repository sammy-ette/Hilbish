import gleam/io
import gleam/list
import gleam/option
import gleam/result
import gleam/string
import pages/blog
import pages/donate
import pages/page
import util

import glaml
import lustre/attribute
import lustre/element
import lustre/element/html
import lustre/ssg
import lustre/ssg/djot
import simplifile

import conf
import pages/doc
import pages/index
import post

pub fn main() {
  let assert Ok(files) = simplifile.get_files("./content")
  let posts =
    list.map(files, fn(path: String) {
      let assert Ok(ext) = path |> string.split(".") |> list.last
      let slug =
        path
        |> string.replace("./content", "")
        |> string.drop_end({ ext |> string.length() } + 1)
      let assert Ok(name) = slug |> string.split("/") |> list.last

      let slug = case name {
        "_index" -> slug |> string.drop_end({ "_index" |> string.length() } + 1)
        _ -> slug
      }

      let assert Ok(content) = simplifile.read(path)

      let metadata = case djot.frontmatter(content) {
        Ok(frontmatter) -> {
          let assert Ok([metadata]) = glaml.parse_string(frontmatter)
          metadata |> glaml.document_root
        }
        Error(_) -> glaml.NodeMap([])
      }
      let title = case metadata |> glaml.select_sugar("title") {
        Ok(glaml.NodeStr(title)) -> title
        _ -> ""
      }
      let description = case metadata |> glaml.select_sugar("description") {
        Ok(glaml.NodeStr(description)) -> description
        _ -> ""
      }

      let assert Ok(filename) = path |> string.split("/") |> list.last
      let content = djot.content(content)
      #(slug, post.Post(name, description, title, slug, metadata, content))
    })

  let doc_pages =
    list.filter(posts, fn(page) {
      let isdoc = is_doc_page(page.0)
      //io.debug(page.0)
      //io.debug(isdoc)
      isdoc
    })
    |> list.filter(fn(page) {
      case { page.1 }.metadata != glaml.NodeMap([]) {
        False -> {
          echo { page.1 }.slug <> " is missing metadata"
          False
        }
        True -> True
      }
    })
    |> list.sort(util.sort_weight)

  let build =
    ssg.new("./public")
    |> ssg.add_static_dir("static")
    |> ssg.add_static_route("/", create_page(index.page(), False))
    |> ssg.add_static_route("/donate", create_page(donate.page(), False))
    |> ssg.add_static_route(
      "/blog",
      create_page(
        blog.page(
          list.filter(posts, fn(page: #(String, post.Post)) {
            case { page.1 }.slug {
              "/blog" <> _ -> True
              _ -> False
            }
          }),
        ),
        False,
      ),
    )
    |> list.fold(posts, _, fn(config, post) {
      let page = case is_doc_page(post.0) {
        True -> doc.page(post.1, post.0, doc_pages)
        False -> page.page(post.1)
      }
      ssg.add_static_route(
        config,
        post.0,
        create_page(page, is_doc_page(post.0)),
      )
    })
    |> ssg.use_index_routes
    |> ssg.build

  case build {
    Ok(_) -> io.println("Website successfully built!")
    Error(e) -> {
      echo e
      io.println("Website could not be built.")
    }
  }
}

fn is_doc_page(slug: String) {
  case slug {
    "/docs" <> _ -> True
    _ -> False
  }
}

fn create_page(
  content: element.Element(a),
  doc_page: Bool,
) -> element.Element(a) {
  let description =
    "Something Unique. Hilbish is the new interactive shell for Lua fans. Extensible, scriptable, configurable: All in Lua."

  html.html(
    [
      attribute.class(
        "bg-stone-50 dark:bg-neutral-900 text-black dark:text-white",
      ),
    ],
    [
      html.head([], [
        html.meta([
          attribute.name("viewport"),
          attribute.attribute(
            "content",
            "width=device-width, initial-scale=1.0",
          ),
        ]),
        html.link([
          attribute.rel("stylesheet"),
          attribute.href(conf.base_url_join("/tailwind.css")),
        ]),
        html.title([], "Hilbish"),
        html.meta([attribute.name("theme-color"), attribute.content("#ff89dd")]),
        html.meta([
          attribute.content(conf.base_url_join("/hilbish-flower.png")),
          attribute.attribute("property", "og:image"),
        ]),
        html.meta([
          attribute.content("Hilbish"),
          attribute.attribute("property", "og:title"),
        ]),
        // this should be same as title
        html.meta([
          attribute.content("Hilbish"),
          attribute.attribute("property", "og:site_name"),
        ]),
        html.meta([
          attribute.content("website"),
          attribute.attribute("property", "og:type"),
        ]),
        html.meta([
          attribute.content(description),
          attribute.attribute("property", "og:description"),
        ]),
        html.meta([
          attribute.content(description),
          attribute.name("description"),
        ]),
        html.meta([
          attribute.name("keywords"),
          attribute.content("Lua,Shell,Hilbish,Linux,zsh,bash"),
        ]),
        html.meta([
          attribute.content(conf.base_url),
          attribute.attribute("property", "og:url"),
        ]),
        // disable dark reader
        html.meta([attribute.name("darkreader-lock")]),
      ]),
      html.body([attribute.class("flex flex-col min-h-screen")], [
        util.nav(),
        content,
        // case doc_page {
      //   True -> element.none()
      //   False -> util.footer()
      // },
      ]),
    ],
  )
}
