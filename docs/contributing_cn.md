# 如何贡献

欢迎来到 Baetyl 开源边缘计算项目，如果您想要向 Baetyl 贡献代码或文档，请遵循以下流程。

## 贡献流程

Baetyl 使用通用的 [Git 分支构建模型](http://nvie.com/posts/a-successful-git-branching-model/)。下面将为您提供通用的 Github 代码贡献方式。

1. Fork 代码库

   我们的开发社区非常活跃，感兴趣的开发者日益增多，因此，我们鼓励开发者采用 **fork** 方式向我们提交代码。关于如何 fork 一个代码库，请参考 Github 提供的官方帮助页面并点击 ["Fork" 按钮](https://help.github.com/articles/fork-a-repo/).

2. 准备开发环境

   如果您想要向 Baetyl 贡献代码，请参考如下命令准备相关本地开发环境：

   ```bash
   go get github.com/baetyl/baetyl # 获取 baetyl 代码库
   cd $GOPATH/src/github.com/baetyl/baetyl # 进入 baetyl 代码库目录
   git checkout master  # 校验当前处于 master 主分支
   git remote add fork https://github.com/<your_github_account>/baetyl  # 指定远程提交代码仓库
   ```

3. 提交代码到 fork 仓库

   这里，将改动的需求或修复的 bug 提交到步骤 2 中 fork 的远程仓库，具体请参考如下命令：

   ```bash
   git status   # 查看当前代码改变状态
   git add .
   git commit -c "modify description"  # 提交代码到本地仓库，并提交代码改动描述信息
   git push fork # 推送已提交本地仓库的代码要远程仓库
   ```

4. 创建代码合入请求

   基于 fork 的仓库地址直接向 Baetyl 官方仓库 [https://github.com/baetyl/baetyl](https://github.com/baetyl/baetyl) 提交 **pull request**（具体请参考[如何创建一个提交请求](https://help.github.com/articles/creating-a-pull-request/)），即可完成向 Baetyl 官方仓库的代码合入请求。一旦 Baetyl 代码仓库评审人员通过了您的代码提交、合入请求，您即可在 Baetyl 官方代码仓库中看到您贡献的代码。

## 代码评审规范

- Golang 的代码风格请参照 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 请在代码 CI 测试通过后及时通过 Email 向你的代码评审人发送代码提交请求URL
- 请及时回答评审人的每一个 comment，如果您采纳评审人给出的建议，请直接回复 **好的** 或是 **Done**；如果您不同意，请给出您的理由
- 如果您不想您的代码评审人被邮件通知频繁打扰，您可以通过 **交互框** 回复评审人提出的每一个建议，具体请参考 [如何使用交互框回复评审人信息](https://help.github.com/articles/reviewing-proposed-changes-in-a-pull-request/)
- 尽可能减少不必要的代码提交。一些开发者总是频繁提交代码。如果您想要向提交的代码中增加一个微小的改动，请使用命令 `git commit --amend` 代替 `git commit`

## 代码合入规范

无规矩不成方圆。这里规定，凡是提交 Baetyl 代码合入请求的代码，一律要求遵循以下规范：

- 代码提交前 **必须** 进行单元测试（提交代码应包含）和竞争检测，参考执行命令 `make test`
- 仅有提交代码通过单元测试和竞争检测，才允许向 Baetyl 官方仓库提交
- 所有向 Baetyl 官方仓库提交的代码，**必须至少** 有 **1** 个代码评审员评审通过后，才可以将提交代码合入 Baetyl 官方代码仓库


**注意**：以上所有代码提交步骤要求及规范，同样适用文档贡献。
