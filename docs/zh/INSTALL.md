[release 页面]: /../../releases
[Go]: https://golang.org
[脚本]: https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh
[快速开始]: https://www.arp242.net/curl-to-sh.html

## 安装

适用于 Linux、MacOS 和 Windows 的各版本的二进制可执行文件可以在 [release 页面] 上手动下载。

如果希望[快速开始]，可以执行:

```bash
curl -s "https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
```

这个[脚本]会：

- 尝试检测您的操作系统
- 在临时目录中下载并解压 tar 文件
- 将 kustomize 二进制可执行文件复制到您当前的工作目录中
- 删除临时目录

## 尝试 `go`

这种方式只是为了更好的展示如何使用 `go` 语言来安装 kustomize。实际使用中，我们并不推荐此方法。kustomize 的开发者应该拉取此 repo（详见下一部分），而 CI/CD 脚本中应直接下载可执行文件，而不要依赖 `go` 语言工具。

将 kustomize 的最新版本 v3 安装到 `$GOPATH/bin`:

```bash
GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v3
```

安装指定版本

```bash
GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3@v3.3.0
```

## 本地源码构建 kustomize CLI

```bash
# 需要 go 1.13 或更高版本
unset GOPATH
# 详见 https://golang.org/doc/go1.13#modules
unset GO111MODULES

# 拉取 repo
git clone git@github.com:kubernetes-sigs/kustomize.git
# 进入目录
cd kustomize

# 如果您不想从 HEAD 开始构建， 则可以选择切换特定的标签
git checkout kustomize/v3.2.3

# 开始构建
(cd kustomize; go install .)

# 运行
~/go/bin/kustomize version
```

### 其他方式

#### macOS

```bash
brew install kustomize
```

#### windows

```bash
choco install kustomize
```

有关软件包管理器 chocolatey 的使用以及对之前版本的支持，请参考以下链接：
- [Choco Package](https://chocolatey.org/packages/kustomize)
- [Package Source](https://github.com/kenmaglio/choco-kustomize)
