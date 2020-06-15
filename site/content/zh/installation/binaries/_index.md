---
title: "可执行文件"
linkTitle: "可执行文件"
weight: 3
type: docs
description: >
    下载编译好的二进制文件来安装 Kustomize。
---

适用于 Linux、MacOS 和 Windows 的各版本的二进制可执行文件可以在 [releases 页面] 上手动下载。

下面的[脚本]会检测你的操作系统，并下载相应的 kustomize 二进制文件到你当前的工作目录中。

```bash
curl -s "https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
```

[releases 页面]: https://github.com/kubernetes-sigs/kustomize/releases
[脚本]: https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh
