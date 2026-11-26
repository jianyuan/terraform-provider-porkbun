resource "porkbun_webhook" "example" {
  url    = "https://example.com/porkbun/webhook"
  events = ["domain.registered", "dns.*"]
}
