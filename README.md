## tfcheck

Experimental go binary that runs terraform check with pretty terminal UI.

![](examples/demo.gif)

## Usage

```sh
$ go install cmd/tfcheck
$ which tfcheck

# Inside a terraform directory
$ tfcheck
```

## TODO

1. Use context cancellation instead of capturing input.
2. Figure out how to set color properly.
