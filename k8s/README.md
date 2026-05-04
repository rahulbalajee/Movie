# Kubernetes Infrastructure

Application services (`metadata`, `rating`, `movie`) are deployed via the
Makefile (`make k8s-apply`). Infrastructure dependencies — Consul, MySQL,
Kafka — are installed manually via Helm so we can swap chart versions or
reconfigure them without touching the build pipeline.

The `*-values.yaml` files in this directory are the working overrides.
Pass them to `helm install` as shown below.

## Consul

Already installed in the `consul` namespace via the HashiCorp chart:

```bash
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install consul hashicorp/consul -n consul --create-namespace
```

Reachable cluster-internally at `consul-server.consul.svc.cluster.local:8500`.

## MySQL

```bash
kubectl create namespace mysql

# Schema bootstrap: ConfigMap mounted into /docker-entrypoint-initdb.d.
# Bitnami's init script runs it on first start (only when data dir is empty).
kubectl -n mysql create configmap mysql-init-scripts \
  --from-file=init.sql=schema/schema.sql

helm install mysql oci://registry-1.docker.io/bitnamicharts/mysql \
  -n mysql -f k8s/mysql-values.yaml --wait
```

Reachable at `mysql.mysql.svc.cluster.local:3306`. Root password is
`password`, database is `movieexample` (matches the DSN baked into the
service Deployment manifests).

Verify schema loaded:

```bash
kubectl exec -n mysql mysql-0 -- \
  mysql -uroot -ppassword movieexample -e "SHOW TABLES;"
```

If the schema didn't load (init only runs when the data dir is empty),
nuke and reinstall:

```bash
helm uninstall mysql -n mysql
kubectl delete pvc -n mysql --all
# then re-run the install
```

## Kafka

```bash
kubectl create namespace kafka

helm install kafka oci://registry-1.docker.io/bitnamicharts/kafka \
  -n kafka -f k8s/kafka-values.yaml --wait
```

Reachable at `kafka.kafka.svc.cluster.local:9092`. KRaft mode (no Zookeeper),
single combined controller+broker node, plaintext listeners.

The `ratings` topic auto-creates on first subscribe (chart default).

## Notes on Bitnami images

As of August 2025, Bitnami moved free community images from `docker.io/bitnami/*`
to `docker.io/bitnamilegacy/*`. The chart defaults still point at the old paths,
so both values files override `image.repository` to use `bitnamilegacy/*`.

MySQL is pinned to `8.4.5-debian-12-r0` (LTS) because the chart's default of
MySQL 9.x rejects the `source` client command Bitnami's init script uses for
`/docker-entrypoint-initdb.d/*.sql` files.

## Tear down

```bash
helm uninstall mysql -n mysql && kubectl delete namespace mysql
helm uninstall kafka -n kafka && kubectl delete namespace kafka
```
