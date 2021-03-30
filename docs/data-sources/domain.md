---
page_title: "googleworkspace_domain Data Source - terraform-provider-googleworkspace"
subcategory: ""
description: |-
  Domain data source in the Terraform Googleworkspace provider.
---

# Data Source `googleworkspace_domain`

Domain data source in the Terraform Googleworkspace provider.



## Schema

### Required

- **domain_name** (String) The domain name of the customer.

### Optional

- **id** (String) The ID of this resource.

### Read-only

- **creation_time** (Number) Creation time of the domain. Expressed in Unix time format.
- **domain_aliases** (List of String) asps.list of domain alias objects.
- **etag** (String) ETag of the resource.
- **is_primary** (Boolean) Indicates if the domain is a primary domain.
- **verified** (Boolean) Indicates the verification state of a domain.


