import lustre/attribute
import lustre/element
import lustre/element/html
import lustre/element/svg
import util

pub fn page() {
  html.div(
    [
      attribute.class(
        "p-8 h-full flex-1 mx-auto flex flex-col justify-center items-center md:w-1/2",
      ),
    ],
    [
      svg.svg(
        [
          attribute.class("fill-pink-500 w-42 h-42"),
          attribute.attribute("viewBox", "0 0 256 256"),
          attribute.attribute("xmlns", "http://www.w3.org/2000/svg"),
        ],
        [
          svg.path([
            attribute.attribute(
              "d",
              "M240,98a57.63,57.63,0,0,1-17,41L133.7,229.62a8,8,0,0,1-11.4,0L33,139a58,58,0,0,1,82-82.1L128,69.05l13.09-12.19A58,58,0,0,1,240,98Z",
            ),
          ]),
        ],
      ),
      html.span([attribute.class("flex flex-row items-center justify-center")], [
        html.img([
          attribute.src("./hilbish-flower.png"),
          attribute.class("h-20"),
        ]),
        html.p([attribute.class("text-4xl font-bold")], [
          element.text("Hilbish"),
        ]),
      ]),
      html.p([attribute.class("flex flex-wrap flex-row items-center gap-1")], [
        element.text(" is developed in the free time of the developer,"),
        util.link("https://sammyette.party", " sammyette.", False),
        element.text(
          " Between working on other projects that interest me more, or playing games, or dealing with life and other things, it is hard to stay motivated putting work towards Hilbish specifically. If you like my work, want a feature worked on, or just want to help me out, you can send a couple dollars via ",
        ),
        util.link("https://ko-fi.com/sammyette", "Ko-fi", True),
      ]),
    ],
  )
}
