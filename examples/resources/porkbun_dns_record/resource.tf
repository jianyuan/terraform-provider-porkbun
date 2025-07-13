resource "porkbun_dns_record" "example" {
  domain = "jiancodes.com"

  type    = "A"
  content = "127.0.0.1"
}

resource "porkbun_dns_record" "example" {
  domain = "jiancodes.com"

  subdomain = "www"
  type      = "A"
  content   = "127.0.0.1"
}

resource "porkbun_dns_record" "example" {
  domain = "jiancodes.com"

  type    = "ALIAS"
  content = "pixie.porkbun.com"
}

resource "porkbun_dns_record" "example" {
  domain = "jiancodes.com"

  subdomain = "*"
  type      = "CNAME"
  content   = "pixie.porkbun.com"
}

resource "porkbun_dns_record" "example" {
  domain = "jiancodes.com"

  subdomain = "*"
  type      = "CNAME"
  content   = "pixie.porkbun.com"
}

import {
  to = porkbun_dns_record.example
  id  = "123456789_jiancodes.com_CNAME"
}