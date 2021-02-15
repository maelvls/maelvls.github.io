---
title: "Using mitmproxy to understand what kubectl does under the hood"
description: "Mitmproxy is an excellent tool that helps us understand what network calls are made by programs. And kubectl is one of these interesting programs, but it uses a mutual TLS authentication which is tricky to get right."
date: 2020-07-01T19:17:24+02:00
url: /mitmproxy-kubectl
images: [mitmproxy-kubectl/cover-mitmproxy-kubectl.png]
tags: [kubernetes, mitmproxy, networking, kind]
author: Maël Valais
devtoId: 377876
devtoPublished: true
---

{{< youtube 30a0WrfaS2A >}}

In Oct 2019, Ahmet Alp Balkan wrote [this blog post](https://ahmet.im/blog/kubectl-man-in-the-middle) that explains how to use `mitmproxy` to observe the requests made by `kubectl`. But I couldn't use the tutorial for two reasons:

- I use `kind` to create local clusters which means I hit the Go `net/http` limitation (skips proxying for hosts `localhost` and `127.0.0.1`, see [this blog post](<(https://maelvls.dev/go-ignores-proxy-localhost/)>))
- I use client certs authentication, which can't work with the method presented by Ahmet; it can only work for header-based authentication (e.g. token) but not for client certs.

In the following, I detail how I managed to make all that work.

```sh
# Let's use a separate kubeconfig file to avoid messing with ~/.kube/config.
export KUBECONFIG=~/.kube/kind-kind

# Now, let's create a cluster.
kind create cluster

# Let's see what the content of $KUECONFIG is:
cat $KUBECONFIG
```

We can see that the kubeconfig uses a client cert as a way to authenticate to the apiserver.

```yaml
apiVersion: v1
clusters:
  - cluster:
      certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01EWXlPVEE1TURJeU4xb1hEVE13TURZeU56QTVNREl5TjFvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBS25qCmhmRzBvenZVb05jMXY1STkvYm13dFBqb2QvK0RyczF4TFZOcWgxQjhFcTY1S3lnQjBSbDdQcTJueUhyRVRnTmsKZ2VPQ2RzRURObGpKOE12SC9GbE9nRkUvS0dqK2hwN2hwc0dHRExReWFUOExUY25JN0FNL2t3KzY5R0hidlN3Nwo2V0Y1VERTMDU1RkRqdnRveGdHcERycmZQZTk5bXN5Zmk3aWtteDk5MmRyMHFQd0xxanJpZHNkWU52MUZqU1Y1CndkRlFISGxBS2hBcmlUWmpQMnhNL3poOENBOFhndjF0UUxVVk5IS1hrSG5UYWlkeFY1MkduaUVaQmd0d2tSK3oKQ3hQempVZFAxQ1JRYzU0YmxDYW9FQVdTc0NYTUVPREhTSnowSi9CWHJCU2JaeGZIakd0Y0k0bEhBUmx2aURNawp3U3lOLy9qdE9tbWhDT29BTzBzQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFIMDBrYnZpaWNNT3IxdFJoTkQweVpGa3ZkT1AKUzFGOEZKK1BFd1o1WExUTVVyVG1yekVlZmgza21xWkxYUnlyK3c0Snk1a0grK1o3enBpdlp6Q3BGOEtwclJaWAp5N2Z6TkJOeWMrOHFKN3dCek0xZ21BdTRha3BlNVBYbkw3akZIcU9aUmNXTmZKcUpTci9BS05sK2c2SjAzV2pnCmJDc0NQNTNrZ1czSDZYMkRoS1JzZFRWQlc1UGVlek1YRVlON2VYbGpnWE4xcElMdEU1VE5oWVZIcTZaUDVCaDkKMmU2Y1U4bEFuNlhYQ0Ira2RsTkhudGp1cTlSL2lPQVJRaWpVSVlWSFdCT1NyWGp6TnhkTklEakRRd3lKUnViMApDVEhrcDg4eUlNaEV1Nk1OV1VCUU82ZnVqbURHbUJJbTlzdzJpSnI1UlArSUVUVnQyWFc0Q2dHT1cwVT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
      server: https://127.0.0.1:53744
    name: kind-helix
contexts:
  - context:
      cluster: kind-helix
      user: kind-helix
    name: kind-helix
current-context: kind-helix
kind: Config
preferences: {}
users:
  - name: kind-helix
    user:
      client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJYTZXc25HT0RGZW93RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TURBMk1qa3dPVEF5TWpkYUZ3MHlNVEEyTWprd09UQXlNekJhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXRkbkxsb0V1cnFoSmRrMGgKME5VcXpFSHhUbmVHS040QTNtWDBobmFLcE1TT3hISlBQcTB6V0t4WEVxNTZPTkdhSkhvS0VTRUM2Yjh2MGcwTQo5dXVGOXJyQ1VORkpXMmdHcFArRG96TEpVS3RtN2F0WFZRYXNVQ2tjbFRtQ003aHlqOGZTTlRDd2lxT3RNOUFvCnFVVTJYd1k0a2xhL0RuMkRBc1V1VmNiUTdITDA1N0tFbjZvVzlFVy9mQ2hMVE5OTWhmelpkNzFISGs3MFNOZDMKTG5jTXNjNmp2eS9kN2MwbjM4ek45SjVzWVZOTjNsS1YvOU5MT2FaTEQ0bSsyVU5XV1FVazRuci8rZjlaZk1IMwphVlFibDZSWEZwaW1jYWg4UjRJTmhRNkhYbmVEbUI2dUl3RGdjQnhhQUtoNFVPVlZwODB1Uy9pYUVLV1VSWE1PCkg4YXBZUUlEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFJaWtmbis5TkJQc2FNSzNVMU5pWmtBaWxsM1pWVXJDdzNsWgpIeGRGbm9MZGxZYmtPeFVEN25EK3lrNXhZWW4yaDc3WUU4NlU5czJ3aitkdFJTR2Ezc3p4WVJQbkt5MDN3L1M4CkhqZzhwSkZOZFhHWlNPUWJDQTBmT1BONXl5MlpYRlVWd0JwZ3VTMC9nTkNUUkRPUDl3QmNjQXhiSGRmSjMxQzQKa2hCemF4QXI5UEliMVBzUlhqS2ZSRnkvcllwYWRBYVhhYmZMOFRvbTJ4cGpLVDYreGxoL3lJb2tZbVhxSnlXRgp1SGFmWG1qUEdyRDFoZUo2UnR5U0xRM0dKUXQxWmVPK3BaYlJlRjcxU0ZRSjRQdFhtTHhGMjZQL3dlV3F0NWFJClhTK0ZnanRYbzZqUG1mWWRweDZ3bWp3UHRaZ1pRdXhSL2tsejVvQnErNnFaYUZDWXg4RT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
      client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBdGRuTGxvRXVycWhKZGswaDBOVXF6RUh4VG5lR0tONEEzbVgwaG5hS3BNU094SEpQClBxMHpXS3hYRXE1Nk9OR2FKSG9LRVNFQzZiOHYwZzBNOXV1RjlyckNVTkZKVzJnR3BQK0RvekxKVUt0bTdhdFgKVlFhc1VDa2NsVG1DTTdoeWo4ZlNOVEN3aXFPdE05QW9xVVUyWHdZNGtsYS9EbjJEQXNVdVZjYlE3SEwwNTdLRQpuNm9XOUVXL2ZDaExUTk5NaGZ6WmQ3MUhIazcwU05kM0xuY01zYzZqdnkvZDdjMG4zOHpOOUo1c1lWTk4zbEtWCi85TkxPYVpMRDRtKzJVTldXUVVrNG5yLytmOVpmTUgzYVZRYmw2UlhGcGltY2FoOFI0SU5oUTZIWG5lRG1CNnUKSXdEZ2NCeGFBS2g0VU9WVnA4MHVTL2lhRUtXVVJYTU9IOGFwWVFJREFRQUJBb0lCQVFDVkNJOXZJeVB0QkFKZwpyOG44NmhhUEc2UDFtTU1jandUTFAyZHRJNDF3aDU0eHBUVUl1czJQNkgzYjA1NWJIbnhqVkprWGZLUjBpTGxhClBsUFhzU0l6R00vVGlCSEVsYmFNVnRPOVZndml6dllsNWZ4R3RKZFhncm5vR2g5NDM3c1Qxc0dSMGZ0OVE3TFkKK2NtNUgvMzFWcFhhYUxsZjJNRWI3aG1STnNWV1lXWm9MUHJ2QUJPemVmUnpKU0RONHU0M3M0eVpnNHJoL3IzWQo3YUhYOS9CSHpJRk93WTRNL1BRQzdYaFhDMXBIeWNPY2lSWVhFTkpuMTdQN2NOSkdoMnJ4dnc2OVQ1QW9rdXY5Ckcxd2lUNmVrVmZuaVYrSlVnL1lwcTNSRGxNdVFETDRZdCtGNG1zVGJtN3NFZk5yVVMzWGZFMFdmcjB2Wk1YN2sKMnE3dXkxNEJBb0dCQU5jb1BxVHNBa3N5YzR6UWtmQmNzb0Y1ekdJM3NQTUFGTHRSMGF6S2x1Q1V3VGMzYVk5dgpiTE05OTBVK3VpVzdtRHF0VHpZMVNVSzZGNW0yMzJoSW5hTnBCTUVJcDZHQ0I3dis5WWZIQURVSkljaVdodmJhCmpIY0M5Qjl4SG5uUTBydU5aNUErdWVtRE9EZkRKUllIVUhrMEh3MlJpTE1qRFhVRmhBanJBUGJ4QW9HQkFOaGUKL3BHb2FWOWtUREtNOWZwTTZUQkZwMFVtb2lSb3crVG9TWjhWeVBmMkl1VlNpa0dtaXUydHU3MzZCZkJtVzNtQgpGWmFRc01rNkMydExkK1NBQ0lSbktQTUN4czYzR05WNkJ0aXRIRFlBaGNtRCtvU3VpaFRrd3l1WXlpQml4aFovCms1UDY3UHFKVkQ2YnhkNDV6YlFsS2VNaUVtMUhnQUVGSEF0UFBqbHhBb0dBYzNnaHhwankwakNOV3ZGRW9WN2UKWGlaanpnSmRjTXlHVTlHaFdiNlFJbzh5OHROR1Q3aFkrZ2t6ZjNJZXJNbDA5V2kxcmo0Q3gxRGdBWnJuWXl3MQpqZEY2djY1SmFLQkVUbHlTb1AvbjJJN0NGc2pTUGdFa2lXcUlZYWR2MTZoK3NERS9kMlp5bUNQWU0vVURIa05tCnFPV1VGTkFhTVNtS3UxYnVlV3JGNWNFQ2dZQnJvNTVySWVnQjM2aVVnVkdoU24rN1Z2dG15RmhqV29jUnFvbHQKamUzamhWeEl6eTRlaU5hV2RSWnY1U0R0UGs2RmZMVWJxVEY1ZWRuU2I4SGVOOStFMXJrbFk1MDVteGJNcEo4aApUY1U2RERxQ1RKamxSdHRFbDZXTVc3ODZLMGsyU2hORnk4LzJ0empreUtPLzhPdW5rZEZyd0RpQWl0QmdNWVdKCkRzdjYwUUtCZ1FDR2htYnZrODRaYm1sdERxYW9hOVR6YU5UT3FrN3dtb0tDby9FVEg5NEtNVVZMVmk3bHhpN1kKcXFQYUdiSm9aYkZGd0NvWW12Rk5jU21nYk9SNjdCVVBkdUEzMHo0VythV1pmQVkwZGpUR08yTC8zVVhoWHdHegpVUGpZTXZZQU5sdzFmZTVmZmk3UExJSGJvZXFhTmhhaDliRVJXdFNRbm95NnYzV2hlNTh6VkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
```

Let's run mitmproxy using the above client cert and keys:

```sh
mitmproxy -p 9000 --ssl-insecure --set client_certs=<(kubectl config view --minify --flatten -o=go-template='{{(index (index .users 0).user "client-key-data")}}' | base64 -d && kubectl config view --minify --flatten -o=go-template='{{(index (index .users 0).user "client-certificate-data")}}' | base64 -d)
```

Now, let's make sure that we have some DNS alias to 127.0.0.1 to work around [the Go proxy limitation](https://maelvls.dev/go-ignores-proxy-localhost/) which skips proxying for `127.0.0.1` and `localhost`:

```sh
grep "^127.0.0.1.*local$" /etc/hosts || sudo bash -c 'echo "127.0.0.1 local" >> /etc/hosts'
```

Finally you can run `kubectl` to see the logs of `kube-scheduler`:

```sh
% HTTPS_PROXY=:9000 kubectl --kubeconfig=<(sed "s|127.0.0.1|local|" $KUBECONFIG) --insecure-skip-tls-verify logs -n kube-system -l component=kube-scheduler,tier=control-plane
I0701 11:11:11.585249       1 registry.go:150] Registering EvenPodsSpread predicate and priority function
I0701 11:11:11.585329       1 registry.go:150] Registering EvenPodsSpread predicate and priority function
I0701 11:11:12.976581       1 serving.go:313] Generated self-signed cert in-memory
W0701 11:11:17.407371       1 authentication.go:349] Unable to get configmap/extension-apiserver-authentication in kube-system.  Usually fixed by 'kubectl create rolebinding -n kube-system ROLEBINDING_NAME --role=extension-apiserver-authentication-reader --serviceaccount=YOUR_NS:YOUR_SA'
W0701 11:11:17.407465       1 authentication.go:297] Error looking up in-cluster authentication configuration: configmaps "extension-apiserver-authentication" is forbidden: User "system:kube-scheduler" cannot get resource "configmaps" in API group "" in the namespace "kube-system"
W0701 11:11:17.407483       1 authentication.go:298] Continuing without authentication configuration. This may treat all requests as anonymous.
W0701 11:11:17.407513       1 authentication.go:299] To require authentication configuration lookup to succeed, set --authentication-tolerate-lookup-failure=false
I0701 11:11:17.478349       1 registry.go:150] Registering EvenPodsSpread predicate and priority function
I0701 11:11:17.478483       1 registry.go:150] Registering EvenPodsSpread predicate and priority function
W0701 11:11:17.491616       1 authorization.go:47] Authorization is disabled
W0701 11:11:17.491729       1 authentication.go:40] Authentication is disabled
I0701 11:11:17.491883       1 deprecated_insecure_serving.go:51] Serving healthz insecurely on [::]:10251
I0701 11:11:17.500576       1 secure_serving.go:178] Serving securely on 127.0.0.1:10259
I0701 11:11:17.500678       1 tlsconfig.go:240] Starting DynamicServingCertificateController
I0701 11:11:17.500919       1 configmap_cafile_content.go:202] Starting client-ca::kube-system::extension-apiserver-authentication::client-ca-file
I0701 11:11:17.500931       1 shared_informer.go:223] Waiting for caches to sync for client-ca::kube-system::extension-apiserver-authentication::client-ca-file
I0701 11:11:17.718085       1 shared_informer.go:230] Caches are synced for client-ca::kube-system::extension-apiserver-authentication::client-ca-file
I0701 11:11:17.801300       1 leaderelection.go:242] attempting to acquire leader lease  kube-system/kube-scheduler...
I0701 11:11:34.659176       1 leaderelection.go:252] successfully acquired lease kube-system/kube-scheduler
```

It works!! Here is what mitmproxy is showing:

```sh
18:17:42 GET  HTTPS   local /api/v1/namespaces/kube-system/pods/kube-scheduler-helix-control-plane      200 …plication/json  5.2k 119ms
18:17:42 GET  HTTPS   local /api/v1/namespaces/kube-system/pods/kube-scheduler-helix-control-plane/log  200      text/plain 2.46k 227ms
```
