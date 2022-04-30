variable "server" {
  type = string
}
variable "name" {
  type = string
}
variable "username" {
  type = string
}
variable "password" {
  type = string
}

data "dependencies_nexus_raw" "dependency" {
  nexus_server = var.server
  name = var.name
  destination = "${path.root}/dependencies"
  username = var.username
  password = var.password
}

output "size" {
  value = data.dependencies_nexus_raw.dependency.asset_size
}
