---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ghostwriter_server_role Data Source - ghostwriter"
subcategory: ""
description: |-
  Search an existing server role in ghostwriter.
---

# ghostwriter_server_role (Data Source)

Search an existing server role in ghostwriter.

## Example Usage

```terraform
data "ghostwriter_server_role" "test" {
  name = "Team Server / C2 Server"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the server role to be returned.

### Read-Only

- `id` (Number) The identifier of the server role.
