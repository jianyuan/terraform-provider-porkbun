---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "porkbun Provider"
description: |-
  The Porkbun provider is used to interact with the Porkbun service.
  If you find this provider useful, please consider supporting me through GitHub Sponsorship or Ko-Fi to help with its development.
  Github-sponsors https://github.com/sponsors/jianyuan
  Ko-Fi https://ko-fi.com/L3L71DQEL
---

# porkbun Provider

The Porkbun provider is used to interact with the Porkbun service.

If you find this provider useful, please consider supporting me through GitHub Sponsorship or Ko-Fi to help with its development.

[![Github-sponsors](https://img.shields.io/badge/sponsor-30363D?style=for-the-badge&logo=GitHub-Sponsors&logoColor=#EA4AAA)](https://github.com/sponsors/jianyuan)
[![Ko-Fi](https://img.shields.io/badge/Ko--fi-F16061?style=for-the-badge&logo=ko-fi&logoColor=white)](https://ko-fi.com/L3L71DQEL)

## Example Usage

```terraform
provider "porkbun" {
  api_key    = "pk1_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  secret_key = "sk1_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String, Sensitive) The API key for the Porkbun account. It can be sourced from the `PORKBUN_API_KEY` environment variable.
- `base_url` (String) The base URL for the Porkbun API. Defaults to `https://api.porkbun.com/api/json`. It can be sourced from the `PORKBUN_BASE_URL` environment variable.
- `secret_key` (String, Sensitive) The secret API key for the Porkbun account. It can be sourced from the `PORKBUN_SECRET_KEY` environment variable.
