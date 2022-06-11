---
page_title: "Nexus Raw Data Source"
description: |-
  Retrieves the asset at a Nexus server (https).
---

# `dependencies_nexus_raw` Data Source

The `dependencies_nexus_raw` data source makes HTTPS GET requests to the given Nexus Server to retrieve information about the desired asset in the raw-trusted repository and if the file does not exist locally or the MD5 does not match it will download it.

## Example Usage

```hcl
data "dependencies_nexus_raw" "dependency" {
  nexus_server = "repository.company.com"
  name = "com/company/product.exe"
  destination = "${path.root}/dependencies"
  
  # Optional authentication information
  username = var.username
  password = var.password

}
```

## Argument Reference

The following arguments are supported:

* `nexus_server` - (Required) The server hostname to request data from.

* `name` - (Required) The name of the component to download.

* `destination` - (Required) The directory where the file will be saved.

* `basic_auth` - (Optional) The basic authentication header **base64 encoded** without "Basic " prefix, i.e.: base64("${username}:${password}"). May also be stored in Secret Manager and referenced with gcp_secret!projects/`project`/secrets/`secret-name`/versions/`version`.

* `username` - (Optional) The username to use when authentication is required. (Only used if password is given as well)

* `password` - (Optional) The password to use when authentication is required. (Only used if username is given as well). May also be stored in Secret Manager and referenced with gcp_secret!projects/`project`/secrets/`secret-name`/versions/`version`.

## Attributes Reference

The following attributes are exported:

* `asset_id` - The ID of the asset as provided by Nexus.

* `asset_url` - The URL of the asset where it can be downloaded from Nexus.

* `asset_md5` - The MD5 hash of the asset downloaded.

* `asset_content_type` - The MIME Type of the asset downloaded.

* `asset_path` - Path where the asset was downloaded (`destination`/`name`).

* `asset_size` - The size of the file downloaded.
