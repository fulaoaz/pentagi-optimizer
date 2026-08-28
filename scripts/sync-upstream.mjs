import { spawnSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repositoryDirectory = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const checkOnly = process.argv.includes("--check");
const remoteArgument = process.argv.find((argument) =>
  argument.startsWith("--remote="),
);
const branchArgument = process.argv.find((argument) =>
  argument.startsWith("--branch="),
);
const remote = remoteArgument?.slice("--remote=".length) || "origin";
const upstreamBranch = branchArgument?.slice("--branch=".length) || "main";
const upstreamRef = `${remote}/${upstreamBranch}`;

const run = (command, arguments_, options = {}) => {
  const result = spawnSync(command, arguments_, {
    cwd: options.cwd ?? repositoryDirectory,
    encoding: "utf8",
    shell: process.platform === "win32",
    stdio: options.capture ? "pipe" : "inherit",
  });

  if (result.status !== 0 && !options.allowFailure) {
    process.exit(result.status ?? 1);
  }

  return options.capture ? result.stdout.trim() : "";
};

const currentBranch = run("git", ["branch", "--show-current"], {
  capture: true,
});

if (currentBranch !== "main") {
  console.error(
    `请在 main 分支运行此脚本（当前分支：${currentBranch || "游离 HEAD"}）。`,
  );
  process.exit(1);
}

run("git", ["fetch", remote, upstreamBranch, "--prune"]);

const behind = Number(
  run("git", ["rev-list", "--count", `HEAD..${upstreamRef}`], {
    capture: true,
  }),
);

if (behind === 0) {
  console.log(`中文默认的 main 已与 ${upstreamRef} 保持同步。`);
  process.exit(0);
}

console.log(`${upstreamRef} 有 ${behind} 个新提交。`);

if (checkOnly) {
  process.exit(0);
}

const worktreeStatus = run("git", ["status", "--porcelain"], { capture: true });

if (worktreeStatus) {
  console.error("同步上游前，请先提交或暂存当前工作区改动。");
  process.exit(1);
}

run("git", ["merge", "--no-ff", "--no-commit", upstreamRef]);

const i18nCheck = spawnSync("node", ["scripts/check-ui-i18n.mjs"], {
  cwd: path.join(repositoryDirectory, "frontend"),
  encoding: "utf8",
  shell: process.platform === "win32",
  stdio: "inherit",
});

if (i18nCheck.status !== 0) {
  console.error(
    "上游改动已合并但尚未提交。请翻译或复核报告中的前端文案，再完成提交。",
  );
  process.exit(i18nCheck.status ?? 1);
}

const providerTestLabelsCheck = spawnSync(
  "pnpm",
  ["exec", "vitest", "run", "src/pages/settings/provider-test-labels.test.ts"],
  {
    cwd: path.join(repositoryDirectory, "frontend"),
    encoding: "utf8",
    shell: process.platform === "win32",
    stdio: "inherit",
  },
);

if (providerTestLabelsCheck.status !== 0) {
  console.error(
    "上游改动已合并但尚未提交。请为新增的提供商测试名称补充中文标签，再完成提交。",
  );
  process.exit(providerTestLabelsCheck.status ?? 1);
}

const installerI18nCheck = spawnSync(
  "go",
  ["test", "./cmd/installer/wizard/locale"],
  {
    cwd: path.join(repositoryDirectory, "backend"),
    encoding: "utf8",
    shell: process.platform === "win32",
    stdio: "inherit",
  },
);

if (installerI18nCheck.status !== 0) {
  console.error(
    "上游改动已合并但尚未提交。请翻译或复核报告中的安装器文案，再完成提交。",
  );
  process.exit(installerI18nCheck.status ?? 1);
}

const apiResponseI18nCheck = spawnSync(
  "go",
  ["test", "./pkg/server/response"],
  {
    cwd: path.join(repositoryDirectory, "backend"),
    encoding: "utf8",
    shell: process.platform === "win32",
    stdio: "inherit",
  },
);

if (apiResponseI18nCheck.status !== 0) {
  console.error(
    "上游改动已合并但尚未提交。请复核 REST 与 GraphQL 的中文错误响应，再完成提交。",
  );
  process.exit(apiResponseI18nCheck.status ?? 1);
}

const graphqlResponseI18nCheck = spawnSync(
  "go",
  [
    "test",
    "./pkg/server/services",
    "-run",
    "^TestLocalizedGraphQLErrorPresenter$",
  ],
  {
    cwd: path.join(repositoryDirectory, "backend"),
    encoding: "utf8",
    shell: process.platform === "win32",
    stdio: "inherit",
  },
);

if (graphqlResponseI18nCheck.status !== 0) {
  console.error(
    "上游改动已合并但尚未提交。请复核 GraphQL 的中文错误响应，再完成提交。",
  );
  process.exit(graphqlResponseI18nCheck.status ?? 1);
}

const regressionCheck = spawnSync(
  "go",
  ["test", "./pkg/config", "./pkg/tools", "./pkg/providers/embeddings"],
  {
    cwd: path.join(repositoryDirectory, "backend"),
    encoding: "utf8",
    shell: process.platform === "win32",
    stdio: "inherit",
  },
);

if (regressionCheck.status !== 0) {
  console.error(
    "上游改动已合并但尚未提交。请复核配置、搜索工具和嵌入模型回归测试，再完成提交。",
  );
  process.exit(regressionCheck.status ?? 1);
}

run("git", ["diff", "--check"]);

run("git", ["commit", "--no-edit"]);
console.log(`${upstreamRef} 已合并；推送前请运行完整测试。`);
