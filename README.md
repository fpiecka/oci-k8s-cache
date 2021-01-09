# oci-k8s-cache

Oracle Cloud Infrastructure Kubernetes token cache - because no one wants to wait 3s for their API to respond

## Installation

```bash
cd ~ # make sure that you don't add this binary to your Go Modules project :)
GO111MODULE=on go get -u github.com/pyr-sh/oci-k8s-cache@latest
```

Make sure that your `$GOPATH/bin` is in your `PATH` variable. Alternatively note the full path of the binary and use it
directly as the `command` in your `~/.kube/config`.

## Usage

After adding an OCI cluster to your `~/.kube/config`, change the cluster's user's `command` to `oci-k8s-cache`, or the
full path to the binary.

## Benchmarks

Ran on an i7-5960X from Poland against the `us-ashburn-1` cluster.

```
# directly using the OCI tool
~ time oci ce cluster generate-token --cluster-id ... --region us-ashburn-1
2.41s user 0.46s system 69% cpu 4.114 total
~ time kubectl get pods
kubectl get pods  2.46s user 0.52s system 62% cpu 4.747 total

# cached
~ time oci-k8s-cache ce cluster generate-token --cluster-id ... --region us-ashburn-1
0.00s user 0.00s system 102% cpu 0.002 total
~ time kubectl get pods
kubectl get pods  0.08s user 0.06s system 26% cpu 0.491 total
```
