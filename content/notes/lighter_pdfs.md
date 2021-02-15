---
title: Get lighter PDFs
date: 2017-01-30
tags: []
author: Maël Valais
devtoId: 365828
devtoPublished: false
---

```shell
convert -resample 72x72 -compress JPEG -quality 20 a.jpg a.pdf
```

```shell
convert -resample 72x72 -compress JPEG -quality 20 a.jpg a.pdf
```

WARNING: convert (imagemagick) often bugs as it uses external tools such as `gs` and `ffmpeg`; reinstalling Ghostcript can help.
