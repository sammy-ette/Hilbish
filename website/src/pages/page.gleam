import lustre/attribute
import lustre/element
import lustre/element/html
import post
import render

pub fn page(p: post.Post) {
  html.div([attribute.class("mt-4 mx-auto md:w-1/2")], [
    html.h1([attribute.class("text-5xl mb-2")], [element.text(p.title)]),
    html.hr([attribute.class("my-4 text-pink-600")]),
    ..render.render_doc(p.contents)
  ])
}
