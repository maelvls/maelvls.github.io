package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/ori-edge/deploy-controllers/cmd"
	"github.com/ori-edge/deploy-controllers/pkg/clientcfg"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	clusterv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clusterv1alpha3.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
}

func main() {
	kubeconfigData := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01EVXdOekV3TWpFME9Gb1hEVE13TURVd05URXdNakUwT0Zvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTUJlCjZEaUtrZGNLR0R1RzNGNVpUTUFmZ0toMTczb3VVT0VRRy9RZ3g3ZjE0eTBlOVdrRnJiOVBVaWxKWUtRZyt4OEkKNFBYbE5nbWhMSmcrbWdLVXMrVVdQNGlHbjdEVmdXWHpXTHJSeHN4ek9UMFdTZ2lXSGJGc2FyU25ZamNyQU90YQptNnUwUkJqb1k1K0tnT1U3eW12WFBLRXpYc3JoVkRBbEVNeXQ5NElIME5xdFY5SUI5ZnpaODdPSnVaUmFJYy9SCjhXUkJJVU5TSUZjTEpwbk9vYTdrRWlYTS8xS3B2bVEwRFBWcWl4WTd1clBOYUtwQXBxZ0E5eTZXMThHMWhLejUKLzlWS09xR0dWdjhiVWhDWUgyaWQrRlpUc0ZaR3hpOXh5UGNMeno1Tk5aY01MTU1qYm1nZ21ieTJZK0JOUGtUTQpZNC8rRGR1OU9LU3ZtSk5Fb0QwQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFGQ0E1QSttV1R4bjNBNnZad0NKMXVRTi9YUloKM1d5TndUcjRSRGd0QXJieTRSdXpja1RYNGpyaEdIeERLM3NUeXFNRFp0UVFZVWZyS3dIV2pkYU1EOXA5cHkySQphYzR1cWZhTVFWeUpESVozeGNSNkJtYU4reDNXRlpIUisxMjh1aVg3dU9YLzh4R1hmbG1yMTBsaWl6bTYzMnVZCmJNbXRvKzRqK3BYTnowcmRQTWRZc2ZkY2FrU1hxb01YTHV0UmJKTlVCRUw0aWNPbTZsQk0wSGoxbmdHQnE0QjkKY01UMUUwWVA2RjFRc0czd2pHa2djdGtNUzJhSXc4ZVFMdHdEYUwxb0QyVWZOd2JwMS9CSVh6U3FSVk5PeXNzQgp3UVU2d0dhWTMxeVhQU1NBVmIwRUw4SFc1cFgvM0JsRkk3NlJ4Q3FIQmFwNHFlMFMvU3JuUGo4N01sWT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://127.0.0.1:32768
  name: kind-capd
contexts:
- context:
    cluster: kind-capd
    user: kind-capd
  name: kind-capd
current-context: kind-capd
kind: Config
preferences: {}
users:
- name: kind-capd
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJYW1ZbENKOTBNZ013RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TURBMU1EY3hNREl4TkRoYUZ3MHlNVEExTURjeE1ESXhOVGRhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQSt3eGpJdDRXVTN1UTNpby8KZnRzTGRCZ2t3RTUvRG9xTEYwM0VTMnFWeFdRcTZtRFEyZXhOTVdJdFlyNjdnUU5LRXcydEJtSXM3eUhBdnBUMgpsNjRoNmJDY0pKbkQyeHZjOWdRVzZ1YkRXcW1lZkdEWGEvNUpNQlZ6QTVpT0FIV3Jldm9CRC9KL2tnSVc4UXZjCjZPejE1aXJNaTROT2s1Y2JCRHhlY09LMmlGUk1Ca0VxWktUVEZDb3RvVHYxRjI1UzhlbmRCeFBZUjBwTFdKb2gKbkdPSlB3NkVoMEF1ZGRwaWl2RzlpV3kxRzBCcW4xUEhobGc3aTVEdytIMDdSdFVQZEJ0T1V5R0xINUhqQk9yWQorNCtON1VQSnZvcWgzbkpnRjRuYkxCRXJnNG5VRkJaVk5jbXVwVEo5OTl3S3ZBa1VnN1FxTGJwU0p6MGpSUWh5Ckk0OEx0UUlEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDbTNVTnk1VmRMbEtSbzdTSXZUQ3FHTnJaNkRtdmR6aTRGegpVSGtwMmUyVUp6NzVGWng5QkREazJhMVBxcHp2dlRQQ2U0ZmZYVnlVUzl3a3FxMHN3KytPUWRhMTdTZkg1ajVpCjhBZXNpb3JsSTg4U1pHdTJ1TlJmMko2ZFNRSmc2eHp1cWlZQ29lWXZEOUcxWXB0YitOVGEzeFBBd2tPL2tXeWYKRHM3OXFKS3ZJZ1dJUEhGMDk5eENOaHNLVWZCS09FdXhWSXRGZGh4ZXY5clg1Sy9kZ01PdU1WRFpFTTkyVm5iNgpnN0hoTFRhb2NvaHFTeWRVNnlmSE9SNlhSejhiWldVQ2pxY0JhY1ppWWQ0clBMTmFWOWFNSC9mL2NUQzNteW1LCjNKc2xFeGNvQiswWjAzS2pFZklwcDl2QWFWeENjT21iUENxdDBiQThoRVRRckFPV1I5ST0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBK3d4akl0NFdVM3VRM2lvL2Z0c0xkQmdrd0U1L0RvcUxGMDNFUzJxVnhXUXE2bURRCjJleE5NV0l0WXI2N2dRTktFdzJ0Qm1Jczd5SEF2cFQybDY0aDZiQ2NKSm5EMnh2YzlnUVc2dWJEV3FtZWZHRFgKYS81Sk1CVnpBNWlPQUhXcmV2b0JEL0ova2dJVzhRdmM2T3oxNWlyTWk0Tk9rNWNiQkR4ZWNPSzJpRlJNQmtFcQpaS1RURkNvdG9UdjFGMjVTOGVuZEJ4UFlSMHBMV0pvaG5HT0pQdzZFaDBBdWRkcGlpdkc5aVd5MUcwQnFuMVBICmhsZzdpNUR3K0gwN1J0VVBkQnRPVXlHTEg1SGpCT3JZKzQrTjdVUEp2b3FoM25KZ0Y0bmJMQkVyZzRuVUZCWlYKTmNtdXBUSjk5OXdLdkFrVWc3UXFMYnBTSnowalJRaHlJNDhMdFFJREFRQUJBb0lCQVFEZVRLbThSa3dld0Z3WApYZkc3c3RzQmdoK0k2Zms0TnhYVEhObWtya3pRN1ZIVEdNZlhNSmRxRXpWOUtzZCtCaHVobzRxRERJd2RkQlhvCnJKOXUwSkxYQzd3MzdMQ3haSXJVamVwOU5ybmxuaXpvbGhncldKQVdNK2dVVnhIbTlrdFNLZTZtNEdSMk5jTjYKenJaZXl2VXpTdEswOXlDdE1EQ01INmpBN1FBVTFrOURWZ2FXVzZhRUQ5cUgxWnRocTBxN1hUdWNXL2Z5M2tTRgprbmRjR0Rqb3hSMVdYQU9tbUxDckRkR21tNlJwY08zb0sxdmxnZy9pSWlyMVBGV1N3Q1BlZWpXSUR1QjVUSDZOCnB4YU1Ub001ZzR0NURLM2FBUjBRamxpVXB0Wi9HcEd2THlSYzV2TGFhSmFBZjN1alpKcVNhTGxvZHhlYUN6QmMKNlphaUk3WkJBb0dCQVAwdzdRandhS2JlMC9ITGp6aXEzT2JmVW84czFLaXN3TVFoRjFyZnNjNFIveHhHaXV5MAoyM0JGTFJ6K3RNNFZLdXUxNGxNQ0xEVHhsc1hxNTg3V2xvLzQ4VGJheGJWcFVnMG5qelRGNTQxWGVzczg2ZklLCmxKSzk4c3NQbzZOQVVUZDlUeDQ2VUZjeHJUQ3FrVnhRTkhvRUJVcTFYalpPTHZJd3dab3gwem12QW9HQkFQM1YKWURvYVgrYWRzdnYrcnFDblF6dkQ3WVkwZk1qcGo3VHYxdXhkRHM1d3diMkpGV3Bsdi9TVjBIL1h6anB0ME1tTQpGOTUwcHlyS21WZkU3RU1VNHRxSnBHUms3ZWJCc216dmVQd2p0TWVYR0tpeHFHdWR2QnA5Q3RqbkpDWU9vZzZ0Ck9ETlVXMnFBbVRPMkZFZWtPV2M1dU5mcHBDd2xzZVU5R1lPd1FqM2JBb0dBZEd6akVwRTZEa0c0eEI4T3BNZ3MKL0IwRkljRkRxS3lIbDZoL3pOSEFPVG9kVFN0REJzWERna1RORWVBdDAvWDMzcHVzanU4WTFOK2lyUy92bURVawoxdDlxVEFjZGt1WHpUUWs3Mk5DSVFYNVFnTlJwMzFydUp1d2hrUzZIMkxIaXB0bUFZQzRBYzVmc1E4eXJPdi9HCm9iVG5tZ3I4WDR4a0dncEJmRjRjK3hFQ2dZRUFvNnEraGhoVmQ2eDlLTkM1cG1yVEJpazU4UXZNM2ZzREp5WnkKVFN0ZmphclVzVEkvdGIvdnVuUVM0U3UwRkthVU5qQjNmMzkxL2podUVWS3ZDRDNpWEFqZUQ4R29SOTdpL2l5Vwp0UFVNN3BpMVZLaGdzU3NlaTNITzJiYUg3MllHQmpLWWh6aEFUWGFuMGRqNFVJMUtXZzIwNnJzQ21WaWcwTy9KCmtNakluWHNDZ1lFQTMyRWpNZ0hQWGJ0YzEveERoekhQWGcxb3U1SDZ2c0d1WFpsWTRndU5XZWpna002UmUxRXEKYmxnUjE2aC9JWjNtaWtkL1VIeHo5cm9XbkZoZGo3S3hiN0F1ZDVrN1c5b01jZm5Ja1hKWUY0UmkxUEVjcGZ2bgpIN1lIVEVsQWpMM1lkUkI3VHpMNXFtZzVmbUlwTFUvVytzS2ZJRTF5ZjhqVnltdVdLdU90NC9rPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
`

	klog.InitFlags(flag.CommandLine)
	defer klog.Flush()

	opts := cmd.DefaultOptions
	opts.FillOptionsUsingFlags(flag.CommandLine)

	flag.Parse()

	userAgent := fmt.Sprintf("deploy-controllers/%s/%s", "version", "commit")

	cfg, err := clientcfg.RestConfig(opts.Kubeconfig, opts.KubeconfigContext, userAgent)
	if err != nil {
		panic(err)
	}
	kclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	kclient.CoreV1().Secrets("default").Get("capd-kubeconfig", v1.GetOptions{})
	err = ioutil.WriteFile("/tmp/kubeconfig", []byte(kubeconfigData), 0666)
	if err != nil {
		panic(err)
	}

	rule := clientcmd.NewDefaultClientConfigLoadingRules()
	rule.ExplicitPath = "/tmp/kubeconfig"
	apicfg, err := rule.Load()
	if err != nil {
		panic(err)
	}
	restcfg, err := clientcmd.NewDefaultClientConfig(*apicfg, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		panic(err)
	}

	c, err := client.New(restcfg, client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	cluster := clusterv1alpha3.Cluster{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: "capd", Namespace: "default"}, &cluster)
	//err = c.List(context.TODO(), &cluster)
	if err != nil {
		panic(err)
	}
	fmt.Printf("cluster: %v", cluster)

	d := appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "capd-deployment-1",
			Labels: map[string]string{
				"ori.co/cluster-name": "capd",
			},
			Namespace: "default",
		},
	}
	err = c.Create(context.TODO(), &d)
	if err != nil {
		panic(err)
	}
}
