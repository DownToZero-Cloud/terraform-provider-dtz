# DownToZero Terraform Provider

## How to use this

1. make sure you have go and terraform installed
2. clone this repo
3. create a new terraformrc file, see [this guide](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install)
4. `dev_overrides` inside the file needs to be `"hashicorp.com/edu/downtozero" = "/home/<user>/go/bin"`
    - the right side needs to be an absolute path and it needs to be the **folder** your go binaries are saved in!
5. go into this repo and `go mod tidy` to install all dependencies `go install .` to install this provider as a binary
6. go into `examples/provider-install-verification` and `terraform plan`
7. if terraform is angry about Missing API Endpoints, Username etc., you did everything right and can start developing

---
Any change inside the go module (main.go/provider.go etc.) is only visible after rebuild with `go install .`! It's easiest to have two consoles open.

## References

[Hashicorp Guide](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider)
