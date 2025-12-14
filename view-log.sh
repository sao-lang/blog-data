#!/bin/bash

# 使用说明: ./view-log.sh logs/app-2025-12-14.log
# 依赖: 需要安装 jq (sudo apt install jq)

LOG_FILE=$1

if [ -z "$LOG_FILE" ]; then
    echo "使用方法: ./view-log.sh [日志文件路径]"
    exit 1
fi

# 实时查看并转换 JSON 到你习惯的分隔符格式
tail -f $LOG_FILE | jq -r '
    "\n\u001b[32m============ \( .time ) ============\u001b[0m",
    "url: \( .url )",
    "method: \( .method )",
    "code: \( .code )",
    "ip: \( .ip )",
    "duration: \( .duration )",
    "requestData: \( .requestData )",
    "responseData: \( .responseData )",
    "\u001b[32m===========================================\u001b[0m"
'