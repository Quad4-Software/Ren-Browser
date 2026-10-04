#!/usr/bin/env bash
# Run CI jobs outside GitHub Actions. Same entry points for local shells,
# Forgejo, Gitea, and thin workflow wrappers.
#
# Usage:
#   bash build/scripts/ci/run.sh <job>
#
# Jobs:
#   pr            Default PR gate (check + contract + gofmt on Linux)
#   go            Go tests (matrix-friendly, hard tests on Linux only)
#   frontend      Frontend quality gates
#   brand         Generated brand file drift
#   coverage      Go and frontend coverage
#   e2e           Playwright E2E and optional visual (visual may fail soft)
#   reproducible  Reproducible server binary check
#   benchmark     Go benchmarks and compare when baseline exists
#   security      gosec and frontend audit (no Trivy action dependency)
#   full          pr + coverage + e2e + reproducible
#   list          Print job names
#
# Env:
#   CI_SKIP_SETUP=1     Skip setup.sh (when a workflow already ran prep)
#   CI_SKIP_TREE_VERIFY=1   Skip renbrowser.rsm verify (local quick runs)
#   RNS_INVENTORY_OUT   Inventory path for end-of-job clean check
set -euo pipefail

. "$(dirname "$0")/env.sh"

run_setup() {
	if ci_truthy "${CI_SKIP_SETUP:-}"; then
		bash "$ROOT/build/scripts/ci-prep-go.sh"
		return 0
	fi
	bash "$ROOT/build/scripts/ci/setup.sh"
}

job_go() {
	run_setup
	require_task
	go list -mod=vendor ./...
	if ci_on_linux; then
		bash "$ROOT/build/scripts/gofmt-project.sh" check
	fi
	task test:go
	if ci_on_linux; then
		task common:install:frontend:deps
		task test:go:hard
		go test ./internal/plugins/...
		task test:contract
	fi
	verify_workspace_clean
}

job_pr() {
	export CI_INSTALL_FRONTEND_DEPS=1
	run_setup
	require_task
	if ci_on_linux; then
		bash "$ROOT/build/scripts/gofmt-project.sh" check
	fi
	task check
	task test:contract
	verify_workspace_clean
}

job_frontend() {
	export CI_INSTALL_FRONTEND_DEPS=1
	run_setup
	require_task
	task frontend:check
	verify_workspace_clean
}

job_brand() {
	run_setup
	require_task
	task common:check:brand
}

job_coverage() {
	export CI_INSTALL_FRONTEND_DEPS=1
	export CI_INSTALL_LINUX_DEPS=1
	run_setup
	require_task
	task coverage:go
	task coverage:frontend
}

job_e2e() {
	export CI_INSTALL_FRONTEND_DEPS=1
	run_setup
	require_task
	task test:e2e
	if ! task test:visual; then
		echo "ci: visual regression failed (non-fatal for portable CI)" >&2
	fi
}

job_reproducible() {
	export CI_INSTALL_FRONTEND_DEPS=1
	run_setup
	require_task
	task check:reproducible
	verify_workspace_clean
}

job_benchmark() {
	export CI_INSTALL_LINUX_DEPS=1
	run_setup
	require_task
	mkdir -p bench/baseline
	task bench:go
	if [ -f bench/baseline/go.txt ]; then
		task bench:compare BENCH_BASELINE=bench/baseline/go.txt
	else
		echo "ci: no bench/baseline/go.txt, skipping compare"
	fi
	cp bench/go.txt bench/baseline/go.txt
}

job_security() {
	export CI_INSTALL_FRONTEND_DEPS=1
	run_setup
	require_task
	task gosec
	pnpm --dir frontend run audit
	pnpm --dir frontend run lint
	if command -v trivy >/dev/null 2>&1; then
		task sbom 2>/dev/null || trivy fs --severity HIGH,CRITICAL --exit-code 1 \
			--skip-dirs vendor,frontend/node_modules,bin,.cache,third_party,build/android/.gradle,build/android/build .
	else
		echo "ci: trivy not installed, skipping filesystem scan (install via build/scripts/ci/setup-trivy.sh)" >&2
	fi
}

job_full() {
	job_pr
	job_coverage
	job_e2e
	job_reproducible
}

usage() {
	sed -n '6,20p' "$0" | sed 's/^# \?//'
}

job="${1:-pr}"
case "$job" in
list)
	echo "pr go frontend brand coverage e2e reproducible benchmark security full"
	;;
go) job_go ;;
pr) job_pr ;;
frontend) job_frontend ;;
brand) job_brand ;;
coverage) job_coverage ;;
e2e) job_e2e ;;
reproducible) job_reproducible ;;
benchmark) job_benchmark ;;
security) job_security ;;
full) job_full ;;
-h | --help | help)
	usage
	;;
*)
	echo "ci: unknown job: $job" >&2
	usage >&2
	exit 2
	;;
esac
