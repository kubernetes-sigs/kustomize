# 使用 kustomize 对 helm charts 进行修改

[last mile]: https://testingclouds.wordpress.com/2018/07/20/844/
[stable chart]: https://github.com/helm/charts/tree/master/stable
[Helm charts]: https://github.com/helm/charts
[_minecraft_]: https://github.com/helm/charts/tree/master/stable/minecraft
[插件]: ../../docs/plugins

kustomize 并不会读取 [Helm charts] ，但可以使用 generator 来访问 [Helm charts] 。

使用 [last mile] 模式来结合 kustomize 和 helm ，使用一个 inflated chart 作为基础，然后使用 kustomize 在部署到集群的途中进行修改。

以下示例中使用的 generator 仅适用于 [stable chart] 仓库中的 chart。该示例虽然使用 [_minecraft_] ，但可以应用于任何 chart。

假设 `helm` 已在你的 `$PATH` 中，建立一个工作空间：

<!-- @makeWorkplace @test -->
```bash
DEMO_HOME=$(mktemp -d)
mkdir -p $DEMO_HOME/base
mkdir -p $DEMO_HOME/dev
mkdir -p $DEMO_HOME/prod
```

## 使用远程 chart

定义 _development_ variant（环境）。

这可能涉及许多 kustomizations（参见其他示例），但在本示例中，将 `dev-` 名称前缀添加到所有资源：

<!-- @writeKustDev @test -->
```bash
cat <<'EOF' >$DEMO_HOME/dev/kustomization.yaml
namePrefix:  dev-
resources:
- ../base
EOF
```

同上，使用 `namePrefix: prod-` 定义生产 variant ：

<!-- @writeKustProd @test -->
```bash
cat <<'EOF' >$DEMO_HOME/prod/kustomization.yaml
namePrefix:  prod-
resources:
- ../base
EOF
```

这两个 variants 指向同一个 base。

定义这个 base：

<!-- @writeKustDev @test -->
```bash
cat <<'EOF' >$DEMO_HOME/base/kustomization.yaml
generators:
- chartInflator.yaml
EOF
```

base 指向一个名为 `chartInflator.yaml` 的生成配置文件。

此文件允许指定 [stable chart] 的名称及其他内容，例如 values 文件的路径，默认为 `values.yaml` 。

创建配置文件 `chartInflator.yaml`，指定 chart 名称为 _minecraft_：

<!-- @writeGeneratorConfig @test -->
```bash
cat <<'EOF' >$DEMO_HOME/base/chartInflator.yaml
apiVersion: someteam.example.com/v1
kind: ChartInflator
metadata:
  name: notImportantHere
chartName: minecraft
EOF
```

因为这个特定的 YAML 文件列在 kustomization文件的 `generators:` 字段中，所以它被视为生成器插件（由 _apiVersion_ 和 _kind_ 字段标识）与配置插件的其他字段之间的绑定。

将插件下载到 `DEMO_HOME` 并赋予其执行权限：

<!-- @installPlugin @test -->
```bash
plugin=plugin/someteam.example.com/v1/chartinflator/ChartInflator
curl -s --create-dirs -o \
"$DEMO_HOME/kustomize/$plugin" \
"https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/$plugin"

chmod a+x $DEMO_HOME/kustomize/$plugin
```

检查目录布局：

<!-- @tree -->
```bash
tree $DEMO_HOME
```

将会得倒类似的目录及文件：

> ```bash
> /tmp/whatever
> ├── base
> │   ├── chartInflator.yaml
> │   └── kustomization.yaml
> ├── dev
> │   └── kustomization.yaml
> ├── kustomize
> │   └── plugin
> │       └── someteam.example.com
> │           └── v1
> │               └── chartinflator
> │                  └── ChartInflator
> └── prod
>    └── kustomization.yaml
> ```

运行 kustomize 定义一个 helper function 来传入正确的环境和常见标志：

<!-- @defineKustomizeIt @test -->
```
function kustomizeIt {
  XDG_CONFIG_HOME=$DEMO_HOME \
  kustomize build --enable_alpha_plugins \
    $DEMO_HOME/$1
}
```

最终构建 `prod` variant。这里要注意的是，所有资源名称现在都具有 `prod-` 前缀：

<!-- @doProd @test -->
```bash
clear
kustomizeIt prod
```

比较 `dev` 和 `prod`：

<!-- @doCompare -->
```bash
diff <(kustomizeIt dev) <(kustomizeIt prod) | more
```

在 base上 运行 kustomize 查看未修改但已展开的 chart。
这里的每次调用都是重新下载并重新展开 chart。

<!-- @showBase @test -->
```bash
kustomizeIt base
```

## 使用本地 chart

上面的示例由于未在配置中指定本地 chart 的主目录，所以kustomize会取得远程chart的副本并存在临时目录中。

要禁止 fetch，请明确指定 `charHome` ，并确保chart 已经被保存在该目录下

要进行演示，并且不会干扰您现有的 helm 环境，请执行以下操作：

<!-- @helmInit @test -->
```bash
helmHome=$DEMO_HOME/dothelm
chartHome=$DEMO_HOME/base/charts

function doHelm {
  helm --home $helmHome $@
}

# 在新位置创建 helm 配置文件。
# 初始化命令比较复杂
doHelm init --client-only >& /dev/null
```

现在下载 chart ； 可以再次使用的 [_minecraft_] （也可以使用其他的 chart ）：

<!-- @fetchChart @test -->
```bash
doHelm fetch --untar \
    --untardir $chartHome \
    stable/minecraft
```

使用 tree 查看更多信息；helm 配置数据和完整的 chart 副本：

<!-- @tree -->
```bash
tree $DEMO_HOME
```

将 `chartHome` 字段添加到生成器的配置文件中，以便可以查找本地 chart：

<!-- @modifyGenConfig @test -->
```bash
echo "chartHome: $chartHome" >>$DEMO_HOME/base/chartInflator.yaml
```

更改 values 文件，用来展示本地 chart 的更改：

<!-- @valueChange @test -->
```
sed -i 's/CHANGEME!/SOMETHINGELSE/' $chartHome/minecraft/values.yaml
sed -i 's/LoadBalancer/NodePort/' $chartHome/minecraft/values.yaml
```

最后进行构建：

<!-- @finalProd @test -->
```bash
kustomizeIt prod
```

观察结果中 `LoadBalancer` 变为 `NodePort`，并且加密的密码也有所不同。
