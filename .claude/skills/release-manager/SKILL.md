---
name: release-manager
description: KubeApp release and GitOps promotion workflows. Handles dev/staging/prod flows, image tags, RC creation, hotfixes, and pre-release validation checklists.
user-invocable: true
---

# Release Manager — KubeApp Release Skill

When this skill is invoked, assist the user with KubeApp release workflows using the domain knowledge below.

## How to handle requests

**Handle directly** (no sub-agent needed):
- Explaining the release flow, conventions, or tag formats
- Status checks: `gh run list`, `git log --tags`, `gh pr list`
- Answering "how do I..." or "what is the format for..." questions
- Helping the user craft a single command or diagnose a single failure

**Delegate to the `release-manager` sub-agent** (via Task tool with `subagent_type: release-manager`):
- End-to-end release orchestration (e.g. "promote dev to staging")
- Troubleshooting a failed pipeline or stalled GitOps promotion
- Running the pre-production validation checklist
- Hotfix flows that require multiple sequential steps
- Any operation requiring multiple gh/kubectl/aws commands in sequence

When delegating, pass the user's full request and any context you already have (app name, current tags, open PRs) to the sub-agent so it doesn't have to re-gather state.

---

## Platform Context

- **GitOps repo**: defined by `KUBECORE_KUBEPROJECT_REPO` (env var / GitHub repo variable)
- **Registry**: defined by `KUBECORE_REGISTRY` (includes full repo path, e.g. `registry.example.com/org/app`)
- **App name**: the GitHub repository name (`APP_NAME`)
- **Environments**: `dev`, `staging`, `preview-pr-{N}`, `production`
- **Deployment system**: ArgoCD — GitOps repo dispatches drive environment updates
- **Branch protection**: `main` and `dev` are protected. Never push directly.

---

## Branch Conventions

| Branch prefix       | Purpose                          | Version bump  |
|---------------------|----------------------------------|---------------|
| `feature/*`         | New functionality                | minor         |
| `bug/*`, `bugfix/*` | Bug fixes                        | patch         |
| `breaking/*`        | Breaking changes                 | major         |
| `maintenance/*`     | Non-feature maintenance          | patch         |
| `chore/*`           | CI, tooling, no-op changes       | none          |
| `hotfix/*`          | Critical production fix          | patch (from prod tag) |
| `rc/vX.Y.Z`         | Release candidate branch         | none (version from branch name) |

**Protected branches**: `main`, `dev`, `master` — never delete, never push directly.

---

## Image Tag Conventions

| Environment | Tag format                                      | Example                                 |
|-------------|-------------------------------------------------|-----------------------------------------|
| dev         | `dev-vX.Y.Z-YYYYMMDD-HHMMSS-{sha7}`            | `dev-v1.2.0-20240506-152301-abc1234`    |
| preview     | `preview-{PR_NUMBER}-{full_sha}`                | `preview-42-a1b2c3d4...`                |
| staging/RC  | `vX.Y.Z-rcN`                                    | `v1.2.0-rc3`                            |
| production  | `vX.Y.Z`                                        | `v1.2.0`                                |
| hotfix      | `vX.Y.Z-hotfix-{sha7}` → promoted to `vX.Y.Z` | `v1.1.5-hotfix-abc1234` → `v1.1.5`     |

Git tags follow the same format. Dev tags: `dev-vX.Y.Z-YYYYMMDD-HHMMSS`. Prod tags: `vX.Y.Z`.

---

## Release Pipelines

### 1. Dev Release (`dev-release.yaml`)
**Trigger**: PR opened/synced/merged targeting `dev`

**On PR open/sync**:
- Gate check: skip if source branch is `main` or `rc/*` (merge-back)
- Validate branch naming against allowed prefixes
- Comment version bump impact on PR

**On PR merge** (not chore, not rc, not merge-back):
1. Determine branch type → calculate next version (semver bump from latest `dev-v*` tag)
2. Set image tag: `dev-vX.Y.Z-{timestamp}-{sha7}`
3. Create and push git tag `dev-vX.Y.Z-{timestamp}`
4. Build and push Docker image
5. Dispatch `dev-environment-update` event to GitOps repo
6. If label `create-rc🚀` was present on the PR → automatically create `rc/vX.Y.Z` branch and PR to `main`
7. If label absent → comment with instructions on how to create RC manually (workflow or CLI)

**Skip conditions**: `chore/*` and `rc/*` branches produce no version bump, no image, no GitOps dispatch.

---

### 2. Preview Request (`preview-request.yaml`) — ⚠️ Not Yet Supported

> Preview environments are **coming soon** and are not yet available on the platform.
> If the user asks for a preview environment or the `preview` label, let them know:
> *"Preview environments are not yet supported but are coming soon. Changes can be validated on the dev environment after merging to dev."*


**Trigger**: PR to `dev` opened/synced/labeled/unlabeled/closed

**Gate**: Only acts when `preview` label is present. Skip if source is `main` or `rc/*`.

**On label `preview` added** (or PR sync with label present):
1. Create per-PR GitHub environment `preview-pr-{N}` if it doesn't exist
2. Build image: `preview-{PR_NUMBER}-{sha}`
3. Create GitHub Deployment → set status `pending`
4. Dispatch `preview-environment-update` to GitOps repo (ArgoCD deploys to `{app}-preview-{N}` namespace)
5. Comment on PR with preview URL: `https://{PR_NUMBER}-{app}.preview.{domain}`

**On PR close or `preview` label removed**:
- Find active deployment for `preview-pr-{N}` environment
- Set deployment status to `inactive` → triggers ArgoCD cleanup
- Comment on PR confirming cleanup

**Merge gate**: Branch protection requires `preview-deployment-validation` status check to pass.
For merge-backs from `main`/`rc/*`, this check is auto-bypassed.

---

### 3. RC Release / Staging Deploy (`rc-release.yaml`)
**Trigger**: PR opened/synced/reopened targeting `main` from `rc/vX.Y.Z` branch

**Gate**: PR head must match `rc/vX.Y.Z` pattern, else skip.

**Steps**:
1. Extract semantic version from branch name `rc/vX.Y.Z` → `X.Y.Z`
2. Calculate RC number: scan `vX.Y.Z-rc*` tags → increment. First RC = rc1.
3. Set image tag: `vX.Y.Z-rcN`
4. Create or update PR to `main` titled `🚀 Release vX.Y.Z-rcN to Production`
5. Build and push RC image
6. Create GitHub Deployment for `staging` environment (transient, non-production)
7. Create pre-release on GitHub tagged `vX.Y.Z-rcN` with auto-generated release notes
8. Create or update RC tracking issue (label: `release-candidate`, `staging-deployment`)
9. Dispatch `staging-environment-update` to GitOps repo

**Each new push to the RC branch increments the RC number** (rc1 → rc2 → rc3...).

---

### 4. Production Release (`prod-release.yaml`)
**Trigger**: PR closed (merged) to `main` from `rc/vX.Y.Z` branch

**Gate**: PR must be merged AND head branch must match `rc/vX.Y.Z`.

**Steps**:
1. Extract version from RC branch name
2. Get previous production tag for changelog range
3. Get RC base commit (`merge-base dev rc-branch`) for changelog scope
4. Create annotated git tag `vX.Y.Z` on `main`
5. Build production image tagged `vX.Y.Z` and `latest`
6. Generate full changelog: categorise PRs merged to `dev` since last release by branch type
7. Dispatch `production-environment-update` to GitOps repo
8. Create GitHub Release (not pre-release) with full changelog
9. Create sync PR from RC branch → `dev` (or auto-merge if no conflicts)

**Changelog categories**: 🚨 Breaking, ✨ Features, 🐛 Bug Fixes, 📚 Docs, 🔧 Maintenance.

---

### 5. Hotfix (`hotfix-release.yaml`)
**Trigger (build)**: Push to `hotfix/*`
**Trigger (deploy)**: PR merged to `main` from `hotfix/*`

**On push to `hotfix/*`**:
1. Calculate hotfix version: latest `vX.Y.Z` prod tag → increment patch → `vX.Y.(Z+1)`
2. Build image: `vX.Y.Z-hotfix-{sha7}`
3. Create (or update) PR to `main` with labels `hotfix`, `critical`, `expedited`

**On merge to `main`**:
1. Create annotated production tag `vX.Y.Z` (from PR title or recalculated)
2. Build final production image tagged `vX.Y.Z` and `latest`
3. Dispatch `hotfix-production-update` to GitOps repo (expedited flag)
4. Create GitHub Release
5. Create sync PR from `main` → `dev`

**Important**: Hotfix skips staging. It goes directly to production. Always create sync PR promptly.

---

### 6. Manual RC Creation (`manual-create-rc.yaml`)
**Trigger**: `workflow_dispatch` from `dev` branch only

**Steps**:
1. Validate execution is from `dev`
2. Get version: from optional `custom_version` input OR from latest `dev-v*` tag
3. Create branch `rc/vX.Y.Z` from current `dev` HEAD
4. Create PR from `rc/vX.Y.Z` → `main`

**Use this when**: You want to promote `dev` to staging without the `create-rc🚀` label flow.

---

### 7. Stale Branch Cleanup (`cleanup-stale-branches.yaml`)
**Trigger**: Daily cron at midnight + manual dispatch

**Cleaned prefixes**: `feature/`, `bugfix/`, `bug/`, `chore/`, `maintenance/`, `rc/v`

**Rules**:
- Skip if branch has open PRs
- Skip if last commit < 7 days ago
- Skip protected branches: `main`, `dev`, `master`
- Creates a GitHub issue summary when branches are deleted

---

## Release Flow Decision Tree

```
New work needed?
├── Bug/feature/maintenance → Create branch from dev → PR to dev
│   ├── Want preview? → Add `preview` label to PR
│   ├── Merge → dev build + tag created
│   └── Want RC now? → Add `create-rc🚀` label before merge
│
├── Ready to stage/release?
│   ├── With label: automatic rc/vX.Y.Z branch + PR created
│   ├── Without label: run manual-create-rc workflow or `git checkout -b rc/vX.Y.Z`
│   └── RC branch pushed → staging deploy triggered automatically (rc-release.yaml)
│
├── RC approved? → Merge rc/vX.Y.Z PR to main → prod-release.yaml fires
│
└── Critical prod bug?
    └── Create hotfix/* from main → hotfix-release.yaml fires → PR to main → deploy
```

---

## Validation Checklist — Before Production Release

- [ ] `vX.Y.Z-rcN` image exists in registry
- [ ] Staging deployment is `success` (GitHub Deployment status)
- [ ] RC tracking issue checklist is complete
- [ ] RC PR to `main` has no failing checks
- [ ] No blocking comments on RC PR
- [ ] Reviewer approval on RC PR

## Validation Checklist — Before Hotfix Production Merge

- [ ] `hotfix/*` image built and pushed
- [ ] PR to `main` has `hotfix`, `critical` labels
- [ ] Reviewer approval
- [ ] Rollback plan documented in PR body
- [ ] Sync PR to `dev` will be created automatically post-merge

---

## Do NOT

- Push directly to `main` or `dev`
- Delete `main`, `dev`, or `master` branches
- Create RC branches with names other than `rc/vX.Y.Z`
- Skip the GitOps dispatch — it is the deployment trigger
- Manually tag `vX.Y.Z` unless the pipeline fails and a recovery tag is needed
- Merge RC to main before staging deployment succeeds
- Use `gh pr merge --admin` — branch protection rules cannot be bypassed, not even by admins; the flag will fail
- Use `gh pr merge --auto` without first verifying auto-merge is enabled for the repo
- Add the `create-rc🚀` label without asking the user first — RC is not always the desired next step
- Try to add a label that doesn't exist — always create it first with `gh label create`

---

## Recovery Patterns

**RC pipeline didn't fire**: Push an empty commit to the `rc/vX.Y.Z` branch to retrigger.
```bash
git commit --allow-empty -m "chore: retrigger RC pipeline"
git push origin rc/vX.Y.Z
```

**Stale RC branch needs cleanup**: Ensure no open PRs, then it will be cleaned by nightly job.
Or delete manually: `git push origin --delete rc/vX.Y.Z`

**Sync PR conflicts**: Resolve on `rc/vX.Y.Z` branch locally, push to branch, re-trigger sync.

**Dev tag missing**: Recalculate and push manually:
```bash
git tag -a "dev-vX.Y.Z-YYYYMMDD-HHMMSS" -m "Dev release vX.Y.Z" <sha>
git push origin <tag>
```

---

## Environment URLs

| Environment    | URL pattern                                              |
|----------------|----------------------------------------------------------|
| Preview        | `https://{PR_NUMBER}-{app}.preview.{KUBECORE_DOMAIN}`   |
| Staging        | `https://staging.{app}.{KUBECORE_DOMAIN}` (ArgoCD managed) |
| ArgoCD         | `https://argocd.{KUBECORE_DOMAIN}`                       |
| Production     | Defined in KubeProject GitOps config                     |

---

## Secrets & Variables Required per App Repo

| Name                        | Type     | Description                          |
|-----------------------------|----------|--------------------------------------|
| `KUBECORE_APP_ID`           | Secret   | GitHub App ID for bot token          |
| `KUBECORE_APP_PKEY`         | Secret   | GitHub App private key               |
| `KUBECORE_REGISTRY_UNAME`   | Secret   | Container registry username          |
| `KUBECORE_REGISTRY_PWD`     | Secret   | Container registry password          |
| `KUBECORE_REGISTRY`         | Variable | Full registry path (host/org/app)    |
| `KUBECORE_KUBEPROJECT_REPO` | Variable | GitOps repo name (owner/repo)        |
| `KUBECORE_KUBEPROJECT_DOMAIN` | Variable | Base domain for environment URLs   |
