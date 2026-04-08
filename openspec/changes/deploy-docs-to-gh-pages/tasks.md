## 1. Documentation Build

- [ ] 1.1 Install documentation dependencies with `npm install`
- [ ] 1.2 Build the VuePress documentation using `npx vuepress build docs`
- [ ] 1.3 Verify the `docs/.vuepress/dist/` directory exists and contains static files

## 2. GitHub Pages Branch Preparation

- [ ] 2.1 Create the `gh-pages` branch if it doesn't already exist
- [ ] 2.2 Add a `.nojekyll` file to the `docs/.vuepress/dist/` directory

## 3. Deployment to gh-pages

- [ ] 3.1 Use `git subtree push --prefix docs/.vuepress/dist/ origin gh-pages` to deploy the static site to the `gh-pages` branch
- [ ] 3.2 Verify the `gh-pages` branch on remote contains only the static files at the root
- [ ] 3.3 Verify the documentation website at `https://strings77wzq.github.io/go-flashdb/` is no longer returning 404
