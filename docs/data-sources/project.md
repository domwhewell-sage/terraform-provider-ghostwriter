---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ghostwriter_project Data Source - ghostwriter"
subcategory: ""
description: |-
  Search an existing project in ghostwriter.
---

# ghostwriter_project (Data Source)

Search an existing project in ghostwriter.

## Example Usage

```terraform
data "ghostwriter_project" "testproject" {
  code_name = "Test Project"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `code_name` (String) The project codename

### Read-Only

- `client_id` (Number) The ID of the client
- `complete` (Boolean) If the project is complete
- `end_date` (String) The end date of the project
- `end_time` (String) The end time of the project
- `id` (Number) The identifier of the project.
- `note` (String) The note asociated with the project
- `operator_id` (Number) The ID of the assigned operator
- `project_type_id` (Number) The ID of the project type
- `slack_channel` (String) The projects slack channel
- `start_date` (String) The start date of the project
- `start_time` (String) The start time of the project
- `timezone` (String) The projects timezone
