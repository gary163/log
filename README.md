# log
目前只支持console和file两种模式(思路参考了beego的log模块)
在file模式下：
1，会根据指定的文件大小自动切分log文件
2，自动删除过期的log文件
