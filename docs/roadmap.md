# Roadmap

参照 [clash-verge-rev](https://github.com/clash-verge-rev/clash-verge-rev) 的功能面，
按「先能上网、再能看、再做系统集成」的顺序切里程碑。每个里程碑都能单独发布，
后面的没做完不影响前面的可用性。

## 三条原则

1. **核心编进二进制，不开子进程。** mihomo 当 Go 库用（已验证：v1.19.33，`CGO_ENABLED=0`，
   stripped 49 MB，启动 4 ms，配置热重载 4 ms，反复 start/stop 干净）。好处是单文件分发、
   没有 IPC、没有僵尸进程、能直接读内存里的连接表和流量计数；代价是升级核心要跟着发版，
   以及**整个项目必须是 GPL-3.0**（mihomo 的许可证）。
2. **UI 线程不碰核心的阻塞调用。** 拉延迟、读连接表这类可能耗时的动作一律在 goroutine 里做，
   结果用 `mygo.RunOnMain` 送回主线程；界面只读快照，不持有 mihomo 的对象。
3. **磁盘上的 YAML 是唯一真相。** UI 上的开关只是往 overrides 里写一行，然后重新合成
   runtime 配置并 `ApplyConfig`。所以用户可以随时手改文件，也可以外部工具同步。

## 里程碑总览

| # | 里程碑 | 一句话 | 版本 | 状态 |
|---|---|---|---|---|
| M0 | 工程骨架 | CI、hooks、签名更新、install.sh | 0.1.0 | ✅ 已完成 |
| M1 | 核心接线 | mihomo 跑在应用里，能开关、切模式、看流量 | 0.2.0 | 已完成 |
| M2 | 订阅与配置 | 粘贴机场链接就能上网，定时更新、多配置切换 | 0.3.0 | |
| M3 | 代理与延迟 | 分组、节点、测速、选择持久化 | 0.4.0 | |
| M4 | 可观测性 | 连接表、规则、日志 | 0.5.0 | |
| M5 | 系统集成 | 托盘、自启、系统代理、TUN（需要提权 helper） | 0.6.0 | |
| M6 | 打磨 | 备份恢复、i18n、应用内更新、正式签名 | 1.0.0 | |

---

## M0 工程骨架（已完成）

- [x] `git` + Conventional Commits：`.githooks/` + `scripts/check-commit-message.sh`，CI 同一份规则
- [x] GitHub Actions：PR 校验 / gofmt+vet / 双平台测试 / 三平台构建
- [x] 发布流水线：tag → 三平台并行 → Ed25519 签名更新 → draft → 自动发布
- [x] `install.sh` / `install.ps1` 一行安装，绕开 Gatekeeper 的 quarantine
- [x] README、Makefile、dependabot

## M1 核心接线（已完成）

**目标**：应用自己把 mihomo 跑起来，能启停、切模式、看到流量。没有 profile 的概念，
先用内置的一份最小配置验证链路。

- [x] `internal/core`：`Manager` 封装 mihomo 的全局单例
  - `New(dataDir)`、`Start(runtimeYAML)`、`Reload`、`Stop`、`Running`
  - 启动前置：`constant.SetHomeDir` + `constant.SetConfig` 指向应用数据目录，
    再 `config.Init`，最后 `executor.ParseWithBytes` + `executor.ApplyConfig(cfg, true)`
  - 数据目录 `~/Library/Application Support/clash-go`（Win `%AppData%`，Linux `~/.config`）
- [x] 数据快照：`Status()` 一次取回 `Now()`/`Total()`/`Snapshot()`，含速率、总量、连接数、内存
- [x] 内置最小配置：mixed-port 7890、`mode: rule`、`MATCH,DIRECT`，首次启动就能用
- [x] 界面外壳：侧栏（仪表盘/代理/订阅/连接/规则/日志/设置）+ 顶部模式切换 + 底部状态栏；
      仪表盘的上行/下行速率、流量曲线、运行时间、内存、连接数都来自真实快照，每秒刷新
- [x] 退出/停止时不残留端口：`executor.Shutdown()` 只关 TUN，入站监听器不会自己关，
      所以 `Stop` 先应用一份「所有端口为 0」的配置，让 `updateListeners` 把 mixed/http/
      socks 监听器都关掉，再 `Shutdown()`
- [ ] 其余页面（代理、订阅、连接、规则、日志）目前是设计稿数据，分别在 M3、M2、M4 接真数据
- [ ] 启动失败的错误要显示在界面上（端口被占用是最常见的）
- [ ] `go test ./internal/core`：启动 → 经 mixed-port 请求 → 切模式 → 停止，全程无泄漏

**验收**：`make dev` 起来后，终端里 `curl -x http://127.0.0.1:7890 https://example.com` 有响应；
点「停止」后 curl 立刻失败；再点「开始」又能通；关闭窗口进程干净退出。

**不做**：profile、系统代理、TUN、托盘。

## M2 订阅与配置

**目标**：用户粘一条订阅链接就能上网，并且配置能自己更新。

- [ ] profile 存储：`config/profiles/<uid>/`，元数据 `profiles.yaml`
- [ ] 导入：URL / 本地文件 / 剪贴板（`clash://install-config?url=`）
- [ ] 订阅更新：`interval`（默认 24h）、手动更新、`subscription-userinfo` 头解析
      （流量总额/已用/到期），UA 可配（很多机场按 UA 返回不同格式）
- [ ] 更新失败保留旧配置并在界面标红，不静默回滚到空规则
- [ ] 列表：激活/切换/删除/重排、最后更新时间、文件大小、来源域名
- [ ] 合成器：profile 原文 → 注入 overrides（mixed-port、allow-lan、mode、log-level、
      dns、sniffing）→ runtime YAML，写到 `config/<uid>.runtime.yaml`
- [ ] 覆盖层（merge）：用户自定义 rules / proxies / proxy-groups，持久化，切 profile 不丢
- [ ] 文本编辑器查看/修改 profile（只读 + 「另存为本地配置」两档）
- [ ] 切换 profile 时 `Reload`，失败自动回退到上一个可用配置

**验收**：粘贴订阅 → 3 秒内可用；改端口后自动重载且旧端口释放；断网时点更新有明确报错且原配置仍可用。

**不做**：脚本增强（`script`）、WebDAV。

## M3 代理与延迟

**目标**：看得清每个分组、每个节点，选得动，测得了速。

- [ ] 分组渲染：`select` / `url-test` / `fallback` / `load-balance` / `relay`，
      组内节点 + 嵌套组，展示当前选中、类型图标、`udp` 标记
- [ ] 延迟测试：单节点 / 整组 / 全部，并发 8~16，超时 5s，结果写 `DelayHistory`
- [ ] 列表状态：延迟色阶（<200ms 绿 / <500ms 黄 / 超时红）、按延迟排序、
      搜索过滤、隐藏不可用节点
- [ ] 选择：点节点即 `SelectAble.Set`，并且落盘「组名 → 节点名」，
      重启/订阅更新后尽量恢复（节点消失则回落到组默认）
- [ ] 组顺序稳定：以配置里的顺序为准，核心返回的顺序不覆盖它
- [ ] 代理提供者（`proxy-providers`）单独展示，支持手动刷新
- [ ] 全局模式下的节点列表单独一屏

**验收**：500 个节点整组测速时界面不卡；杀掉应用重开，选中项还在；订阅更新后新增节点出现在组里。

## M4 可观测性

**目标**：出问题时能自己诊断。

- [ ] 连接表：主机、类型(TCP/UDP)、规则、链路（`链 → 链`）、上传/下载、时长、进程名，
      每秒刷新；关单个 / 关全部
- [ ] 连接详情：进程路径、sniff 结果、DNS 模式、源地址
- [ ] 规则页：规则列表、规则提供者、当前匹配模式；高亮命中的规则
- [ ] 日志：`log.Subscribe` 推流，级别过滤、关键字搜索、暂停/清屏、复制
- [ ] 流量图：`ui.Chart` 画最近 60s 上下行
- [ ] 诊断：端口占用、DNS 解析、连通性自检，一键复制诊断信息

**验收**：打开一个视频页面，连接表能立刻看到对应条目并能关掉；日志页在 10k 行时滚动不卡。

## M5 系统集成

**目标**：从「一个能上网的窗口」变成「系统里的一个代理」。

- [ ] 托盘：图标 + 模式切换 + 系统代理开关 + TUN 开关 + 退出；关窗口进托盘
- [ ] 单实例 + `clash://` / `mihomo://` 深链导入
- [ ] 开机自启（macOS `SMAppService`/Login Items、Windows 注册表、Linux autostart）
- [ ] **提权 helper**（一次性设计，两个功能共用）
  - macOS：`launchd` daemon + 受限 UNIX socket，安装时用一次管理员授权
  - Windows：Windows 服务（`sc create`，需要 UAC 一次）
  - Linux：`setcap` 或 systemd unit，按需选择
- [ ] 系统代理：读/写系统设置（macOS `networksetup`、Windows WinINet 注册表、Linux `gsettings`），
      旁路列表、退出时恢复
- [ ] TUN 模式：stack（gvisor/system）、`auto-route`、`dns-hijack`、`auto-detect-interface`，
      Mac 上要处理 `utun` 与 DNS 劫持
- [ ] 端口占用检测与自动换端口

**验收**：打开系统代理后 Safari 走代理；TUN 下不设系统代理也能全局；关应用后系统代理设置被还原。

**风险**：helper 必须签名才能长期免密码；ad-hoc 签名下每次安装都会弹一次授权。
这条和「免费分发」是冲突的，先做不需要 helper 的部分，把 helper 留到有证书之后。

## M6 打磨

- [ ] 备份/恢复：导出全部配置为 zip，可选 WebDAV 同步
- [ ] 从 Clash Verge 迁移（读它的 `profiles.yaml` + `verge.yaml`）
- [ ] i18n：中文/英文，`i18n` 包 + `mygo` 的本地化机制
- [ ] 应用内更新：`mygo.Updater` 接上已经发布的签名清单
- [ ] 崩溃日志与「上次异常退出」提示
- [ ] 正式签名与公证（Apple Developer ID、Windows 证书）
- [ ] 主题、字体、窗口尺寸记忆、快捷键

---

## 技术决策记录

**核心当库用，而不是子进程。** 子进程的好处是崩溃隔离和内核独立升级，但代价是：
要管进程生命周期、要自己起 IPC、连接表和日志得再走一遍 HTTP API，多一个可执行文件要签名。
Clash Verge 选子进程是因为 Rust 侧本来就没法内嵌 Go；我们用 Go，就没有这个约束。

**不走 RESTful 控制器。** 库模式下 `tunnel.Proxies()`、`statistic.DefaultManager`、
`log.Subscribe` 都是进程内调用，不需要 HTTP 往返和 secret。但配置里保留
`external-controller` 选项，方便用户接第三方面板或脚本。

**配置合成而不是配置生成。** 用户给的 YAML 原样保存，只在内存里合成 runtime 配置，
这样订阅更新不会覆盖用户手改的内容，出问题也能 diff。

**mihomo 的全局单例是硬约束。** 一个进程只能跑一个核心，测试不能并行跑核心相关用例；
所有核心状态都收在 `internal/core` 里，别的地方不许 import mihomo。

## 目录规划

```
main.go                 窗口与页面装配
internal/core            mihomo 的生命周期、快照、配置合成（唯一 import mihomo 的地方）
internal/config         profile 与 overrides 的读写
internal/sysproxy       系统代理（M5）
internal/helper         提权 helper 的客户端（M5）
internal/sub            订阅解析与 userinfo（M2）
docs/roadmap.md         本文件
```