# Kubernetes Production Deployment

The manifests under `deployments/kubernetes/` are grouped by component category
(`core/`, `security/`, `database/`, `messaging/`, `cache/`, `observability/`,
`networking/`, `services/`) and describe the active Java Vert.x services and
their supporting infrastructure. The root `kustomization.yaml` is the entrypoint
for `kubectl apply -k deployments/kubernetes`; the production overlay lives in
`overlays/production/`. They are intentionally not a complete cluster bootstrap:
the following prerequisites must exist before an ArgoCD sync.

```text
deployments/kubernetes/
├── kustomization.yaml            # entrypoint (staging): kubectl apply -k deployments/kubernetes
├── core/                         # namespace, app-config/app-secrets (ExternalSecret)
├── security/                     # namespace-level NetworkPolicies (DNS, SMTP)
├── database/                     # postgres, pgbouncer, postgres-exporter, clickhouse, migrate Job
├── messaging/                    # kafka
├── cache/                        # redis
├── observability/                # prometheus, grafana, loki, jaeger, otel, node-exporter
├── networking/                   # nginx
├── services/                     # one folder per service (apigateway … user, stats_reader, stats_writer)
├── components/image-pull-secret/ # shared imagePullSecrets patch (root + overlay)
└── overlays/production/          # ArgoCD entrypoint (image pinning, GHCR_OWNER, sync-waves)
```

## Required cluster prerequisites

1. **External Secrets Operator** must be installed and serving
   `external-secrets.io/v1beta1`.
2. A `ClusterSecretStore` named `ecommerce-secrets` must exist and provide these
   remote keys:
   `ecommerce/DB_USERNAME`, `DB_PASSWORD`, `POSTGRES_USER`,
   `POSTGRES_PASSWORD`, `GF_SECURITY_ADMIN_PASSWORD`, `SECRET_KEY`,
   `REDIS_PASSWORD`, `SMTP_USER`, `SMTP_PASS`, `ALERTMANAGER_CONFIG`, and
   `GHCR_DOCKERCONFIGJSON`. `ALERTMANAGER_CONFIG` must contain the complete
   Alertmanager YAML (including SMTP settings) as one secret value.
3. The `GHCR_DOCKERCONFIGJSON` value must be a valid Docker config JSON with
   `read:packages` access to the GHCR image repo. The image owner is templated:
   it is read from the `app-config` ConfigMap key `GHCR_OWNER`
   (`deployments/kubernetes/core/configmaps.yaml`) and substituted into
   the `<owner>` segment of every image reference by the `replacements` block in
   `overlays/production/kustomization.yaml`. Keep `GHCR_OWNER` equal to
   `github.repository_owner` (lowercase) of the CI repo — on a fork, change only
   that ConfigMap value.
4. The cluster must have metrics-server installed before enabling the HPAs.
5. Apply and verify the External Secrets resources server-side before syncing
   application Deployments:

```bash
kubectl apply --server-side -k deployments/kubernetes/core
kubectl -n ecommerce get externalsecret app-secrets ghcr-pull-secret
kubectl -n ecommerce get secret app-secrets ghcr-pull-secret
```

## Image promotion

CI publishes each application image to:

```text
ghcr.io/<owner>/vertx-ecommerce/<service>:<git-sha>
ghcr.io/<owner>/vertx-ecommerce/<service>:latest
```

where `<owner>` is `github.repository_owner` (lowercase) and must match the
`GHCR_OWNER` value in the `app-config` ConfigMap (see prerequisites above).

The manifests currently use `:latest`, and the production overlay pins an
immutable tag per release. This pinning is **automated**: after every push
build, the `update-manifests` job in `.github/workflows/ci.yml` rewrites every
`newTag` in `overlays/production/kustomization.yaml` to the commit SHA it just
pushed (and commits it back with `[skip ci]`), so ArgoCD always rolls out the
exact artifacts the pipeline produced. Manual `:latest` → SHA replacement is
no longer needed. To verify a pinned tag before rollout, confirm it exists in
GHCR and perform a server-side dry run against the target cluster:

```bash
kubectl apply --server-side --dry-run=server -k deployments/kubernetes
kubectl diff -k deployments/kubernetes
```

Do not treat offline `kubectl kustomize` rendering as proof that CRDs,
ClusterSecretStore, image credentials, or admission policies exist in the
cluster.

> **Note:** the root `deployments/kubernetes` renders image references with the
> `__GHCR_OWNER__` placeholder (the real owner is substituted by the
> `replacements` block in `overlays/production/kustomization.yaml`). Always
> build/render through the production overlay — do not apply the root directly
> for application workloads.
