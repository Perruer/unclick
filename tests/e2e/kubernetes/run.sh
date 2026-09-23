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
tofu plan -input=false -no-color -out=plan.bin | tee plan.txt
grep -E "^Plan: [1-9][0-9]* to import, 0 to add, [0-9]+ to change, 0 to destroy\.$" plan.txt

# The Kubernetes provider does not set wait_for_rollout on import, so the
# first plan stores its default (true). It is a client-side setting and
# changes nothing in the cluster; any other change is a failure.
changed=$(tofu show -json plan.bin | jq -r '
  .resource_changes[]
  | select(.change.actions != ["no-op"])
  | .address as $addr
  | .change.before as $b
  | .change.after as $a
  | ($a | keys[]) as $k
  | select($b[$k] != $a[$k])
  | "\($addr).\($k)"')
unexpected=$(printf '%s\n' "$changed" | grep -v -e '\.wait_for_rollout$' -e '^$' || true)
if [ -n "$unexpected" ]; then
  echo "unexpected changes:"
  echo "$unexpected"
  exit 1
fi
