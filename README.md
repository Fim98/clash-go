# clash-go

用 Go 与 [MyGo](https://github.com/egoist/mygo) 写的 Clash 桌面客户端。

界面是 MyGo 的**原生 UI**（`ui` 包，GPU 直绘），没有前端工程、没有 webview、没有 npm 依赖：
一个几 MB 的纯 Go 单二进制，可交叉编译到 macOS / Windows / Linux。

## 环境

- Go 1.27.3 及以上。`go.mod` 的 `tool` 指令已经固定了 mygo CLI 的版本，
  所以 `go tool mygo` 随时可用，不用额外安装。

## 常用命令

```sh
make setup   # 一次性：启用 git hooks、拉取依赖（克隆后先跑这个）
make dev     # 开发模式运行，改代码自动重建
make test    # CGO_ENABLED=0 go test ./...
make vet     # go vet + go tool mygo vet
make check   # fmt + vet + test，和 CI 跑的一致
make build   # 打包 PLATFORMS 里的目标到 build/
```

直接记 Go 命令也可以：`go tool mygo dev`、`go test ./...`、`go tool mygo build -platform darwin/arm64`。

## 提交规范

采用 [Conventional Commits](https://www.conventionalcommits.org/zh-hans/)：

```
<type>(<scope>): <subject>
```

- `type`：`build` `chore` `ci` `docs` `feat` `fix` `perf` `refactor` `revert` `style` `test`
- `scope`：可选，小写，例如 `feat(window): ...`、`fix(updater): ...`
- 破坏性变更在 `:` 前加 `!`，例如 `feat(api)!: drop the v0 config`
- subject 用祈使句、不加句号，整行不超过 100 字符

规则只有一份实现：`scripts/check-commit-message.sh`。`commit-msg` hook 和 CI 都调它，
所以本地和远端不会各说各话。

## Git hooks

不需要 husky / commitlint / Node：直接用 git 自己的 `core.hooksPath`，脚本在 `.githooks/` 里，
零依赖、跨平台、跟着仓库走。

| hook | 做什么 |
| --- | --- |
| `commit-msg` | 按上面的规范校验提交信息（`Merge` / `Revert` / `fixup!` 放行） |
| `pre-commit` | 暂存的 Go 文件必须 gofmt 过；跑 `go vet ./...` 与 `go tool mygo vet .` |
| `pre-push` | 跑 `go test ./...` |

启用（每个克隆一次，`make setup` 已包含）：

```sh
git config core.hooksPath .githooks
```

hooks 只是本地顺手，**CI 才是保证**：忘了配 hooks 也拦不住坏提交，但 PR 上的
`Commit messages` 任务会拦。确实需要绕过时用 `git commit --no-verify`。

## CI 与发布

`.github/workflows/ci.yml`（推到 `main` 或开 PR 时）：

- **Commit messages**：校验 PR 里每个提交的标题，以及 PR 标题（squash 合并后它就是提交标题）
- **Format and vet**：`gofmt -l`、`go vet ./...`、`go tool mygo vet .`
- **Test**：Ubuntu 与 macOS 各跑一遍 `go test ./...`
- **Build**：macOS / Windows / Linux 三个平台行，验证每个目标都还能编出来

`.github/workflows/release.yml`（推 tag 时）：

```sh
# 1. 改 mygo.json 里的 version
# 2. 提交，然后
git tag v0.1.0
git push origin v0.1.0
```

tag 必须等于 `v` + `mygo.json` 的 `version`（工作流第一步会校验），否则不出包。
三个平台并行构建，产物上传到一个 draft release，最后自动发布；tag 带 `-`
（如 `v0.2.0-beta.1`）会标成 prerelease。

`.github/dependabot.yml` 每周检查 Go 依赖和 Actions 版本，并开一个 `chore:` / `ci:` 的 PR。

## 目录

```
main.go              应用入口与窗口
main_test.go         用 ui.NewTester 无窗口点按/输入
mygo.json            应用名、identifier、版本、打包输出
resources/           随应用发布的资源（icon.png 等）
.githooks/           commit-msg / pre-commit / pre-push
scripts/             提交信息校验（hook 与 CI 共用）
.github/workflows/   CI 与发布
```

## 还没做，但发布前要定

- **identifier**：现在是 `io.github.fim98.clashgo`，一旦发布就不要再改（系统按它存偏好、权限、注册项）。
- **签名**：目前是 ad-hoc 签名 / 无签名，别人下载会被 Gatekeeper、SmartScreen 拦。
  有 Apple Developer ID 与 Windows 证书后，按 [distribution.md](https://github.com/egoist/mygo/blob/main/docs/distribution.md)
  配置 `macos.signingIdentity`、`macos.notarize`、`windows.certificate`，再加 `MACOS_CERTIFICATE`、
  `APPLE_*`、`WINDOWS_CERTIFICATE*` 这些 secrets。
- **自动更新**：`go tool mygo keygen` 生成密钥，`mygo.json` 加 `updates`
  （`publicKey` + `github`），把私钥放进 `MYGO_UPDATER_PRIVATE_KEY` secret，
  然后把发布工作流里的 `gh release upload` 换成 `go tool mygo build ... -upload`，
  就能带上 Ed25519 签名、delta 更新和 `install.sh` / `install.ps1` 一行安装。