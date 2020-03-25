# podcast内容下载器

## 用法

sorry,看代码更快


## 辅助脚本

### 用wget下载媒体
```shell
cat wget.task|awk '{print $1}'|wget -c -i
```
### 重命名wget下载好的媒体

```shell 
cat wget.task |sed  's|^.*/\(.*\.mp3\) \(.*\)|mv \1 \2|g' |bash
```