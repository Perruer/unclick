# Publishing checklist

Nothing here is published automatically. Steps, in order:

1. Create the public repository `Perruer/unclick` (empty, no README).
2. `git remote add origin https://github.com/Perruer/unclick.git`
3. `git push -u origin main` (never push the upstream `0.8.x` tags).
4. Repository settings:
   - Description: *Turn existing cloud resources into OpenTofu / Terraform code. Continuation of Terraformer: protocol 5 and 6 providers, OpenTofu, import blocks, provider-driven scan.*
   - Website: leave empty until there is a docs site.
   - Topics: `terraform`, `opentofu`, `infrastructure-as-code`, `iac`, `terraformer`, `reverse-terraform`, `import`, `aws`, `gcp`, `azure`, `kubernetes`, `devops`, `cli`, `golang`.
   - Features: tick **Sponsorships** so FUNDING.yml shows the Sponsor button.
   - Social preview: upload [social-preview.png](social-preview.png) (1280×640; source in social-preview.svg).
5. Wait for the `tests` workflow to pass on `main` (unit, acceptance, e2e-aws, e2e-kubernetes).
6. `git tag -a v1.0.0 -m "Unclick 1.0.0"` and `git push origin v1.0.0`. The `release` workflow builds the binaries and opens a **draft** release.
7. Paste [github-release-v1.0.0.md](github-release-v1.0.0.md) into the draft, check the assets and checksums, publish.
8. Optional: a Homebrew tap (`Perruer/homebrew-tap`) and a Scoop bucket; GoReleaser can update both once they exist.
