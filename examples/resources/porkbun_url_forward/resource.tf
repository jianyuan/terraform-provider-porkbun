resource "porkbun_url_forward" "root" {
  domain       = "jian.fyi"
  type         = "temporary"
  wildcard     = false
  include_path = false
  location     = "https://www.jiancodes.com"
}

resource "porkbun_url_forward" "wildcard" {
  domain       = "jian.fyi"
  type         = "temporary"
  wildcard     = true
  include_path = false
  location     = "https://www.jiancodes.com"
}

