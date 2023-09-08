terraform {
  required_providers {
    downtozero = {
      source = "hashicorp.com/edu/downtozero",
    }
  }
}

provider "downtozero" {
  apikey = "<add-me>"
}

data "downtozero_containerservices" "test" {}

output "downtozero_containerservices" {
  description = "Testing output"
  value       = data.downtozero_containerservices.test
}
