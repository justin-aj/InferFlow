$mockJob = Start-Job -ScriptBlock {
    Set-Location $using:PWD
    go run ./cmd/mock-backend
}

Start-Sleep -Seconds 1

$env:INFERFLOW_BACKENDS = "http://localhost:9000"
go run ./cmd/router

Stop-Job $mockJob | Out-Null
Remove-Job $mockJob | Out-Null
