---
title: Some stuff about gpg
date: 00-00-00
---

1. [Sign my message](#sign-my-message)
2. [Things on securiry](#things-on-securiry)
3. [Public-key protocol is slow](#public-key-protocol-is-slow)
4. [Errors I got](#errors-i-got)

To list the keys I have in my keyring, use `gpg --list-key`. Here, I can
see one public key in the file `pubring.pgp`:

```plain
> gpg --list-key
/Users/mvalais/.gnupg/pubring.gpg
-------------------------------
pub   4096R/27F4C016 2016-03-04
uid                  Maël Valais <mael.valais@gmail.com>
sub   4096R/A3D4828E 2016-03-04
```

The identifier for my public key is `27F4C016` and the sub identifier
`A3D4828E`. 4096R = 4096-bits RSA encrypted.

When seeing the whole primary key fingerprint:

```plain
F105 E235 8FEB FC3D A4A4 9499 0075 8A3D 27F4 C016
                                        ^^^^^^^^^
                                        identifier
```

To export my public key to the keys server:

```plain
> gpg --send-keys 27F4C016
gpg: envoi de la clef 27F4C016 au serveur hkp keys.gnupg.net
```

Check that my public key has been pushed to the keys server:

```plain
> gpg --search-keys "mael.valais@gmail.com"
gpg: recherche de « mael.valais@gmail.com » sur le serveur hkp
keys.gnupg.net
(1) Maël Valais <mael.valais@gmail.com>
To  4096 bit RSA key 27F4C016, créé : 2016-03-04
```

## Sign my message

Say I have writen the message in `message.txt`. Now I want to sign it and I want
the signature to be in ASCII-format `.asc` (not a binary mess `.sig`).

    gpg2 --armor --detach-sign message.txt

This signature will be in text format in `message.asc`.

## Things on securiry

encrypt: AES 256 (symetric)
salting: hash(pwd+salt), salt is dynamically generated
key derivation: PBKDF2
hash: MD5, SHA-1, SHA-256

## Public-key protocol is slow

With gpg, the public-key protocol is only used for encrypting
the symmetric key and the hash of the message (hash with SHA-256).
The message is then encrypted using the symmetric encryption.

```plain
gpg:
    rsa public key ---encrypts--> symmetric key --encrpyts--> message
                   ---encrypts--> SHA-256 hash
```

With SSL protocol, the only purpose of the public-key encryption is to
encrypt the AES symmetric key.

```plain
ssl:
    <---client--->             <-----server---->
    rsa public key ---sends--> AES symmetric key
```

What about ssh? Two pairs of pub/priv keys?

## Errors I got

- Error message with pinentry
  1. Install `brew install pinentry-mac`
  2. edit ~/.gnupg/gpg-agent.conf to use pinentry-mac instead of pinentry:
     pinentry-program /usr/local/bin/pinentry-mac
