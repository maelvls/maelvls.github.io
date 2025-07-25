---
title: "Logging into Synology NAS with Personal Google Accounts"
description: "Learn how to simplify access to your Synology NAS by enabling single sign-on (SSO) with personal Google accounts."
date: 2024-04-17T19:50:33+02:00
url: /synology-sso-with-personal-google-account
images: [synology-sso-with-personal-google-account/cover-synology-sso-with-personal-google-account.png]
draft: false
tags: [Synology, NAS, Single Sign-On, Google, OIDC]
author: Maël Valais
devtoId: 1826751
devtoPublished: true
devtoUrl: https://dev.to/maelvls/logging-into-synology-nas-with-personal-google-accounts-16go
---

## Introduction

My parents and I share a DS923+. We mainly use it for storing photos. To save energy, the NAS is stopped at night and restarted in the morning. Restarting the NAS means that you need to log in to the UI at least once every day.

And since my parents don't use the Synology often, they would first struggle remembering what the username thing is about. And then, they would forget about their password, which would lead to me having to reset it because Synology hasn't implemented a way to reset the password over email:

![Synology login screen where you are asked for your username but don't remember it](syno-login-username-1.jpeg)

![Synology login screen where you are asked to enter your password after having entered your username](syno-login-username-2.jpeg)

Over the past two years, my parents forgot their passwords a couple of times. It led me to look for an alternative way to log into the NAS... why not use their Google accounts using the single sign-on mechanism?

## Challenges with Local Accounts and SSO

A couple of years ago, Synology introduced SSO (single sign-on). Since 7.2, DSM supports generic OIDC providers, and supports logging into local users (it used to be only possible for LDAP users).

Since my parents are always signed in into their Google account, I figured it would be possible to use OIDC with Google... Except it won't work with local accounts.

Here is what I tried: I created a My OIDC configuration for Google looked like this:

![Credentials page in API and Services in GCP's Console](credentials-page-in-api-and-services-gcp-console.png)

Then, I configured my Synology to use Google's OIDC endpoint:

![Configuration of the SSO Client using Google OIDC in DSM](config-sso-client-using-google-oidc.png)

The problem arose with the "Username claim". I want to log into my local account `mael.valais`, but none of the claims in Google's ID tokens contain that username. Here is an example of a Google ID token:

```json
{
  "iss": "https://accounts.google.com",
  "aud": "1234987819200.apps.googleusercontent.com",
  "sub": "10769150350006150715113082367",
  "email": "jsmith@example.com",
  "email_verified": "true",
  "iat": 1353601026,
  "exp": 1353604926
}
```

## Forking Dex to use it as an OIDC middleware for Google OIDC

I figured I could use Dex to act as a middleware between Synology's OIDC client and Google's OIDC server. My goal was to "augment" Google's JWTs with Synology's usernames by looking up the user by email.

Dex isn't as flexible as I would have hoped. To make it work, I had to fork it to change the internals of the Google OIDC connector.

Fork: <https://github.com/maelvls/dex/tree/google-to-synology-sso>

This fork is a fork of the fork presented in <https://github.com/dexidp/dex/pull/2954>. It builds on the idea of the `ExtendPayload` interface, which I slightly adjusted to pass the original claims since I needed access to the email contained in the JWT provided by Google.

With this fork, you will need to set three more environment variables:

```bash
SYNO_PASSWD=redacted
SYNO_USER=mael.valais
SYNO_URL=http://127.0.0.1:5000
```

When the OIDC flow with Google is done and before Dex issues its own JWT, I added some code to add the claim `username`. With this modified Dex, the JWT looks like this:

```json
{
  "at_hash": "-j6HZYvzDaqkQB2KxIgSyw",
  "aud": "caddy",
  "c_hash": "8SK3tobDYgaI3cnDzkmi5g",
  "email": "mael65@gmail.com",
  "email_verified": true,
  "exp": 1713387587,
  "iat": 1713301187,
  "iss": "https://login.mysynodomain.dev/dex",
  "name": "Maël Valais",
  "nonce": "MFZFSkESL1XqdQmbvr0T43Kn7v0CzLap",
  "sub": "ChUxMDAzNjk3OTQzNjg3MDAwOTk5MTISBmdvb2dsZQ",
  "username": "mael.valais"
}
```

Here is the updated configuration in DSM:

![Configuration of the SSO Client using Dex in DSM](./config-sso-client-with-dex-in-dsm.png)

## Using the fork of Dex

Create a file `dex.yaml` on your NAS:

```yaml
issuer: https://login.mysynodomain.dev/dex

storage:
  type: sqlite3
  config:
    file: dex.sqlite

web:
  http: 0.0.0.0:5556

logger:
  level: debug

oauth2:
   skipApprovalScreen: true
   alwaysShowLoginScreen: false

staticClients:
- id: synology
  name: 'Synology'
  redirectURIs:
  - 'https://mysynodomain.dev/'
  secret: foo # Use openssl rand -hex 16 to generate this.

connectors:
- type: google
  id: google
  name: Google
  config:
    issuer: https://accounts.google.com
    clientID: $GOOGLE_CLIENT_ID
    clientSecret: $GOOGLE_CLIENT_SECRET
    redirectURI: https://login.mysynodomain.dev/dex/callback

# I have disabled email login.
enablePasswordDB: false
```

Finally, run Dex:


```bash
docker run --name dex -d \
  -v $HOME/dex.yaml:/dex.yaml \
  -v $HOME/dex.sqlite:/dex.sqlite \
  -e SYNO_PASSWD=redacted \
  -e SYNO_USER=mael.valais \
  -e SYNO_URL=http://127.0.0.1:5000 \
  -e GOOGLE_CLIENT_ID=207842732284-l7nhetlsvimmds80fa2knir8fundp3h4.apps.googleusercontent.com \
  -e GOOGLE_CLIENT_SECRET=redacted \
  -p 5556:5556 \
  ghcr.io/maelvls/dex:google-to-synology-sso-v2@sha256:252713d98c8369612994fbbed6f257d79dc35ff84b2cbb6952a11d63c57b64bb serve /dex.yaml
```

> With this command, you will be using Dex images that I built:
>
> ```text
> ghcr.io/maelvls/dex
> ```
>
> I wouldn't recommend using random Docker images from the internet, especially since this is about authentication. I might be a malicious actor trying to steal your Synology credentials! But if you still want to proceed, here is an image! Note that I am not monitoring the image for security vulnerabilities, and do not guarantee that it is secure. Use at your own risk!

## The Docker image

The Docker image is available on GitHub Container Registry: <https://ghcr.io/maelvls/dex>.

### Rebuilding the image yourself and pushing it to your Synology NAS

First, install `zig` and `ko`. That will allow you to cross-compile Dex to `linux/amd64` on macOS without Buildx (cross-compiling is required because Dex's sqlite library needs CGO)

```bash
brew install ko zig
```

Clone the fork:

```bash
git clone https://github.com/maelvls/dex --branch google-to-synology-sso
```

Then, build the image:

```bash
CC="zig cc -target x86_64-linux" CXX="zig c++ -target x86_64-linux" CGO_ENABLED=1 \
  KO_DOCKER_REPO=ghcr.io/maelvls/dex \
  KO_DEFAULTBASEIMAGE=alpine \
  ko build ./cmd/dex --bare --tarball /tmp/out.tar --push=false
```

Then, copy the image to your NAS:

```bash
ssh yournas /usr/local/bin/docker load </tmp/out.tar
```

### (Just so that I don't forget) Here is how I pushed `ghcr.io/maelvls/dex` to GitHub Container Registry

```bash
export VERSION=google-to-synology-sso-v6
git tag $VERSION -m "Release $VERSION"
git push maelvls $VERSION
```

Then:

```bash
CC="zig cc -target x86_64-linux" CXX="zig c++ -target x86_64-linux" CGO_ENABLED=1 \
  KO_DOCKER_REPO=ghcr.io/maelvls/dex \
  KO_DEFAULTBASEIMAGE=alpine \
  ko build ./cmd/dex --bare --push=true --tags $VERSION \
    --image-annotation "org.opencontainers.image.created=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --image-annotation "org.opencontainers.image.url=https://maelvls.dev/synology-sso-with-personal-google-account/" \
    --image-annotation "org.opencontainers.image.source=https://github.com/maelvls/dex" \
    --image-annotation "org.opencontainers.image.version=$VERSION" \
    --image-annotation "org.opencontainers.image.revision=$(git rev-parse HEAD)" \
    --image-annotation "org.opencontainers.image.vendor=Maël Valais" \
    --image-annotation "org.opencontainers.image.title=google-to-synology-sso" \
    --image-annotation "org.opencontainers.image.description=Fork of Dex to use Synology SSO with Google accounts" \
    --image-annotation "org.opencontainers.image.documentation=https://maelvls.dev/synology-sso-with-personal-google-account/" \
    --image-annotation "org.opencontainers.image.authors=Maël Valais <mael.valais@gmail.com>" \
    --image-annotation "org.opencontainers.image.licenses=Apache-2.0" \
    --image-annotation "org.opencontainers.image.ref.name=google-to-synology-sso"
```

See the history below to know the image hashes.

### History

#### June 21st, 2025: v5

Release: [google-to-synology-sso-v5](https://github.com/maelvls/google-to-synology-sso/releases/tag/google-to-synology-sso-v5)

I somehow didn't realize that I was hardcoding the Synology URL. In this
version, I've added `SYNO_URL` (I thought I had already added it, but I
hadn't!).

I've also renamed the fork to google-to-synology-sso to help with
discoverability.

The image:

```text
ghcr.io/maelvls/dex:google-to-synology-sso-v5@sha256:e805a95be565268421ccdb2271dfc0d85ae12b6b53cf82c47b294d34891ff3d1
```

#### June 14th, 2025: v4

Release: [google-to-synology-sso-v4](https://github.com/maelvls/google-to-synology-sso/releases/tag/google-to-synology-sso-v4)

The dex container kept crashing due to i/o timeouts when Dex was trying to connect to the Synology API. I fixed that by adding a retry mechanism with an exponential backoff and a maximum of 10 retries and maximum of 1 hour between retries.

The image:

```text
ghcr.io/maelvls/dex:google-to-synology-sso-v4@sha256:f8bf15901c2b994337994c4f60c48c154437af656cbe85701cb8d1d7d94127ba
```

#### June 4th, 2025: v3

Release: [google-to-synology-sso-v3](https://github.com/maelvls/google-to-synology-sso/releases/tag/google-to-synology-sso-v3)

Reduced the Synology SSO loading time from 10 seconds to 1 second. The reason it was so slow is that I wasn't caching the Synology users and was fetching them every time someone was logging into Synology. The image:

```text
ghcr.io/maelvls/dex:google-to-synology-sso-v3@sha256:d0d889e32400ef70529daef32e7a77bf9da021cbaff9954589db2204a5c49335
```

#### June 1st, 2025: v2

Release: [google-to-synology-sso-v2](https://github.com/maelvls/google-to-synology-sso/releases/tag/google-to-synology-sso-v2)

The `google-to-synology-sso-v1` tag was buggy, the `ExtendPayload` func wasn't being called correctly. I've pushed `google-to-synology-sso-v2` to fix that. Here is the new image:

```text
ghcr.io/maelvls/dex:google-to-synology-sso-v2@sha256:252713d98c8369612994fbbed6f257d79dc35ff84b2cbb6952a11d63c57b64bb
```

#### Apr 12nd, 2024: v1

Release: [google-to-synology-sso-v1](https://github.com/maelvls/google-to-synology-sso/releases/tag/google-to-synology-sso-v1)

Image:

```text
ghcr.io/maelvls/dex:google-to-synology-sso-v1@sha256:345c8fec6b222c308759f21864c6af3b16c373801fd5e0b7ad4b131a743d3b07
```

## Conclusion

With this method, my parents can log into the NAS with their Google account and no longer have to remember their Synology username and password.

Although it works, I wish I didn't have to fork Dex to customize the claims it puts into the JWT payload. I came across a couple of designs that would aim to make Dex more extendable, but none have been implemented yet.

The login flow is much smoother now: click "Login with Google", select the Google account, and you're in! Just two screens:

![Synology login screen that shows a button that says Login with Google](syno-login-sso-google-account-1.jpeg)
![Google screen allowing you to select a Google account](syno-login-sso-google-account-2.png)
