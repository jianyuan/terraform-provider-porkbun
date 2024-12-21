resource "porkbun_domain_nameservers" "example" {
  domain = "jiancodes.com"

  nameservers = [
    "gabe.ns.cloudflare.com",
    "ivy.ns.cloudflare.com",
  ]
}
