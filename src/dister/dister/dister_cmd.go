package dister

import (
    "strings"
    "fmt"
    "encoding/json"
    "strconv"
    "gitee.com/johng/gf/g/os/gfile"
    "gitee.com/johng/gf/g/net/ghttp"
    "gitee.com/johng/gf/g/os/gconsole"
    "gitee.com/johng/gf/g/encoding/gjson"
    "gitee.com/johng/gf/g/os/glog"
)

// 显示帮助信息
func cmd_help () {
    //fmt.Printf("dister version %s\n", gVERSION)
    fmt.Printf("Version  : %s\n", gVERSION)
    fmt.Printf("Usage    : dister [command] [option]\n")
    fmt.Printf("Commands : \n")
    fmt.Printf("    ?,help                      : this help\n")
    fmt.Printf("    nodes                       : show all nodes of this group\n")
    fmt.Printf("    kvs                         : show all key-value sets\n")
    fmt.Printf("    services                    : show all services\n")
    fmt.Printf("    addnode    IP/DOMAIN        : add ip/domain to this group\n")
    fmt.Printf("    delnode    IP/DOMAIN,...    : remove ip/domain from this group, multiple ips/domains seperated by ','\n")
    fmt.Printf("    addkv      KEY VALUE        : add key-value set to this group\n")
    fmt.Printf("    delkv      KEY,...          : remove keys from this group, multiple keys seperated by ','\n")
    fmt.Printf("    addservice CONFIG           : add service to this group, CONFIG specifies the service config file path\n")
    fmt.Printf("    delservice SERVICE_NAME,... : remove service from this group, multiple service names seperated by ','\n")
    fmt.Printf("\n")
}

// 查看集群节点
// 使用方式：dister nodes
func cmd_nodes () {
    r, _ := ghttp.Get(fmt.Sprintf("http://127.0.0.1:%d/node", gPORT_API))
    if r == nil {
        fmt.Println("ERROR: connect to local dister api failed")
        return
    }
    defer r.Close()
    content := r.ReadAll()
    peers   := make([]NodeInfo, 0)
    j, err  := gjson.DecodeToJson(content)
    if err != nil {
        glog.Error(err)
    } else {
        if err := j.GetToVar("data", &peers); err != nil {
            glog.Error(err)
        } else {
            fmt.Printf("%12s %25s %25s %15s %12s %12s %10s\n", "Id", "Name", "Group", "Ip", "Type", "Role", "Status")
            for _,v := range peers {
                status := "alive"
                if v.Status == 0 {
                    status = "dead"
                }
                fmt.Printf("%12s %25s %25s %15s %12s %12s %10s\n", v.Id, v.Name, v.Group, v.Ip, roleName(v.Role), raftRoleName(v.RaftRole), status)
            }
        }
    }
}

// 添加集群节点
// 使用方式：dister addnode IP1,IP2,IP3,...
func cmd_addnode () {
    nodes := gconsole.Value.Get(2)
    if nodes != "" {
        params := make([]string, 0)
        list   := strings.Split(strings.TrimSpace(nodes), ",")
        for _, v := range list {
            if v != "" {
                params = append(params, v)
            }
        }
        if len(params) > 0 {
            b, _ := gjson.Encode(params)
            r, e := ghttp.Request("post", fmt.Sprintf("http://127.0.0.1:%d/node", gPORT_API), b)
            if e != nil {
                glog.Error("ERROR: connect to local dister api failed,", e.Error())
                return
            }
            defer r.Close()
            data, err := gjson.DecodeToJson(r.ReadAll())
            if err != nil {
                glog.Error(err)
                return
            } else {
                if data.GetInt("result") != 1 {
                    fmt.Println(data.GetString("message"))
                    return
                }
            }
        }
    }
    fmt.Println("ok")
}

// 删除集群节点
// 使用方式：dister delnode IP1,IP2,IP3,...
func cmd_delnode () {
    nodes := gconsole.Value.Get(2)
    if nodes != "" {
        params := make([]string, 0)
        list   := strings.Split(strings.TrimSpace(nodes), ",")
        for _, v := range list {
            if v != "" {
                params = append(params, v)
            }
        }
        if len(params) > 0 {
            b, _ := gjson.Encode(params)
            r, e := ghttp.Request("delete", fmt.Sprintf("http://127.0.0.1:%d/node", gPORT_API), b)
            if e != nil {
                glog.Error("ERROR: connect to local dister api failed,", e.Error())
                return
            }
            defer r.Close()
            data, err := gjson.DecodeToJson(r.ReadAll())
            if err != nil {
                glog.Error(err)
                return
            } else {
                if data.GetInt("result") != 1 {
                    fmt.Println(data.GetString("message"))
                    return
                }
            }
        }
    }
    fmt.Println("ok")
}

// 查看所有kv
// 使用方式：dister kvs
func cmd_kvs () {
    r, e := ghttp.Get(fmt.Sprintf("http://127.0.0.1:%d/kv", gPORT_API))
    if e != nil {
        glog.Error("ERROR: connect to local dister api failed,", e.Error())
        return
    }
    defer r.Close()
    data, err := gjson.DecodeToJson(r.ReadAll())
    if err != nil {
        glog.Error(err)
        return
    }
    if data.GetInt("result") != 1 {
        glog.Error("ERROR: " + data.GetString("message"))
        return
    }
    m := data.GetMap("data")
    if len(m) > 0 {
        // 自动计算key的宽度
        length := 0
        for k, _ := range m {
            if len(k) > length {
                length = len(k)
            }
        }
        lenstr := strconv.Itoa(length)
        format1 := "%-" + lenstr + "s : %s\n"
        format2 := "%-" + lenstr + "s : %.100s\n"
        fmt.Printf(format1, "K", "V")
        for k, v := range m {
            fmt.Printf(format2, k, v)
        }
    } else {
        fmt.Println("it's empty")
    }
}


// 查询kv
// 使用方式：dister getkv 键名
func cmd_getkv () {
    k    := gconsole.Value.Get(2)
    r, e := ghttp.Get(fmt.Sprintf("http://127.0.0.1:%d/kv?k=%s", gPORT_API, k))
    if e != nil {
        glog.Error("ERROR: connect to local dister api failed,", e.Error())
        return
    }
    defer r.Close()
    data, err := gjson.DecodeToJson(r.ReadAll())
    if err != nil {
        glog.Error(err)
        return
    } else {
        if data.GetInt("result") != 1 {
            fmt.Println(data.GetString("message"))
            return
        }
    }
    fmt.Println(data.GetString("data"))
}

// 设置kv
// 使用方式：dister addkv 键名 键值
func cmd_addkv () {
    k := gconsole.Value.Get(2)
    v := gconsole.Value.Get(3)
    if k != "" && v != ""{
        b, _ := gjson.Encode(map[string]string{k: v})
        r, e := ghttp.Request("post", fmt.Sprintf("http://127.0.0.1:%d/kv", gPORT_API), b)
        if e != nil {
            glog.Error("ERROR: connect to local dister api failed,", e.Error())
            return
        }
        defer r.Close()
        data, err := gjson.DecodeToJson(r.ReadAll())
        if err != nil {
            glog.Error(err)
            return
        } else {
            if data.GetInt("result") != 1 {
                fmt.Println(data.GetString("message"))
                return
            }
        }
    }
    fmt.Println("ok")
}

// 删除
// 使用方式：dister delkv 键名1,键名2,键名3,...
func cmd_delkv () {
    keys := gconsole.Value.Get(2)
    if keys != "" {
        params := make([]string, 0)
        list   := strings.Split(strings.TrimSpace(keys), ",")
        for _, v := range list {
            if v != "" {
                params = append(params, v)
            }
        }
        if len(params) > 0 {
            b, _ := gjson.Encode(params)
            r, e := ghttp.Request("delete", fmt.Sprintf("http://127.0.0.1:%d/kv", gPORT_API), b)
            if e != nil {
                glog.Error("ERROR: connect to local dister api failed,", e.Error())
                return
            }
            defer r.Close()
            data, err := gjson.DecodeToJson(r.ReadAll())
            if err != nil {
                glog.Error(err)
                return
            } else {
                if data.GetInt("result") != 1 {
                    fmt.Println(data.GetString("message"))
                    return
                }
            }
        }
    }
    fmt.Println("ok")
}

// 查看所有Service
// 使用方式：dister services
func cmd_services () {
    r, e := ghttp.Get(fmt.Sprintf("http://127.0.0.1:%d/service", gPORT_API))
    if e != nil {
        glog.Error("ERROR: connect to local dister api failed,", e.Error())
        return
    }
    defer r.Close()
    data, err := gjson.DecodeToJson(r.ReadAll())
    if err != nil {
        glog.Error(err)
        return
    } else {
        if data.GetInt("result") != 1 {
            fmt.Println(data.GetString("message"))
            return
        }
    }
    services := data.GetMap("data")
    if services != nil {
        s, _ := json.MarshalIndent(services, "", "    ")
        fmt.Println(string(s))
    }
}

// 查看Service
// 使用方式：dister getservice Service名称
func cmd_getservice () {
    name := gconsole.Value.Get(2)
    r, e := ghttp.Get(fmt.Sprintf("http://127.0.0.1:%d/service?name=%s", gPORT_API, name))
    if e != nil {
        glog.Error("ERROR: connect to local dister api failed,", e.Error())
        return
    }
    defer r.Close()
    data, err := gjson.DecodeToJson(r.ReadAll())
    if err != nil {
        glog.Error(err)
        return
    } else {
        if data.GetInt("result") != 1 {
            fmt.Println(data.GetString("message"))
            return
        }
    }
    service := data.GetMap("data")
    if service != nil {
        s, _ := json.MarshalIndent(service, "", "    ")
        fmt.Println(string(s))
    }
}

// 添加Service
// 使用方式：dister addservice Service文件路径
func cmd_addservice () {
    path := gconsole.Value.Get(2)
    if path == "" {
        fmt.Println("please sepecify the service config file path")
        return
    }
    if !gfile.Exists(path) {
        fmt.Println("service config file does not exist")
        return
    }
    r, e := ghttp.Post(fmt.Sprintf("http://127.0.0.1:%d/service", gPORT_API), gfile.GetContents(path))
    if e != nil {
        glog.Error("ERROR: connect to local dister api failed,", e.Error())
        return
    }
    defer r.Close()
    data, err := gjson.DecodeToJson(r.ReadAll())
    if err != nil {
        glog.Error(err)
        return
    } else {
        if data.GetInt("result") != 1 {
            fmt.Println(data.GetString("message"))
            return
        }
    }
    fmt.Println("ok")
}

// 删除Service
// 使用方式：dister delservice Service名称1,Service名称2,Service名称3,...
func cmd_delservice () {
    s := gconsole.Value.Get(2)
    if s != "" {
        params := make([]string, 0)
        list  := strings.Split(strings.TrimSpace(s), ",")
        for _, v := range list {
            if v != "" {
                params = append(params, v)
            }
        }
        if len(params) > 0 {
            b, _ := gjson.Encode(params)
            r, e := ghttp.Request("delete", fmt.Sprintf("http://127.0.0.1:%d/service", gPORT_API), b)
            if e != nil {
                glog.Error("ERROR: connect to local dister api failed,", e.Error())
                return
            }
            defer r.Close()
            data, err := gjson.DecodeToJson(r.ReadAll())
            if err != nil {
                glog.Error(err)
                return
            } else {
                if data.GetInt("result") != 1 {
                    fmt.Println(data.GetString("message"))
                    return
                }
            }
        }
    }
    fmt.Println("ok")
}


// 负载均衡查询
// 使用方式：dister balance Service名称
func cmd_balance () {
    name := gconsole.Value.Get(2)
    r, e := ghttp.Get(fmt.Sprintf("http://127.0.0.1:%d/balance?name=%s", gPORT_API, name))
    if e != nil {
        glog.Error("ERROR: connect to local dister api failed,", e.Error())
        return
    }
    defer r.Close()
    data, err := gjson.DecodeToJson(r.ReadAll())
    if err != nil {
        glog.Error(err)
        return
    } else {
        if data.GetInt("result") != 1 {
            fmt.Println(data.GetString("message"))
            return
        }
    }
    service := data.GetMap("data")
    if service != nil {
        s, _ := json.MarshalIndent(service, "", "    ")
        fmt.Println(string(s))
    }
}
