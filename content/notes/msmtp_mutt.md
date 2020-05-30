---
title: Using mutt and msmtp with Gmail
date: 00-00-00
---

On macOS, install msmtp:

    brew install mutt msmtp

I read this thread:
<https://www.engadget.com/2010/05/04/msmtp-a-free-tool-to-send-email-from-terminal/>

In order to use msmtp from command-line 'mail'

1. add to `~/.mailrc`:

   ```conf
   set sendmail=/usr/local/bin/msmtp
   ```

2. In Keychain Access.app, go to File > New Password Item (Cmd+N)
   and fill it with

   ```plain
       smtp://smtp.gmail.com
       email@gmail.com
       mypassword
   ```

3. In Keychain Access.app, select System Roots, select all certificates
   and File > Export Items (Shift+Cmd+E) > Privacy Enhanced Mail (.pem)
   and save it as 'Certificates.pem' in your \$HOME.

4. Test the configuration with

   ```shell
   echo "Hello world" | mail -s "msmtp test at `date`" cheprugeph@qwfox.com
   ```

Troubleshooting:

- Check in keychain Access.app that there is only ONE Internet password using the
  email@gmail.com and smtp://smtp.gmail.com.
- If the SMTP of Google says that you should go to some adress, with the error:

      send-mail: server message: 534-5.7.14 <https://accounts.google.com/signin/continue?sarp=1&scc=1&plt=AKgnsbtk
      send-mail: server message: 534-5.7.14 z36uL0SoqtWkRHK4lg9Q60tKoOB5B_vNgRZ04wP0SbjkVSvkI3ByYM3aNcsUhrxrskG2Vf
      send-mail: server message: 534-5.7.14 DEQUXboD5d_Px9ehRUYcaOhSdjo2hgTeUvPFDt7PuYRIZTv9CZ8RSybOQ84OAZAPNkbqhd
      send-mail: server message: 534-5.7.14 LWOrMuYScmxwWyjg9PsUoUal67TvV2TYMyM9L_9-KcSyp5aZ-SRDH45cKU1vjGMdi7g7yn
      send-mail: server message: 534-5.7.14 OARlP4_M6aIDqqDF2vbvirlleKMQs> Please log in via your web browser and
      send-mail: server message: 534-5.7.14 then try again.
      send-mail: server message: 534-5.7.14  Learn more at
      send-mail: server message: 534 5.7.14  https://support.google.com/mail/answer/78754 m62sm13851787wmi.19 - gsmtp
      send-mail: could not send mail (account default from /Users/mvalais/.msmtprc)

  then you need to go to <https://accounts.google.com/b/0/DisplayUnlockCaptcha>

## Set up Mutt

I followed the instructions of <https://gist.github.com/chrismytton/3976435> except
for the keychain stuff:

1. We are going to re-use the Keychain thing we created earlier. In order to show
   the password, use

   ```shell
   security find-internet-password -s smtp.gmail.com -a email@gmail.com -w
   ```

   where '-w' tells him to only show the password, '-s' means server (in my case I
   had used 'smtp://smtp.gmail.com' but you must remove the protocol smtp://) and
   '-a' means account.
