# 安全政策

## 支持范围

我们会对当前 `master` 分支上报告的漏洞进行评估与修复。

## 报告漏洞

**请不要在公开 Issue 中披露漏洞细节。**

请通过 GitHub 的 [Security Advisories](https://github.com/praise579/fit_sfm/security/advisories/new)（或私有邮件，见维护者资料）私下报告，说明：

- 影响版本 / 分支
- 漏洞描述与复现步骤
- 预期影响与缓解建议

我们会在评估后回复，并在修复发布时公开致谢（如你愿意）。

## 已知边界

本项目会读取远端订阅与模板文件并执行本地模板渲染，请只配置可信来源；`auth.password` 通过 URL 查询参数传输，生产环境应配合 HTTPS 或放在访问控制后。
