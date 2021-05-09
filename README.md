# ks
Kubernetes Secret Reader From Cluster and Local File

# Usage

```
go get github.com/pavansh/ks
ks local -f <secret.yaml>
ks k8s -n <namespace> -s <secretname>
```