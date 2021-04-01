---
title: Using Matlab at IRIT
date: 2017-03-20
tags: []
author: Maël Valais
devtoSkip: true
---

Token server at IRIT:

    matlab -nodecktop -nodisplay -c 27000@licence.irit.fr

or

    SERVER licence.irit.fr 0050568A1251 27000
    USE_SERVER

in `/Applications/MATLAB_R2016a.app/licenses/irit.lic`

What toolbox can I access? `ver` Know what licence file is used: `matlab -e` or `matlab -n`

## Which licenses at IRIT

connecto to bali and do

     /usr/local/matlab/etc/lmstat -a

## On cauchy.math.ups-tlse.fr

    /opt/MATLABR2016a/toolbox/distcomp/bin

    ./mdce start
    ./admincenter
    ./nodestatus

## Matlab R2015a hangs at startup

Edit the shell script and go to line

    cd /Applications/MATLAB_R2015a.app/bin/

I did a `diff matlab matlab_patched`; you can patch it:

```shell
patch matlab << EOF
497c497
<     arglist=""
---
>     arglist="-c \$MATLABdefault/licenses/license.lic"
EOF
```

Note the `\$` because of the shell treatment.
