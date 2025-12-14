# 用法: .\view-log.ps1 .\logs\app-2025-12-14.log
param (
    [Parameter(Mandatory=$true)][string]$LogFile
)

Write-Host ">>> 正在监控日志文件: $LogFile (按 Ctrl+C 停止)" -ForegroundColor Cyan

# 模拟 tail -f 效果
Get-Content $LogFile -Wait -Tail 10 | ForEach-Object {
    if ($_.Trim()) {
        try {
            # 将 JSON 行解析为 PowerShell 对象
            $obj = $_ | ConvertFrom-Json
            
            # 打印分隔符和内容
            Write-Host "`n============ $($obj.time) ============" -ForegroundColor Green
            Write-Host "`"url`": `"$($obj.url)`""
            Write-Host "`"method`": `"$($obj.method)`""
            Write-Host "`"code`": $($obj.code)"
            Write-Host "`"duration`": `"$($obj.duration)`""
            Write-Host "`"ip`": `"$($obj.ip)`""
            
            # 格式化打印 Headers
            $reqH = $obj.req_headers | ConvertTo-Json -Compress
            $respH = $obj.resp_headers | ConvertTo-Json -Compress
            Write-Host "`"req_headers`": $reqH"
            Write-Host "`"requestData`": $($obj.requestData)"
            Write-Host "`"resp_headers`": $respH"
            Write-Host "`"responseData`": $($obj.responseData)"
            Write-Host "===========================================" -ForegroundColor Green
        }
        catch {
            # 如果不是 JSON 行（比如普通的错误信息），直接原样输出
            Write-Host $_ -ForegroundColor Gray
        }
    }
}