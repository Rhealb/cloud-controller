# cloud-provider-alibaba-cloud

本项目fork自 https://github.com/kubernetes/cloud-provider-alibaba-cloud，对loadbalancer功能做了部分修改，使一个slb可以支持添加多个svc。

## 准备

1，kubernetes配置修改：

在kube-apiserver，kube-controller-manager,kubelet中添加参数：

1）kube-apiserver：

```
--cloud-provider=external
```

2）kube-controller-manager

```
--cloud-provider=external
```

3）kube-proxy

```
--cloud-provider=external
```

4）kubelet

```
--cloud-provider=external --hostname-override=cn-shanghai.i-uf63wpxafsmdnyaihkjq --provider-id=cn-shanghai.i-uf63wpxafsmdnyaihkjq 
```

其中机器号是通过以下命令获取的

```
$ echo `curl -s http://100.100.100.200/latest/meta-data/region-id`.`curl -s <http://100.100.100.200/latest/meta-data/instance-id`>
```

## 使用

### 1，创建公网loadbalance

```
apiVersion: v1

kind: Service

metadata:

  name: nginx

  namespace: default

spec:

  ports:

  \- port: 29007

    nodePort: 29007

    protocol: TCP

    targetPort: 80

  selector:

    run: nginx

  type: LoadBalancer


```

创建以上service，就会创建一个名为ali-slb（默认外网slb名字）的loadbalance，若ali-slb已存在，就会在slb中添加对应的监听端口29007。

```
# kubectl get svc 

NAME         TYPE           CLUSTER-IP     EXTERNAL-IP    PORT(S)           AGE

nginx        LoadBalancer   10.254.235.239   139.224.166.231  29007:29007/TCP   6s
```

再创建一个svc，加入同样的loadbalance（ali-slb）：

```
apiVersion: v1

kind: Service

metadata:

  name: nginx-2

  namespace: default

spec:

  ports:

  \- port: 29008

    nodePort: 29008

    protocol: TCP

    targetPort: 80

  selector:

    run: nginx

  type: LoadBalancer
```

```
# kubectl get svc 

NAME         TYPE           CLUSTER-IP     EXTERNAL-IP    PORT(S)           AGE

nginx        LoadBalancer   10.254.235.239   139.224.166.231   29007:29007/TCP   1m

nginx-2      LoadBalancer   10.254.86.56     139.224.166.231   29008:29008/TCP   5s
```

### 2, 创建内网loadbalance

```
apiVersion: v1

kind: Service

metadata:

  annotations:

    service.beta.kubernetes.io/alicloud-loadbalancer-address-type: "intranet"

  name: nginx-intranet

  namespace: default

spec:

  ports:

  \- port: 29009

    nodePort: 29009

    protocol: TCP

    targetPort: 80

  selector:

    run: nginx

  type: LoadBalancer
```

创建一个名为ali-slb-internal（默认内网slb名字）的loadbalance，若ali-slb-internal已存在，就会在slb中添加对应的监听端口29009。

```
# kubectl get svc 

NAME             TYPE           CLUSTER-IP       EXTERNAL-IP    PORT(S)           AGE

nginx            LoadBalancer   10.254.235.239   139.224.166.231   29007:29007/TCP   4m

nginx-2          LoadBalancer   10.254.86.56     139.224.166.231   29008:29008/TCP   2m

nginx-intranet   LoadBalancer   10.254.29.45     172.16.1.205      29009:29009/TCP   20s
```

再创建一个svc，加入同样的loadbalance（ali-slb-internal）：

```
apiVersion: v1

kind: Service

metadata:

  annotations:

    service.beta.kubernetes.io/alicloud-loadbalancer-address-type: "intranet"

  name: nginx-intranet-2

  namespace: default

spec:

  ports:

  \- port: 29010

nodePort: 29010

    protocol: TCP

    targetPort: 80

  selector:

    run: nginx

  type: LoadBalancer
```

```
# kubectl get svc

NAME               TYPE           CLUSTER-IP       EXTERNAL-IP       PORT(S)           AGE

nginx              LoadBalancer   10.254.235.239   139.224.166.231   29007:29007/TCP   5m

nginx-2            LoadBalancer   10.254.86.56     139.224.166.231   29008:29008/TCP   4m

nginx-intranet     LoadBalancer   10.254.29.45     172.16.1.205      29009:29009/TCP   2m

nginx-intranet-2   LoadBalancer   10.254.211.36    172.16.1.205      29010:29010/TCP   2s
```

### 3，自定义slb

即不在默认的slb ali-slb loadbalance上添加svc，自定义创建slb

```
apiVersion: v1

kind: Service

metadata:

  annotations:

    enndata.io/ali-load-balancer-name: "test-slb"

  name: nginx-test

  namespace: default

spec:

  ports:

  \- port: 29011

nodePort: 29011

    protocol: TCP

targetPort: 80

  selector:

    run: nginx

  type: LoadBalancer 
```

会创建一个新的slb，test-slb，如果test-slb已存在，就会在test-slb中添加监听端口，同用法1.

```
# kubectl get svc

NAME         TYPE           CLUSTER-IP       EXTERNAL-IP       PORT(S)           AGE

nginx        LoadBalancer   10.254.235.239   139.224.166.231   29007:29007/TCP   11m

nginx-2      LoadBalancer   10.254.86.56     139.224.166.231   29018:29018/TCP   9m

nginx-intranet     LoadBalancer   10.254.29.45     172.16.1.205      29009:29009/TCP   5m

nginx-intranet-2   LoadBalancer   10.254.211.36    172.16.1.205      29010:29010/TCP   4m

nginx-test   LoadBalancer   10.254.156.110   47.102.62.60      29011:29011/TCP   16s
```

所有的slb，除了ali-slb，若listeners为0就会被删除。

### 4，指定规格loadbalance

在annotations中添加参数：

```
 service.beta.kubernetes.io/alicloud-loadbalancer-spec: "slb.s1.small"
```

以下内容引用自https://help.aliyun.com/document_detail/27657.html?spm=5176.8009612.101.7.4fe171b3tmliZC

阿里云负载均衡性能保障型实例开放了如下六种实例规格（各地域因资源情况不同，开放的规格可能略有差异，请以控制台购买页为准）。

| **规格** | **规格**                 | **最大连接数** | **每秒新建连接数 (CPS)** | **每秒查询数(QPS)** |
| -------- | ------------------------ | -------------- | ------------------------ | ------------------- |
| 规格 1   | 简约型I (slb.s1.small)   | 5,000          | 3,000                    | 1,000               |
| 规格 2   | 标准型I (slb.s2.small)   | 50,000         | 5,000                    | 5,000               |
| 规格 3   | 标准型II (slb.s2.medium) | 100,000        | 10,000                   | 10,000              |
| 规格 4   | 高阶型I (slb.s3.small)   | 200,000        | 20,000                   | 20,000              |
| 规格 5   | 高阶型II (slb.s3.medium) | 500,000        | 50,000                   | 30,000              |
| 规格 6   | 超强型I (slb.s3.large)   | 1,000,000      | 100,000                  | 50,000              |

 

 

 

 