# Tene CLI Release & Deployment Guide

## Release Checklist

### Before Releasing

1. `go test ./...` — all tests pass
2. `go build ./cmd/tene` — build succeeds
3. `golangci-lint run` — no lint issues

### CLI Release (GoReleaser → S3 + GitHub Releases)

```
git tag v0.x.x
git push origin v0.x.x
```

This triggers `.github/workflows/release.yml`:
1. GoReleaser builds multi-platform binaries (darwin/linux x amd64/arm64 + windows/amd64)
2. Uploads to S3 (`tene-releases` bucket) and GitHub Releases (dual publish)
3. Updates `LATEST_VERSION` file on S3

### Landing Page (Next.js → Vercel)

- `git push origin main` → Vercel auto-deploy
- Root Directory: `apps/web`
- Domain: `tene.sh`

## S3 Release Infrastructure

| Item | Value |
|------|-------|
| Bucket | `tene-releases` |
| Region | `ap-northeast-2` |
| Access | Public Read (s3:GetObject) |
| Auth | OIDC (GitHub Actions → IAM role) |

```
s3://tene-releases/
├── LATEST_VERSION           # "0.3.1" (text, max-age=60)
├── v0.3.1/
│   ├── tene_0.3.1_darwin_amd64.tar.gz
│   ├── tene_0.3.1_darwin_arm64.tar.gz
│   ├── tene_0.3.1_linux_amd64.tar.gz
│   ├── tene_0.3.1_linux_arm64.tar.gz
│   ├── tene_0.3.1_windows_amd64.zip
│   └── checksums.txt
└── ...
```

## Update Flow

- `install.sh`: Downloads from S3, verifies SHA-256 checksum
- `tene update`: Checks S3 first, falls back to GitHub API. Verifies checksums.

## GoReleaser Snapshot (Dry-run)

```bash
goreleaser release --snapshot --clean
```
