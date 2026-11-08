## tfcheck

Experimental go binary that runs terraform (or OpenTofu) checks with pretty terminal UI: `fmt`, `init -backend=false`,
`validate` and (if it's in `PATH`) `tflint`, in each directory with `.tf` files.

![](examples/demo.gif)

## Usage

```sh
$ go install cmd/tfcheck
$ which tfcheck

# Inside a terraform directory
$ tfcheck
```

`tfcheck` uses `terraform` if it's in `PATH`, and otherwise `tofu`. To choose, use `--binary tofu` (or
`TFCHECK_BINARY=tofu`).

`init` with `tofu` in a directory with a `.terraform.lock.hcl` written by `terraform` rewrites its providers from
`registry.terraform.io` to `registry.opentofu.org`, so don't commit those changes if the repo uses `terraform`.
