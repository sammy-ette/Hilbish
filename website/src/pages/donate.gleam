import lustre/attribute
import lustre/element
import lustre/element/html
import lustre/element/svg
import util

pub fn page() {
  html.main([attribute.class("flex-1")], [
    html.section(
      [
        attribute.class("border-b border-b-pink-500/20 px-6 py-20"),
      ],
      [
        html.div(
          [
            attribute.class(
              "mx-auto grid max-w-5xl items-center gap-12 md:grid-cols-2",
            ),
          ],
          [
            html.div([], [
              html.h1(
                [
                  attribute.class(
                    "text-5xl sm:text-6xl font-bold mb-6 text-neutral-900 dark:text-neutral-50",
                  ),
                ],
                [
                  element.text("Help keep "),
                  util.logo_image(
                    "inline-block h-12 w-auto align-middle sm:h-14",
                  ),
                  element.text(" moving."),
                ],
              ),
              html.p(
                [
                  attribute.class(
                    "text-xl text-neutral-600 dark:text-neutral-300 mb-8 leading-relaxed",
                  ),
                ],
                [
                  element.text(
                    "Hilbish is built in my spare time. If it makes your terminal a little more comfortable, a small contribution helps make room for more features, fixes, and experiments.",
                  ),
                ],
              ),
              html.div([attribute.class("flex flex-col gap-3 sm:flex-row")], [
                html.a(
                  [
                    attribute.href("https://ko-fi.com/sammyette"),
                    attribute.target("_blank"),
                    attribute.rel("noopener noreferrer"),
                    attribute.class(
                      "inline-flex items-center justify-center gap-2 rounded-lg bg-pink-600 px-5 py-3 font-semibold text-white transition-colors hover:bg-pink-700",
                    ),
                  ],
                  [
                    element.text("Support on Ko-fi"),
                    util.external_link_icon("size-4"),
                  ],
                ),
                html.a(
                  [
                    attribute.href("https://github.com/sammy-ette/Hilbish"),
                    attribute.target("_blank"),
                    attribute.rel("noopener noreferrer"),
                    attribute.class(
                      "inline-flex items-center justify-center gap-2 rounded-lg border border-pink-500/50 px-5 py-3 font-semibold text-pink-700 transition-colors hover:border-pink-500 hover:bg-pink-50 dark:text-pink-200 dark:hover:bg-pink-950/40",
                    ),
                  ],
                  [
                    element.text("Visit GitHub"),
                    util.external_link_icon("size-4"),
                  ],
                ),
              ]),
            ]),
            html.div(
              [
                attribute.class(
                  "mx-auto flex size-64 items-center justify-center rounded-full border border-pink-300/30 bg-neutral-50 dark:border-pink-500/30 dark:bg-neutral-800 sm:size-72",
                ),
              ],
              [
                svg.svg(
                  [
                    attribute.class("size-40 fill-pink-500 sm:size-48"),
                    attribute.attribute("viewBox", "0 0 256 256"),
                    attribute.attribute("xmlns", "http://www.w3.org/2000/svg"),
                    attribute.attribute("aria-hidden", "true"),
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
              ],
            ),
          ],
        ),
      ],
    ),
    html.section(
      [
        attribute.class(
          "py-20 px-6 border-b border-b-pink-500/20 bg-white dark:bg-neutral-900",
        ),
      ],
      [
        html.div([attribute.class("max-w-5xl mx-auto")], [
          html.h2(
            [
              attribute.class(
                "text-4xl font-bold mb-12 text-neutral-900 dark:text-neutral-50",
              ),
            ],
            [element.text("What your support makes possible")],
          ),
          html.div([attribute.class("grid md:grid-cols-3 gap-8")], [
            support_point(
              "A little more motivation",
              "Donations are a good reminder that people care about Hilbish, which makes it easier to keep coming back to it.",
            ),
            support_point(
              "Your feature gets a closer look",
              "If there is something you would like to see in Hilbish, supporting the project gives that idea more weight when I choose what to work on next.",
            ),
            support_point(
              "Every bit helps",
              "There is no set amount. Even a small contribution is appreciated and keeps the project going.",
            ),
          ]),
        ]),
      ],
    ),
  ])
}

fn support_point(title: String, description: String) -> element.Element(a) {
  html.div(
    [
      attribute.class(
        "p-6 border border-pink-300/30 dark:border-pink-500/30 rounded-lg hover:border-pink-400/50 dark:hover:border-pink-500/50 transition-colors bg-neutral-50 dark:bg-neutral-800",
      ),
    ],
    [
      html.h3(
        [
          attribute.class(
            "text-xl font-semibold text-pink-700 dark:text-pink-300 mb-3",
          ),
        ],
        [element.text(title)],
      ),
      html.p(
        [
          attribute.class(
            "text-neutral-600 dark:text-neutral-400 leading-relaxed",
          ),
        ],
        [element.text(description)],
      ),
    ],
  )
}
