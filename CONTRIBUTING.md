# Contributing

Welcome to OpenEdge of Baidu Open Source Project. To contribute to OpenEdge, please follow the process below.

We sincerely appreciate your contribution. This document explains our workflow and work style.

## Workflow

OpenEdge use this [Git branching model](https://nvie.com/posts/a-successful-git-branching-model/). The following steps guide usual contributions.

1. Fork

   Our development community has been growing fast, so we encourage developers to submit code. And please file Pull Requests from your fork. To make a fork, please refer to Github page and click on the ["Fork" button](https://help.github.com/articles/fork-a-repo/).

2. Prepare for the development environment

   ```bash
   go get github.com/baidu/openedge # clone openedge official repository
   cd $GOPATH/src/github.com/baidu/openedge # step into openedge
   git checkout master  # verify master branch
   git remote add fork https://github.com/<your_github_account>/openedge  # specify remote repository
   ```

3. Push changes to your forked repository

   ```bash
   git status   # view current code change status
   git add .    # add all local changes
   git commit -c "modify description"  # commit changes with comment
   git push fork # push code changes to remote repository which specifies your forked repository
   ```

4. Create pull request

   You can push and file a pull request to OpenEdge official repository [https://github.com/baidu/openedge](https://github.com/baidu/openedge). To create a pull request, please follow [these steps](https://help.github.com/articles/creating-a-pull-request/). Once the OpenEdge repository reviewer approves and merges your pull request, you will see the code which contributed by you in the OpenEdge official repository.

## Code Review

- About Golang format, please refer to [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Please feel free to ping your reviewers by sending them the URL of your pull request via email. Please do this after your pull request passes the CI.
- Please answer reviewers' every comment. If you are to follow the comment, please write "Done"; please give a reason otherwise.
- If you don't want your reviewers to get overwhelmed by email notifications, you might reply their comments by [in a batch](https://help.github.com/articles/reviewing-proposed-changes-in-a-pull-request/).
- Reduce the unnecessary commits. Some developers commit often. It is recommended to append a sequence of small changes into one commit by running `git commit --amend` instead of `git commit`.

## Merge Rule

- Please run command `govendor fmt +local` before push changes, more details refer to [govendor](https://github.com/kardianos/govendor)
- Must run command `make test` before push changes(unit test should be contained), and make sure all unit test and data race test passed
- Only the passed(unit test and data race test) code can be allowed to submit to OpenEdge official repository
- At least one reviewer approved code can be merged into OpenEdge official repository
