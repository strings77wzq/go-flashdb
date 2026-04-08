## Why

The current documentation deployment for `go-flashdb` is failing with a 404 error because the GitHub Pages configuration is set to "Deploy from a branch" (main branch, /docs folder), but the `/docs` folder contains VuePress source files instead of built static HTML. To align with the successful deployment pattern of `claude-code-Go`, we will switch to deploying from a dedicated `gh-pages` branch.

## What Changes

- Switch GitHub Pages deployment source from `main` branch to `gh-pages` branch.
- Implement a localized build and deploy strategy (Manual or script-based for now, similar to the user's preferred "Plan A").
- Create or update the `gh-pages` branch to contain only the built static assets.
- Ensure the documentation is accessible at `https://strings77wzq.github.io/go-flashdb/`.

## Capabilities

### New Capabilities
- `gh-pages-deployment`: Capability to build VuePress docs locally and push the static output to a dedicated `gh-pages` branch for hosting.

### Modified Capabilities
- None

## Impact

- GitHub Repository Settings: Pages deployment source will change.
- Git Branching: A new `gh-pages` branch will be managed.
- Documentation workflow: Developers will need to build and push to `gh-pages` to update the live site.
