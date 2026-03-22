import glaml
import gleam/option

pub type Post {
  Post(
    name: String,
    description: String,
    title: String,
    slug: String,
    metadata: glaml.Node,
    contents: String,
  )
}
