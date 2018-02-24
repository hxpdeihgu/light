# fmt.Sscanf()
例子:
```golang
fmt.Sscanf("abc def\n","%s%s",&c,&b)
fmt.Println(c,"---",b)
echo abc --- def
```
# 函数:
```golang
func Sscanf(str string, format string, a ...interface{}) (n int, err error)
```
###str中空格取参数个数,\n 回车完成提交

