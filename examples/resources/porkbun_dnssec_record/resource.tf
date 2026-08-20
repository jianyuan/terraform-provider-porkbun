resource "porkbun_dnssec_record" "example" {
  domain = "jiancodes.com"

  alg         = "13"
  digest      = "18C7F73D40B152446315468F12E61A6F5D4DF3DC7EB899BAF7F56F6BBDB33C72"
  digest_type = "2"
  key_tag     = "35175"
}

import {
  to = porkbun_dnssec_record.example
  id = "jiancodes.com_35175"
}
