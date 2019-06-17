# How To Windows

This is the PowerShell script to run all go tests for Kustomize on a windows based platform which mimics /build/pre-commit.sh

## Pre-Reqs:
  - (obviously) PowerShell installed
    - PowerShell Core is supported
  - go installed
  - golangci-lint installed
  - mdrip installed

This script should output to the current console and return an exit code if all tests are successful(0) or any failed(1).

### If you are tryin to run these tests locally you can follow these instructions.

Assume: 
  - Running a stock Windows 10 system
  - Local Admin rights.
  - You can open [PowerShell as administrator](http://lmgtfy.com/?iie=1&q=How+to+open+powershell+as+administrator)
  - You should be knowledgeable enough to pull source for packages into your GO ```src``` directory
    -  Yes, this means you also need to know a bit about **git** usually


#### Step 1 - Install Go
  - [Install Go](https://golang.org/dl/) - please use the msi
    - If you use chocolatey - it's using the zip not msi and assumptions on where go is located are made for you.
#### Step 2 - Install Go Packages
  - Open new PowerShell Administrative window
    - Install golangci-lint
      - ```go get -u github.com/golangci/golangci-lint/cmd/golangci-lint```
    - Install mdrip
      - ```go get github.com/monopole/mdrip```

You should now be able to issue these commands and see comparable responses

```
C:\...> golangci-lint --help
Smart, fast linters runner. Run it in cloud for every GitHub pull request on https://golangci.com
...

C:\...> mdrip --help
Usage:  C:\_go\bin\mdrip.exe {fileName}...
...
```

#### Step 3 - Get Source and Test
- In your GoRoot src
  - ```Example: C:\_go\src```
- Navigate to the Kustomize `travis` directory
  - ```Example: C:\_go\src\sigs.k8s.io\kustomize\travis```
- Now Execute:
  - ```.\Invoke-PreCommit.ps1```

This should run all pre-commit tests thus defined in the script.
