import { GitContributors } from "/home/strin/go/src/devLearn/aiLab/goflashdb/node_modules/@vuepress/plugin-git/dist/client/components/GitContributors.js";
import { GitChangelog } from "/home/strin/go/src/devLearn/aiLab/goflashdb/node_modules/@vuepress/plugin-git/dist/client/components/GitChangelog.js";

export default {
  enhance: ({ app }) => {
    app.component("GitContributors", GitContributors);
    app.component("GitChangelog", GitChangelog);
  },
};
