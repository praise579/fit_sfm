#!/bin/sh
echo "docker images clean shell"
# 项目名取的是当前目录名，故先锚定到仓库根（脚本位于 scripts/，需回退一级）
cd "$(dirname "$0")/.." || exit 1
projectName=$(pwd | awk -F "/" '{print $NF}')
dockerrm=$(docker images | grep "${projectName}" | awk '{print $3}' | awk '!a[$0]++')

if [  -n "$dockerrm"  ]; then
    docker rmi -f ${dockerrm}
    echo "docker images ${projectName} clean OK"
fi

dockerrm=$(docker images | grep "none" | awk '{print $3}' | awk '!a[$0]++')
if [  -n "$dockerrm" ]; then
    docker rmi -f ${dockerrm}
    echo "docker images none clean OK"
fi
