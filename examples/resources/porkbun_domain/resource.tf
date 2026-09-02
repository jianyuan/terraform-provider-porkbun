resource "porkbun_domain" "example" {
  domain = "jiancodes.com"

  # Refuse the registration if it is quoted above this, in US cents.
  max_cost = 1200
}
