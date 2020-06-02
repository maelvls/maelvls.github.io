---
title: How client-server SSH authentication works
date: 2019-05-28
tags: []
author: Maël Valais
---

1. [Difference between CA and cert](#difference-between-ca-and-cert)
2. [Ssh authentication themes](#ssh-authentication-themes)
   1. [Case A: unknown host](#case-a-unknown-host)
   2. [Case B: host already in known_host but ip changed](#case-b-host-already-in-known_host-but-ip-changed)
   3. [Case C: host already known (ip/fqdn in `~/.ssh/known_hosts`)](#case-c-host-already-known-ipfqdn-in-sshknown_hosts)
   4. [Case D: host has client's pub key in authorized_keys](#case-d-host-has-clients-pub-key-in-authorized_keys)
3. [Ssh + certificates](#ssh--certificates)
4. [Reason why `ssh host cmd` not using .bashrc/.login/.zshrc](#reason-why-ssh-host-cmd-not-using-bashrcloginzshrc)
5. [Passwordless connexion to a server](#passwordless-connexion-to-a-server)
6. [Glossary](#glossary)

**WARNING:** many things in this memo are wrong or very wrong. Do not take
any of it as good information.

## Difference between CA and cert

These confusing names (CA cert, client cert, cert key...) are due to the fact
that you need 4 keys in order to have two parties communicate:

```plain
+-----------+                                   +-----------+
|           |-------what's your ca-cert?------->|           |
|           |                                   |           |
|  client   |<------------ca-cert-------------- |  server   |
|           |                                   |           |
|           |------let's use this sym key-----> |           |
+-----------+                                   +-----------+
```

1. ca-cert is the server pub key that is offered to the client
2. cert (or client-cert) is the client pub key which allows 

From
<https://docs.gitlab.com/runner/configuration/advanced-configuration.html>

- tls-ca-file certificates is a pub key that is used to **verify** the remote
  peer.
- tls-cert-file certificate to **authenticate** with the remote peer
- tls-key-file private key to authenticate with the remote peer

So:

- a **CA** is a certificate of authority. Example: AlphaSSL CA - SHA256 - G2.
  I will define a certificate of authority as a certificate that is a root
  CA (i.e., it is installed in browsers and /etc/ssl) or a sub-CA, i.e.,
  any certificate signed by a root CA.
- a **cert** is a certificate

## Ssh authentication themes

### Case A: unknown host

--- everything in plain text---

- **CLIENT**: hi server, can you send mme your public key?
- **SERVER**: sure, here you go.
- **CLIENT**: (Hmm... should I trust this pub key? This pub key isn't in my
  `~/.ssh/known_hosts`. Let's ask the user if he recognises the sha256 sum.)
- **HUMAN**: whatever I don't care (types 'yes' and enter)
- **CLIENT**: OK, SERVER, I trust you.

--- only SERVER -> CLIENT encrypted ---

- **SERVER**: perfect. (SERVER generates a pub/priv key for this specific
  connexion). Take that id_rsa2.pub I just generated so that you can
  encrypt stuff to me.

--- both ways encrypted ---

- **CLIENT**: Thanks, now we can speak privatly.

### Case B: host already in known_host but ip changed

--- plain text ---

- **CLIENT**: Hi SERVER, I don't know you (i.e., I don't have this ip/fqhn
  in my `~/.ssh/known_hosts`), can you send me your pub key?
- **SERVER**: Sure, here you go.
- **CLIENT**: (Should I trust him? Oh wait, this pub key already exists in
  my known hosts! But with a different IP!!! Tell the user to do something)
- **HUMAN**: Oh crap, why is ssh throwing this error. Let's remove this
  faulty line from `~/.ssh/known_hosts`!

Now, we can go to Case B.

### Case C: host already known (ip/fqdn in `~/.ssh/known_hosts`)

--- only CLIENT -> SERVER encrypted ---

- **CLIENT**: Hi again, SERVER. Can you give me a new pub key so that I can
  tell you stuff secretely?
- **SERVER**: Sure, here you go.

--- both ways encrypted

- **CLIENT**: Perfect.
- **CLIENT**: (Oh! I know this host, lets use the pub key in known_hosts)

### Case D: host has client's pub key in authorized_keys

--- plain text ---

- **CLIENT**: Hi, SERVER, remember me?

--- only SERVER -> CLIENT encrypted ---

- **SERVER**: (Oh, I know this gay as I have his pub key in my
  `authorized_keys`) I guess we can speak privately from now on! Here is a
  generated a pub key so that you can talk to me privatly!

--- both ways encrypted ---

- **CLIENT**: perfect.

## Ssh + certificates

My take on that:

As we saw in [#case-a-unknown-host], the client has a very basic and quite
dumb way of guessing wether or not the server should be trusted: it asks
the user if he recognises/trusts the sha256sum of the public key! Here is
an example of prompt:

```shell
> ssh 123.10.9.23
RSA key fingerprint is 96:a9:23:5c:cc:d1:0a:d4:70:22:93:e9:9e:1e:74:2f.
Are you sure you want to continue connecting (yes/no)? yes
```

Naturally, no one is checking and it is practically impossible to enforce
trust in production servers. Users should not be expected to check wether a
public key is valid or not, and most importantly, requiring a human
prevents automation.

The trick is to use the same system as for standard HTTPS: a chain of trust
based on certification entities. A trusted entity will sign the public key
so that `ssh` will be able to check by himself (if root certificates are
installed, obviously).

NOTE : the root certificate is simple public key (in the gpg sense, not the
ssh one) that allows people to check if the signed version (the signature, called 'certificate')
of a public key (i.e., the result of the root entity using his private key
to sign the public key of someone else) matches. To sum up,

1. [creating certificate] the root entity creates a signature file (called
   certificate) using his private key to sign the public key of entity B
2. [checking certificate] a client that wants to check if the pub key of
   entity B can be trusted will use the public key of the root entity (CA
   file? pem/crt file?) and will know if the signature of the pub key (=
   certificate) matches the public key itself (or something like that).
3. [deploying certificate] now, how to make sure every client that wants to
   connect to entoty B has the pub key of the root entity? These pub keys
   must be installed somewhere: in /etc/ssl/certs for example.

## Reason why `ssh host cmd` not using .bashrc/.login/.zshrc

See: <https://superuser.com/questions/1224938/ssh-host-is-a-login-shell-but-ssh-host-command-is-not>

ssh is using a non-login, non-interactive shell, meaning that

- .login (for login shell) .bashrc, .zshrc (interactive shell) are not
  called. In order to have some environment set, I have to use
  .zshenv (which is called before anything and in any case)

## Passwordless connexion to a server

```shell
$ ssh-keygen -t rsa -b 4096 -C "mael.valais@gmail.com"
ssh-copy-id -i ~/.ssh/id_rsa.pub <remote_server>
```

## Glossary

- **fqdn**: fully-quallified domain name, e.g., dl.bintray.com
