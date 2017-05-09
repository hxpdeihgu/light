package light

import (
	"net/http"
	"sync"
	"reflect"
	"strings"
	"errors"
	"encoding/json"
	"fmt"
)

var handlerMap map[string]reflect.Value = make(map[string]reflect.Value)
var lock  sync.Mutex

type Light struct {
	response http.ResponseWriter
	request *http.Request
	Parm string
}

func (this *Light)ServeHTTP(w http.ResponseWriter, r *http.Request){
	defer func() {
		if err := recover();err !=nil {
			fmt.Println(err)
			http.NotFound(w,r)
		}
	}()
	this.response = w
	this.request = r
	
	h,m,err := getHandle(this.request)
	if err!=nil{
		panic(err)
	}
	
	beforeFun:=h.Elem().MethodByName("Before")
	this.invoke(beforeFun)
	mfun:=h.Elem().MethodByName(m)
	this.invoke(mfun)
	afterFun:=h.Elem().MethodByName("After")
	this.invoke(afterFun)
}

func (this *Light) invoke(mfun reflect.Value){
	var in []reflect.Value
	if !mfun.IsValid() {
		return
	}
	switch mfun.Type().NumIn() {
	case 0:
		value:=mfun.Call(nil)
		outValue(value,this.response)
	case 1:
		in = append(in,reflect.ValueOf(this))
		value:=mfun.Call(nil)
		outValue(value,this.response)
	default:
		in = append(in,reflect.ValueOf(this.response))
		in = append(in,reflect.ValueOf(this.request))
		value:=mfun.Call(in)
		outValue(value,this.response)
	}
}

func outValue(v []reflect.Value,w http.ResponseWriter){
	for _,out:=range v {
		switch out.Kind() {
		case reflect.Int:
			w.WriteHeader(int(out.Int()))
		case reflect.String:
			w.Write([]byte(out.String()))
		default:
			o,err:=json.Marshal(out.Interface())
			if err == nil {
				w.Write(o)
			}
		}
	}
}

func getHandle(r *http.Request) (h reflect.Value,m string,e error) {
	path := strings.Trim(r.URL.Path,"/")
	var pkg string
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			pkg = path[:i]
			m = strings.Title(path[i+1:])
			break
		}
	}
	lock.Lock()
	h,ok:=handlerMap[pkg]
	lock.Unlock()
	if ok {
		return h,m,nil
	}
	
	return reflect.Value{},"",errors.New("handler not found")
}

type Handler interface {}

func (this *Light) Add(h Handler){
	rv:=reflect.ValueOf(h)
	verify(rv)
	pkgPath := rv.Elem().Type().PkgPath()
	lock.Lock()
	handlerMap[pkgPath] = rv
	lock.Unlock()
}

func verify(rv reflect.Value) {
	if rv.Type().Kind() != reflect.Ptr && rv.Type().Kind()!= reflect.Struct{
		panic("添加路由错误："+rv.String())
	}
}

func Run(l *Light)  {
	http.ListenAndServe(":8080",l)
}
