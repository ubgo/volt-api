# GitHub Actions deploy over NetBird VPN — technical summary

Use this document as context when applying the same pattern to another repository: CI joins a NetBird network, then SSH (or other access) targets the VPS only on the private overlay—no need to expose SSH to the public internet.

## Problem and approach

- **Risk:** Opening TCP 22 on a public IP increases attack surface.
- **Approach:** Run NetBird on the VPS and in the workflow. Each Actions run registers a short‑lived peer, connects to the overlay, then uses **`VPS_HOST` set to the VPS’s NetBird address** (WireGuard / Magic DNS name, depending on your NetBird setup) for SSH.

Flow:

1. Runner executes **NetBird connect** (setup key).
2. Runner waits until management connection is up (the connect action loops; optional extra `netbird status` / `ping` for logs).
3. Runner runs **SSH deploy** (e.g. `appleboy/ssh-action`) to `VPS_HOST` over the tunnel.

Reference workflow shape (public example): [khanakia/netbird-github-action-test](https://github.com/khanakia/netbird-github-action-test).  
Connect action: [Alemiz112/netbird-connect](https://github.com/Alemiz112/netbird-connect) (`Alemiz112/netbird-connect@v1`).

## NetBird dashboard setup

1. **Setup key** (Dashboard → Setup Keys → Create):
   - Prefer **Reusable** (each CI run = new peer).
   - Prefer **Ephemeral** so peers auto‑expire after inactivity (~10 minutes) and do not clutter the dashboard.
   - Assign to a **group** whose ACLs allow reaching the VPS peer on **SSH** (port 22 or your custom SSH port).
2. Ensure the **VPS** runs the NetBird client and is online on the same network.
3. Note the **management URL** if self‑hosted; the action defaults to NetBird Cloud (`management-url` input on the action if needed).

## GitHub secrets

| Secret | Purpose |
|--------|---------|
| `NETBIRD_SETUP_KEY` | Setup key string from NetBird (CI only; never ship to app containers). |
| `VPS_HOST` | Address SSH should use **after** VPN is up (NetBird IP or DNS name for the VPS peer—not necessarily the old public IPv4 if 22 is closed publicly). |
| `VPS_USER`, `VPS_PORT` | SSH user and port (often `22`). |
| SSH private key | Repo-specific naming: e.g. raw `VPS_SSH_KEY`, or base64 `VPS_DEPLOY_KEY_B64` if your generator encodes keys. |

Add `NETBIRD_SETUP_KEY` to **each** GitHub Environment (or repo secrets) that runs deploy, if environments are split (prod/stag).

## Workflow YAML changes (generic)

Insert **before** the SSH/deploy step in the `deploy` job (same job that runs `ssh-action` or raw `ssh`):

```yaml
      - name: Connect to NetBird
        uses: Alemiz112/netbird-connect@v1
        with:
          setup-key: ${{ secrets.NETBIRD_SETUP_KEY }}
          hostname: 'gh-actions-<repo-or-env>-${{ github.run_id }}'
          # management-url: 'https://your-management:443'   # only if self-hosted

      - name: NetBird connectivity check
        run: |
          sudo netbird status || true
          ping -c 2 ${{ secrets.VPS_HOST }} || echo "ICMP may be blocked; continuing to SSH..."
```

Replace `<repo-or-env>` with a stable label (e.g. `myapp-prod`) so peers are identifiable in the NetBird UI.

**Requirements:**

- **Linux** runners only (the action enforces this).
- Connect step must run in the **same job** as SSH so the tunnel exists on that runner.

## Action version baseline (refresh periodically)

Pin majors you trust; bump when maintainers release new majors (Node 24 / runner compatibility is usually documented on each release).

| Action | Typical current major (check releases) |
|--------|----------------------------------------|
| `actions/checkout` | v6 |
| `docker/setup-qemu-action` | v4 |
| `docker/setup-buildx-action` | v4 |
| `docker/login-action` | v4 |
| `docker/metadata-action` | v6 |
| `docker/build-push-action` | v7 |
| `appleboy/ssh-action` | v1.2.x |
| `Alemiz112/netbird-connect` | v1 |

Verify on each action’s GitHub **Releases** page before upgrading across major versions.

## Repo-specific integration patterns

### A) Multi-environment Go monorepo (`secrets_build` + `deploy.yml.gotmpl`)

- **Source of truth for workflow:** `.github/workflows/deploy.yml.gotmpl` (delimiters `[[ ]]` to avoid clashing with `${{ }}`).
- **Per-env files:** `.env.<env>` with `GITHUB_SECRET_*` vars; `GITHUB_SECRET_NETBIRD_SETUP_KEY` becomes secret `NETBIRD_SETUP_KEY` and is **excluded** from container `ENV_B64`.
- **Regenerate:** `go run ./scripts gh:secrets:update` (or project’s equivalent).
- **Hostname in template:** e.g. `gh-actions-[[ .Name ]]-${{ github.run_id }}`.

### B) Single generated `deploy.yml` + `.secrets.prod`

- Edit **`.github/workflows/deploy.yml.gotmpl`**, then run the project’s **secrets update** command to rewrite `deploy.yml`.
- Document `NETBIRD_SETUP_KEY` in **`.secrets.example`** (or equivalent).
- Add the key to **`.secrets.prod`** (often gitignored) before `gh secrets push`.

### C) Raw `deploy.yml` only

- Patch `deploy.yml` directly; duplicate the NetBird steps in every workflow file that SSHes to the VPS.

## Troubleshooting (short)

- **NetBird connect fails:** Expired or wrong setup key; group not allowed by ACL; management URL wrong for self‑hosted.
- **SSH timeout after VPN “up”:** `VPS_HOST` still set to public IP while firewall blocks 22; use NetBird/Magic DNS target. Confirm `netbird status` on the VPS.
- **SSH auth fails:** Key secret malformed (include `BEGIN`/`END` for PEM); authorized_keys on server.

---

## Copy-paste prompt for other repositories

Use the block below as LLM or human context when asking for the same update in another repo.

```text
Goal: GitHub Actions deploy should reach the VPS over NetBird VPN instead of relying on publicly reachable SSH.

Requirements:
1. In the deploy job(s) that SSH to the VPS, BEFORE any ssh-action or ssh command, add:
   - Alemiz112/netbird-connect@v1 with setup-key from GitHub secret NETBIRD_SETUP_KEY and a unique hostname per run (e.g. gh-actions-<app>-${{ github.run_id }}).
   - Optional diagnostic step: sudo netbird status; ping VPS_HOST (non-fatal if ICMP blocked).
2. Document that VPS_HOST must be the address reachable over NetBird (peer IP or Magic DNS), not necessarily the public IP if SSH is closed on the public interface.
3. I will add NETBIRD_SETUP_KEY to GitHub (environment or repo secrets). Same reusable+ephemeral setup key can be shared across repos if NetBird ACLs allow those CI peers to reach each VPS.
4. Align other actions (checkout, docker/setup-qemu-action, docker/setup-buildx-action, docker/login-action, docker/metadata-action, docker/build-push-action, appleboy/ssh-action) to current stable majors per their GitHub Releases pages.
5. If the repo uses a deploy template (e.g. deploy.yml.gotmpl) and a script regenerates deploy.yml, edit the template only and run the regeneration command; do not hand-edit generated YAML.
6. Do not commit real secrets; use placeholders in example env/secrets files only.

Adapt file paths and secret names (e.g. VPS_SSH_KEY vs base64 SSH key) to this repository’s existing conventions.
```
