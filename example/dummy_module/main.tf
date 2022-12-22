terraform {
  required_providers {
    random = {
      source = "hashicorp/random"
      version = ">=3.4.0"
    }
  }
}

variable "key" {
    type = string
}
resource "random_pet" "server" {
  keepers = {
    key = var.key
  }
}

output "name" {
  value = "My pet is called: ${random_pet.server.id}"  
}