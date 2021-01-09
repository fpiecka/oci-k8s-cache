# oci-k8s-cache

Oracle Cloud Infrastructure Kubernetes token cache - because no one wants to wait 3s for their API to respond

## Installation

```bash
cd ~
GO111MODULE=on go get -u github.com/pyr-sh/oci-k8s-cache@latest
```

Make sure that your `$GOPATH/bin` is in your `PATH` variable. Alternatively note the full path of the binary and use it
directly as the `command` in your `~/.kube/config`.

## Usage

After adding an OCI cluster to your `~/.kube/config`, change the cluster's user's `command` to `oci-k8s-cache`, or the
full path to the binary.
