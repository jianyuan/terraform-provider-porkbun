---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "porkbun_dns_record Data Source - terraform-provider-porkbun"
subcategory: ""
description: |-
  Retrieve a single record for a particular record ID.
---

# porkbun_dns_record (Data Source)

Retrieve a single record for a particular record ID.

## Example Usage

```terraform
data "porkbun_dns_record" "test" {
  domain = "jiancodes.com"
  id     = "123456"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `domain` (String) The domain name.
- `id` (String) The record ID.

### Read-Only

- `content` (String)
- `name` (String)
- `notes` (String)
- `priority` (Number)
- `ttl` (Number)
- `type` (String)