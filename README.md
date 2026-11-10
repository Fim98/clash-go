# clash-go

用 Go 与 [MyGo](https://github.com/egoist/mygo) 写的 Clash 桌面客户端。

界面是 MyGo 的**原生 UI**（`ui` 包，GPU 直绘），没有前端工程、没有 webview、没有 npm 依赖：
一个几 MB 的纯 Go 单二进制，可交叉编译到 macOS / Windows / Linux。

## 现在能做什么

M1 已完成：应用把 mihomo 当库跑在自己的进程里，点「启动核心」就在 `7890` 起一个
混合端口（规则模式、`MATCH,DIRECT`，还没有订阅，所以流量直连），底部状态栏和仪表盘
实时显示上下行速率、流量曲线、运行时间、内存与连接数，右上角可以切规则/全局/直连，
模式会同步到内核。

代理、订阅、连接、规则、日志页面已经按设计稿排好版，但数据还是示例值 —— 它们分别在
M3、M2、M4 接上真实数据。进度见 [docs/roadmap.md](docs/roadmap.md)。

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

每次构建都用 `updates.publicKey` 对应的私钥（`MYGO_UPDATER_PRIVATE_KEY` secret）签名，
所以发布里除了安装包还有 `update-<os>-<arch>.json` 清单、跨版本的 delta 更新，
以及 `install.sh` / `install.ps1`。

## 安装：绕过 Gatekeeper 与 SmartScreen

没有付费签名证书时，从浏览器下载的 `.dmg` 会被 macOS 标记为来自网络，直接打开会弹
「无法验证开发者」。三条免费的路，体验从好到差：

1. **一行安装（推荐，macOS / Linux）**

   ```sh
   curl -fsSL https://github.com/Fim98/clash-go/releases/latest/download/install.sh | sh
   ```

   `curl` 下载的文件不带 quarantine 属性，Gatekeeper 根本不会评估签名，装进
   `/Applications` 后点开就能用。脚本会按清单里的 SHA-256 与 Ed25519 签名校验
   安装包，安装路径和普通 app 一样。
2. **下载 `.dmg` 后手动放行**：拖动安装，第一次打开被拦 → 「系统设置 → 隐私与安全性 →
   仍要打开」。macOS 15 起 Apple 取消了右键「打开」的绕过方式，只能这样点。
3. **自己编**：`make build`，本机编译出的二进制没有 quarantine，直接能跑。

Windows 同理：签名的 exe 才会让 SmartScreen 快速闭嘴，未签名的要么点
「更多信息 → 仍要运行」，要么等下载量积累出信誉。

彻底解决只有一条路：Apple Developer Program（99 美元/年）的 Developer ID 证书 + notarize。
证书绑在你自己的 Team 下，私钥不能外借，所以只能你自己申请；申请下来后把证书和
`APPLE_*` 密码存成 secrets，`release.yml` 里加回 `-sign` / 去掉 `-skip-notarize` 即可，
`mygo` 会自动签名、公证、staple。Windows 侧可以先试
[SignPath Foundation](https://signpath.org/)（对开源项目免费）或 Azure Trusted Signing（约 10 美元/月）。

`.github/dependabot.yml` 每周检查 Go 依赖和 Actions 版本，并开一个 `chore:` / `ci:` 的 PR。

## 目录

```
main.go              应用入口、窗口、内核轮询与生命周期
types.go             界面状态模型与示例数据
components.go        卡片、区块标题、延迟徽章等小部件
view_*.go            侧栏外壳与各页面
main_test.go         用 ui.NewTester 无窗口点按/输入
internal/core        mihomo 的启动/重载/停止、状态快照、模式切换
docs/roadmap.md      里程碑与验收标准
mygo.json            应用名、identifier、版本、打包输出
resources/           随应用发布的资源（icon.png 等）
.githooks/           commit-msg / pre-commit / pre-push
scripts/             提交信息校验（hook 与 CI 共用）
.github/workflows/   CI 与发布
```

## 还没做

- **签名**：目前 macOS 是 ad-hoc 签名、Windows 无签名，靠上面的 `install.sh` 绕开
  Gatekeeper。有证书后按 [distribution.md](https://github.com/egoist/mygo/blob/main/docs/distribution.md)
  加 `macos.signingIdentity`、`macos.notarize`、`windows.certificate` 与
  `MACOS_CERTIFICATE`、`APPLE_*`、`WINDOWS_CERTIFICATE*` 这些 secrets。
- **应用内更新**：清单已经在发了，但 `main.go` 还没调 `mygo.Updater`；接上之后
  应用能自己下载签名过的更新并替换自己。更新说明取自 `CHANGELOG.md` 里对应版本的
  `## 0.1.0` 小节（文件一旦存在，每个版本都必须有这一节，否则构建会停）。
- **更新私钥**：`~/Library/Application Support/mygo/update-keys/mygo-update.key`，
  同时存在 GitHub secret 里。放进密码管理器备份：丢了就再也没法给已安装的用户发更新了。
- **界面真数据**：代理/订阅/连接/规则/日志四个页面现在展示示例数据（M2–M4）。
- **系统代理 / TUN**：开关已经画出来了，但还没有接系统设置和提权 helper（M5）。
- **identifier**：`io.github.fim98.clashgo` 一旦发布就不要再改（系统按它存偏好、权限、注册项）。