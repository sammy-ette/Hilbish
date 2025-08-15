//pub const base_url = "https://rosettea.github.io/Hilbish/versions/new-website"

pub const base_url = "http://localhost:9080"

pub fn base_url_join(cont: String) -> String {
  base_url <> cont
}
