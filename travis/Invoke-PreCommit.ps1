<#

Please reference this document:
  /docs/howtowindows.md
 
#>

#####################################################################################
#  Start of process
#####################################################################################
# stop on any error
$ErrorActionPreference = 'Stop'


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
    $lint = 0
    $result = "SUCCESS"
  } else {
    $lint = 1
    $result = "FAILURE"
  }
  Write-Host ("============== end Test-GoLangCILint : {0} code={1}`n`n`n" -f $result, $lint)


  Write-Host "============== begin Test-GoTest"
  Test-GoTest
  if ($LASTEXITCODE -eq 0) {
    $tests = 0
    $result = "SUCCESS"
  } else {
    $tests = 1
    $result = "FAILURE"
  }
  Write-Host ("============== end Test-GoTest : {0} code={1}`n`n`n" -f $result, $tests)
 

  Write-Host "============== skipping Test-Examples for Windows Testing "
  
  #Write-Host "============== begin Test-Examples"
  #Test-Examples
  #if ($LASTEXITCODE -eq 0) {
  #  $examples = 0
  #  $result = "SUCCESS"
  #} else {
  #  $examples = 1
  #  $result = "FAILURE"
  #}
  #Write-Host ("============== end Test-Examples : {0} code={1}`n`n`n" -f $result, $examples)

  #calc final return code
  #$rc = $lint -AND $tests -AND $examples
  
  #calc final return code - omit mdrip testing
  $rc = $lint -AND $tests

  Pop-Location

  Exit $rc

} catch {
  Write-Host "Error: $_"
  exit 1
}

Pop-Location
