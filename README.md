# Kubernetes Manifest Check

Tools for checking Kubernetes YAML files.

At the moment the tools check Syntax and Resources in all pods, rc and deployments are set to a none zero value and cpu/memory request is below limit.

## Install

```bash
go get github.com/bborbe/k8s-manifest-check
```

## Check all Kubernetes manifest files


```bash
find . \
-type f \
-name "*.yaml" \
-exec k8s-manifest-check "{}" +
```
