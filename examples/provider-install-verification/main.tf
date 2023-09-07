terraform {
  required_providers {
    downtozero = {
      source = "hashicorp.com/edu/downtozero",
    }
  }
}

provider "downtozero" {

}

data "downtozero_coffees" "test" {}