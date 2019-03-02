

# stop on any error
$ErrorActionPreference = 'Stop'


# You can run this function if you need to - to install pre-reqs locally.
function Register-Prereqs {
  try {
    # install Chocolatey    
    Set-ExecutionPolicy Bypass -Scope Process -Force
    iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))
    refreshenv
    choco install golang -y
    refreshenv
    go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
    refreshenv
    go get github.com/monopole/mdrip
    refreshenv
  } catch {
    Write-Error "Could not install pre-reqs"
  }
}


#####################################################################################
#  Start of process
#####################################################################################

Push-Location

try{
  $scriptPath = $MyInvocation.MyCommand.Path
  Write-Host "Script Root: $scriptPath"
  $baseDir = Split-Path (Split-Path $scriptPath -Parent) -Parent
  Write-Host "Changing Directory: $baseDir"
  
  Set-Location $baseDir

  $rc = $false

  function Test-GoLangCILint {
    golangci-lint -v run ./...
  }

  function Test-GoTest {
    go test -v ./...
  }

  function Test-Examples {
    mdrip --mode test --label test README.md ./examples
  }

  # unfortunately because go test hides output in windows if we try to call it 
  # using Invoke-Express ( calling the function dynamically )
  # we have to call them in-line here instead of using a function
  
  Write-Host "============== begin Test-GoLangCILint"
  Test-GoLangCILint
  if ($LASTEXITCODE -eq 0) {
    $lint = $true
    $result = "SUCCESS"
  } else {
    $lint = $false
    $result = "FAILURE"
  }
  Write-Host ("============== end Test-GoLangCILint : {0} code={1}`n`n`n" -f $result, $lint)


  Write-Host "============== begin Test-GoTest"
  Test-GoTest
  if ($LASTEXITCODE -eq 0) {
    $tests = $true
    $result = "SUCCESS"
  } else {
    $tests = $false
    $result = "FAILURE"
  }
  Write-Host ("============== end Test-GoTest : {0} code={1}`n`n`n" -f $result, $tests)
 

  Write-Host "============== begin Test-Examples"
  Test-Examples
  if ($LASTEXITCODE -eq 0) {
    $examples = $true
    $result = "SUCCESS"
  } else {
    $examples = $false
    $result = "FAILURE"
  }
  Write-Host ("============== end Test-Examples : {0} code={1}`n`n`n" -f $result, $examples)

  #calc final return code
  $rc = $lint -AND $tests -AND $examples

  Pop-Location

  Exit $rc

} catch {
  Write-Host "Error: $_"
  exit 1
}

Pop-Location
