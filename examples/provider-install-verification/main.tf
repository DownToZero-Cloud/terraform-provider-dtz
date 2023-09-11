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

data "downtozero_container_services" "test" {}

output "downtozero_containerservices" {
  description = "Testing output"
  value       = data.downtozero_container_services.test
}

resource "downtozero_container_service" "test1" {
  domains = [{
    name = "36f51551-1cac-4e61-a2da-f58c6192eaff.containers.dtz.dev"
    routing = [{
      prefix = "/",
      service_definition = {
        container_image = "docker.io/library/nginx"
      }
    }]
  }]
}

resource "downtozero_container_service" "test2" {
  domains = [{
    name = "08605de8-8baf-49bf-a820-4977ba694edc.containers.dtz.dev"
    routing = [{
      prefix = "/",
      service_definition = {
        container_image = "docker.io/library/nginx"
      }
    }]
  }]
}
