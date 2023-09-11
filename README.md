# DownToZero Terraform Provider

## READ THIS

* Read "how to use this" if you are new!
* The model for `containerservices` is based on one users response, the model might currently be incomplete!
* Any change inside the go module (main.go/provider.go etc.) is only visible after rebuild with `go install .`! It's easiest to have two consoles open.

## How to use this

1. make sure you have go and terraform installed
2. clone this repo
3. create a new terraformrc file
    * mac: `echo 'provider_installation { dev_overrides { "hashicorp.com/edu/downtozero" = "/Users/<Username>/go/bin"} direct {}}' > ~/.terraformrc`
    * linux: `echo 'provider_installation { dev_overrides { "hashicorp.com/edu/downtozero" = "/home/<Username>/go/bin"} direct {}}' > ~/.terraformrc`
    * the right side needs to be an absolute path and it needs to be the **folder** your go binaries are saved in!
4. go into this repo and `go mod tidy` to install all dependencies `go install .` to install this provider as a binary
5. Make sure you have an API Key from DTZ for a project with containerservices enabled
6. go into `examples/provider-install-verification` and add your APIKEY to `main.tf`
7. `terraform plan` (do not use `terraform init`, it has no effect)
8. Terraform should print out the containerservices for this apikey

## References

* [Hashicorp Guide](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider)
* logging (to log out a struct): `tflog.Info(ctx, fmt.Sprintf("%+v", yourstruct))`
