resource "porkbun_email_forward" "hello" {
  domain     = "example.com"
  address    = "hello" # the part before the @
  forward_to = "me@example.org"
}

# Allow Porkbun to delete other apex MX records and rewrite an existing SPF
# record when setting up forwarding.
resource "porkbun_email_forward" "info" {
  domain              = "example.com"
  address             = "info"
  forward_to          = "me@example.org"
  confirm_dns_changes = true
}
