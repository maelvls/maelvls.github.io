---
title: Migrating from GKE to Civo's K3s
description: "My free trial on GKE was ending in 2 days and I had to find a
way to migrate away. I decided to switch to Civo's managed K3s."
date: 2020-03-22
url: /from-gke-to-civo-k3s
images: [from-gke-to-civo-k3s/civo-k3s.png]
tags: [kubernetes]
---

A few days ago, reality hit me hard with this message:

<img alt="Only 2 days left on my GCP 1-year trial" src="2-days-free-trial.png" width="60%">

I had only 2 days to find a plan and migrate everything away from GKE! My
current setup was only using a single `n1-standard-1` on `us-west-1` and
with no network load balancer. But that was still around €15 a month and I
just didn't want to pay.

Note that I will still use Google's CloudDNS service for now.

I chose to migrate to Civo's managed K3s since they are in beta and I
really wanted to try K3s.

So I went ahead and create a two-nodes cluster:

![My K3s cluster on Civo](civo-k3s.png)

Note that since this is K3s, the master can also be scheduled with pods. So
I could have gone with a single node.

## Migrating external-dns, cert-manager and traefik

That's the easy part. I just had to go through my README.md and play the
`helm` commands one by one (I know I know, I should automate that 😁).

In the process, I decided to go with `*.k.maelvls.dev` instead of
`*.k.maelvls.dev`. The shorter the better!

After creating Traefik, I knew that its service would automatically be
populated with the `status.loadBalancer` field thanks to K3s'
[servicelb](https://github.com/rancher/k3s/blob/master/pkg/servicelb/controller.go).
Let's see:

```sh
% kubectl -n traefik get services
NAME     TYPE          CLUSTER-IP        EXTERNAL-IP      PORT(S)
traefik  LoadBalancer  192.168.255.134   91.211.152.190   443:32164/TCP,80:32684/TCP
```

Great! Now, Traefik will propagate this external IP to the ingresses, and
ExternalDNS will use the `status.loadBalancer` from these ingresses in
order to set `A` records.

If you want to know more about how servicelb works, you can take a look at
[The Packet's-Eye View of a Kubernetes
Service](https://maelvls.dev/packets-eye-kubernetes-service/) where I
describe how Akrobateo works (K3s' servicelb has the exact same behavior).

## Migrating MinIO

I use minio for various uses. It is great if you want a S3-compatible
storage solution. In order to migrate, I followed
[this](https://www.scaleway.com/en/docs/how-to-migrate-object-storage-buckets-with-minio).
It's quite painless:

```sh
$ kubectl -n minio run a --generator=run-pod/v1 -it --rm --restart=Never --image=alpine

% wget https://dl.minio.io/client/mc/release/linux-amd64/mc && install mc /usr/bin
% mc config host add old https://minio.kube.maelvls.dev AKIAIOSFODNN7EXAMPLE "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" --api S3v4
% mc config host add new http://minio:9000 AKIAIOSFODNN7EXAMPLE "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" --api S3v4

% mc ls old/         # List buckets since I had to create manually each bucket.
% mc mb new/bucket1  # Then create each bucket one by one.
% mc cp --recursive old/bucket1/ new/bucket1/
```

I also decided to change the access key & secret key. Again, quite
painless. As mentioned in the
[documentation](https://github.com/minio/minio/tree/master/docs/config),
I changed the secret stored in Kubernetes:

```sh
kubectl -n minio edit secret minio
```

and then I temporarily added the `MINIO_ACCESS_KEY_OLD` and
`MINIO_SECRET_KEY_OLD` to the deployment by editing it.

```sh
kubectl -n minio edit deployment minio
```

After that, the pods get recreated and MinIO picks up the new secret. Note:
I also had to edit the deployment again in order to remove the temporary
`_OLD` environment variables.

---

To recap, the whole migration was painless. The only data I migrated was
MinIO. Note that I didn't have any SLA to comply with, but if I had planned
a bit better I could have moved over with almost zero downtime.

In order to get a almost-zero-downtime, I would have made sure to keep the
old and new MinIO instances replicated until the move was over. The only
problem with the whole migration is the DNS change: I cannot precisely know
how much time that will take: after the migration was over and the DNS
entries propagated, if some people were hitting the old IP (e.g., old DNS
entries), the old and new clusters would have become out-of-sync. To
mitigate that, I could have chosen to "cut" the old cluster just to make
sure that case never happens.

The repo for my Kubernetes playground cluster (`*.k.maelvls.dev`) is
available [here](https://github.com/maelvls/k.maelvls.dev).


