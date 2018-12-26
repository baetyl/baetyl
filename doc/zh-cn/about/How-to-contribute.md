# 贡献

欢迎来到OpenEdge百度开源边缘计算项目，如果您想要向OpenEdge贡献代码或文档，请遵循以下流程。

## 贡献流程

Openedge使用通用的[Git 分支构建模型](http://nvie.com/posts/a-successful-git-branching-model/)。下面将为您提供通用的Github代码贡献方式。

1. Fork代码库

   我们的开发社区非常活跃，感兴趣的开发者日益增多，每一个人都向OpenEdge官方仓库提交代码没有意义。因此，我们鼓励开发者采用“**fork**”方式向我们提交代码。关于如何fork一个代码库，请参考Github提供的官方帮助页面并点击["Fork" 按钮](https://help.github.com/articles/fork-a-repo/).

2. 克隆代码库

   请拷贝或克隆代码库至您的本地计算机上，请执行如下命令：
   ```bash
   go get github.com/your-github-account/openedge (推荐)
   cd openedge
   ```
   或
   ```bash
   git clone https://github.com/your-github-account/openedge
   cd openedge
   ```

3. 创建本地分支

   日常简单的，像增加一个新的功能或者修复一个bug，请在开始编写代码前创建一个新的分支：

   ```bash
   git checkout -b local-cool-branch
   ```

4. 推送本地分支到GitHub远程

   推送本地分支到GitHub远程是为了向代码仓库提交新的代码，同时又不会影响代码库主分支结构，具体可参照如下命令：

   ```bash
   git push origin remote-cool-branch:local-cool-branch
   ```

   _**注意**：在将本地分支推送到远程分支时，你依然可以更改远程分支的名称。当然，这里建议您将远程新分支的名称与本地新分支名称保持一致。_

5. 提交代码

   在您提交代码前，请保持您处于新创建的分支，然后执行下述命令提交代码：

   ```bash
   git status # 检查代码更改状态
   git add .  # 提交所有更改代码，当然，您也可以选择某些需要提交的代码进行提交
   git commit -m "description about commiting code"  # 提交代码到本地，并撰写代码提交描述信息
   git push --set-upstream origin remote-cool-branch # 推送已提交到本地的代码到远程仓库
   ```
   
   至此，即可完成本地编写代码到远程仓库的提交。
   
   **需要注意的是**，完成以上操作，代码仅仅提交到远程新创建的分支（remote-cool-branch），而非master主分支。如果您觉得无伤大雅，可以省略步骤3和步骤4，直接在本地基于master主分支进行代码开发和提交。但是，如果您参照上述步骤执行，想要将新提交的代码合入master主分支，请继续执行以下步骤。

6. 创建代码合入请求

   通过浏览器打开[GitHub主页](https://github.com)，登录您的账号，进入openedge代码仓库，点击“New pull request”按钮，并选择新提交代码的分支与master主分支进行比较，然后继续点击“Create pull request”按钮，同时完善相应的代码提交描述信息，接着点击“Commit”按钮，至此完成分支代码提交合入master主分支的提交请求。

   如果代码审核或校验无误，即可合入master主分支。这里，由于您处于自己的代码仓库，拥有评审、合入代码的权限，可以直接将分支代码合入master主分支。但是，如果您想要将新提交的代码合入OpenEdge官方代码仓库的master主分支，则还需重复执行步骤6，直至完成向OpenEdge官方仓库提交代码合入请求。一旦，OpenEdge代码仓库评审人员通过了您的代码提交、合入请求，您即可在OpenEdge官方代码仓库中看到您贡献的代码。

## 代码评审规范

待补充。（@ludanfeng）

## 代码合入规范

无规矩不成方圆。这里规定，凡是提交OpenEdge代码合入请求的代码，一律要求遵循以下规范：

> + 代码提交前**必须**进行单元测试（提交代码应包含）和竞争检测
> + 仅有提交代码通过单元测试和竞争检测，才允许向OpenEdge官方仓库提交
> + 所有向OpenEdge官方仓库提交的代码，**必须至少**有**1**个代码评审员评审通过后，才可以将提交代码合入OpenEdge官方代码仓库

**注意**：以上所有代码提交步骤要求及规范，同样适用文档贡献。