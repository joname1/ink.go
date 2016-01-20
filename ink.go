package ink

import (
    "net/http"
    "strings"
    "regexp"
    "fmt"
)

/* data structure */

type MatchMap map[string]string

type Handle func(ctx *Context)

type Context struct {
    http.ResponseWriter
    Res http.ResponseWriter
    Req *http.Request
    Param MatchMap
    Ware map[string]interface{}
    Stop func()
}

type Web struct {
    route map[string][]Handle
    patternAry [][]string
    patternRegx regexp.Regexp
}

/* private method */

func (web *Web) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    matchMap, pattern := web.match(r.URL.Path)
    ctx := &Context{w, w, r, matchMap, make(map[string]interface{}), nil}
    allHandle := make([]Handle, 0)
    handleAry1, ok1 := web.route[r.Method + ":" + pattern]
    if ok1 {
        allHandle = append(allHandle, handleAry1...)
    }
    handleAry2, ok2 := web.route[r.Method + ":*"]
    if ok2 {
        allHandle = append(allHandle, handleAry2...)
    }
    if len(allHandle) != 0 {
        for _, handle := range allHandle {
            keep := true
            ctx.Stop = func() {
                keep = false
            }
            handle(ctx)
            if !keep {
                return
            }
        }
    } else {
        http.NotFound(w, r)
    }
}

func (web *Web) match(path string) (matchMap MatchMap, pattern string) {
    pathAry := web.patternRegx.FindAllString(path, -1)
    matchMap = make(map[string]string)
    for _, patternItem := range web.patternAry {
        if len(pathAry) != len(patternItem) {
            continue
        }
        for j, patternKey := range patternItem {
            if j > len(pathAry) - 1 {
                break
            }
            pathKey := pathAry[j]
            if strings.HasPrefix(patternKey, ":") {
                name := strings.TrimPrefix(patternKey, ":")
                matchMap[name] = pathKey
            } else {
                if pathKey != patternKey {
                    break
                }
            }
            // match success
            if j == len(patternItem) - 1 {
                pattern = strings.Join(patternItem, "/")
                return
            }
        }
    }
    return
}

func (web *Web) addHandle(method string, pattern string, handle Handle) {
    path := method + ":" + pattern
    if _, ok := web.route[path]; !ok {
        web.route[path] = make([]Handle, 0)
    }
    web.route[path] = append(web.route[path], handle)
    web.patternAry = append(web.patternAry, strings.Split(pattern, "/"))
}

/* public api */

func (web *Web) Use(handle Handle) {
    methods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"}
    for _, method := range methods {
        web.addHandle(method, "*", handle)
    }
}

func (web *Web) Get(pattern string, handle Handle) {
    web.addHandle("GET", pattern, handle)
}

func (web *Web) Post(pattern string, handle Handle) {
    web.addHandle("POST", pattern, handle)
}

func (web *Web) Put(pattern string, handle Handle) {
    web.addHandle("PUT", pattern, handle)
}

func (web *Web) Delete(pattern string, handle Handle) {
    web.addHandle("DELETE", pattern, handle)
}

func (web *Web) Options(pattern string, handle Handle) {
    web.addHandle("OPTIONS", pattern, handle)
}

func (web *Web) Head(pattern string, handle Handle) {
    web.addHandle("HEAD", pattern, handle)
}

func (web *Web) Listen(addr string) {
    err := http.ListenAndServe(addr, web)
    if err != nil {
        fmt.Println(err)
    }
}

func New() (web Web) {
    web = Web{}
    web.route = make(map[string][]Handle)
    web.patternAry = make([][]string, 0)
    patternRegx, _ := regexp.Compile("([^/])*")
    web.patternRegx = *patternRegx
    return
}
