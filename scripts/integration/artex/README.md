# Isolated ARTEX contract lab

This lab runs the real ARTEX backend and PostgreSQL 16 on an internal Docker
network. A deterministic local model fixture records protocol metadata and
returns synthetic responses. It validates API contracts and lifecycle behavior;
it is **not** real-model validation and does not discover vulnerabilities.

The fixed-target control proxy exposes only loopback ports 18787/18788. The
ARTEX container itself has no external network, Docker socket or host workspace
mount. PostgreSQL and the ARTEX workspace use independent named volumes.

## Build the pinned backend

Check out `Autumn-27/ARTEX` at
`bf7f414477832b77d2152539c0723dc691086522` outside this working tree. From that
checkout, build with Go 1.26.5 or a compatible toolchain:

```powershell
$env:GOOS='linux'
$env:GOARCH='amd64'
$env:CGO_ENABLED='0'
$env:GOMAXPROCS='2'
go build -p 2 -trimpath -o F:/CTF/PentAGI/.integration-lab/artex-bf7f414-linux ./cmd/artex
```

## Start and probe (from the PentAGI repository)

The runtime directory must be outside Git. Startup generates local fixture
credentials there and restricts their Windows ACL to the current user. Do not
copy credentials, environment files, JWT signing keys or database volumes into
test reports. The sanitized reports contain assertions and public IDs only.

```powershell
./scripts/integration/artex/start-lab.ps1 `
  -BinaryPath F:/CTF/PentAGI/.integration-lab/artex-bf7f414-linux `
  -LabRoot F:/CTF/PentAGI/.integration-lab

./scripts/integration/artex/run-probe.ps1 `
  -BinaryPath F:/CTF/PentAGI/.integration-lab/artex-bf7f414-linux `
  -LabRoot F:/CTF/PentAGI/.integration-lab -Mode bootstrap

./scripts/integration/artex/run-probe.ps1 `
  -BinaryPath F:/CTF/PentAGI/.integration-lab/artex-bf7f414-linux `
  -LabRoot F:/CTF/PentAGI/.integration-lab -Mode verify-restart -Recreate
```

The bootstrap creates fixture-owned profiles and tasks and seeds one synthetic
finding directly into the lab database. It tests actual worker/planner model
selection, in-flight cancellation, key rotation, independent retester binding,
and an explicitly inconclusive retest. It pauses its tasks afterward. It only
reuses named fixture profiles; unrelated profiles are not changed.

The recreation probe verifies the pre-existing login token remains valid and
that task state, model bindings, rotated credentials and retest history survive
replacement of the ARTEX container. ARTEX keeps its signing key beside the
binary in this source version, so the lab explicitly mounts that key as a
read-only file instead of assuming the workspace volume includes it.

ARTEX task deadlines continue during pause. Run the recreation probe within the
fixture's 3600-second wall-clock deadline, or bootstrap fresh tasks. A timed-out
task remaining timed out after recreation is not loss of paused state.

Reports go under `docs/merge-baseline/DATE-artex-live-*.json`; each run replaces
that date/mode's report. Archive unexpected failures separately before rerunning.

## Stop

Set `ARTEX_LAB_RUNTIME` and `ARTEX_LAB_BINARY` to the paths used above, then:

```powershell
docker compose -f scripts/integration/artex/compose.yml stop
```

Stopping preserves test volumes and signing material for the next verification.
No script removes volumes or touches the normal PentAGI deployment.
