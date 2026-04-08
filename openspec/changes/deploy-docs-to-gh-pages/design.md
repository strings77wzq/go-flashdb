## Context

The `go-flashdb` documentation currently serves the root `docs/` folder of the `main` branch, which contains source Markdown files. This results in a 404 error because GitHub Pages expects rendered HTML. We are adopting "Plan A": building docs locally (or via script) and deploying only the static output to a dedicated `gh-pages` branch, mimicking the `claude-code-Go` repository structure.

## Goals / Non-Goals

**Goals:**
- Successfully serve the VuePress documentation at `https://strings77wzq.github.io/go-flashdb/`.
- Switch deployment source to the `gh-pages` branch.
- Maintain the `main` branch for source code and documentation source files only.

**Non-Goals:**
- Implementing a full CI/CD pipeline via GitHub Actions (unless specifically asked, as "Plan A" focuses on the branch deployment method).
- Redesigning the current documentation content.

## Decisions

- **Branch Selection**: Use `gh-pages` branch for static assets. This is the community standard and aligns with the user's reference.
- **Build Tool**: Use `vuepress build docs` to generate the `docs/.vuepress/dist/` directory.
- **Deployment Folder**: The content of `docs/.vuepress/dist/` will be pushed to the root of the `gh-pages` branch.
- **No Jekyll**: Add a `.nojekyll` file to the root of `gh-pages` to prevent GitHub from ignoring files starting with an underscore (like VuePress's `_assets`).

## Risks / Trade-offs

- [Risk] → `gh-pages` branch might get out of sync with `main`. Mitigation: Perform a build and push whenever documentation in `main` is updated.
- [Risk] → Manual deployment error. Mitigation: Use a simple script or precise `git` commands to automate the subtree push.
- [Trade-off] → Larger repository size due to committing build artifacts in a separate branch. This is acceptable for the benefit of a working documentation site.
