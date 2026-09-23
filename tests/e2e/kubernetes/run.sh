#!/usr/bin/env bash
# End-to-end test against a disposable cluster (kind): create a namespace
# with a config map and a deployment, import them with Unclick and check that
# `tofu plan` only imports.
#
# Needs: kubectl pointing at a throwaway cluster, tofu in PATH and an unclick
# binary built with the kubernetes provider ($UNCLICK).
set -euo pipefail

UNCLICK="${UNCLICK:-unclick}"
ns=unclick-e2e

kubectl create namespace "$ns"
kubectl -n "$ns" create configmap settings --from-literal=mode=production
kubectl -n "$ns" create deployment web --image=nginx:1.27 --replicas=1

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
cd "$work"

# The provider plugin reads the cluster from KUBE_CONFIG_PATH, the importer
# from the usual kubeconfig.
export KUBE_CONFIG_PATH="${KUBECONFIG:-$HOME/.kube/config}"

"$UNCLICK" import kubernetes --resources=configmaps,deployments \
  --filter="Name=metadata.0.namespace;Value=$ns"
cd generated/kubernetes

tofu init -input=false -no-color >/dev/null
tofu plan -input=false -no-color | tee plan.txt
grep -E "^Plan: [1-9][0-9]* to import, 0 to add, 0 to change, 0 to destroy\.$" plan.txt
