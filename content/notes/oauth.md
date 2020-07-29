---
title: Notes on OAuth
date: 2019-02-17
tags: []
author: Maël Valais
---

I know two main ways of using OAuth2

- password-based client grant (2-leg oauth flow: on the project I worked on,
  the OAuth client was not third party server, but instead, it was the
  front-end.)
- authorization code based grant (3-legs oauth flow)


## What is "the client"?

The [RFC 6749][rfc6749] that presents OAuth2 is very readable, but some terms
like "client" or "code" are confusing. What I found is that trying to understand
"why" the main flow (3-legged oauth, or "authorization code flow") has to be
like this.

The most important aspect that I had to realize is that "client" is hidden from
the end-user (e.g. as a nodejs server serving /callback). The "client" has a
single purpose: receive the code when /callback is called by the third-party
Authorize screen, POST /token to the third party using the `client_secret` using
that code.

[rfc6749]: https://tools.ietf.org/html/rfc6749

## Authorization code (3 legged oauth)

1. User clicks the "Login with Google"; this URL is public and forwards the user
   to an "Authorize" form.

   ```http
   GET /o/oauth2/auth?client_id=foo&redirect_uri=http%3A%2F%2Flocalhost%3A8042%2Fcallback&response_type=code&scope=calendar.readonly&state=something HTTP/1.1
   Host: https://accounts.google.com
   ```

Since this URL is quite long, let's see what we have:
- `access_type` is `offline` (huh??)
- `client_id` (remember that the "client" is my CLI which asks the permission to
  Google's servers to access the scopes.
- `redirect_uri` is <http://localhost:8042/oauth2>
- `response_type` is `code` which makes sense since we want a code so that we
  can, eventually, get a token. The `code` and `state` values will be given to
  us in the callback. When pressing "Authorize" in the Google authentication
  page, we get something like:

2. After authorizing, the user is redirected to the "client" endpoint:

   ```http
   GET /callback?code=f1a2bc&state=something HTTP/1.1
   Host: http://localhost:8042
   ```

   Response is

3. The client gets the token using the code it received.

   ```http
   POST /token HTTP/1.1
   Host: https://oauth2.googleapis.com
   Content-Type: application/x-www-form-urlencoded

   grant_type=authorization_code&code=f1a2bc&client_id=foo&client_secret=<the_secret>&redirect_url=<same as above>
   ```

   The token looks like this:

   ```json
   {
     "access_token": "ya29.a0AfH6SMCkF6Kd0bZPf60Knhq8XyMKTgmQ6zE5lP9pjdZfW-9ebV5V9wifFagdiioN5JWovHmfVfdukAE0-jcHRmjzsycQCYPj7zzSup55X0n_gz8rkglYGBaeG5Tyde8a8rAIu1CimhtSdsoq0_HCh2VBXOLmrq7oKSg",
     "token_type": "Bearer",
     "refresh_token": "1//03E6r0qdOHKqrCgYIARAAGAMSNwF-L9Ir37zr9-GH8po_A5XSwsSiEw8XmiHPnCbKPaFHNCckIF-vmJRRKddbjWLbX9ZrbOzffts",
     "expiry": "2020-07-21T19:26:47.346143+02:00"
   }
   ```

   Note: when we say "an oauth token", what we actually mean is the access token
   the refresh token.

