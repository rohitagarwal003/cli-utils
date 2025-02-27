[kind]: https://github.com/kubernetes-sigs/kind

# Demo: Lifecycle directives

This demo shows how it is possible to use a lifecycle directive to 
change the behavior of prune and delete for specific resources.

First define a place to work:

<!-- @makeWorkplace @testE2EAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

Alternatively, use

> ```
> DEMO_HOME=~/hello
> ```

## Establish the base

<!-- @createBase @testE2EAgainstLatestRelease -->
```
BASE=$DEMO_HOME/base
mkdir -p $BASE
OUTPUT=$DEMO_HOME/output
mkdir -p $OUTPUT
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

function expectedOutputLine() {
  if ! grep -q "$@" "$OUTPUT/status"; then
    echo -e "${RED}Error: output line not found${NC}"
    echo -e "${RED}Expected: $@${NC}"
    exit 1
  else
    echo -e "${GREEN}Success: output line found${NC}"
  fi
}

function expectedNotFound() {
  if grep -q "$@" $OUTPUT/status; then
    echo -e "${RED}Error: output line found:${NC}"
    echo -e "${RED}Found: $@${NC}"
    exit 1
  else
    echo -e "${GREEN}Success: output line not found found${NC}"
  fi
}
```

In this example we will just use three ConfigMap resources for simplicity, but
of course any type of resource can be used.

- the first ConfigMap resource does not have any annotations;
- the second ConfigMap resource has the **cli-utils.sigs.k8s.io/on-remove** annotation with the value of **keep**;
- the third ConfigMap resource has the **client.lifecycle.config.k8s.io/deletion** annotation with the value of **detach**.

These two annotations tell the kapply tool that a resource should not be deleted, even
if it would otherwise be pruned or deleted with the destroy command.

<!-- @createFirstCM @testE2EAgainstLatestRelease-->
```
cat <<EOF >$BASE/configMap1.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: firstmap
data:
  artist: Ornette Coleman
  album: The shape of jazz to come
EOF
```

This ConfigMap includes the **cli-utils.sigs.k8s.io/on-remove** annotation

<!-- @createSecondCM @testE2EAgainstLatestRelease-->
```
cat <<EOF >$BASE/configMap2.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: secondmap
  annotations:
    cli-utils.sigs.k8s.io/on-remove: keep
data:
  artist: Husker Du
  album: New Day Rising
EOF
```


This ConfigMap includes the **client.lifecycle.config.k8s.io/deletion** annotation

<!-- @createSecondCM @testE2EAgainstLatestRelease-->
```
cat <<EOF >$BASE/configMap3.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: thirdmap
  annotations:
    client.lifecycle.config.k8s.io/deletion: detach
data:
  artist: Husker Du
  album: New Day Rising
EOF
```

## Run end-to-end tests

The following requires installation of [kind].

Delete any existing kind cluster and create a new one. By default the name of the cluster is "kind"
<!-- @deleteAndCreateKindCluster @testE2EAgainstLatestRelease -->
```
kind delete cluster
kind create cluster
```

Use the kapply init command to generate the inventory template. This contains
the namespace and inventory id used by apply to create inventory objects. 
<!-- @createInventoryTemplate @testE2EAgainstLatestRelease-->
```
kapply init $BASE | tee $OUTPUT/status
expectedOutputLine "namespace: default is used for inventory object"

```

Apply the three resources to the cluster.
<!-- @runApply @testE2EAgainstLatestRelease -->
```
kapply apply $BASE --reconcile-timeout=1m | tee $OUTPUT/status
```

Use the preview command to show what will happen if we run destroy. This should
show that secondmap and thirdmap will not be deleted even when using the destroy
command.
<!-- @runDestroyPreview @testE2EAgainstLatestRelease -->
```
kapply preview --destroy $BASE | tee $OUTPUT/status

expectedOutputLine "configmap/firstmap deleted"

expectedOutputLine "configmap/secondmap delete skipped"

expectedOutputLine "configmap/thirdmap delete skipped"
```

We run the destroy command and see that the resource without the annotations (firstmap)
has been deleted, while the resources with the annotations (secondmap and thirdmap)  are still in the
cluster.
<!-- @runDestroy @testE2EAgainstLatestRelease -->
```
kapply destroy $BASE | tee $OUTPUT/status

expectedOutputLine "configmap/firstmap deleted"

expectedOutputLine "configmap/secondmap delete skipped"

expectedOutputLine "configmap/thirdmap delete skipped"

expectedOutputLine "1 resource(s) deleted, 2 skipped"
expectedNotFound "resource(s) pruned"

kubectl get cm --no-headers | awk '{print $1}' | tee $OUTPUT/status
expectedOutputLine "secondmap"

kubectl get cm --no-headers | awk '{print $1}' | tee $OUTPUT/status
expectedOutputLine "thirdmap"
```

Apply the resources back to the cluster so we can demonstrate the lifecycle
directive with pruning.
<!-- @runApplyAgain @testE2EAgainstLatestRelease -->
```
kapply apply $BASE --inventory-policy=adopt --reconcile-timeout=1m | tee $OUTPUT/status
```

Delete the manifest for secondmap and thirdmap
<!-- @runDeleteManifest @testE2EAgainstLatestRelease -->
```
rm $BASE/configMap2.yaml

rm $BASE/configMap3.yaml
```

Run preview to see that while secondmap and thirdmap would normally be pruned, they
will instead be skipped due to the lifecycle directive.
<!-- @runPreviewForPrune @testE2EAgainstLatestRelease -->
```
kapply preview $BASE | tee $OUTPUT/status

expectedOutputLine "configmap/secondmap prune skipped"

expectedOutputLine "configmap/thirdmap prune skipped"
```

Run apply and verify that secondmap and thirdmap are still in the cluster.
<!-- @runApplyToPrune @testE2EAgainstLatestRelease -->
```
kapply apply $BASE | tee $OUTPUT/status

expectedOutputLine "configmap/secondmap prune skipped"

expectedOutputLine "configmap/thirdmap prune skipped"

kubectl get cm --no-headers | awk '{print $1}' | tee $OUTPUT/status
expectedOutputLine "secondmap"

kubectl get cm --no-headers | awk '{print $1}' | tee $OUTPUT/status
expectedOutputLine "thirdmap"

kind delete cluster;
```
