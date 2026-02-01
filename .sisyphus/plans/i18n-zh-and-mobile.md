# i18n(zh) + Mobile UX Work Plan

## TL;DR

> **Quick Summary**: 在不动后端的前提下，把前端 next-intl 从仅 `en` 扩展到 `en+zh`，补齐中文 `messages/zh.json`，提供可用的语言切换器，并针对移动端做“触摸目标+下拉宽度+admin 栅格”三类关键优化。
>
> **Deliverables**:
> - `frontend/messages/zh.json`（全量翻译 + 新增 Auth/Theme/Language keys）
> - `frontend/i18n/routing.ts` + `frontend/i18n/navigation.ts`（统一 locale 配置与导航封装）
> - `frontend/components/LanguageSwitcher.tsx` + 接入 `frontend/app/[locale]/page.tsx`
> - Login/Register/Theme 文案迁移到 next-intl
> - 移动端优化：touch target >= 44px、dropdown 宽度自适应、admin 页响应式优化
>
> **Estimated Effort**: Medium
> **Parallel Execution**: YES (3 waves + final verification)
> **Critical Path**: 路由/locale 配置 → messages keys 对齐 → 组件迁移/语言切换 → 回归验证

---

## Context

### Original Request
- 全网站中文适配：创建完整中文翻译，支持中英文切换
- UI 手机端优化：改善移动端体验，确保触摸友好

### Repo Facts (verified)
- `frontend/package.json` 显示 `next: 16.1.3`（与“Next.js 15”描述略有出入，但不影响本计划的 next-intl/App Router 实施路径）
- next-intl 入口：`frontend/middleware.ts`, `frontend/i18n/request.ts`, `frontend/app/[locale]/layout.tsx`
- 仅存在 `frontend/messages/en.json`（132 行）
- 硬编码英文残留：
  - `frontend/components/LoginForm.tsx`
  - `frontend/components/RegisterForm.tsx`
  - `frontend/components/mode-toggle.tsx`
  - `frontend/app/[locale]/page.tsx`（aria-label: "GitHub"）
- 移动端风险点：
  - `frontend/components/AccountMenu.tsx` dropdown `w-56`
  - `frontend/components/mode-toggle.tsx` 图标按钮 `h-9 w-9`
  - `frontend/components/PriceFeed.tsx` 图标按钮 `h-9 w-9`
  - `frontend/app/[locale]/admin/page.tsx` `lg:grid-cols-[1.2fr_0.8fr]`

---

## Work Objectives

### Core Objective
在现有 App Router 结构（`app/[locale]/*`）下，新增 `zh` locale 并完成全站中文化 + 语言切换，同时让关键交互在移动端更易点、更不挤。

### Definition of Done
- `frontend` 侧：`npm run lint`、`npx tsc --noEmit`、`npm run build` 全通过
- 本地启动后（`npm run dev`）：
  - `/en` 与 `/zh` 页面可访问
  - LanguageSwitcher 可在不丢失当前路径的情况下切换语言
  - Login/Register/Theme 菜单项中文显示正确
  - 375x812 视口下：Navbar 图标按钮可点（>=44px），AccountMenu dropdown 不溢出屏幕，Admin 页面不拥挤/不横向滚动

### Must NOT Have (guardrails)
- 不改 Go 后端接口、不调整鉴权协议
- 不引入大规模重构：只在“触碰到的文件”内逐步改用 next-intl navigation wrapper
- 不为“未来多语言”做超范围设计（本次只做 `en` + `zh`）

---

## Verification Strategy

### Test / QA Decision
- **Existing test infra**: NO（`frontend/package.json` 无 `test` script）
- **Strategy**: Automated checks via Playwright skill + build/lint/tsc commands

### Required Commands (agent-executable)
```bash
cd frontend
npm run lint
npx tsc --noEmit
npm run build
```

### UI Verification (agent-executable via Playwright skill)
在 dev server 运行后（`cd frontend && npm run dev`），用 Playwright 自动化：
- 访问 `http://localhost:3000/en` 与 `http://localhost:3000/zh`
- 打开 LanguageSwitcher，切换语言后 URL 与页面文案同步变化
- 访问 `http://localhost:3000/zh/auth/login`、`/zh/auth/register` 确认表单文案为中文
- 打开 ModeToggle dropdown，选项为中文（浅色/深色/跟随系统）
- 模拟手机视口（375x812）：检查 dropdown 不超出视口并截图存证

---

## Parallel Execution Waves

Wave 1 (foundation; parallelize all)
- Task 0: 全仓扫描（硬编码英文 + 固定宽度/触摸目标）
- Task 1: i18n routing + navigation wrappers
- Task 2: middleware + request.ts 支持 zh
- Task 3: 新增 zh.json（覆盖 en.json keys）
- Task 4: 补齐 en.json 缺失 keys（Auth/Theme/Language/aria）

Wave 2 (feature integration; depends on Wave 1 keys)
- Task 5: LanguageSwitcher 组件 + 接入 Navbar
- Task 6: LoginForm i18n 迁移
- Task 7: RegisterForm i18n 迁移
- Task 8: ModeToggle i18n 迁移

Wave 3 (mobile improvements; can overlap with Wave 2 but prefer after strings land)
- Task 9: AccountMenu dropdown 宽度 + touch target
- Task 10: Admin page 响应式栅格优化
- Task 11: PriceFeed 控制栏 touch target 与小屏布局复核
- Task 12: SubmitForm 手机端密度复核（小改）

Wave 4 (final)
- Task 13: 全量回归（lint/tsc/build + Playwright 路径）

---

## TODOs (with agent profiles)

> 说明：每个 Task 都包含可执行验收；执行时建议每波次内并行，波次间按依赖顺序推进。

- [ ] 0. Repo-wide audit: remaining hardcoded strings + fixed widths

  **What to do**:
  - 扫描残留硬编码英文（尤其是 aria-label / 错误提示 / 按钮文案）
  - 扫描固定宽度与潜在移动端问题（`w-56`, `w-[160px]`, `h-9 w-9` 等）
  - 产出一份清单（建议写入：`.sisyphus/evidence/i18n-mobile-audit.md`），并在后续任务中逐项消灭

  **Recommended Agent Profile**:
  - Category: `unspecified-low`
  - Skills: `desktop-commander`

  **References**:
  - `frontend/components/**/*.tsx` - 主要 UI 文案与交互集中区
  - `frontend/app/[locale]/*` - 页面级文案与 aria

  **Acceptance Criteria**:
  - 清单文件中至少包含：文件路径、原文、建议 key（或修复建议）、影响（i18n/移动端）
  - 明确标注本计划已覆盖的项（Login/Register/ModeToggle/AccountMenu/Admin/PriceFeed/SubmitForm）与新增发现项

- [ ] 1. Create next-intl routing + navigation wrappers

  **What to do**:
  - 新增 `frontend/i18n/routing.ts`，配置 `locales: ['en','zh']`, `defaultLocale: 'en'`
  - 新增 `frontend/i18n/navigation.ts`，用 next-intl 的 navigation wrapper 生成 `Link/useRouter/usePathname` 等

  **Recommended Agent Profile**:
  - Category: `ultrabrain`（需要统一路由/导航策略，避免后续手写 locale 前缀）
  - Skills: `context7-auto-research`（核对 next-intl API）, `desktop-commander`

  **References**:
  - `frontend/i18n/request.ts` - 当前 locale 校验与 messages import 方式
  - `frontend/middleware.ts` - 当前 locale/matcher 配置

  **Acceptance Criteria**:
  - 新文件存在：`frontend/i18n/routing.ts`, `frontend/i18n/navigation.ts`
  - `cd frontend && npx tsc --noEmit` → PASS

- [ ] 2. Update middleware matcher + request locale validation for zh

  **What to do**:
  - `frontend/middleware.ts`：把 locales 扩展到 `['en','zh']`，并使用更稳健的 matcher（排除 `api`, `_next`, 静态文件）
  - `frontend/i18n/request.ts`：允许 `zh`，并对 `requestLocale` 做白名单校验

  **Recommended Agent Profile**:
  - Category: `quick`
  - Skills: `desktop-commander`

  **References**:
  - `frontend/middleware.ts` - 现 matcher 仅覆盖 `/` 和 `/en/:path*`
  - `frontend/i18n/request.ts` - 现仅允许 `en`

  **Acceptance Criteria**:
  - `cd frontend && npm run build` → PASS
  - Playwright: 访问 `http://localhost:3000/zh` 不应 404

- [ ] 3. Create `frontend/messages/zh.json` (full coverage)

  **What to do**:
  - 新增 `frontend/messages/zh.json`
  - 覆盖 `frontend/messages/en.json` 的所有 keys
  - 为本次新增的 Auth/Theme/Language/aria keys 提供中文

  **Recommended Agent Profile**:
  - Category: `writing`（高质量翻译 + 一致术语）
  - Skills: `humanizer-zh`（让中文更自然）, `desktop-commander`

  **References**:
  - `frontend/messages/en.json` - 现有全量 keys

  **Acceptance Criteria**:
  - `frontend/messages/zh.json` 可被 `request.ts` 正常加载（`npm run build` PASS）
  - Playwright: `/zh` 页面中 Navbar/Feed/Submit/Admin 等关键区块有中文文案（至少 10 个断言点）

- [ ] 4. Add missing message namespaces/keys (Auth, Theme, LanguageSwitcher)

  **What to do**:
  - 在 `frontend/messages/en.json` 补齐：
    - Login/Register 表单文案（标题、label、placeholder、错误提示、按钮、captcha）
    - Theme（Light/Dark/System/Toggle theme aria）
    - LanguageSwitcher（切换语言 aria、菜单项）
    - Navbar 的 GitHub aria（建议纳入以满足“全站中文适配”）

  **Recommended Agent Profile**:
  - Category: `quick`
  - Skills: `desktop-commander`

  **References**:
  - `frontend/components/LoginForm.tsx` - 需要抽取的硬编码
  - `frontend/components/RegisterForm.tsx` - 需要抽取的硬编码
  - `frontend/components/mode-toggle.tsx` - 需要抽取的硬编码
  - `frontend/app/[locale]/page.tsx` - aria-label "GitHub"

  **Acceptance Criteria**:
  - `cd frontend && npm run build` → PASS（确保新 keys 不导致运行期缺 key 报错）

- [ ] 5. Implement LanguageSwitcher and wire into Navbar

  **What to do**:
  - 新增 `frontend/components/LanguageSwitcher.tsx`：
    - 使用 `DropdownMenu`（shadcn）
    - 使用 `@/i18n/navigation` 的 router/pathname 保持当前路径切换 locale
    - 移动端触摸目标：触发按钮 `h-11 w-11`，菜单项 `min-h-[44px]`
  - 在 `frontend/app/[locale]/page.tsx` Navbar actions 区域接入

  **Recommended Agent Profile**:
  - Category: `visual-engineering`
  - Skills: `frontend-ui-ux`, `desktop-commander`

  **References**:
  - `frontend/app/[locale]/page.tsx` - Navbar actions 插槽（现有 `ModeToggle`, `AccountMenu`）
  - `frontend/components/AccountMenu.tsx` - dropdown 风格参考

  **Acceptance Criteria**:
  - Playwright:
    - 从 `/en/auth/login` 切到 zh 后，URL 变为 `/zh/auth/login` 且页面文案变中文
    - 截图：`.sisyphus/evidence/lang-switcher-desktop.png`

- [ ] 6. Migrate `frontend/components/LoginForm.tsx` to i18n

  **What to do**:
  - 用 `useTranslations()` 替换硬编码：标题/描述/labels/placeholders/错误提示/按钮/aria-label（show/hide password）
  - 逐步替换 `next/navigation` 为 `@/i18n/navigation`（仅在本文件内）以减少手写 `/${locale}`
  - 维持既有逻辑：captcha 加载、signIn、admin 检测跳转

  **Recommended Agent Profile**:
  - Category: `quick`
  - Skills: `desktop-commander`

  **References**:
  - `frontend/components/LoginForm.tsx` - 当前硬编码与跳转
  - `frontend/messages/en.json` - 新增 Auth/Login keys 落点

  **Acceptance Criteria**:
  - Playwright: 访问 `/zh/auth/login`，页面出现中文标题与按钮文字
  - `cd frontend && npm run build` → PASS

- [ ] 7. Migrate `frontend/components/RegisterForm.tsx` to i18n

  **What to do**:
  - 与 LoginForm 同步策略：labels/placeholders/错误提示/aria
  - “Passwords do not match / Captcha is required / Registration failed”等错误改为 i18n
  - 保持注册成功后跳转 `/[locale]/auth/login`

  **Recommended Agent Profile**:
  - Category: `quick`
  - Skills: `desktop-commander`

  **References**:
  - `frontend/components/RegisterForm.tsx`

  **Acceptance Criteria**:
  - Playwright: `/zh/auth/register` 文案为中文；密码不一致时显示中文错误

- [ ] 8. Migrate `frontend/components/mode-toggle.tsx` to i18n + touch target

  **What to do**:
  - 把 "Toggle theme"/"Light"/"Dark"/"System" 抽到 i18n
  - 图标按钮移动端触摸目标 >=44px：建议 `h-11 w-11 md:h-9 md:w-9`
  - dropdown item 设置 `min-h-[44px]`

  **Recommended Agent Profile**:
  - Category: `visual-engineering`
  - Skills: `frontend-ui-ux`, `desktop-commander`

  **References**:
  - `frontend/components/mode-toggle.tsx`

  **Acceptance Criteria**:
  - Playwright: `/zh` 打开 Theme 菜单项中文显示
  - 手机视口截图：`.sisyphus/evidence/theme-toggle-mobile.png`

- [ ] 9. Fix AccountMenu dropdown width + touch target (mobile-safe)

  **What to do**:
  - `frontend/components/AccountMenu.tsx`：
    - dropdown content 从 `w-56` 改为“最小宽 + 视口自适应”，避免中文更长导致溢出
    - trigger icon button 同样提升到移动端 >=44px（可用 `h-11 w-11 md:h-9 md:w-9`）
    - 菜单项 `min-h-[44px]`

  **Recommended Agent Profile**:
  - Category: `visual-engineering`
  - Skills: `frontend-ui-ux`, `desktop-commander`

  **References**:
  - `frontend/components/AccountMenu.tsx` - `DropdownMenuContent` 现为 `w-56`

  **Acceptance Criteria**:
  - Playwright（375x812）：打开 AccountMenu，dropdown 不超出右侧视口且可完整点击
  - 截图：`.sisyphus/evidence/account-menu-mobile.png`

- [ ] 10. Improve Admin page responsive grid

  **What to do**:
  - `frontend/app/[locale]/admin/page.tsx`：把双栏拆分从 `lg` 推迟到更宽断点（例如 `xl`），避免中等屏幕过窄
  - 处理 table 区域：保持 `overflow-x-auto`，避免整体横向滚动

  **Recommended Agent Profile**:
  - Category: `visual-engineering`
  - Skills: `frontend-ui-ux`, `desktop-commander`

  **References**:
  - `frontend/app/[locale]/admin/page.tsx` - `lg:grid-cols-[1.2fr_0.8fr]`

  **Acceptance Criteria**:
  - Playwright（768x1024）：admin 页面不出现“右侧栏太窄导致按钮挤压/换行异常”
  - 截图：`.sisyphus/evidence/admin-tablet.png`

- [ ] 11. PriceFeed controls touch target + layout check

  **What to do**:
  - `frontend/components/PriceFeed.tsx`：刷新 icon button 从 `h-9 w-9` 提升移动端触摸目标
  - SelectTrigger 高度与按钮一致（移动端 >=44px）

  **Recommended Agent Profile**:
  - Category: `visual-engineering`
  - Skills: `frontend-ui-ux`, `desktop-commander`

  **References**:
  - `frontend/components/PriceFeed.tsx` - controls bar 区域

  **Acceptance Criteria**:
  - Playwright（375x812）：controls bar 不溢出、不挤压；刷新按钮易点

- [ ] 12. SubmitForm mobile density pass (minimal tweaks only)

  **What to do**:
  - `frontend/components/SubmitForm.tsx`：只做必要的移动端间距/字体/换行微调（若实际观感过挤）

  **Recommended Agent Profile**:
  - Category: `artistry`
  - Skills: `frontend-ui-ux`, `desktop-commander`

  **References**:
  - `frontend/components/SubmitForm.tsx` - 已采用 `h-11`，优先保持现设计语言

  **Acceptance Criteria**:
  - Playwright（375x812）：SubmitForm 不横向滚动；输入区与按钮有足够间距

- [ ] 13. Final verification (build/lint/tsc + Playwright paths)

  **What to do**:
  - 运行：lint / tsc / build
  - Playwright 自动化：
    - `/en` ↔ `/zh` 切换
    - `/zh/auth/login`、`/zh/auth/register`
    - 手机/平板视口截图

  **Recommended Agent Profile**:
  - Category: `visual-engineering`
  - Skills: `playwright`, `desktop-commander`

  **Acceptance Criteria**:
  - `cd frontend && npm run lint` → PASS
  - `cd frontend && npx tsc --noEmit` → PASS
  - `cd frontend && npm run build` → PASS
  - 截图文件输出到 `.sisyphus/evidence/`（至少 3 张：lang switch、mobile navbar dropdown、admin tablet）

---

## Commit Strategy (suggested)
- Commit 1: `feat(i18n): add zh locale + language switcher`
- Commit 2: `feat(i18n): translate auth + theme strings`
- Commit 3: `fix(ui): improve mobile touch targets and responsive layouts`
