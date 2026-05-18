# `netdebug bench` — in-cluster recipes

Run the echo server as a Deployment and `bench` as a one-shot Job. The
Deployment fronts multiple replicas behind a ClusterIP Service so the
per-backend breakdown has something to show; the Job captures the JSON
summary to its pod log, which is easy to pull back with `kubectl logs`.

The Dockerfile at the repo root produces a single image whose entrypoint is
`/app/netdebug`. Build and push it to a registry the cluster can reach, then
substitute that reference for `IMAGE` in the manifests below.

For the bench flag recipes themselves (payload-shape, compression,
conn-model), see [`bench-local.md`](bench-local.md). This doc only covers
how to wrap them for Kubernetes.

## Echo target

A Deployment with three replicas and a ClusterIP Service in front. Three
replicas is the minimum that exercises the per-backend breakdown and backend
skew.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo
  labels:
    app: echo
spec:
  replicas: 3
  selector:
    matchLabels:
      app: echo
  template:
    metadata:
      labels:
        app: echo
    spec:
      containers:
        - name: echo
          image: IMAGE
          args: ["echo", "--mode=grpc", "--host=0.0.0.0", "--port=8080"]
          ports:
            - name: grpc
              containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: echo
spec:
  selector:
    app: echo
  ports:
    - name: grpc
      port: 8080
      targetPort: grpc
```

```
kubectl apply -f echo.yaml
kubectl rollout status deployment/echo
```

The Service DNS name in the bench commands below is `echo.default.svc`;
adjust for the namespace you deployed into.

## Bench Job

A one-shot Job with `restartPolicy: Never` and `backoffLimit: 0` so a single
failed run does not get retried implicitly. `--output=json` writes the
summary to stdout, and the `--label` flags carry run metadata through to the
JSON so the log is self-identifying.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: bench
spec:
  backoffLimit: 0
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: bench
          image: IMAGE
          args:
            - "bench"
            - "--target=echo.default.svc:8080"
            - "--plaintext"
            - "--payload=embedding-float"
            - "--embedding-dim=1024"
            - "--concurrency=8"
            - "--duration=30s"
            - "--output=json"
            - "--label=run=2026-05-17a"
```

```
kubectl apply -f bench-job.yaml
kubectl wait --for=condition=complete --timeout=2m job/bench
kubectl logs -l job-name=bench > bench.json
kubectl delete job bench
```

## Paired mesh-on / mesh-off

The pattern from `bench-local.md` maps onto two runs of the same Job with a
different `sidecar.istio.io/inject` (or `linkerd.io/inject`) annotation. Use
the `--label` flag to tag each run so the summaries are self-identifying.

```yaml
# Mesh-off: explicitly disable injection.
apiVersion: batch/v1
kind: Job
metadata:
  name: bench-mesh-off
spec:
  backoffLimit: 0
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      restartPolicy: Never
      containers:
        - name: bench
          image: IMAGE
          args:
            - "bench"
            - "--target=echo.default.svc:8080"
            - "--plaintext"
            - "--payload=embedding-float"
            - "--embedding-dim=1024"
            - "--concurrency=8"
            - "--duration=30s"
            - "--output=json"
            - "--label=mesh=off"
            - "--label=run=2026-05-17a"
---
# Mesh-on: opt in. Same target, same flags.
apiVersion: batch/v1
kind: Job
metadata:
  name: bench-mesh-on
spec:
  backoffLimit: 0
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "true"
    spec:
      restartPolicy: Never
      containers:
        - name: bench
          image: IMAGE
          args:
            - "bench"
            - "--target=echo.default.svc:8080"
            - "--plaintext"
            - "--payload=embedding-float"
            - "--embedding-dim=1024"
            - "--concurrency=8"
            - "--duration=30s"
            - "--header=x-trace-id=mesh-on-001"
            - "--output=json"
            - "--label=mesh=on"
            - "--label=run=2026-05-17a"
```

```
kubectl apply -f bench-paired.yaml
kubectl wait --for=condition=complete --timeout=2m job/bench-mesh-off job/bench-mesh-on
kubectl logs -l job-name=bench-mesh-off > mesh-off.json
kubectl logs -l job-name=bench-mesh-on  > mesh-on.json
kubectl delete job bench-mesh-off bench-mesh-on
```

For a mesh-on run the per-backend breakdown's `source` will read
`peer-addr` pointing at the local sidecar; the `kubernetes.pod_name`
fallback exposes the real backend distribution. Under Envoy/Istio the
summary's `upstream` block reflects the
`x-envoy-upstream-service-time` header. Under linkerd2-proxy the upstream
block reads `n/a`; the proxy's latency signal lives on its `:4191/metrics`
endpoint, which bench does not scrape.

## Ad-hoc runs

For one-off interactive runs without committing a Job manifest, attach a
short-lived pod with `kubectl run`. The pod prints its summary to the
terminal and goes away when the run completes.

```
kubectl run bench --rm -it --restart=Never --image=IMAGE -- \
    bench --target=echo.default.svc:8080 --plaintext \
    --payload=embedding-float --embedding-dim=1024 \
    --concurrency=4 --duration=15s
```

Add `--annotations sidecar.istio.io/inject=true` (or the linkerd equivalent)
to put the ad-hoc pod behind a sidecar for a quick mesh-on sanity check.
