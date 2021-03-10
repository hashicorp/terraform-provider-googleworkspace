---
page_title: "scaffolding_data_source Data Source - terraform-provider-googleworkspace"
subcategory: ""
description: |-
  Sample data source in the Terraform provider scaffolding.
---

# Data Source `scaffolding_data_source`

Sample data source in the Terraform provider scaffolding.

## Example Usage

```terraform
data "scaffolding_data_source" "example" {
  sample_attribute = "foo"
}
```

## Schema

### Required

- **sample_attribute** (String) Sample attribute.

### Optional

- **id** (String) The ID of this resource.


